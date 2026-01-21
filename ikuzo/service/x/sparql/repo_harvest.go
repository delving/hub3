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

// HarvestError is a custom error type that can be serialized to TOML.
// It preserves the original error for runtime use while serializing
// only the message string for persistence.
type HarvestError struct {
	Message string
	Err     error `toml:"-"` // original error, not serialized
}

func (e *HarvestError) Error() string {
	return e.Message
}

// Unwrap returns the original error for use with errors.Is/errors.As
func (e *HarvestError) Unwrap() error {
	return e.Err
}

func (e HarvestError) MarshalText() ([]byte, error) {
	return []byte(e.Message), nil
}

func (e *HarvestError) UnmarshalText(text []byte) error {
	e.Message = string(text)
	return nil
}

// NewHarvestError creates a HarvestError from any error
func NewHarvestError(err error) *HarvestError {
	if err == nil {
		return nil
	}
	return &HarvestError{
		Message: err.Error(),
		Err:     err,
	}
}

// retryLogger implements retryablehttp.LeveledLogger to log retry events
type retryLogger struct{}

func (l *retryLogger) Error(msg string, keysAndValues ...interface{}) {
	slog.Error(msg, keysAndValues...)
}

func (l *retryLogger) Info(msg string, keysAndValues ...interface{}) {
	slog.Info(msg, keysAndValues...)
}

func (l *retryLogger) Debug(msg string, keysAndValues ...interface{}) {
	slog.Debug(msg, keysAndValues...)
}

func (l *retryLogger) Warn(msg string, keysAndValues ...interface{}) {
	slog.Warn(msg, keysAndValues...)
}

// HarvestMetrics tracks timing and progress statistics during harvest
type HarvestMetrics struct {
	StartTime       time.Time
	TotalFetchTime  time.Duration
	TotalParseTime  time.Duration
	RequestCount    int64
	SuccessCount    int64
	FailCount       int64
	TotalTriples    int64
	mu              sync.Mutex
}

func (m *HarvestMetrics) Add(fetchTime, parseTime time.Duration, triples int, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalFetchTime += fetchTime
	m.TotalParseTime += parseTime
	m.RequestCount++
	m.TotalTriples += int64(triples)
	if success {
		m.SuccessCount++
	} else {
		m.FailCount++
	}
}

func (m *HarvestMetrics) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	elapsed := time.Since(m.StartTime)
	avgFetch := time.Duration(0)
	avgParse := time.Duration(0)
	if m.RequestCount > 0 {
		avgFetch = m.TotalFetchTime / time.Duration(m.RequestCount)
		avgParse = m.TotalParseTime / time.Duration(m.RequestCount)
	}
	return fmt.Sprintf(
		"elapsed=%v requests=%d success=%d fail=%d triples=%d avgFetch=%v avgParse=%v totalFetch=%v totalParse=%v",
		elapsed.Round(time.Second), m.RequestCount, m.SuccessCount, m.FailCount, m.TotalTriples,
		avgFetch.Round(time.Millisecond), avgParse.Round(time.Millisecond),
		m.TotalFetchTime.Round(time.Second), m.TotalParseTime.Round(time.Second),
	)
}

// DereferenceCache provides thread-safe caching for dereferenced graphs
type DereferenceCache struct {
	cache map[string]*rdf.Graph
	mu    sync.RWMutex
	hits  int64
	miss  int64
}

func NewDereferenceCache() *DereferenceCache {
	return &DereferenceCache{
		cache: make(map[string]*rdf.Graph),
	}
}

func (c *DereferenceCache) Get(uri string) (*rdf.Graph, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g, ok := c.cache[uri]
	if ok {
		atomic.AddInt64(&c.hits, 1)
	} else {
		atomic.AddInt64(&c.miss, 1)
	}
	return g, ok
}

func (c *DereferenceCache) Set(uri string, g *rdf.Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[uri] = g
}

func (c *DereferenceCache) Stats() (size int, hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache), atomic.LoadInt64(&c.hits), atomic.LoadInt64(&c.miss)
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
	From                   time.Time // the time of the last harvest
	GraphMimeType          string    // if the subject can be harvested directly which mime-type to use
	MaxSubjects            int
	PageSize               int
	TotalSizeSubjects      int
	HarvestErrors          map[string]*HarvestError
	DereferencePredicates  []string // List of predicate URIs whose object IRIs should be dereferenced
	DereferenceMaxDepth    int      // Maximum depth for recursive dereferencing (default 1)
	ForceHarvest           bool     // Force harvest regardless of revision check
	rw                     sync.RWMutex
	client                 *http.Client
	OutputDir              string
	LastCheck              time.Time
	TargetDatasets         map[string]int
	Tags                   []string
	repo                   *sparql.Repo
	AboutTypeURI           string
	Metrics                *HarvestMetrics   `toml:"-"` // runtime only, not persisted
	dereferencePredicates  map[string]bool   // internal lookup map
	dereferenceCache       *DereferenceCache // cache for dereferenced graphs
}

func (cfg *HarvestConfig) AddError(subject string, err error) {
	cfg.rw.Lock()
	if len(cfg.HarvestErrors) == 0 {
		cfg.HarvestErrors = map[string]*HarvestError{}
	}
	cfg.HarvestErrors[subject] = NewHarvestError(err)
	cfg.rw.Unlock()
}

// initDereferencePredicates initializes the internal lookup map for dereference predicates
func (cfg *HarvestConfig) initDereferencePredicates() {
	cfg.dereferencePredicates = make(map[string]bool)
	for _, p := range cfg.DereferencePredicates {
		cfg.dereferencePredicates[p] = true
	}
	if cfg.DereferenceMaxDepth == 0 {
		cfg.DereferenceMaxDepth = 1 // default to 1 level of dereferencing
	}
	// Initialize dereference cache
	cfg.dereferenceCache = NewDereferenceCache()
}

// shouldDereference checks if a predicate URI should have its objects dereferenced
func (cfg *HarvestConfig) shouldDereference(predicateURI string) bool {
	if cfg.dereferencePredicates == nil {
		return false
	}
	return cfg.dereferencePredicates[predicateURI]
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
								?o5 ?p6 ?o6 .
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

	slog.Info("executing count query", "query", countQuery)

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

	slog.Info("count query result", "totalIDs", totalIDs, "previousTotal", cfg.TotalSizeSubjects)

	if totalIDs == 0 {
		slog.Warn("no subjects found in SPARQL endpoint", "query", countQuery)
		return nil
	}

	cfg.TotalSizeSubjects = totalIDs
	slog.Info("found subjects to harvest", "total", totalIDs)

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

		slog.Info("executing paging query", "offset", offSet, "limit", pageSize, "query", pagingQuery)

		resp, err := repo.Query(pagingQuery)
		if err != nil {
			return err
		}

		subjects, ok := resp.Bindings()[cfg.Queries.SubjectVar]
		if !ok {
			return fmt.Errorf("invalid SPARQL query: %q", pagingQuery)
		}

		slog.Debug("received subjects from SPARQL query", "count", len(subjects), "offset", offSet, "pageSize", pageSize)

		for _, subject := range subjects {
			if subject.String() == "" {
				continue
			}
			if cfg.MaxSubjects > 0 && seen >= cfg.MaxSubjects {
				break harvestLoop
			}

			seen++

			if seen%100 == 0 {
				slog.Debug("sending subjects to harvest", "sent", seen, "total", totalIDs)
			}

			select {
			case <-ctx.Done():
				slog.Warn("producer context cancelled", "seen", seen, "total", totalIDs, "error", ctx.Err())
				return ctx.Err()
			case ids <- subject.String():
			}
		}

		// Only break if we got zero results, meaning we've exhausted the dataset
		if len(subjects) == 0 {
			slog.Debug("no more subjects returned from SPARQL", "offset", offSet)
			break
		}

		// If we got fewer than pageSize, we might be on the last page, but continue
		// until we get zero results to ensure we don't miss any data
		offSet += pageSize
	}

	slog.Info("finished sending all subjects to harvest", "total", seen)
	return
}

func HarvestGraph(ctx context.Context, cfg *HarvestConfig, subject string) (*rdf.Graph, error) {
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
	// Track already-fetched URIs to avoid infinite loops
	fetched := make(map[string]bool)
	fetched[subject] = true

	return harvestSubjectWithDereference(ctx, subject, cfg, fetched, 0)
}

func harvestSubjectWithDereference(ctx context.Context, subject string, cfg *HarvestConfig, fetched map[string]bool, depth int) (*rdf.Graph, error) {
	// Create a per-request context with timeout to prevent hanging
	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	fetchStart := time.Now()
	body, err := getSubject(reqCtx, cfg.client, subject, cfg.GraphMimeType)
	fetchDuration := time.Since(fetchStart)
	if err != nil {
		slog.Debug("HTTP fetch failed", "subject", subject, "duration", fetchDuration, "error", err)
		if cfg.Metrics != nil {
			cfg.Metrics.Add(fetchDuration, 0, 0, false)
		}
		return nil, fmt.Errorf("unable to retrieve rdf: %w", err)
	}
	defer body.Close()

	parseStart := time.Now()
	g, err := ntriples.Parse(body, nil)
	parseDuration := time.Since(parseStart)
	if err != nil {
		slog.Debug("parsing failed", "subject", subject, "fetchDuration", fetchDuration, "parseDuration", parseDuration, "error", err)
		if cfg.Metrics != nil {
			cfg.Metrics.Add(fetchDuration, parseDuration, 0, false)
		}
		return nil, fmt.Errorf("unable to parse rdf: %w", err)
	}

	tripleCount := g.Len()
	if cfg.Metrics != nil {
		cfg.Metrics.Add(fetchDuration, parseDuration, tripleCount, true)
	}

	slog.Debug("harvested subject", "subject", subject, "fetchDuration", fetchDuration, "parseDuration", parseDuration, "triples", tripleCount, "depth", depth)

	// Dereference objects for configured predicates
	if len(cfg.DereferencePredicates) > 0 && depth < cfg.DereferenceMaxDepth {
		if err := dereferenceObjects(ctx, g, cfg, fetched, depth); err != nil {
			slog.Warn("error dereferencing objects", "subject", subject, "error", err)
			// Continue even if dereferencing fails - we still have the main subject data
		}
	}

	s, err := rdf.NewIRI(subject)
	if err != nil {
		return nil, fmt.Errorf("unable to parse subject; %w", err)
	}

	g.Subject = rdf.Subject(s)

	return g, nil
}

// dereferenceObjects fetches and merges triples for object IRIs of configured predicates
func dereferenceObjects(ctx context.Context, g *rdf.Graph, cfg *HarvestConfig, fetched map[string]bool, depth int) error {
	// Collect URIs to dereference (avoid modifying graph while iterating)
	var toFetch []string
	var fromCache []string

	for _, resource := range g.Resources() {
		for _, predicate := range resource.SortedPredicates() {
			predicateURI := predicate.IRI().RawValue()
			if !cfg.shouldDereference(predicateURI) {
				continue
			}

			for _, obj := range predicate.Objects() {
				if obj.Type() != rdf.TermIRI {
					continue
				}

				objURI := obj.RawValue()
				if fetched[objURI] {
					continue // Already processed in this record
				}

				// Mark as processed for this record
				fetched[objURI] = true

				// Check cache first
				if cfg.dereferenceCache != nil {
					if cachedGraph, ok := cfg.dereferenceCache.Get(objURI); ok {
						// Merge cached triples into the main graph
						for _, triple := range cachedGraph.Triples() {
							g.Add(triple)
						}
						fromCache = append(fromCache, objURI)
						continue
					}
				}

				toFetch = append(toFetch, objURI)
			}
		}
	}

	if len(toFetch) == 0 && len(fromCache) == 0 {
		return nil
	}

	if len(fromCache) > 0 {
		slog.Debug("dereference cache hits", "count", len(fromCache), "depth", depth+1)
	}

	if len(toFetch) == 0 {
		return nil
	}

	slog.Debug("dereferencing objects", "count", len(toFetch), "fromCache", len(fromCache), "predicates", cfg.DereferencePredicates, "depth", depth+1)

	// Fetch each object URI and merge into the graph
	for _, uri := range toFetch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		objGraph, err := harvestSubjectWithDereference(ctx, uri, cfg, fetched, depth+1)
		if err != nil {
			slog.Debug("failed to dereference object", "uri", uri, "error", err)
			continue // Continue with other objects even if one fails
		}

		// Store in cache for future use
		if cfg.dereferenceCache != nil {
			cfg.dereferenceCache.Set(uri, objGraph)
		}

		// Merge the fetched triples into the main graph
		for _, triple := range objGraph.Triples() {
			g.Add(triple)
		}

		slog.Debug("dereferenced and merged object", "uri", uri, "triples", objGraph.Len())
	}

	return nil
}

func getSubject(ctx context.Context, c *http.Client, uri, mimeType string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", mimeType)

	// Generate curl command equivalent for logging
	curlCmd := fmt.Sprintf("curl -X GET -L '%s' -H 'Accept: %s'", uri, mimeType)
	slog.Debug("SPARQL request as curl command", "curl", curlCmd)

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
	// Initialize metrics
	cfg.Metrics = &HarvestMetrics{StartTime: time.Now()}

	// Initialize dereference predicates lookup map
	cfg.initDereferencePredicates()

	subjects := make(chan string, 100) // buffered channel to prevent blocking
	g, gCtx := errgroup.WithContext(ctx) // Use errgroup context for proper cancellation propagation

	// Produce
	g.Go(func() error {
		slog.Debug("producer starting")
		if len(cfg.HarvestErrors) == 0 {
			err := HarvestSubjects(gCtx, cfg, subjects)
			if err != nil {
				slog.Error("HarvestSubjects returned error", "error", err)
			} else {
				slog.Debug("HarvestSubjects completed successfully")
			}
			return err
		}
		oldErrors := cfg.HarvestErrors
		cfg.HarvestErrors = map[string]*HarvestError{}
		for subject := range oldErrors {
			subjects <- subject
		}
		close(subjects)
		return nil
	})

	graphs := make(chan *rdf.Graph, 100) // buffered channel to prevent workers from blocking

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = &retryLogger{} // Custom logger for retry events
	retryClient.RetryMax = 3

	retryClient.HTTPClient.Timeout = 30 * time.Second // Increased from 8s to handle slow responses

	cfg.client = retryClient.StandardClient()

	// Map
	nWorkers := 2 // Reduced from 4 to avoid rate limiting on handle.net
	workers := int32(nWorkers)
	for i := 0; i < nWorkers; i++ {
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					slog.Debug("last worker closing graphs channel")
					close(graphs)
				}
			}()

			var err error
			var processed int
			var failed int
			var succeeded int
			slog.Debug("worker starting to consume subjects")
			for subject := range subjects {
				processed++
				remaining := cfg.TotalSizeSubjects - int(cfg.Metrics.RequestCount)
				slog.Debug("worker received subject from channel", "subject", subject, "processed", processed, "total", cfg.TotalSizeSubjects, "remaining", remaining)
				if processed%10 == 0 {
					slog.Info("worker progress", "processed", processed, "succeeded", succeeded, "failed", failed, "total", cfg.TotalSizeSubjects, "metrics", cfg.Metrics.String())
				}

				// Check for context cancellation before processing
				select {
				case <-gCtx.Done():
					slog.Info("context cancelled, worker stopping", "processed", processed, "error", gCtx.Err())
					return gCtx.Err()
				default:
				}

				var g *rdf.Graph
				switch {
				case cfg.GraphMimeType != "":
					g, err = harvestSubject(gCtx, subject, cfg)
					if err != nil {
						failed++
						cfg.AddError(subject, err)
						slog.Debug("failed to harvest subject", "uri", subject, "error", err)
						continue
					}
					succeeded++
				default:
					g, err = HarvestGraph(gCtx, cfg, subject)
					if err != nil {
						failed++
						cfg.AddError(subject, err)
						slog.Debug("failed to harvest subject", "uri", subject, "error", err)
						continue
					}
					succeeded++
				}

				slog.Debug("attempting to send graph to channel", "subject", subject, "graphsChannelLen", len(graphs))
				select {
				case <-gCtx.Done():
					slog.Info("context cancelled while sending graph", "subject", subject, "error", gCtx.Err())
					return gCtx.Err()
				case graphs <- g:
					slog.Debug("sent graph to channel", "subject", subject)
				}
			}

			slog.Info("worker finished", "processed", processed, "succeeded", succeeded, "failed", failed)
			return nil
		})
	}

	// Progress reporter - logs stats every 5 seconds to detect hangs
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		var lastCount int64
		for {
			select {
			case <-done:
				return
			case <-gCtx.Done():
				slog.Info("harvest interrupted", "error", gCtx.Err(), "metrics", cfg.Metrics.String())
				return
			case <-ticker.C:
				currentCount := cfg.Metrics.RequestCount
				rate := float64(currentCount-lastCount) / 5.0
				remaining := cfg.TotalSizeSubjects - int(currentCount)
				slog.Info("harvest progress",
					"processed", currentCount,
					"remaining", remaining,
					"total", cfg.TotalSizeSubjects,
					"rate", fmt.Sprintf("%.1f/s", rate),
					"metrics", cfg.Metrics.String(),
				)
				lastCount = currentCount
			}
		}
	}()

	// Reduce
	g.Go(func() error {
		var consumed int
		slog.Debug("reducer starting")
		for graph := range graphs {
			slog.Debug("reducer waiting for graph", "consumed", consumed, "graphsChannelLen", len(graphs))
			if graph != nil {
				consumed++
				slog.Debug("consuming graph", "subject", graph.Subject.RawValue(), "consumed", consumed)
				cbStart := time.Now()
				if err := cb(graph); err != nil {
					slog.Error("callback error in reducer", "error", err, "consumed", consumed)
					return err
				}
				slog.Debug("callback completed", "subject", graph.Subject.RawValue(), "duration", time.Since(cbStart))
			}
		}

		slog.Debug("reduce finished", "totalConsumed", consumed)
		return nil
	})

	err = g.Wait()
	close(done) // Stop the progress reporter

	// Always print final metrics
	slog.Info("harvest finished", "metrics", cfg.Metrics.String())

	// Log cache stats if dereferencing was used
	if cfg.dereferenceCache != nil {
		size, hits, misses := cfg.dereferenceCache.Stats()
		if size > 0 || hits > 0 || misses > 0 {
			hitRate := float64(0)
			if hits+misses > 0 {
				hitRate = float64(hits) / float64(hits+misses) * 100
			}
			slog.Info("dereference cache stats", "cached", size, "hits", hits, "misses", misses, "hitRate", fmt.Sprintf("%.1f%%", hitRate))
		}
	}

	if err != nil {
		slog.Error("errgroup returned error", "error", err)
		return err
	}

	slog.Info("HarvestGraphs completed successfully")
	return nil
}
