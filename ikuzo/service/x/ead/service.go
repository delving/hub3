// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"

	eadHub3 "github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
	indexHub3 "github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh/harvest"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	PaccessKey = "processAccessTime"
)

type CreateTreeFn func(cfg *eadHub3.NodeConfig, n *eadHub3.Node, hubID string, id string) *fragments.Tree

type PreStoreFn func(b []byte) []byte

type DaoFn func(cfg *eadHub3.DaoConfig) error

var _ domain.Service = (*Service)(nil)

type ProcessFn func(s *Service, parentCtx context.Context, t *Task) error

type Service struct {
	index                   *index.Service
	dataDir                 string
	M                       Metrics
	CreateTreeFn            CreateTreeFn
	PreStoreFn              PreStoreFn
	DaoFn                   DaoFn
	ProcessFn               ProcessFn
	DaoClient               *eadHub3.DaoClient
	processDigital          bool
	processDigitalIfMissing bool
	tasks                   map[string]*Task
	rw                      sync.RWMutex
	workers                 int
	cancel                  context.CancelFunc
	group                   *errgroup.Group
	postHooks               map[string][]domain.PostHookService
	log                     zerolog.Logger
	orgs                    domain.OrgConfigRetriever
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		tasks:     make(map[string]*Task),
		workers:   1,
		postHooks: map[string][]domain.PostHookService{},
		ProcessFn: Process,
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

	if s.PreStoreFn == nil {
		s.PreStoreFn = addHeader
	}

	daoClient := eadHub3.NewDaoClient(s.index)
	daoClient.HttpFallback = true

	s.DaoClient = &daoClient

	if s.DaoFn == nil {
		s.DaoFn = daoClient.DefaultDaoFn
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

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Publish(ctx context.Context, messages ...*domainpb.IndexMessage) error {
	return s.index.Publish(ctx, messages...)
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}

func (s *Service) StartWorkers() error {
	// create errgroup and add cancel to service
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx

	s.cancel = cancel
	s.group = g

	ticker := time.NewTicker(1 * time.Second)
	heartbeat := time.NewTicker(5 * time.Minute)

	for i := 0; i < s.workers; i++ {
		worker := i

		g.Go(func() error {
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				case <-heartbeat.C:
					log.Trace().Str("svc", "eadProcessor").Int("worker", worker).Msg("worker heartbeat")
				case <-ticker.C:
					s.rw.Lock()
					task := s.findAvailableTask()
					if task == nil {
						s.rw.Unlock()
						continue
					}

					task.Next()
					s.rw.Unlock()

					if err := s.ProcessFn(s, gctx, task); err != nil {
						return err
					}
				}
			}
		})
	}

	return nil
}

func (s *Service) Metrics() Metrics {
	return s.M
}

func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	s.handleUpload(w, r)
}

// Call this function each night at 00:01 in a cron job to check and clear tree node restrictions.
func (s *Service) ClearRestrictions(w http.ResponseWriter, r *http.Request) {
	s.clearRestrictions(w, r)
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

	err = ead.SaveDescription(cfg, unitInfo)
	if err != nil {
		return fmt.Errorf("unable to create index representation of the description; %w", err)
	}

	return nil
}

func getEAD(r io.Reader) (*eadHub3.Cead, error) {
	// parse EAD
	ead := new(eadHub3.Cead)

	if err := xml.NewDecoder(r).Decode(ead); err != nil {
		return ead, err
	}

	return ead, nil
}

func Process(s *Service, parentCtx context.Context, t *Task) error {
	// return immediately with invalid states
	if !t.isActive() {
		return nil
	}

	if t.InState == StateStarted {
		s.M.IncStarted()
		t.Next()
	}

	// wrap parent so both will stop
	_ = parentCtx
	g, gctx := errgroup.WithContext(t.ctx)

	f, err := os.Open(t.Meta.getSourcePath())
	if err != nil {
		errMsg := fmt.Errorf("unable to find EAD source file: %w", err)
		return t.finishWithError(errMsg)
	}

	ead, err := getEAD(f)
	if err != nil {
		errMsg := fmt.Errorf("error during EAD parsing; %w", err)

		f.Close()

		return t.finishWithError(errMsg)
	}

	if closeErr := f.Close(); closeErr != nil {
		errMsg := fmt.Errorf("unable to close ead source file; %w", err)
		return t.finishWithError(errMsg)
	}

	meta, created, err := eadHub3.GetOrCreateMeta(t.Meta.DatasetID)
	if err != nil {
		ErrorfMsg := fmt.Errorf("unable to get ead Meta for %s", t.Meta.DatasetID)
		return t.finishWithError(ErrorfMsg)
	}

	meta.Revision++

	t.Meta.Created = created
	t.Meta.Revision = meta.Revision

	// create a dataset
	ds, _, datasetErr := models.GetOrCreateDataSet(t.Meta.OrgID, t.Meta.DatasetID)
	if datasetErr != nil {
		ErrorfMsg := fmt.Errorf("unable to get ead dataset for %s; %w", t.Meta.DatasetID, datasetErr)
		return t.finishWithError(ErrorfMsg)
	}

	ds.Revision = int(t.Meta.Revision)
	ds.RecordType = "ead"

	if err := ds.Save(); err != nil {
		log.Error().Err(err).Msg("unable to save revision in dataset")
	}

	// set basics for ead
	meta.Label = ead.Ceadheader.GetTitle()
	meta.Period = ead.Carchdesc.GetPeriods()

	cfg := eadHub3.NewNodeConfig(gctx)
	cfg.CreateTree = s.CreateTreeFn
	cfg.DaoFn = s.DaoFn
	cfg.Spec = t.Meta.DatasetID
	cfg.OrgID = t.Meta.OrgID
	cfg.IndexService = s.index
	cfg.RetrieveDao = true
	cfg.Tags = t.Meta.Tags
	cfg.Revision = t.Meta.Revision
	cfg.ProcessDigital = t.Meta.ProcessDigital
	cfg.ProcessDigitalIfMissing = t.Meta.ProcessDigitalIfMissing
	cfg.ProcessAccessTime = t.Meta.ProcessAccessTime

	cfg.Nodes = make(chan *eadHub3.Node, 2000)

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
		// xml.Decoder is not used anymore so it can be garbage collected
		ead = nil

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

				fg, _, err := n.FragmentGraph(cfg)
				if err != nil {
					return err
				}

				if s.index == nil {
					continue
				}

				m, err := fg.IndexMessage()
				if err != nil {
					return fmt.Errorf("unable to marshal fragment graph: %w", err)
				}

				if err := s.Publish(context.Background(), m); err != nil {
					return err
				}

				atomic.AddUint64(&cfg.RecordsCreatedCounter, 1)

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
				t.moveState(StateCanceled)
				t.Next()

				return nil
			}

			t.Interrupted = true
			// save the task

			return nil
		}

		digitalObjects, countErr := t.s.DaoClient.GetDigitalObjectCount(t.Meta.DatasetID)
		if countErr != nil {
			return t.finishWithError(fmt.Errorf("unable to count digitalobjects: %w", countErr))
		}

		cfg.MetsCounter.IncrementDigitalObject(uint64(digitalObjects))

		meta.MetsFiles = int(cfg.MetsCounter.GetCount())
		meta.Inventories = int(cfg.Counter.GetCount())

		stats := models.DaoStats{
			DuplicateLinks: map[string]int{},
		}
		stats.ExtractedLinks = cfg.MetsCounter.GetCount()
		stats.RetrieveErrors = cfg.MetsCounter.GetErrorCount()
		stats.DigitalObjects = cfg.MetsCounter.GetDigitalObjectCount()
		stats.ErrorsPerInventory = cfg.MetsCounter.GetErrors()
		uniqueLinks := cfg.MetsCounter.GetUniqueCounter()
		stats.UniqueLinks = uint64(len(uniqueLinks))

		for k, v := range uniqueLinks {
			if v > 1 {
				stats.DuplicateLinks[k] = v
			}
		}

		t.Meta.Clevels = cfg.Counter.GetCount()
		t.Meta.DaoLinks = cfg.MetsCounter.GetCount()
		t.Meta.DaoErrors = cfg.MetsCounter.GetErrorCount()
		t.Meta.DaoErrorLinks = cfg.MetsCounter.GetErrors()
		t.Meta.TotalRecordsPublished = atomic.LoadUint64(&cfg.RecordsCreatedCounter)
		t.Meta.DigitalObjects = cfg.MetsCounter.GetDigitalObjectCount()

		metrics := map[string]uint64{
			"description":       1,
			"inventories":       t.Meta.Clevels,
			"mets-files":        t.Meta.DaoLinks,
			"records-published": t.Meta.TotalRecordsPublished,
			"digital-objects":   t.Meta.DigitalObjects,
		}

		t.Transitions[len(t.Transitions)-1].Metrics = metrics

		if dropErr := t.dropOrphans(cfg.Revision); dropErr != nil {
			return t.finishWithError(fmt.Errorf("error during dropping orphans: %w", dropErr))
		}

		t.finishTask()

		meta.DaoStats = stats
		meta.DigitalObjects = int(cfg.MetsCounter.GetDigitalObjectCount())
		meta.RecordsPublished = int(t.Meta.TotalRecordsPublished)
		meta.Revision = cfg.Revision

		if meta.OrgID == "" {
			meta.OrgID = t.Meta.OrgID
		}

		countErr = meta.Write()
		if countErr != nil {
			return fmt.Errorf("unable to save ead meta for %s; %w", meta.DatasetID, countErr)
		}
	} else {
		return t.finishWithError(fmt.Errorf("error during invertory processing; %w", err))
	}

	return nil
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

	s.M.IncSubmitted()

	// TODO(kiivihal): finish this later. Add multi tenancy middleware first
	var orgID string
	if id := domain.GetOrganizationID(r); id != "" {
		orgID = string(id)
	} else {
		log.Printf("unable to find orgID in request")
	}

	meta, err := s.SaveEAD(in, header.Size, "", orgID)
	if err != nil {
		if errors.Is(err, ErrTaskAlreadySubmitted) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		s.M.IncFailed()
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	task, metaErr := s.CreateTask(r, meta)
	if metaErr != nil {
		http.Error(w, metaErr.Error(), http.StatusConflict)
		return
	}

	render.JSON(w, r, task)
}

func (s *Service) CreateTask(r *http.Request, meta Meta) (*taskResponse, error) {
	if orgID := r.Header.Get("orgID"); orgID != "" {
		meta.OrgID = orgID
	}

	if processDigital := r.FormValue("mets"); processDigital != "" {
		if b, convErr := strconv.ParseBool(processDigital); convErr == nil {
			meta.ProcessDigital = b
		}
	}

	if processDigital := r.FormValue("ifmissing"); processDigital != "" {
		if b, convErr := strconv.ParseBool(processDigital); convErr == nil {
			meta.ProcessDigitalIfMissing = b
		}
	}

	if processAccessTime := r.FormValue(PaccessKey); processAccessTime != "" {
		if t, convErr := time.Parse(time.RFC3339, processAccessTime); convErr == nil {
			meta.ProcessAccessTime = t
		}
	}

	t, err := s.NewTask(&meta)
	if err != nil {
		s.M.IncAlreadyQueued()
		return nil, err
	}

	if forTags := r.FormValue("tags"); forTags != "" {
		for _, tag := range strings.Split(forTags, ",") {
			meta.Tags = append(meta.Tags, strings.TrimSpace(tag))
		}
	}

	tr := &taskResponse{
		TaskID:    t.ID,
		OrgID:     t.Meta.OrgID,
		DatasetID: t.Meta.DatasetID,
		Status:    string(t.InState),
	}

	return tr, nil
}

// TODO: move this to elasticsearch package later
func (s *Service) clearRestrictions(w http.ResponseWriter, r *http.Request) {
	org, _ := domain.GetOrganization(r)

	search := indexHub3.ESClient().Search(org.Config.GetIndexName())

	format := "02-01-2006"
	today := time.Now()

	if r.URL.Query().Get(PaccessKey) != "" {
		td, err := time.Parse(format, r.URL.Query().Get(PaccessKey))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		today = td
	}

	query := elastic.NewBoolQuery().
		Filter(
			elastic.NewMatchPhraseQuery(elasticsearch.PathOrgID, org.RawID()),
			elastic.NewMatchPhraseQuery("tree.hasRestriction", true),
			elastic.NewMatchPhraseQuery("tree.type", "file"),
			elastic.NewMatchPhraseQuery("tree.access", today.Format(format)),
		)
	aggsKey := "meta.spec"
	agg := elastic.NewTermsAggregation().Field(aggsKey).Size(10000)
	search.Query(query).Aggregation(aggsKey, agg)
	response, err := search.Do(r.Context())
	clearLogger := log.WithLevel(zerolog.InfoLevel).
		Str("component", "hub3").
		Str("svc", "eadClearRestrictions").
		Str("eventType", "read")

	if err != nil {
		clearLogger.Err(err).Msgf("bad ead clear request %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if response.Hits == nil || response.Hits.Hits == nil || len(response.Hits.Hits) == 0 {
		render.JSON(w, r, struct{ Message string }{Message: "No hits"})
		return
	}

	terms, ok := response.Aggregations.Terms(aggsKey)
	if !ok {
		clearLogger.Msg("error with aggregations")
		http.Error(w, "terms could not load", http.StatusBadRequest)

		return
	}

	trSlice := make([]*taskResponse, 0)

	for _, el := range terms.Buckets {
		spec := el.Key.(string)

		meta, err := s.LoadEAD(org.RawID(), spec)
		if err != nil {
			clearLogger.Err(err).Msgf("could not load spec %s from bucket: %v", spec, err)
			continue
		}

		meta.ProcessAccessTime = today

		task, metaErr := s.CreateTask(r, meta)
		if metaErr != nil {
			clearLogger.Err(err).Msgf("could not handle spec meta %s from bucket: %v", spec, metaErr)
			continue
		}

		trSlice = append(trSlice, task)
	}

	render.JSON(w, r, trSlice)
}

func (s *Service) LoadEAD(orgID, spec string) (Meta, error) {
	var meta Meta

	// TODO(kiivihal): later fix this with calling root Meta eadhub3.GetMeta
	meta.DatasetID = spec
	meta.OrgID = orgID

	if _, err := s.findTask(meta.OrgID, meta.DatasetID, true); !errors.Is(err, ErrTaskNotFound) {
		return meta, ErrTaskAlreadySubmitted
	}

	meta.basePath = s.getDataPath(meta.DatasetID)

	f, err := os.Stat(meta.getSourcePath())
	if err != nil {
		errMsg := fmt.Errorf("unable to find EAD source file: %w", err)
		return meta, errMsg
	}

	meta.FileSize = uint64(f.Size())
	meta.ProcessDigital = s.processDigital
	meta.ProcessDigitalIfMissing = s.processDigitalIfMissing

	return meta, nil
}

// func (s *Service) SaveEAD(r io.Reader, size int64) (*bytes.Buffer, Meta, error) {

// var meta Meta

// // TODO(kiivihal): remove this step
// buf, tmpFile, err := s.storeEAD(r, size)
// if err != nil {
// return nil, meta, err
// }

// meta, err = s.moveTmpFile(buf, tmpFile)
// if err != nil {
// return nil, meta, err
// }

// meta.FileSize = uint64(size)
// meta.ProcessDigital = s.processDigital

// return buf, meta, nil
// }

func (s *Service) GetName(b []byte) (string, error) {
	var (
		dataset string
		inElem  bool
	)

	xmlDec := xml.NewDecoder(bytes.NewReader(b))

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
	if err := os.MkdirAll(s.dataDir, os.ModePerm); err != nil {
		return nil, "", fmt.Errorf("unable to create data directory; %w", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))

	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, "", fmt.Errorf("unable to copy ead to tmpfile; %w", err)
	}

	b := buf.Bytes()

	if s.PreStoreFn != nil {
		b = s.PreStoreFn(b)
	}

	f, err := ioutil.TempFile(s.dataDir, "*")
	if err != nil {
		return nil, "", fmt.Errorf("unable to create ead tmpFiles; %w", err)
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return nil, "", fmt.Errorf("unable to write to ead tmpFiles; %w", err)
	}

	return buf, f.Name(), nil
}

// SaveEAD stores the source EAD in the revision store
func (s *Service) SaveEAD(r io.Reader, size int64, datasetID, orgID string) (Meta, error) {
	var meta Meta

	b, tmpFile, err := s.storeEAD(r, size)
	if err != nil {
		return meta, err
	}

	meta, err = s.moveTmpFile(b, tmpFile)
	if err != nil {
		return meta, err
	}

	meta.FileSize = uint64(size)
	meta.ProcessDigital = s.processDigital
	meta.ProcessDigitalIfMissing = s.processDigitalIfMissing
	meta.OrgID = orgID

	return meta, nil
}

// DeleteEAD removes the dataset from the store.
func (s *Service) DeleteEAD(datasetID, orgID string) error {
	return models.DeleteDataSet(orgID, datasetID, context.Background())
}

// moveTmpFile retrieves Meta from EAD and moves it to the right location
func (s *Service) moveTmpFile(buf *bytes.Buffer, tmpFile string) (Meta, error) {
	var (
		meta Meta
		err  error
	)

	b := buf.Bytes()

	// get ead identifier
	meta.DatasetID, err = s.GetName(b)
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

// AddPostHook adds posthook to the EAD service
func (s *Service) AddPostHook(hook domain.PostHookService) error {
	s.postHooks[hook.OrgID()] = append(s.postHooks[hook.OrgID()], hook)

	if s.index != nil {
		if err := s.index.AddPostHook(hook); err != nil {
			return err
		}
	}

	return nil
}

// getModifiedDatePath returns the filepath of the EAD modified date.
func (s *Service) getModifiedDatePath(spec string) string {
	return fmt.Sprintf("%s/modified", s.getDataPath(spec))
}

// LoadModifiedEADDate loads the modified date of the EAD or else zero time if not available.
func (s *Service) LoadModifiedEADDate(spec string) time.Time {
	s.rw.Lock()
	defer s.rw.Unlock()

	t := time.Time{}
	timePath := s.getModifiedDatePath(spec)

	b, err := os.ReadFile(timePath)
	if err != nil {
		return t
	}

	pt, pErr := time.Parse(harvest.DateFormat, string(b))
	if pErr != nil {
		return t
	}

	return pt
}

// SaveModifiedEADDate stores the modified EAD date.
func (s *Service) SaveModifiedEADDate(spec, modified string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	timePath := s.getModifiedDatePath(spec)

	err := os.WriteFile(timePath, []byte(modified), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ResyncCacheDir(orgID string) error {
	dirs, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	var seen int

	for _, ead := range dirs {
		if !ead.IsDir() {
			continue
		}

		spec := ead.Name()

		meta, err := s.LoadEAD(orgID, spec)
		if err != nil {
			log.Warn().Err(err).
				Str("orgID", orgID).
				Str("archiveID", spec).
				Msg("unable to get meta information probably mets only archive")

			continue
		}

		meta.ProcessDigital = true

		_, err = s.NewTask(&meta)
		if err != nil && !errors.Is(err, ErrTaskAlreadySubmitted) {
			return err
		}

		seen++
	}

	ticker := time.NewTicker(2 * time.Second)

	// block until done
	for {
		select {
		case <-ticker.C:
			m := s.Metrics()
			if (m.Finished + m.Failed + m.Canceled) == uint64(seen) {
				log.Info().Msgf("updated %d eads", seen)
				return nil
			}
		}
	}
}

func addHeader(b []byte) []byte {
	declaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>`)
	if !bytes.Contains(b, declaration) {
		b = append(declaration, b...)
	}

	return b
}
