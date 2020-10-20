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
	"strings"
	"sync"
	"sync/atomic"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

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

			if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
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

	if config.Config.RDF.RDFStoreEnabled {
		if errs := p.RDFBulkInsert(); errs != nil {
			return errs[0]
		}
	}

	return nil
}

// RDFBulkInsert inserts all triples from the bulkRequest in one SPARQL update statement
func (p *Parser) RDFBulkInsert() []error {
	triplesStored, errs := fragments.RDFBulkInsert(p.sparqlUpdates)
	p.sparqlUpdates = nil
	p.stats.TriplesStored = uint64(triplesStored)

	return errs
}

func (p *Parser) setDataSet(req *Request) {
	ds, _, dsError := models.GetOrCreateDataSet(req.DatasetID)
	if dsError != nil {
		// log error
		return
	}

	p.stats.Spec = req.DatasetID
	p.stats.OrgID = req.OrgID
	req.Revision = ds.Revision
	p.ds = ds
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

	return nil
}

func (p *Parser) process(ctx context.Context, req *Request) error {
	p.once.Do(func() { p.setDataSet(req) })

	if p.ds == nil {
		return fmt.Errorf("unable to get dataset")
	}

	req.Revision = p.ds.Revision
	// TODO(kiivihal): add logger

	switch req.Action {
	case "index":
		if err := p.Publish(ctx, req); err != nil {
			log.Error().Err(err).Msg("unable to publish bulk index request")

			return err
		}
	case "increment_revision":
		ds, err := p.ds.IncrementRevision()
		if err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to increment DataSet")
			return err
		}

		log.Info().Str("datasetID", req.DatasetID).Int("revision", ds.Revision).Msg("Incremented dataset")
	case "clear_orphans":
		// clear triples
		if err := p.dropOrphans(req); err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to drop orphans")
			return err
		}

		log.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("mark orphans and delete them")
	case "disable_index":
		ok, err := p.ds.DropRecords(ctx, nil)
		if !ok || err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to disable index")
			return err
		}

		p.dropPosthook(req.OrgID, req.DatasetID, -1)

		log.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("remove dataset from index")
	case "drop_dataset":
		ok, err := p.ds.DropAll(ctx, nil)
		if !ok || err != nil {
			log.Error().Err(err).Str("datasetID", req.DatasetID).Msg("Unable to drop dataset")
			return err
		}

		p.dropPosthook(req.OrgID, req.DatasetID, -1)

		log.Info().Str("datasetID", req.DatasetID).Int("revision", p.ds.Revision).Msg("dropped dataset")
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
		g := fb.SortedGraph

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
	if err := g.Serialize(&b, "text/turtle"); err != nil {
		return fmt.Errorf("unable to convert RDF graph; %w", err)
	}

	su := fragments.SparqlUpdate{
		Triples:       b.String(),
		NamedGraphURI: req.NamedGraphURI,
		Spec:          req.DatasetID,
		SpecRevision:  req.Revision,
	}

	p.sparqlUpdates = append(p.sparqlUpdates, su)

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
