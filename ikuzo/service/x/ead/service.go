package ead

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/delving/hub3/config"
	eadHub3 "github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
)

type Metrics struct {
	Started  uint64
	Failed   uint64
	Finished uint64
	Canceled uint64
}

type CreateTreeFn func(cfg *eadHub3.NodeConfig, n *eadHub3.Node, hubID string, id string) *fragments.Tree

type Service struct {
	index        *index.Service
	dataDir      string
	m            Metrics
	CreateTreeFn CreateTreeFn
	tasks        map[string]*Task
	rw           sync.RWMutex
	workers      int
	cancel       context.CancelFunc
	group        *errgroup.Group
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		tasks:   make(map[string]*Task),
		workers: 2,
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.CreateTreeFn == nil {
		s.CreateTreeFn = eadHub3.CreateTree
	}

	// create datadir
	if s.dataDir != "" {
		createErr := os.MkdirAll(s.dataDir, os.ModePerm)
		if createErr != nil {
			return nil, createErr
		}
	}

	return s, nil
}

func (s *Service) findAvailableTask() *Task {
	tasks := []*Task{}

	s.rw.RLock()
	for _, task := range s.tasks {
		if task.InState == StatePending || task.Interrupted {
			tasks = append(tasks, task)
		}
	}
	s.rw.RUnlock()

	if len(tasks) == 0 {
		return nil
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].currentTransition().Started.After(tasks[j].currentTransition().Started)
	})

	return tasks[0]
}

func (s *Service) StartWorkers() error {
	// create errgroup and add cancel to service
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx

	s.cancel = cancel
	s.group = g

	ticker := time.NewTicker(1 * time.Second)

	for i := 0; i < s.workers; i++ {
		g.Go(func() error {
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				case <-ticker.C:
					s.rw.RLock()
					task := s.findAvailableTask()
					s.rw.RUnlock()
					if task == nil {
						continue
					}

					if err := s.Process(gctx, task); err != nil {
						// check for context canceled
						return err
					}
				}
			}
		})
	}

	return nil
}

func (s *Service) Metrics() Metrics {
	return s.m
}

func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	s.handleUpload(w, r)
}

func (s *Service) Tasks(w http.ResponseWriter, r *http.Request) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	// TODO(kiivihal): add option to filter by datasetID

	render.JSON(w, r, s.tasks)
}

func (s *Service) findTask(orgID, datasetID string, filterActive bool) (*Task, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	for _, t := range s.tasks {
		if t.Meta.DatasetID == datasetID {
			if filterActive && !t.isActive() {
				continue
			}

			return t, nil
		}
	}

	return nil, ErrTaskNotFound
}

func (s *Service) GetTask(w http.ResponseWriter, r *http.Request) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	id := chi.URLParam(r, "id")

	task, ok := s.tasks[id]
	if !ok {
		http.Error(w, "unknown task", http.StatusNotFound)
		return
	}

	render.JSON(w, r, task)
}

func (s *Service) CancelTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.rw.Lock()
	defer s.rw.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		http.Error(w, "unknown task", http.StatusNotFound)
		return
	}

	task.moveState(StateCancelled)

	task.log().Info().Msg("canceling running ead task")
	task.cancel()

	task.Next()
	// TODO(kiivihal): do we delete or keep it
	// delete(s.tasks, id)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) Shutdown(ctx context.Context) error {
	// cancel workers.
	s.cancel()

	if err := s.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (s *Service) saveDescription(cfg *eadHub3.NodeConfig, t *Task, ead *eadHub3.Cead) error {
	desc, err := eadHub3.NewDescription(ead)
	if err != nil {
		return fmt.Errorf("unable to create description; %w", err)
	}

	t.Meta.Title = desc.Summary.File.Title

	cfg.Title = []string{desc.Summary.File.Title}

	descIndex := eadHub3.NewDescriptionIndex(t.Meta.DatasetID)

	err = descIndex.CreateFrom(desc)
	if err != nil {
		return fmt.Errorf("unable to create DescriptionIndex; %w", err)
	}

	err = descIndex.Write()
	if err != nil {
		return fmt.Errorf("unable to write DescriptionIndex; %w", err)
	}

	err = desc.Write()
	if err != nil {
		return fmt.Errorf("unable to write description; %w", err)
	}

	var unitInfo *eadHub3.UnitInfo
	if desc.Summary.FindingAid != nil && desc.Summary.FindingAid.UnitInfo != nil {
		unitInfo = desc.Summary.FindingAid.UnitInfo
	}

	if s.index != nil {
		err = ead.SaveDescription(cfg, unitInfo, s.index)
		if err != nil {
			return fmt.Errorf("unable to create index representation of the description; %w", err)
		}
	}

	return nil
}

func (s *Service) Process(parentCtx context.Context, t *Task) error {
	// return immediately with invalid states
	if !t.isActive() {
		return nil
	}

	// wrap parent so both will stop
	_ = parentCtx
	g, gctx := errgroup.WithContext(t.ctx)

	// parse EAD
	ead := new(eadHub3.Cead)

	f, err := os.Open(t.Meta.getSourcePath())
	if err != nil {
		errMsg := fmt.Errorf("unable to find EAD source file: %w", err)
		return t.finishWithError(errMsg)
	}

	err = xml.NewDecoder(f).Decode(ead)
	if err != nil {
		errMsg := fmt.Errorf("error during EAD parsing; %w", err)
		return t.finishWithError(errMsg)
	}

	// TODO(kiivihal): replace with dataset service later or store in DescriptionIndex
	ds, created, err := models.GetOrCreateDataSet(t.Meta.DatasetID)
	if err != nil {
		ErrorfMsg := fmt.Errorf("unable to get DataSet for %s", t.Meta.DatasetID)
		return t.finishWithError(ErrorfMsg)
	}

	t.Meta.Created = created

	// set basics for ead
	ds.Label = ead.Ceadheader.GetTitle()
	ds.Period = ead.Carchdesc.GetPeriods()

	// description must be set to empty
	ds.Description = ""

	cfg := eadHub3.NewNodeConfig(gctx)
	cfg.CreateTree = s.CreateTreeFn
	cfg.Spec = t.Meta.DatasetID
	cfg.OrgID = config.Config.OrgID
	cfg.IndexService = s.index

	cfg.Nodes = make(chan *eadHub3.Node, 2000)

	if t.InState == StatePending {
		atomic.AddUint64(&s.m.Started, 1)
		t.Next()
	}

	// create description
	if t.InState == StateProcessingDescription {
		if err := s.saveDescription(cfg, t, ead); err != nil {
			return t.finishWithError(fmt.Errorf("unable to save index: %w", err))
		}

		t.Next()
	}

	if t.InState != StateProcessingInventories {
		return fmt.Errorf("invalid state for processing inventories: %s", t.InState)
	}

	// publish nodes
	g.Go(func() error {
		_, _, err := ead.Carchdesc.Cdsc.NewNodeList(cfg)
		return err
	})

	workers := 8
	workerChan := make(chan int, workers)

	// when to close the HubIDs channel
	g.Go(func() error {
		var workersDone int
		for worker := range workerChan {
			workersDone += worker

			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			if workersDone == workers {
				close(cfg.HubIDs)
				close(workerChan)
				return nil
			}
		}

		return nil
	})

	// gather duplicates
	g.Go(func() error {
		hubIDs := map[string]*eadHub3.NodeEntry{}
		duplicates := map[*eadHub3.NodeEntry]bool{}

		for entry := range cfg.HubIDs {
			dupEntry, ok := hubIDs[entry.HubID]
			if ok {
				duplicates[entry] = true
				duplicates[dupEntry] = true

				continue
			}

			hubIDs[entry.HubID] = entry

			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}
		}

		if len(duplicates) != 0 {
			sortedDups := []*eadHub3.NodeEntry{}
			for dup := range duplicates {
				sortedDups = append(sortedDups, dup)
			}

			sort.Slice(sortedDups, func(i, j int) bool {
				return sortedDups[i].HubID < sortedDups[j].HubID
			})

			for _, dup := range sortedDups {
				t.log().Warn().
					Str("hubID", dup.HubID).
					Str("path", dup.Path).
					Int("sortKey", int(dup.Order)).
					Str("label", dup.Title).
					Msg("duplicate hubIDs discovered")
			}
		}

		return nil
	})

	for i := 0; i < workers; i++ {
		// consume nodes
		g.Go(func() error {
			for n := range cfg.Nodes {
				n := n
				if s.index == nil {
					continue
				}

				fg, _, err := n.FragmentGraph(cfg)
				if err != nil {
					return err
				}

				m, err := fg.IndexMessage()
				if err != nil {
					return fmt.Errorf("unable to marshal fragment graph: %w", err)
				}

				if err := s.index.Publish(context.Background(), m); err != nil {
					return err
				}
				atomic.AddUint64(&cfg.RecordsPublishedCounter, 1)

				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
				}
			}

			workerChan <- 1

			return nil
		})
	}

	// wait for all errgroup goroutines
	if err := g.Wait(); err == nil || errors.Is(err, context.Canceled) {
		if errors.Is(err, context.Canceled) {
			// TODO(kiivihal): deal with canceled that must be restartable
			// this is interrupted
			t.Meta.Clevels = cfg.Counter.GetCount()
			if t.ctx.Err() == context.Canceled {
				t.moveState(StateCancelled)
				t.Next()

				return nil
			}

			t.Interrupted = true
			// save the task

			return nil
		}

		ds.MetsFiles = int(cfg.MetsCounter.GetCount())
		ds.Clevels = int(cfg.Counter.GetCount())

		stats := models.DaoStats{
			DuplicateLinks: map[string]int{},
		}
		stats.ExtractedLinks = cfg.MetsCounter.GetCount()
		stats.RetrieveErrors = cfg.MetsCounter.GetErrorCount()
		stats.DigitalObjects = cfg.MetsCounter.GetDigitalObjectCount()
		stats.Errors = cfg.MetsCounter.GetErrors()
		uniqueLinks := cfg.MetsCounter.GetUniqueCounter()
		stats.UniqueLinks = uint64(len(uniqueLinks))

		for k, v := range uniqueLinks {
			if v > 1 {
				stats.DuplicateLinks[k] = v
			}
		}

		ds.DaoStats = stats

		err = ds.Save()
		if err != nil {
			return fmt.Errorf("unable to save dataset %s; %w", ds.Spec, err)
		}

		t.Meta.Clevels = cfg.Counter.GetCount()
		t.Meta.DaoLinks = cfg.MetsCounter.GetCount()
		t.Meta.RecordsPublished = atomic.LoadUint64(&cfg.RecordsPublishedCounter)
		t.Meta.DigitalObjects = cfg.MetsCounter.GetDigitalObjectCount()

		metrics := map[string]uint64{
			"description":       1,
			"inventories":       t.Meta.Clevels,
			"mets-files":        t.Meta.DaoLinks,
			"records-published": t.Meta.RecordsPublished,
			"digital-objects":   t.Meta.DigitalObjects,
		}

		t.Transitions[len(t.Transitions)-1].Metrics = metrics

		t.finishTask()
	} else {
		return t.finishWithError(fmt.Errorf("error during invertory processing; %w", err))
	}

	return nil
}

// legacy should be deprecated
func (s *Service) eadUpload(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")

	_, err := eadHub3.ProcessUpload(r, w, spec, s.index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type taskResponse struct {
	TaskID    string `json:"taskID"`
	OrgID     string `json:"orgID,omitempty"`
	DatasetID string `json:"datasetID"`
	Status    string `json:"status"`
}

func (s *Service) handleUpload(w http.ResponseWriter, r *http.Request) {
	in, header, err := r.FormFile("ead")
	if err != nil {
		http.Error(w, "cannot find ead form file", http.StatusBadRequest)
		return
	}

	defer in.Close()
	// cleanup upload
	defer func() {
		err = r.MultipartForm.RemoveAll()
	}()

	_, meta, err := s.SaveEAD(in, header.Size)
	if err != nil {
		if errors.Is(err, ErrTaskAlreadySubmitted) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		atomic.AddUint64(&s.m.Failed, 1)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t, err := s.NewTask(&meta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	render.JSON(w, r, taskResponse{
		TaskID:    t.ID,
		OrgID:     t.Meta.OrgID,
		DatasetID: t.Meta.DatasetID,
		Status:    string(t.InState),
	})
}

func (s *Service) SaveEAD(r io.Reader, size int64) (*bytes.Buffer, Meta, error) {
	var meta Meta

	buf, tmpFile, err := s.storeEAD(r, size)
	if err != nil {
		return nil, meta, err
	}

	meta, err = s.moveTmpFile(buf, tmpFile)
	if err != nil {
		return nil, meta, err
	}

	meta.FileSize = uint64(size)

	return buf, meta, nil
}

func (s *Service) GetName(buf *bytes.Buffer) (string, error) {
	var (
		dataset string
		inElem  bool
	)

	xmlDec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))

L:
	for {
		t, tokenErr := xmlDec.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			} else {
				return "", fmt.Errorf("failed to read token: %w", tokenErr)
			}
		}

		switch elem := t.(type) {
		case xml.StartElement:
			if elem.Name.Local == "eadid" {
				inElem = true
			}

		case xml.EndElement:
			if inElem {
				break L
			}
		case xml.CharData:
			if inElem {
				dataset = string(elem)
				if dataset != "" {
					return dataset, nil
				}
			}
		}
	}

	return "", fmt.Errorf("eadid chardata cannot be empty")
}

// getDataPath returns the path to where all files for an EAD should be stored.
func (s *Service) getDataPath(dataset string) string {
	return filepath.Join(s.dataDir, dataset)
}

// storeEAD stores the ead in a tmpFile and returns a io.Reader and name of the tmpFile.
//
// The tempFile must be closed by the calling code. An error is returned when the tmpFile
// cannot be created or written to.
//
// The returned io.Reader is a bytes.Buffer that can be read from multiple times.
func (s *Service) storeEAD(r io.Reader, size int64) (*bytes.Buffer, string, error) {
	f, err := ioutil.TempFile(s.dataDir, "*")
	if err != nil {
		return nil, "", fmt.Errorf("unable to create ead tmpFiles; %w", err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(make([]byte, 0, size))

	_, err = io.Copy(f, io.TeeReader(r, buf))
	if err != nil {
		return nil, "", fmt.Errorf("unable to copy ead to tmpfile; %w", err)
	}

	return buf, f.Name(), nil
}

// moveTmpFile retrieves Meta from EAD and moves it to the right location
func (s *Service) moveTmpFile(buf *bytes.Buffer, tmpFile string) (Meta, error) {
	var (
		meta Meta
		err  error
	)

	// get ead identifier
	meta.DatasetID, err = s.GetName(buf)
	if err != nil {
		return meta, err
	}

	if _, err := s.findTask("", meta.DatasetID, true); !errors.Is(err, ErrTaskNotFound) {
		return meta, ErrTaskAlreadySubmitted
	}

	meta.basePath = s.getDataPath(meta.DatasetID)

	// create dataDir
	if err := os.MkdirAll(meta.basePath, os.ModePerm); err != nil {
		return meta, err
	}

	if err := os.Rename(tmpFile, meta.getSourcePath()); err != nil {
		return meta, err
	}

	return meta, nil
}
