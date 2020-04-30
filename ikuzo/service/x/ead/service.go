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
	"sync/atomic"

	"github.com/delving/hub3/config"
	eadHub3 "github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
)

type metrics struct {
	Started  uint64
	Failed   uint64
	Finished uint64
}

// type CreateTreeFn func(meta *Meta, n *Node, hubID string, id string) *fragments.Tree

type Service struct {
	index   *index.Service
	dataDir string
	m       metrics
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	// create datadir
	if s.dataDir != "" {
		createErr := os.MkdirAll(config.Config.EAD.CacheDir, os.ModePerm)
		if createErr != nil {
			return nil, createErr
		}
	}

	return s, nil
}

func (s *Service) Metrics() metrics {
	return s.m
}

func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	// legacy
	// s.eadUpload(w, r)

	// new with channels
	s.handleUpload(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) Process(ctx context.Context, r io.Reader, size int64) (Meta, error) {
	ctx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx

	defer done()

	atomic.AddUint64(&s.m.Started, 1)

	// save ead and get ead context
	buf, meta, err := s.SaveEAD(r, size)
	if err != nil {
		atomic.AddUint64(&s.m.Failed, 1)
		return meta, err
	}

	// parse EAD
	ead := new(eadHub3.Cead)

	err = xml.Unmarshal(buf.Bytes(), ead)
	// err = xml.NewDecoder(buf).Decode(ead)
	if err != nil {
		atomic.AddUint64(&s.m.Failed, 1)
		return meta, fmt.Errorf("error during EAD parsing; %w", err)
	}

	cfg := eadHub3.NewNodeConfig(context.Background())
	cfg.CreateTree = eadHub3.CreateTree
	cfg.Spec = meta.Dataset
	cfg.OrgID = config.Config.OrgID

	cfg.Nodes = make(chan *eadHub3.Node, 2000)

	// create description
	g.Go(func() error {
		desc, err := eadHub3.NewDescription(ead)
		if err != nil {
			return fmt.Errorf("unable to create description; %w", err)
		}

		// TODO(kiivihal): add mutex later
		// meta.m.Lock()
		meta.Title = desc.Summary.File.Title
		cfg.Title = []string{desc.Summary.File.Title}
		// meta.m.Unlock()

		descIndex := eadHub3.NewDescriptionIndex(meta.Dataset)
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

		// TODO(kiivihal): refactor save description and add them to fg queue
		if s.index != nil {
			err = ead.SaveDescription(cfg, unitInfo, s.index)
			if err != nil {
				return fmt.Errorf("unable to create index representation of the description; %w", err)
			}
		}

		return nil
	})

	// publish nodes
	g.Go(func() error {
		_, _, err := ead.Carchdesc.Cdsc.NewNodeList(cfg)
		return err
	})

	workers := 8

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

				b, err := fg.Marshal()
				if err != nil {
					return fmt.Errorf("unable to marshal fragment graph: %w", err)
				}

				m := &domainpb.IndexMessage{
					OrganisationID: fg.Meta.GetOrgID(),
					DatasetID:      fg.Meta.GetSpec(),
					RecordID:       fg.Meta.GetHubID(),
					IndexName:      config.Config.GetIndexName(),
					Source:         b,
				}

				if err := s.index.Publish(context.Background(), m); err != nil {
					return err
				}

				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
				}
			}

			return nil
		})
	}

	// wait for all errgroup goroutines
	if err := g.Wait(); err == nil || errors.Is(err, context.Canceled) {
		atomic.AddUint64(&s.m.Finished, 1)
		meta.Clevels = cfg.Counter.GetCount()
	} else {
		atomic.AddUint64(&s.m.Failed, 1)
		fmt.Printf("received error: %v", err)
		return meta, err
	}

	return meta, nil
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

	meta, err := s.Process(r.Context(), in, header.Size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, meta)
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
	meta.Dataset, err = s.GetName(buf)
	if err != nil {
		return meta, err
	}

	meta.basePath = s.getDataPath(meta.Dataset)

	// create dataDir
	if err := os.MkdirAll(meta.basePath, os.ModePerm); err != nil {
		return meta, err
	}

	if err := os.Rename(tmpFile, fmt.Sprintf("%s/%s.xml", meta.basePath, meta.Dataset)); err != nil {
		return meta, err
	}

	return meta, nil
}
