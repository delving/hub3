package elasticsearch8

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// Store implements semantic.SearchStore backed by Elasticsearch 8.
type Store struct {
	client *Client
	aggMode AggregationMode
	log     zerolog.Logger

	contextMu sync.RWMutex
	contexts  map[string]*semantic.SearchContext
}

// compile-time interface checks
var (
	_ semantic.SearchStore  = (*Store)(nil)
	_ semantic.SimilarStore = (*Store)(nil)
)

// NewStore creates a new Store wrapping the given Client.
// Defaults to AggNested mode for facets, matching the v2 index's nested
// resources.entries structure.
func NewStore(client *Client, logger zerolog.Logger) *Store {
	return &Store{
		client:   client,
		aggMode:  AggNested,
		log:      logger.With().Str("svc", "es8-store").Logger(),
		contexts: make(map[string]*semantic.SearchContext),
	}
}

// SetAggregationMode sets the aggregation mode (AggFlat or AggNested).
func (s *Store) SetAggregationMode(mode AggregationMode) {
	s.aggMode = mode
}

// resolveIndices returns the list of ES index names to search.
// It always includes the primary org index, plus any context indices.
func (s *Store) resolveIndices(orgID string, contextIndices []string) []string {
	indices := []string{s.client.IndexName(orgID)}

	for _, idx := range contextIndices {
		name := s.client.IndexName(idx)
		if name != indices[0] { // avoid duplicate
			indices = append(indices, name)
		}
	}

	return indices
}

// Search executes a search query against Elasticsearch and returns results.
func (s *Store) Search(ctx context.Context, opts *semantic.QueryOptions, config *semantic.ResourceConfig) (*semantic.SearchResult, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var contextIndices []string
	if opts != nil {
		contextIndices = opts.ContextIndices
	}

	qb := NewQueryBuilder(orgID, contextIndices...)

	body, err := qb.BuildQuery(opts)
	if err != nil {
		return nil, fmt.Errorf("building search query: %w", err)
	}

	// Merge facet aggregations if present.
	if opts != nil && len(opts.Facets) > 0 {
		body, err = s.mergeAggregations(body, opts.Facets)
		if err != nil {
			return nil, fmt.Errorf("merging aggregations: %w", err)
		}
	}

	indices := s.resolveIndices(orgID, contextIndices)

	s.log.Debug().
		Strs("indices", indices).
		Str("orgID", orgID).
		Msg("executing search")

	resp, err := s.client.ES().Search(
		s.client.ES().Search.WithContext(ctx),
		s.client.ES().Search.WithIndex(indices...),
		s.client.ES().Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("executing ES search: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES search error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES search response: %w", err)
	}

	parser := &ResultParser{}

	result, err := parser.ParseSearchResponse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	// Parse facets if requested.
	if opts != nil && len(opts.Facets) > 0 {
		var esResp esSearchResponse
		if err := json.Unmarshal(data, &esResp); err == nil && len(esResp.Aggs) > 0 {
			result.Facets = parser.ParseFacets(esResp.Aggs, opts.Facets)
		}
	}

	return result, nil
}

// FindSimilar returns documents similar to the given document using
// Elasticsearch's More-Like-This (MLT) query.
func (s *Store) FindSimilar(ctx context.Context, id string, opts *semantic.SimilarOptions, config *semantic.ResourceConfig) (*semantic.SearchResult, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = semantic.DefaultSimilarOptions()
	}

	// Cap count at 20.
	count := opts.Count
	if count <= 0 {
		count = 5
	}

	if count > 20 {
		count = 20
	}

	// Translate field names: prefix with "fields." and replace ":" with "_".
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"full_text"}
	} else {
		translated := make([]string, len(fields))
		for i, f := range fields {
			translated[i] = "fields." + strings.ReplaceAll(f, ":", "_")
		}

		fields = translated
	}

	indices := s.resolveIndices(orgID, opts.ContextIndices)

	minTermFreq := opts.MinTermFreq
	if minTermFreq <= 0 {
		minTermFreq = 2
	}

	minDocFreq := opts.MinDocFreq
	if minDocFreq <= 0 {
		minDocFreq = 5
	}

	query := map[string]any{
		"query": map[string]any{
			"more_like_this": map[string]any{
				"fields": fields,
				"like": []map[string]any{
					{"_index": indices[0], "_id": id},
				},
				"min_term_freq":  minTermFreq,
				"min_doc_freq":   minDocFreq,
				"max_query_terms": 15,
			},
		},
		"size":             count,
		"track_total_hits": true,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshaling MLT query: %w", err)
	}

	s.log.Debug().
		Strs("indices", indices).
		Str("id", id).
		Int("count", count).
		Msg("executing MLT search")

	resp, err := s.client.ES().Search(
		s.client.ES().Search.WithContext(ctx),
		s.client.ES().Search.WithIndex(indices...),
		s.client.ES().Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("executing ES MLT search: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES MLT search error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES MLT response: %w", err)
	}

	parser := &ResultParser{}

	return parser.ParseSearchResponse(data)
}

// GetByID retrieves a single document by its ID from Elasticsearch.
// Supports both hubID (ES _id) and full URI (meta.entryURI) lookups.
func (s *Store) GetByID(ctx context.Context, id string, config *semantic.ResourceConfig) (map[string]any, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	index := s.client.IndexName(orgID)

	s.log.Debug().
		Str("index", index).
		Str("id", id).
		Msg("getting document by ID")

	if isURI(id) {
		return s.getByEntryURI(ctx, index, id)
	}

	resp, err := s.client.ES().Get(
		index,
		id,
		s.client.ES().Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("executing ES get: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("document %q not found", id)
		}

		return nil, fmt.Errorf("ES get error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES get response: %w", err)
	}

	parser := &ResultParser{}

	return parser.ParseGetResponse(data)
}

// getByEntryURI searches for a document by its meta.entryURI field.
func (s *Store) getByEntryURI(ctx context.Context, index, uri string) (map[string]any, error) {
	query := fmt.Sprintf(`{"query":{"term":{"meta.entryURI":{"value":%q}}},"size":1}`, uri)

	resp, err := s.client.ES().Search(
		s.client.ES().Search.WithContext(ctx),
		s.client.ES().Search.WithIndex(index),
		s.client.ES().Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("searching by entryURI: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES search by entryURI error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES search response: %w", err)
	}

	parser := &ResultParser{}

	result, err := parser.ParseSearchResponse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing entryURI search response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("document with URI %q not found", uri)
	}

	return result.Results[0], nil
}

// isURI returns true if the id looks like a URI (starts with http:// or https://).
func isURI(id string) bool {
	return strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://")
}

// Aggregate executes aggregation queries and returns only facet results.
func (s *Store) Aggregate(ctx context.Context, facets []semantic.FacetRequest, filters []semantic.Filter, config *semantic.ResourceConfig) ([]semantic.FacetResult, error) {
	orgID, err := s.orgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Build a query with size=0 (we only want aggregations).
	opts := &semantic.QueryOptions{
		Filters: filters,
		Peek:    true, // sets size=0
	}

	qb := NewQueryBuilder(orgID)

	body, err := qb.BuildQuery(opts)
	if err != nil {
		return nil, fmt.Errorf("building aggregation query: %w", err)
	}

	body, err = s.mergeAggregations(body, facets)
	if err != nil {
		return nil, fmt.Errorf("merging aggregations: %w", err)
	}

	index := s.client.IndexName(orgID)

	s.log.Debug().
		Str("index", index).
		Int("facets", len(facets)).
		Msg("executing aggregation")

	resp, err := s.client.ES().Search(
		s.client.ES().Search.WithContext(ctx),
		s.client.ES().Search.WithIndex(index),
		s.client.ES().Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("executing ES aggregation: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES aggregation error: %s", resp.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ES aggregation response: %w", err)
	}

	var esResp esSearchResponse
	if err := json.Unmarshal(data, &esResp); err != nil {
		return nil, fmt.Errorf("parsing aggregation response: %w", err)
	}

	parser := &ResultParser{}

	return parser.ParseFacets(esResp.Aggs, facets), nil
}

// SaveSearchContext saves a search context for detail-level navigation.
func (s *Store) SaveSearchContext(_ context.Context, searchCtx *semantic.SearchContext) error {
	s.contextMu.Lock()
	defer s.contextMu.Unlock()

	s.contexts[searchCtx.Token] = searchCtx

	return nil
}

// GetSearchContext retrieves a saved search context by token.
func (s *Store) GetSearchContext(_ context.Context, token string) (*semantic.SearchContext, error) {
	s.contextMu.RLock()
	defer s.contextMu.RUnlock()

	ctx, ok := s.contexts[token]
	if !ok {
		return nil, semantic.ErrContextNotFound
	}

	return ctx, nil
}

// DeleteSearchContext removes a search context by token.
func (s *Store) DeleteSearchContext(_ context.Context, token string) error {
	s.contextMu.Lock()
	defer s.contextMu.Unlock()

	delete(s.contexts, token)

	return nil
}

// Health checks if the Elasticsearch cluster is reachable and healthy.
func (s *Store) Health(ctx context.Context) error {
	resp, err := s.client.ES().Cluster.Health(
		s.client.ES().Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("ES cluster health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("ES cluster unhealthy: %s", resp.String())
	}

	return nil
}

// orgIDFromContext extracts the organization ID from the context.
func (s *Store) orgIDFromContext(ctx context.Context) (string, error) {
	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok {
		return "", fmt.Errorf("no organization in context")
	}

	return org.ID.String(), nil
}

// mergeAggregations adds aggregation clauses to an existing query body.
func (s *Store) mergeAggregations(body []byte, facets []semantic.FacetRequest) ([]byte, error) {
	var query map[string]any
	if err := json.Unmarshal(body, &query); err != nil {
		return nil, fmt.Errorf("unmarshaling query for aggregation merge: %w", err)
	}

	ab := &AggregationBuilder{Mode: s.aggMode}
	aggs := ab.BuildAggregations(facets)
	query["aggs"] = aggs

	return json.Marshal(query)
}
