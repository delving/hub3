package semantic

import (
	"context"
	"time"
)

// SearchStore defines the interface for search operations.
// This abstraction allows different search backends (Elasticsearch, etc.)
// to be used interchangeably.
type SearchStore interface {
	// Search executes a search query and returns results.
	Search(ctx context.Context, opts *QueryOptions, config *ResourceConfig) (*SearchResult, error)

	// GetByID retrieves a single resource by its ID.
	GetByID(ctx context.Context, id string, config *ResourceConfig) (map[string]any, error)

	// Aggregate executes aggregation queries for facets.
	Aggregate(ctx context.Context, facets []FacetRequest, filters []Filter, config *ResourceConfig) ([]FacetResult, error)

	// SaveSearchContext saves a search context for detail-level navigation.
	SaveSearchContext(ctx context.Context, searchCtx *SearchContext) error

	// GetSearchContext retrieves a saved search context.
	GetSearchContext(ctx context.Context, token string) (*SearchContext, error)

	// DeleteSearchContext removes an expired search context.
	DeleteSearchContext(ctx context.Context, token string) error

	// Health checks if the store is healthy and accessible.
	Health(ctx context.Context) error
}

// SearchResult represents the result of a search operation.
type SearchResult struct {
	// Total number of results
	Total int64

	// Results as JSON-LD documents
	Results []map[string]any

	// Result IDs for detail-level navigation
	ResultIDs []string

	// Facet results
	Facets []FacetResult

	// Query execution time
	Took time.Duration

	// Scroll cursor for cursor-based pagination
	ScrollCursor string

	// Metadata about the search
	Metadata map[string]any
}

// FacetResult represents the result of a facet aggregation.
type FacetResult struct {
	Field       string
	FacetType   string
	TotalValues int64
	Values      []FacetValueResult
	SumOther    int64  // Count of values not included in top values
	Missing     int64  // Count of documents without this field
	Error       string // Error message if facet failed
	NextCursor  string // Opaque cursor for fetching next page of facet values
}

// FacetValueResult represents a single facet value.
type FacetValueResult struct {
	Value    string
	Label    string
	Count    int64
	Selected bool
	Metadata map[string]any
}

// GeoAggregationResult represents the result of a geospatial aggregation.
type GeoAggregationResult struct {
	Clusters []GeoClusterResult
	Bounds   *GeoBounds
}

// GeoClusterResult represents a geographic cluster.
type GeoClusterResult struct {
	Centroid *GeoPoint
	Bounds   *GeoBounds
	Count    int64
	Zoom     int
}

// StoreCapabilities describes what features a store implementation supports.
type StoreCapabilities struct {
	// SupportsFullText indicates if the store supports full-text search.
	SupportsFullText bool

	// SupportsGeoSpatial indicates if the store supports geospatial queries.
	SupportsGeoSpatial bool

	// SupportsFacets indicates if the store supports faceted search.
	SupportsFacets bool

	// SupportsAggregations indicates if the store supports advanced aggregations.
	SupportsAggregations bool

	// SupportsCursorPagination indicates if the store supports cursor-based pagination.
	SupportsCursorPagination bool

	// SupportsHighlighting indicates if the store supports search highlighting.
	SupportsHighlighting bool

	// MaxPageSize is the maximum number of results per page.
	MaxPageSize int

	// MaxFacetSize is the maximum number of facet values per facet.
	MaxFacetSize int
}

// DefaultStoreCapabilities returns a default set of capabilities.
func DefaultStoreCapabilities() StoreCapabilities {
	return StoreCapabilities{
		SupportsFullText:         true,
		SupportsGeoSpatial:       true,
		SupportsFacets:           true,
		SupportsAggregations:     true,
		SupportsCursorPagination: true,
		SupportsHighlighting:     true,
		MaxPageSize:              1000,
		MaxFacetSize:             100,
	}
}

// SearchContextStore manages search contexts for detail-level navigation.
// This can be implemented using Redis, in-memory cache, or database.
type SearchContextStore interface {
	// Save stores a search context with expiration.
	Save(ctx context.Context, searchCtx *SearchContext, ttl time.Duration) error

	// Get retrieves a search context by token.
	Get(ctx context.Context, token string) (*SearchContext, error)

	// Delete removes a search context.
	Delete(ctx context.Context, token string) error

	// Cleanup removes expired contexts.
	Cleanup(ctx context.Context) error
}

// QueryBuilder helps construct store-specific queries from QueryOptions.
// Each store implementation (Elasticsearch, etc.) provides its own QueryBuilder.
type QueryBuilder interface {
	// BuildQuery constructs a store-specific query from QueryOptions.
	BuildQuery(opts *QueryOptions, config *ResourceConfig) (any, error)

	// BuildFacetQuery constructs facet aggregations.
	BuildFacetQuery(facets []FacetRequest, config *ResourceConfig) (any, error)

	// BuildFilterQuery constructs filter queries.
	BuildFilterQuery(filters []Filter, config *ResourceConfig) (any, error)

	// BuildSortQuery constructs sort queries.
	BuildSortQuery(sorts []SortField, config *ResourceConfig) (any, error)
}

// ResultTransformer transforms store-specific results to our domain model.
type ResultTransformer interface {
	// TransformResult transforms a store result to SearchResult.
	TransformResult(storeResult any) (*SearchResult, error)

	// TransformDocument transforms a store document to JSON-LD.
	TransformDocument(storeDoc any, config *ResourceConfig) (map[string]any, error)

	// TransformFacets transforms store facet results to FacetResult.
	TransformFacets(storeFacets any, config *ResourceConfig) ([]FacetResult, error)
}

// CacheStore provides caching capabilities for search results.
type CacheStore interface {
	// Get retrieves a cached result.
	Get(ctx context.Context, key string) (*SearchResult, error)

	// Set stores a result in cache with expiration.
	Set(ctx context.Context, key string, result *SearchResult, ttl time.Duration) error

	// Delete removes a cached result.
	Delete(ctx context.Context, key string) error

	// Invalidate removes all cached results matching a pattern.
	Invalidate(ctx context.Context, pattern string) error
}

// IntrospectionStore provides introspection capabilities for indexed data.
// It can optionally scope results to a query context.
type IntrospectionStore interface {
	// IntrospectClasses returns all RDF classes found in the data.
	// If opts is non-nil, results are scoped to that query's result set.
	IntrospectClasses(ctx context.Context, opts *QueryOptions) ([]ClassInfo, error)

	// IntrospectProperties returns properties for a given class.
	IntrospectProperties(ctx context.Context, classURI string, opts *QueryOptions) ([]PropertyInfo, error)

	// IntrospectField returns value distribution for a specific field.
	IntrospectField(ctx context.Context, field string, opts *QueryOptions) (*PropertyInfo, error)

	// IntrospectPaths returns predicate paths between classes.
	IntrospectPaths(ctx context.Context, opts *QueryOptions) ([]PathInfo, error)
}

// MockIntrospectionStore is a test implementation of IntrospectionStore.
type MockIntrospectionStore struct {
	Classes    []ClassInfo
	Properties map[string][]PropertyInfo
	Paths      []PathInfo
}

func (m *MockIntrospectionStore) IntrospectClasses(_ context.Context, _ *QueryOptions) ([]ClassInfo, error) {
	return m.Classes, nil
}

func (m *MockIntrospectionStore) IntrospectProperties(_ context.Context, classURI string, _ *QueryOptions) ([]PropertyInfo, error) {
	if m.Properties == nil {
		return nil, nil
	}
	return m.Properties[classURI], nil
}

func (m *MockIntrospectionStore) IntrospectField(_ context.Context, field string, _ *QueryOptions) (*PropertyInfo, error) {
	for _, props := range m.Properties {
		for _, p := range props {
			if p.Field == field {
				return &p, nil
			}
		}
	}
	return nil, nil
}

func (m *MockIntrospectionStore) IntrospectPaths(_ context.Context, _ *QueryOptions) ([]PathInfo, error) {
	return m.Paths, nil
}

// MockStore is a simple in-memory implementation for testing.
type MockStore struct {
	results        map[string]*SearchResult
	contexts       map[string]*SearchContext
	capabilities   StoreCapabilities
}

// NewMockStore creates a new mock store.
func NewMockStore() *MockStore {
	return &MockStore{
		results:      make(map[string]*SearchResult),
		contexts:     make(map[string]*SearchContext),
		capabilities: DefaultStoreCapabilities(),
	}
}

// Search implements SearchStore.Search for testing.
func (ms *MockStore) Search(ctx context.Context, opts *QueryOptions, config *ResourceConfig) (*SearchResult, error) {
	// Simple mock implementation
	return &SearchResult{
		Total:     0,
		Results:   []map[string]any{},
		ResultIDs: []string{},
		Facets:    []FacetResult{},
		Took:      time.Millisecond,
	}, nil
}

// GetByID implements SearchStore.GetByID for testing.
func (ms *MockStore) GetByID(ctx context.Context, id string, config *ResourceConfig) (map[string]any, error) {
	return map[string]any{
		"@id":   id,
		"@type": config.ResourceType,
	}, nil
}

// Aggregate implements SearchStore.Aggregate for testing.
func (ms *MockStore) Aggregate(ctx context.Context, facets []FacetRequest, filters []Filter, config *ResourceConfig) ([]FacetResult, error) {
	results := make([]FacetResult, len(facets))
	for i, facet := range facets {
		results[i] = FacetResult{
			Field:     facet.Field,
			FacetType: "enum",
			Values:    []FacetValueResult{},
		}
	}
	return results, nil
}

// SaveSearchContext implements SearchStore.SaveSearchContext for testing.
func (ms *MockStore) SaveSearchContext(ctx context.Context, searchCtx *SearchContext) error {
	ms.contexts[searchCtx.Token] = searchCtx
	return nil
}

// GetSearchContext implements SearchStore.GetSearchContext for testing.
func (ms *MockStore) GetSearchContext(ctx context.Context, token string) (*SearchContext, error) {
	if ctx, ok := ms.contexts[token]; ok {
		return ctx, nil
	}
	return nil, ErrContextNotFound
}

// DeleteSearchContext implements SearchStore.DeleteSearchContext for testing.
func (ms *MockStore) DeleteSearchContext(ctx context.Context, token string) error {
	delete(ms.contexts, token)
	return nil
}

// Health implements SearchStore.Health for testing.
func (ms *MockStore) Health(ctx context.Context) error {
	return nil
}
