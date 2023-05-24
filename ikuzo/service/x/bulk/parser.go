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
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oklog/ulid"
	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"

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
	s             *Service
	graphs        map[string]*fragments.FragmentBuilder // TODO: probably remove this later
	rawRequest    bytes.Buffer
	store         *redisStore
}

func (p *Parser) setUpdateDataset(ds *models.DataSet) {
	p.m.Lock()
	p.ds = ds
	p.m.Unlock()
}

func (p *Parser) dataset() *models.DataSet {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.ds
}

func (p *Parser) Parse(ctx context.Context, r io.Reader) error {
	ctx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx
	defer done()

	workers := 4

	actions := make(chan Request)

	g.Go(func() error {
		defer func() {
			close(actions)
		}()

		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 5*1024*1024)

		for scanner.Scan() {
			var req Request
			b := scanner.Bytes()

			if p.s.logRequests {
				b = append(b, '\n')
				p.rawRequest.Write(b)
			}

			if err := json.Unmarshal(b, &req); err != nil {
				atomic.AddUint64(&p.stats.JSONErrors, 1)
				log.Error().Str("svc", "bulk").Err(err).Msg("json parse error")
				log.Debug().Str("svc", "bulk").Str("raw", scanner.Text()).Err(err).Msg("wrong json input")
				atomic.AddUint64(&p.stats.JSONErrors, 1)
				continue
			}
			if p.dataset() == nil {
				p.once.Do(func() {
					p.setDataSet(&req)
				})
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
					log.Error().Err(err).Msg("unable to process action")
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

	if p.s.logRequests {
		if err := p.storeRequest(); err != nil {
			p.s.log.Warn().Err(err).Msg("unable to store request for debugging")
		}
	}

	p.s.log.Info().Msgf("graphs: %d", len(p.graphs))
	p.s.log.Info().Msgf("%#v", config.Config.RDF)

	if config.Config.RDF.RDFStoreEnabled {
		if config.Config.RDF.StoreSparqlDeltas {
			if err := p.StoreGraphDeltas(); err != nil {
				return err
			}
		} else {
			if errs := p.RDFBulkInsert(); errs != nil {
				return errs[0]
			}
		}
	}

	return nil
}

// TODO: implement this
func (p *Parser) StoreGraphDeltas() error {
	p.s.log.Info().Str("storeName", p.store.revisionSetName(rdfType, false)).Int("graphs", len(p.sparqlUpdates)).Msg("store deltas")
	updates := []fragments.SparqlUpdate{}
	for _, su := range p.sparqlUpdates {
		known, err := p.store.addID(su.HubID, rdfType, su.GetHash())
		if err != nil {
			return err
		}
		if known {
			continue
		}
		p.s.log.Info().Msgf("unknown hubID: %s", su.HubID)
		updates = append(updates, su)
		if err := p.store.storeRDFData(su); err != nil {
			return err
		}
	}

	p.s.log.Info().Int("submitted", len(p.sparqlUpdates)).Int("changed", len(updates)).Msg("after delta check")

	p.sparqlUpdates = updates
	if errs := p.RDFBulkInsert(); errs != nil {
		return errs[0]
	}

	return nil
}

// TODO: remove this
func (p *Parser) StoreGraphDeltasOld() error {
	ids := []string{}
	for _, su := range p.sparqlUpdates {
		ids = append(ids, su.HubID)
	}

	// get previously stored updates
	previousUpdates, err := p.s.getPreviousUpdates(ids)
	if err != nil {
		return err
	}
	lookUp := map[string]*fragments.SparqlUpdate{}
	for _, prev := range previousUpdates {
		lookUp[prev.HubID] = prev
	}

	// check which updates have changed
	changed := []*DiffConfig{}
	for _, current := range p.sparqlUpdates {
		prev, ok := lookUp[current.HubID]
		if !ok {
			// new so add it
			changed = append(changed, &DiffConfig{su: &current})
			continue
		}

		if prev.RDFHash == current.RDFHash {
			// same has so skip it
			continue
		}

		changed = append(changed, &DiffConfig{su: &current, previousTriples: prev.Triples, previousHash: prev.RDFHash})
	}

	var sparqlUpdate strings.Builder
	for _, cfg := range changed {
		update, err := diffAsSparqlUpdate(cfg)
		if err != nil {
			return err
		}
		sparqlUpdate.WriteString(update)
	}

	// update the diffs in the triple store
	errs := fragments.UpdateViaSparql(p.dataset().OrgID, sparqlUpdate.String())
	if errs != nil {
		return errs[0]
	}

	// store the graphs in the S3 bucket
	// if err := p.StoreGraphs(changed); err != nil {
	// 	return err
	// }

	// store the update hashes
	if err := p.s.storeUpdatedHashes(changed); err != nil {
		return err
	}

	// update revision for all seen ids
	if err := p.s.incrementRevisionForSeen(ids); err != nil {
		return err
	}

	return nil
}

// logRequest logs the bulk request to disk for inspection and reuse
func (p *Parser) storeRequest() error {
	u, err := ulid.New(ulid.Now(), nil)
	if err != nil {
		p.s.log.Error().Err(err).Msg("unable to create ulid")
		return err
	}

	path := fmt.Sprintf("/tmp/%s_%s_%d_%s.ldjson", p.ds.OrgID, p.ds.Spec, p.ds.Revision, u.String())
	return os.WriteFile(path, p.rawRequest.Bytes(), os.ModePerm)
}

// RDFBulkInsert inserts all triples from the bulkRequest in one SPARQL update statement
func (p *Parser) RDFBulkInsert() []error {
	triplesStored, errs := fragments.RDFBulkInsert(p.ds.OrgID, p.sparqlUpdates)
	p.stats.GraphsStored = uint64(len(p.sparqlUpdates))
	p.sparqlUpdates = nil
	p.stats.TriplesStored = uint64(triplesStored)

	return errs
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func (p *Parser) setDataSet(req *Request) error {
	ds, _, dsErr := models.GetOrCreateDataSet(req.OrgID, req.DatasetID)
	if dsErr != nil {
		p.s.log.Error().Err(dsErr).Msg("unable to get or create dataset")
		return dsErr
	}

	if ds.RecordType == "" {
		ds.RecordType = "narthex"
	}

	if req.Tags != "" {
		var changed bool

		tags := strings.Split(req.Tags, ",")
		for _, tag := range tags {
			if !containsString(ds.Tags, tag) {
				ds.Tags = append(ds.Tags, tag)
				changed = true
			}
		}

		if changed {
			if err := ds.Save(); err != nil {
				log.Printf("unable to save dataset: %s [%s]", err, ds.Spec)
			}
		}
	}

	p.stats.Spec = req.DatasetID
	p.stats.OrgID = req.OrgID
	p.stats.SpecRevision = uint64(ds.Revision)
	p.setUpdateDataset(ds)

	p.store = &redisStore{
		orgID: req.OrgID,
		spec:  req.DatasetID,
		c:     p.s.rc,
	}

	return nil
}

func (p *Parser) dropGraphOrphans() error {
	p.s.log.Info().Msg("dropping orphans")
	orphans, err := p.store.findOrphans(rdfType)
	if err != nil {
		return err
	}
	p.s.log.Info().Int("orphanCount", len(orphans)).Msgf("%#v", orphans)
	if len(orphans) == 0 {
		return nil
	}

	updateQuery, err := p.store.dropOrphansQuery(orphans)
	if err != nil {
		return err
	}

	errs := fragments.UpdateViaSparql(p.stats.OrgID, updateQuery)
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

func (p *Parser) dropOrphans(req *Request) error {
	m := &domainpb.IndexMessage{
		OrganisationID: req.OrgID,
		DatasetID:      req.DatasetID,
		Revision:       &domainpb.Revision{Number: int32(p.ds.Revision)},
		ActionType:     domainpb.ActionType_DROP_ORPHANS,
	}

	if err := p.bi.Publish(context.Background(), m); err != nil {
		return err
	}

	if config.Config.RDF.RDFStoreEnabled {
		if config.Config.RDF.StoreSparqlDeltas {
			go func() {
				// block for orphanWait seconds to allow cluster to be in sync
				timer := time.NewTimer(time.Second * time.Duration(5))
				<-timer.C
				if err := p.dropGraphOrphans(); err != nil {
					p.s.log.Error().Err(err).Msg("unable to drop graph orphans")
				}
			}()
		} else {
			p.s.log.Info().Msg("wrong orphan")
			_, err := models.DeleteGraphsOrphansBySpec(p.ds.OrgID, p.ds.Spec, p.ds.Revision)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

func (p *Parser) IncrementRevision() (int, error) {
	previous := p.ds.Revision
	ds, err := p.dataset().IncrementRevision()
	if err != nil {
		return 0, err
	}

	if err := p.store.SetRevision(ds.Revision, previous); err != nil {
		return 0, err
	}

	p.setUpdateDataset(ds)

	p.stats.SpecRevision = uint64(ds.Revision)

	return ds.Revision, nil
}

func (p *Parser) process(ctx context.Context, req *Request) error {
	subLogger := addLogger(req.DatasetID)

	if p.dataset() == nil {
		return fmt.Errorf("unable to get dataset")
	}

	req.Revision = p.ds.Revision

	switch req.Action {
	case "index":
		if err := p.Publish(ctx, req); err != nil {
			subLogger.Error().Err(err).Msg("unable to publish bulk index request")

			return err
		}
	case "increment_revision":
		revision, err := p.IncrementRevision()
		if err != nil {
			subLogger.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to increment DataSet")
		}

		subLogger.Info().Str("datasetID", req.DatasetID).Int("revision", revision).Msg("Incremented dataset")
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

		return err
	}

	fb, err := req.createFragmentBuilder(req.Revision)
	if err != nil {
		log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("unable to build fragment builder")
		return err
	}

	_, err = fb.ResourceMap()
	if err != nil {
		log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("unable to build resource map")
		return err
	}

	_ = fb.Doc()

	for _, tag := range fb.FragmentGraph().Meta.Tags {
		if tag == "fragmentsOnly" {
			p.indexTypes = []string{"fragments"}
		}
	}

	p.m.Lock()
	p.graphs[fb.FragmentGraph().Meta.HubID] = fb
	p.m.Unlock()

	for _, indexType := range p.indexTypes {
		switch indexType {
		case "v1":
			if err := req.processV1(ctx, fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("v1 indexing error")
				return err
			}
		case "v2":
			if err := req.processV2(ctx, fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("v2 indexing error")
				return err
			}
		case "fragments":
			if err := req.processFragments(fb, p.bi); err != nil {
				log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("v2 indexing error")
				return err
			}
		default:
			return fmt.Errorf("unknown indexType: %s", indexType)
		}
	}

	// TODO(kiivihal): get the configuration values via injection instead of global config
	if config.Config.RDF.RDFStoreEnabled {
		if err := p.AppendRDFBulkRequest(req, fb.Graph); err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("unable to append bulk request")
			return err
		}
	}

	if p.postHooks != nil {
		subject := strings.TrimSuffix(req.NamedGraphURI, "/graph")
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

// AppendRDFBulkRequest gathers all the triples from an BulkAction to be inserted in bulk.
func (p *Parser) AppendRDFBulkRequest(req *Request, g *rdf.Graph) error {
	var b bytes.Buffer
	if err := serializeNTriples(g, &b); err != nil {
		return fmt.Errorf("unable to convert RDF graph; %w", err)
	}

	su := fragments.SparqlUpdate{
		Triples:       b.String(),
		NamedGraphURI: req.NamedGraphURI,
		Spec:          req.DatasetID,
		HubID:         req.HubID,
		OrgID:         req.OrgID,
		SpecRevision:  req.Revision, // TODO: This can only be removed after the orphan control is fixed
	}

	su.GetHash()

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
	GraphsStored       uint64 `json:"graphsStored"`
	PostHooksSubmitted uint64 `json:"postHooksSubmitted"`
	// ContentHashMatches uint64    `json:"contentHashMatches"` // originally json was content_hash_matches
}

func encodeTerm(iterm rdf.Term) string {
	switch term := iterm.(type) {
	case *rdf.Resource:
		return fmt.Sprintf("<%s>", term.URI)
	case *rdf.Literal:
		return term.String()
	case *rdf.BlankNode:
		return term.String()
	}

	return ""
}

func serializeNTriples(g *rdf.Graph, w io.Writer) error {
	var err error

	triples := []string{}

	for triple := range g.IterTriplesOrdered() {
		s := encodeTerm(triple.Subject)
		if strings.HasPrefix(s, "<urn:private/") {
			continue
		}

		p := encodeTerm(triple.Predicate)
		o := encodeTerm(triple.Object)

		if strings.HasPrefix(o, "<urn:private/") {
			continue
		}

		triples = append(triples, fmt.Sprintf("%s %s %s .", s, p, o))
	}

	sort.Strings(triples)

	_, err = fmt.Fprint(w, strings.Join(triples, "\n"))
	if err != nil {
		return err
	}

	return nil
}
