package bulk

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Parser struct {
	once       sync.Once
	ds         *models.DataSet
	stats      *Stats
	bi         index.BulkIndex
	indexTypes []string
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
				log.Error().Err(err).Msg("json parse error")
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
			// log.Printf("Error scanning bulkActions: %s", scanner.Err())
			return scanner.Err()
		}

		return nil
	})

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for a := range actions {
				a := a

				if err := p.process(&a); err != nil {
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

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error().Err(err).Msg("workers with errors")
		return err
	}

	return nil
}

func (p *Parser) setDataSet(req *Request) {
	ds, _, dsError := models.GetOrCreateDataSet(req.Spec)
	if dsError != nil {
		// log error
		return
	}

	p.stats.Spec = req.Spec
	p.ds = ds
}

func (p *Parser) process(req *Request) error {
	p.once.Do(func() { p.setDataSet(req) })

	if p.ds == nil {
		return fmt.Errorf("unable to get dataset")
	}

	switch req.Action {
	case "index":
		return p.Publish(req)
	case "disable_index":
		// TODO(kiivihal): implement
		// ok, err := p.ds.DropRecords(ctx, nil)
		// if !ok || err != nil {
		// // log.Printf("Unable to drop records for %s\n", req.Spec)
		// return err
		// }
		// log.Printf("remove dataset %s from the storage", action.Spec)
	case "drop_dataset":
		// TODO(kiivihal): implement
		// ok, err := response.ds.DropAll(ctx, action.wp)
		// if !ok || err != nil {
		// log.Printf("Unable to drop dataset %s", action.Spec)
		// return err
		// }
		// log.Printf("remove the dataset %s completely", action.Spec)
	case "increment_revision":
		// TODO(kiivihal): implement
	case "clear_orphans":
		// TODO(kiivihal): implement
	default:
		return fmt.Errorf("unknown bulk action: %s", req.Action)
	}

	return nil
}

func (p *Parser) Publish(req *Request) error {
	if err := req.valid(); err != nil {
		return err
	}

	// TODO(kiivihal): replace revision later
	fb, err := req.createFragmentBuilder(0)
	if err != nil {
		// log.Printf("Unable to build fragmentBuilder: %v", err)
		return err
	}

	_, err = fb.ResourceMap()
	if err != nil {
		// log.Printf("Unable to create resource map: %v", err)
		return err
	}

	for _, indexType := range p.indexTypes {
		switch indexType {
		case "v1":
			if err := req.processV1(fb, p.bi); err != nil {
				return err
			}
		case "v2":
			if err := req.processV2(fb, p.bi); err != nil {
				return err
			}
		case "fragments":
			if err := req.processFragments(fb, p.bi); err != nil {
				return err
			}
		case "rdf":
			// TODO(kiivihal): add RDF triples
		default:
			return fmt.Errorf("unknown indexType: %s", indexType)
		}
	}

	return nil
}

type Stats struct {
	Spec          string `json:"spec"`
	SpecRevision  uint64 `json:"specRevision"`  // version of the records stored
	TotalReceived uint64 `json:"totalReceived"` // originally json was total_received
	RecordsStored uint64 `json:"recordsStored"` // originally json was records_stored
	JSONErrors    uint64 `json:"jsonErrors"`
	TriplesStored uint64 `json:"triplesStored"`
	// ContentHashMatches uint64    `json:"contentHashMatches"` // originally json was content_hash_matches
}
