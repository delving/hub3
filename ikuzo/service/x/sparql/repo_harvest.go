package sparql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/knakk/sparql"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

type HarvestDataSet struct {
	Spec      string
	Revision  int
	LastCheck time.Time
}

type HarvestConfig struct {
	OrgID   string
	Spec    string
	URL     string
	Queries struct {
		NamespacePrefix        string
		WhereClause            string // sparql query for all identier
		SubjectVar             string // for example: `?identifier`
		IncrementalWhereClause string // using the From timestamp to harvest an incremental set
		GetGraphQuery          string // sparql query to get full graph. ?subject is injected for each
	}
	From              time.Time // the time of the last harvest
	GraphMimeType     string    // if the subject can be harvested directly which mime-type to use
	MaxSubjects       int
	PageSize          int
	TotalSizeSubjects int
	HarvestErrors     map[string]error
	rw                sync.RWMutex
	client            *http.Client
	OutputDir         string
	LastCheck         time.Time
	TargetDatasets    map[string]int
	Tags              []string
	repo              *sparql.Repo
}

func (cfg *HarvestConfig) AddError(subject string, err error) {
	cfg.rw.Lock()
	if len(cfg.HarvestErrors) == 0 {
		cfg.HarvestErrors = map[string]error{}
	}
	cfg.HarvestErrors[subject] = err
	cfg.rw.Unlock()
}

func (cfg *HarvestConfig) getRepo() (*sparql.Repo, error) {
	if cfg.repo == nil {
		repo, err := sparql.NewRepo(cfg.URL, sparql.Timeout(time.Second*10))
		if err != nil {
			return nil, err
		}

		cfg.repo = repo
	}

	return cfg.repo, nil
}

func HarvestWithContext(ctx context.Context, cfg *HarvestConfig, subject string) (res *responseWithContext, err error) {
	queryFmt := `
		SELECT *
		WHERE {
		graph ?graph {
			BIND(<%s> as ?s1)
			?s1 ?p1 ?o1 .
			OPTIONAL {
				?o1 ?p2 ?o2 .
				OPTIONAL {
					?o2 ?p3 ?o3 .
					OPTIONAL {
					?o3 ?p4 ?o4 .
						OPTIONAL {
							?o4 ?p5 ?o5 .
							OPTIONAL {
								?o5 ?p6 ?o .

							}
						}
					}
				}
			}
	    }
		}
		LIMIT 2500`

	if cfg.Queries.GetGraphQuery != "" {
		queryFmt = cfg.Queries.GetGraphQuery
	}
	if !strings.Contains(queryFmt, "BIND(<%s> as ?s1)") {
		return nil, errors.New("getGraphQuery should contain 'BIND(<%s> as ?s1)'")
	}

	q := fmt.Sprintf(
		queryFmt,
		subject,
	)
	repo, err := cfg.getRepo()
	if err != nil {
		return nil, err
	}

	resp, err := repo.Query(q)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	replacements := map[string]string{
		"Value":    "value",
		"Type":     "type",
		"DataType": "datatype",
		"Vars":     "vars",
		"Head":     "head",
		"Link":     "link",
		"Results":  "results",
		"Bindings": "bindings",
	}

	for oldKey, newValue := range replacements {
		oldValue := []byte(fmt.Sprintf("\"%s\":", oldKey))
		newValueBytes := []byte(fmt.Sprintf("\"%s\":", newValue))
		b = bytes.ReplaceAll(b, oldValue, newValueBytes)
	}

	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func HarvestSubjects(ctx context.Context, cfg *HarvestConfig, ids chan string) (err error) {
	defer close(ids)

	layout := "2006-01-02T15:04:05.999Z"

	whereClause := cfg.Queries.WhereClause
	if !cfg.From.IsZero() {
		whereClause = cfg.Queries.IncrementalWhereClause
		whereClause = strings.ReplaceAll(whereClause, "~~DATE~~", cfg.From.Format(layout))
	}

	countQuery := fmt.Sprintf(
		`%s
		select (count(distinct ?%s) as ?count)
		where {%s}
	    `,
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

	if totalIDs == 0 {
		return nil
	}

	cfg.TotalSizeSubjects = totalIDs
	var offSet int
	pageSize := 5000
	if cfg.PageSize != 0 {
		pageSize = cfg.PageSize
	}
	var seen int

harvestLoop:
	for offSet <= totalIDs {
		pagingQuery := fmt.Sprintf(
			"%s \n select distinct ?%s where {%s} OFFSET %d LIMIT %d",
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

	return
}

func harvestGraph(ctx context.Context, cfg *HarvestConfig, subject string) (*rdf.Graph, error) {
	resp, err := HarvestWithContext(ctx, cfg, subject)
	if err != nil {
		return nil, fmt.Errorf("unable to harvest with context: %w", err)
	}

	g, err := resp.Graph()
	if err != nil {
		return nil, err
	}

	s, err := rdf.NewIRI(subject)
	if err != nil {
		return nil, fmt.Errorf("unable to parse subject; %w", err)
	}

	g.Subject = rdf.Subject(s)

	return g, nil
}

func harvestSubject(ctx context.Context, subject string, cfg *HarvestConfig) (*rdf.Graph, error) {
	body, err := getSubject(cfg.client, subject, cfg.GraphMimeType)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve rdf: %w", err)
	}

	g, err := ntriples.Parse(body, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to parse rdf: %w", err)
	}

	s, err := rdf.NewIRI(subject)
	if err != nil {
		return nil, fmt.Errorf("unable to parse subject; %w", err)
	}

	g.Subject = rdf.Subject(s)

	return g, nil
}

func getSubject(c *http.Client, uri, mimeType string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", mimeType)

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

func HarvestGraphs(ctx context.Context, cfg *HarvestConfig, cb func(g *rdf.Graph) error) (err error) {
	subjects := make(chan string)
	g, _ := errgroup.WithContext(ctx)

	// Produce
	g.Go(func() error {
		if len(cfg.HarvestErrors) == 0 {
			return HarvestSubjects(ctx, cfg, subjects)
		}
		oldErrors := cfg.HarvestErrors
		cfg.HarvestErrors = map[string]error{}
		for subject := range oldErrors {
			subjects <- subject
		}
		close(subjects)
		return nil
	})

	graphs := make(chan *rdf.Graph)

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryMax = 3

	retryClient.HTTPClient.Timeout = 8 * time.Second

	cfg.client = retryClient.StandardClient()

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

			var err error
			for subject := range subjects {
				var g *rdf.Graph
				switch {
				case cfg.GraphMimeType != "":
					g, err = harvestSubject(ctx, subject, cfg)
					if err != nil {
						cfg.AddError(subject, err)
						slog.Error("unable to harvest subject", "uri", subject, "error", err)
						continue
					}
				default:
					g, err = harvestGraph(ctx, cfg, subject)
					if err != nil {
						cfg.AddError(subject, err)
						slog.Error("unable to harvest subject", "uri", subject, "error", err)
						continue
					}
				}

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
