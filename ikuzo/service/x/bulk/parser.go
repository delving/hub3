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

package bulk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/teris-io/shortid"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/delving/hub3/ikuzo/service/x/pid"

	irdf "github.com/delving/hub3/ikuzo/rdf"
	rdf "github.com/kiivihal/rdf2go"
)

type Parser struct {
	once       sync.Once
	ds         *models.DataSet
	stats      *Stats
	bi         index.BulkIndex
	indexTypes []string
	// TODO(kiivihal): find better solution for this
	sparqlUpdates []fragments.SparqlUpdate // store all the triples here for bulk insert
	postHooks     []*domain.PostHookItem
	m             sync.RWMutex
	debugPath     string
	recDef        RecDefResolver
	s             *Service
	// batch tracks the current indexing batch for persistent state
	batch *models.IndexingBatch
}

func (p *Parser) Parse(ctx context.Context, r io.Reader) error {
	ctx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx

	defer done()

	workers := 4

	actions := make(chan Request)

	g.Go(func() error {
		defer close(actions)

		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 5*1024*1024)

		for scanner.Scan() {
			var req Request
			b := scanner.Bytes()

			if err := json.Unmarshal(b, &req); err != nil {
				atomic.AddUint64(&p.stats.JSONErrors, 1)
				log.Error().Str("svc", "bulk").Err(err).Msg("json parse error")
				log.Debug().Str("svc", "bulk").Str("raw", scanner.Text()).Err(err).Msg("wrong json input")
				atomic.AddUint64(&p.stats.JSONErrors, 1)
				continue
			}

			select {
			case actions <- req:
			case <-gctx.Done():
				return gctx.Err()
			}
			atomic.AddUint64(&p.stats.TotalReceived, 1)
		}

		if scanner.Err() != nil {
			return scanner.Err()
		}

		return nil
	})

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for a := range actions {
				a := a

				if err := p.process(ctx, &a); err != nil {
					log.Error().Err(err).Str("orgID", a.OrgID).Msg("unable to process action")
					return err
				}

				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
					atomic.AddUint64(&p.stats.RecordsStored, 1)
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Error().Err(err).Msg("workers with errors")
			return err
		}

		log.Warn().Err(err).Msg("context canceled during bulk indexing")
	}

	if config.Config.RDF.RDFStoreEnabled {
		if errs := p.RDFBulkInsert(); errs != nil {
			return errs[0]
		}
	}

	return nil
}

// RDFBulkInsert inserts all triples from the bulkRequest in one SPARQL update statement
func (p *Parser) RDFBulkInsert() []error {
	triplesStored, errs := fragments.RDFBulkInsert(p.ds.OrgID, p.sparqlUpdates)
	p.sparqlUpdates = nil
	p.stats.TriplesStored = uint64(triplesStored)

	return errs
}

func (p *Parser) setDataSet(req *Request) {
	ds, _, dsError := models.GetOrCreateDataSet(req.OrgID, req.DatasetID)
	if dsError != nil {
		// log error
		return
	}

	if err := p.debugSetup(req.OrgID, req.DatasetID); err != nil {
		slog.Error("unable to setup debug for bulk API request", "error", err)
		p.debugPath = ""
	}

	if ds.RecordType == "" {
		ds.RecordType = "narthex"
	}

	if req.RecDefID != "" {
		ds.RecDefID = req.RecDefID
	}

	p.stats.Spec = req.DatasetID
	p.stats.DatasetID = req.DatasetID
	p.stats.OrgID = req.OrgID
	req.Revision = ds.Revision
	p.ds = ds

	req.setTags()

	for _, tag := range req.resolvedTags {
		switch tag {
		case "fragmentsOnly":
			p.indexTypes = []string{"fragments"}
		case "rdfOnly":
			p.indexTypes = []string{}
		}
	}
}

func (p *Parser) dropOrphans(req *Request) error {
	// Store expected record count for verification before orphan deletion
	// This allows the index service to verify all documents are indexed
	// before deleting orphans
	expectedCount := int(atomic.LoadUint64(&p.stats.RecordsStored))
	if expectedCount > 0 {
		p.ds.ExpectedRecords = expectedCount
		if err := models.ORM().UpdateField(
			&models.DataSet{Spec: p.ds.Spec},
			"ExpectedRecords",
			expectedCount,
		); err != nil {
			log.Warn().Err(err).Str("datasetID", req.DatasetID).
				Msg("unable to store expected record count for verification")
		}
	}

	// Update batch with expected count and transition to verifying
	if p.batch != nil {
		p.batch.ExpectedCount = expectedCount
		if err := p.batch.StartVerifying(); err != nil {
			log.Warn().Err(err).Str("batchID", p.batch.ID).
				Msg("unable to transition batch to verifying state")
		} else {
			log.Debug().
				Str("batchID", p.batch.ID).
				Int("expectedCount", expectedCount).
				Int("submittedCount", p.batch.SubmittedCount).
				Msg("batch transitioned to verifying")
		}
	}

	m := &domainpb.IndexMessage{
		OrganisationID: req.OrgID,
		DatasetID:      req.DatasetID,
		Revision:       &domainpb.Revision{Number: int32(p.ds.Revision)},
		ActionType:     domainpb.ActionType_DROP_ORPHANS,
	}

	if err := p.bi.Publish(context.Background(), m); err != nil {
		return err
	}

	return nil
}

// dropRecords validates the hubIds belong to this dataset, then asks
// the DataSet to delete them. Idempotent at the storage layer; see
// DataSet.DropRecordsByHubIDs.
func (p *Parser) dropRecords(ctx context.Context, req *Request) error {
	if len(req.HubIDs) == 0 {
		return nil
	}
	prefix := req.OrgID + "_" + req.DatasetID + "_"
	for _, hid := range req.HubIDs {
		if !strings.HasPrefix(hid, prefix) {
			return fmt.Errorf(
				"hubId %q does not belong to dataset %s/%s",
				hid, req.OrgID, req.DatasetID,
			)
		}
	}
	_, err := p.ds.DropRecordsByHubIDs(ctx, req.HubIDs)
	return err
}

func addLogger(datasetID string) zerolog.Logger {
	switch {
	case strings.HasSuffix(datasetID, "ntfoto"):
		return log.With().Str("svc", "ntfoto").Logger()

	case strings.HasPrefix(datasetID, "nt0"):
		return log.With().Str("svc", "nt").Logger()

	default:
		return log.With().Logger()
	}
}

func (p *Parser) debugSetup(orgID, datasetID string) error {
	if p.debugPath == "" {
		return nil
	}

	p.debugPath = filepath.Join(p.debugPath, orgID, datasetID)

	if err := os.MkdirAll(p.debugPath, os.ModePerm); err != nil {
		slog.Error("unable to create debug dir", "error", err)
		return err
	}

	return nil
}

func (p *Parser) debugWrite(req *Request) error {
	if p.debugPath == "" {
		return nil
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(p.debugPath, req.HubID+".json"), b, os.ModePerm)
}

func (p *Parser) process(ctx context.Context, req *Request) error {
	if req.RecDefID == "" {
		req.RecDefID = p.recDef.defaultRecDefID
	}

	p.once.Do(func() { p.setDataSet(req) })

	subLogger := addLogger(req.DatasetID)

	slog.Debug("bulk request", "data", req)

	if writeErr := p.debugWrite(req); writeErr != nil {
		slog.Error("unable to write debug messages from bulk api", "error", writeErr)
	}

	if p.ds == nil {
		return fmt.Errorf("unable to get dataset")
	}

	if len(req.aboutTypeURI) == 0 {
		aboutURI, err := p.recDef.Resolve(req.RecDefID)
		if err != nil {
			return fmt.Errorf("unable to resolve recDefID %q; %w", req.RecDefID, err)
		}
		req.aboutTypeURI = aboutURI
	}

	req.Revision = p.ds.Revision
	// TODO(kiivihal): add logger

	switch req.Action {
	case "index":
		if err := p.Publish(ctx, req); err != nil {
			subLogger.Error().Err(err).Msg("unable to publish bulk index request")

			return err
		}
	case "increment_revision":
		ds, err := p.ds.IncrementRevision()
		if err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to increment DataSet")
			return err
		}

		// Update parser's dataset reference with new revision to ensure
		// subsequent index actions use the correct revision number
		p.ds = ds

		// Create batch for tracking this indexing session
		// ExpectedCount is 0 here since we don't know upfront how many records will be indexed
		batch, batchErr := models.CreateBatch(req.OrgID, req.DatasetID, ds.Revision, 0)
		if batchErr != nil {
			subLogger.Warn().Err(batchErr).Str("datasetID", req.DatasetID).
				Msg("unable to create indexing batch, continuing without batch tracking")
		} else {
			p.batch = batch
			subLogger.Debug().Str("batchID", batch.ID).Int("revision", ds.Revision).
				Msg("created indexing batch")
		}

		subLogger.Info().Str("datasetID", req.DatasetID).Int("revision", ds.Revision).Msg("Incremented dataset")
	case "clear_orphans", "drop_orphans":
		// clear triples
		if err := p.dropOrphans(req); err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to drop orphans")
			return err
		}

		subLogger.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("mark orphans and delete them")
	case "disable_index":
		ok, err := p.ds.DropRecords(ctx, nil)
		if !ok || err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to disable index")
			return err
		}

		p.dropPosthook(req.OrgID, req.DatasetID, -1)

		subLogger.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("remove dataset from index")
	case "drop_dataset":
		ok, err := p.ds.DropAll(ctx, nil)
		if !ok || err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to drop dataset")
			return err
		}

		p.dropPosthook(req.OrgID, req.DatasetID, -1)

		subLogger.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("dropped dataset")
	case "drop_records":
		if err := p.dropRecords(ctx, req); err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to drop records")
			return err
		}
		subLogger.Info().Str("datasetID", req.DatasetID).Int("count", len(req.HubIDs)).Msg("dropped records")
	default:
		return fmt.Errorf("unknown bulk action: %s", req.Action)
	}

	return nil
}

func (p *Parser) dropPosthook(orgID, datasetID string, revision int) {
	if p.postHooks != nil {
		p.m.Lock()
		defer p.m.Unlock()

		p.postHooks = append(
			p.postHooks,
			&domain.PostHookItem{
				Deleted:   true,
				DatasetID: datasetID,
				OrgID:     orgID,
				Revision:  revision,
			},
		)
	}
}

func (p *Parser) Publish(ctx context.Context, req *Request) error {
	if err := req.valid(); err != nil {
		log.Error().
			Err(err).
			Str("datasetID", req.DatasetID).
			Msg("bulk request is not valid")

		return NewBulkError(req, "validation", "bulk request is not valid", err)
	}

	// Track submitted count in batch (if batch tracking is active)
	if p.batch != nil {
		if err := p.batch.IncrementSubmitted(1); err != nil {
			log.Warn().Err(err).Str("batchID", p.batch.ID).
				Msg("unable to increment batch submitted count")
		}
	}

	fb, err := req.createFragmentBuilder(req.Revision)
	if err != nil {
		log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("unable to build fragment builder")
		// Error already wrapped as BulkError in createFragmentBuilder for rdf_parse errors
		if _, ok := err.(*BulkError); ok {
			return err
		}
		return NewBulkError(req, "fragment_builder", "unable to build fragment builder", err)
	}

	_, err = fb.ResourceMap()
	if err != nil {
		log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("unable to build resource map")
		return NewBulkError(req, "resource_map", "unable to build resource map", err)
	}

	_, err = fb.Doc()
	if err != nil {
		log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("unable to build fg doc")
		return NewBulkError(req, "document", "unable to build fragment graph document", err)
	}

	for _, indexType := range p.indexTypes {
		switch indexType {
		case "v1":
			if err := req.processV1(ctx, fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("v1 indexing error")
				return NewBulkError(req, "indexing_v1", "v1 indexing error", err)
			}
		case "v2":
			if err := req.processV2(ctx, fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("v2 indexing error")
				return NewBulkError(req, "indexing_v2", "v2 indexing error", err)
			}
		case "fragments":
			if err := req.processFragments(fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Str("hubID", req.HubID).Msg("fragments indexing error")
				return NewBulkError(req, "indexing_fragments", "fragments indexing error", err)
			}
		default:
			return NewBulkError(req, "indexing", fmt.Sprintf("unknown indexType: %s", indexType), nil)
		}
	}

	g, err := fb.FragmentGraph().Graph()
	if err != nil {
		return fmt.Errorf("unable to get graph: %w", err)
	}

	if err := p.schedulePID(req.HubID, g); err != nil {
		return fmt.Errorf("unable to schedule PID: %w", err)
	}

	// TODO(kiivihal): get the configuration values via injection instead of global config
	if config.Config.RDF.RDFStoreEnabled {
		if err := p.AppendRDFBulkRequest(req, fb.Graph); err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("unable to append bulk request")
			return err
		}
	}

	if p.postHooks != nil {
		subject, err := fb.FragmentGraph().GetAboutURI()
		if err != nil {
			return err
		}
		g := fb.Graph

		p.m.Lock()
		defer p.m.Unlock()

		p.postHooks = append(
			p.postHooks,
			&domain.PostHookItem{
				Graph:     g,
				Deleted:   false,
				Subject:   subject,
				OrgID:     req.OrgID,
				DatasetID: req.DatasetID,
				HubID:     req.HubID,
				Revision:  int(fb.FragmentGraph().Meta.Revision),
			},
		)
	}

	return nil
}

func (p *Parser) schedulePID(hubID string, g *irdf.Graph) error {
	id, err := domain.NewHubID(hubID)
	if err != nil {
		return fmt.Errorf("unable to create hubID: %w", err)
	}

	pidData, err := pid.Extract(id, g)
	if err != nil {
		return fmt.Errorf("unable to extract PID: %w", err)
	}

	// don't enqueue anything if pid is empty
	if pidData == nil {
		return nil
	}

	task, err := pid.NewSavePIDTask(pidData)
	if err != nil {
		return fmt.Errorf("unable to create save PID task: %w", err)
	}

	_, err = p.s.EnqueueTask(task)
	if err != nil {
		log.Error().Err(err).Str("hubID", hubID).Msg("unable to enqueue PID task, continuing without PID processing")
		return nil // Don't fail the whole bulk operation just because PID task failed
	}

	return nil
}

// AppendRDFBulkRequest gathers all the triples from an BulkAction to be inserted in bulk.
func (p *Parser) AppendRDFBulkRequest(req *Request, g *rdf.Graph) error {
	var b bytes.Buffer
	if err := serializeNTriples(g, &b); err != nil {
		return fmt.Errorf("unable to convert RDF graph; %w", err)
	}

	for _, tag := range strings.Split(req.Tags, ",") {
		if strings.EqualFold(tag, "nomergegraph") {
			req.NamedGraphURI = strings.ReplaceAll(req.NamedGraphURI, "/graph", fmt.Sprintf("/%s/graph", req.DatasetID))
			break
		}
	}

	su := fragments.SparqlUpdate{
		Triples:       b.String(),
		NamedGraphURI: req.NamedGraphURI,
		Spec:          req.DatasetID,
		SpecRevision:  req.Revision,
	}

	p.m.Lock()
	p.sparqlUpdates = append(p.sparqlUpdates, su)
	p.m.Unlock()

	return nil
}

type Stats struct {
	OrgID              string `json:"orgID"`
	DatasetID          string `json:"datasetID"`
	Spec               string `json:"spec"`
	SpecRevision       uint64 `json:"specRevision"`  // version of the records stored
	TotalReceived      uint64 `json:"totalReceived"` // originally json was total_received
	RecordsStored      uint64 `json:"recordsStored"` // originally json was records_stored
	JSONErrors         uint64 `json:"jsonErrors"`
	TriplesStored      uint64 `json:"triplesStored"`
	PostHooksSubmitted uint64 `json:"postHooksSubmitted"`
	// ContentHashMatches uint64    `json:"contentHashMatches"` // originally json was content_hash_matches
}

func encodeTerm(iterm rdf.Term, shortID string) string {
	switch term := iterm.(type) {
	case *rdf.Resource:
		return fmt.Sprintf("<%s>", term.URI)
	case *rdf.Literal:
		return term.String()
	case *rdf.BlankNode:
		id := term.RawValue()
		if len(id) < 5 {
			id = id + "-" + shortID
		}
		return fmt.Sprintf("<urn:bnode:%s>", id)
	}

	return ""
}

func serializeNTriples(g *rdf.Graph, w io.Writer) error {
	var err error

	shortID, err := shortid.Generate()
	if err != nil {
		return fmt.Errorf("unable to generated shortID: %w", err)
	}

	slog.Debug("serializing ntriples with bnode seed", "seed", shortID)

	for triple := range g.IterTriples() {
		s := encodeTerm(triple.Subject, shortID)
		if strings.HasPrefix(s, "<urn:private/") {
			continue
		}

		p := encodeTerm(triple.Predicate, shortID)
		o := encodeTerm(triple.Object, shortID)

		if strings.HasPrefix(o, "<urn:private/") {
			continue
		}

		slog.Debug("serialized triple", "s", s, "p", p, "o", o)

		_, err = fmt.Fprintf(w, "%s %s %s .\n", s, p, o)
		if err != nil {
			return err
		}
	}

	return nil
}
