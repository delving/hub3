package sparql

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/knakk/sparql"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

type HarvestConfig struct {
	URL     string
	Queries struct {
		NamespacePrefix        string
		WhereClause            string // sparql query for all identier
		SubjectVar             string // for example: `?identifier`
		IncrementalWhereClause string // using the From timestamp to harvest an incremental set
		GetGraphQuery          string // sparql query to get full graph. ?subject is injected for each
	}
	From          time.Time // the time of the last harvest
	GraphMimeType string    // if the subject can be harvested directly which mime-type to use
	MaxSubjects   int
	PageSize      int
}

func (cfg *HarvestConfig) getRepo() (*sparql.Repo, error) {
	repo, err := sparql.NewRepo(cfg.URL, sparql.Timeout(time.Millisecond*1500))
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func HarvestSubjects(ctx context.Context, cfg HarvestConfig, ids chan string) (err error) {
	whereClause := cfg.Queries.WhereClause
	if !cfg.From.IsZero() {
		whereClause = cfg.Queries.IncrementalWhereClause
	}

	countQuery := fmt.Sprintf(
		"%s \n select (count(?%s) as ?count) where {%s}",
		cfg.Queries.NamespacePrefix,
		cfg.Queries.SubjectVar,
		whereClause,
	)

	repo, err := cfg.getRepo()
	if err != nil {
		return err
	}

	res, err := repo.Query(countQuery)
	if err != nil {
		return err
	}

	totalStr, ok := res.Bindings()["count"]
	if !ok {
		return fmt.Errorf("unable to get count from result bindings: %#v \n %s",
			res.Bindings(),
			countQuery,
		)
	}

	totalIDs, err := strconv.Atoi(totalStr[0].String())
	if err != nil {
		return fmt.Errorf("error converting string to integer: %w", err)
	}

	_ = totalIDs
	var offSet int
	pageSize := 5000
	if cfg.PageSize != 0 {
		pageSize = cfg.PageSize
	}
	var seen int

harvestLoop:
	for offSet <= totalIDs {
		pagingQuery := fmt.Sprintf(
			"%s \n select ?%s where {%s} OFFSET %d LIMIT %d",
			cfg.Queries.NamespacePrefix,
			cfg.Queries.SubjectVar,
			whereClause,
			offSet,
			pageSize,
		)

		resp, err := repo.Query(pagingQuery)
		if err != nil {
			return err
		}

		subjects, ok := resp.Bindings()[cfg.Queries.SubjectVar]
		if !ok {
			return fmt.Errorf("invalid SPARQL query: %q", pagingQuery)
		}

		for _, subject := range subjects {
			if subject.String() == "" {
				continue
			}
			if cfg.MaxSubjects > 0 && seen >= cfg.MaxSubjects {
				break harvestLoop
			}

			seen++

			select {
			case <-ctx.Done():
				return ctx.Err()
			case ids <- subject.String():
			}
		}

		if len(subjects) < pageSize {
			break
		}

		offSet += pageSize
	}

	close(ids)

	return
}

func Harvest(ctx context.Context, repo *Repo, query string) (responses []*responseWithContext, err error) {
	return
}

func getSubject(c *http.Client, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "text/turtle")

	response, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("the HTTP request failed with status code: %d", response.StatusCode)
	}

	return response.Body, nil
}

func HarvestGraphs(ctx context.Context, cfg HarvestConfig, cb func(g *rdf.Graph) error) (err error) {
	subjects := make(chan string)
	g, _ := errgroup.WithContext(ctx)

	// Produce
	g.Go(func() error {
		return HarvestSubjects(ctx, cfg, subjects)
	})

	graphs := make(chan *rdf.Graph)

	c := retryablehttp.NewClient().StandardClient()

	// Map
	nWorkers := 4
	workers := int32(nWorkers)
	for i := 0; i < nWorkers; i++ {
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					close(graphs)
				}
			}()

			for subject := range subjects {
				body, err := getSubject(c, subject)
				if err != nil {
					return err
				}

				g, err := ntriples.Parse(body, nil)
				if err != nil {
					return err
				}

				s, err := rdf.NewIRI(subject)
				if err != nil {
					return err
				}

				g.Subject = rdf.Subject(s)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case graphs <- g:
				}
			}

			return nil
		})
	}

	// Reduce
	g.Go(func() error {
		for graph := range graphs {
			if graph != nil {
				if err := cb(graph); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
