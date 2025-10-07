package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

// SemanticStore implements semantic.SearchStore using Elasticsearch.
type SemanticStore struct {
	client *elastic.Client
	index  string
	log    zerolog.Logger
}

// NewSemanticStore creates a new Elasticsearch-backed semantic store.
func NewSemanticStore(client *elastic.Client, index string, logger zerolog.Logger) *SemanticStore {
	return &SemanticStore{
		client: client,
		index:  index,
		log:    logger.With().Str("component", "semantic_store").Logger(),
	}
}

// Search executes a search query and returns results.
func (s *SemanticStore) Search(
	ctx context.Context,
	opts *semantic.QueryOptions,
	config *semantic.ResourceConfig,
) (*semantic.SearchResult, error) {
	startTime := time.Now()

	// Build the search query
	query := s.buildQuery(opts, config)

	// Build search service
	searchService := s.client.Search(s.index).
		Query(query).
		TrackTotalHits(true)

	// Apply pagination
	if opts.Pagination != nil {
		from := (opts.Pagination.Page - 1) * opts.Pagination.Size
		searchService = searchService.From(from).Size(opts.Pagination.Size)
	}

	// Apply sorting
	for _, sort := range opts.Sort {
		ascending := sort.Direction == semantic.SortAsc || sort.Direction == ""
		searchService = searchService.Sort(sort.Field, ascending)
	}

	// Add facet aggregations
	if len(opts.Facets) > 0 {
		for _, facet := range opts.Facets {
			agg := s.buildFacetAggregation(facet, config)
			if agg != nil {
				searchService = searchService.Aggregation(facet.Field, agg)
			}
		}
	}

	// Execute search
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}

	// Transform results
	result := &semantic.SearchResult{
		Total:     searchResult.TotalHits(),
		Results:   make([]map[string]any, 0, len(searchResult.Hits.Hits)),
		ResultIDs: make([]string, 0, len(searchResult.Hits.Hits)),
		Facets:    make([]semantic.FacetResult, 0),
		Took:      time.Since(startTime),
		Metadata:  make(map[string]any),
	}

	// Extract documents
	for _, hit := range searchResult.Hits.Hits {
		var doc map[string]any
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			s.log.Warn().Err(err).Str("id", hit.Id).Msg("failed to unmarshal document")
			continue
		}

		// Ensure @id is set
		if _, ok := doc["@id"]; !ok {
			doc["@id"] = hit.Id
		}

		result.Results = append(result.Results, doc)
		result.ResultIDs = append(result.ResultIDs, hit.Id)
	}

	// Extract facets
	if len(opts.Facets) > 0 {
		result.Facets = s.extractFacets(searchResult, opts.Facets, config)
	}

	// Store metadata
	result.Metadata["elasticsearch_took_ms"] = searchResult.TookInMillis

	return result, nil
}

// GetByID retrieves a single resource by its ID.
func (s *SemanticStore) GetByID(
	ctx context.Context,
	id string,
	config *semantic.ResourceConfig,
) (map[string]any, error) {
	getResult, err := s.client.Get().
		Index(s.index).
		Id(id).
		Do(ctx)

	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, fmt.Errorf("resource not found: %s", id)
		}
		return nil, fmt.Errorf("elasticsearch get failed: %w", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(getResult.Source, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	// Ensure @id is set
	if _, ok := doc["@id"]; !ok {
		doc["@id"] = id
	}

	return doc, nil
}

// Aggregate executes aggregation queries for facets.
func (s *SemanticStore) Aggregate(
	ctx context.Context,
	facets []semantic.FacetRequest,
	filters []semantic.Filter,
	config *semantic.ResourceConfig,
) ([]semantic.FacetResult, error) {
	// Build base query with filters
	opts := &semantic.QueryOptions{
		Filters: filters,
	}
	query := s.buildQuery(opts, config)

	// Build search with only aggregations
	searchService := s.client.Search(s.index).
		Query(query).
		Size(0) // We only want aggregations

	for _, facet := range facets {
		agg := s.buildFacetAggregation(facet, config)
		if agg != nil {
			searchService = searchService.Aggregation(facet.Field, agg)
		}
	}

	// Execute search
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch aggregation failed: %w", err)
	}

	// Extract facet results
	return s.extractFacets(searchResult, facets, config), nil
}

// SaveSearchContext saves a search context for detail-level navigation.
func (s *SemanticStore) SaveSearchContext(ctx context.Context, searchCtx *semantic.SearchContext) error {
	// For now, we'll skip storing search contexts in ES
	// In production, you'd want to store this in Redis or a dedicated cache
	s.log.Debug().Str("token", searchCtx.Token).Msg("search context save not implemented")
	return nil
}

// GetSearchContext retrieves a saved search context.
func (s *SemanticStore) GetSearchContext(ctx context.Context, token string) (*semantic.SearchContext, error) {
	return nil, semantic.ErrContextNotFound
}

// DeleteSearchContext removes an expired search context.
func (s *SemanticStore) DeleteSearchContext(ctx context.Context, token string) error {
	return nil
}

// Health checks if the store is healthy and accessible.
func (s *SemanticStore) Health(ctx context.Context) error {
	_, err := s.client.ClusterHealth().Do(ctx)
	if err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}
	return nil
}

// buildQuery builds an Elasticsearch query from QueryOptions.
func (s *SemanticStore) buildQuery(opts *semantic.QueryOptions, config *semantic.ResourceConfig) elastic.Query {
	boolQuery := elastic.NewBoolQuery()

	// Add text query
	if opts.Query != nil && opts.Query.Value != "" {
		if len(opts.Query.Fields) > 0 {
			// Multi-field query
			multiMatch := elastic.NewMultiMatchQuery(opts.Query.Value, opts.Query.Fields...)
			if opts.Query.Fuzzy {
				multiMatch = multiMatch.Fuzziness("AUTO")
			}
			if opts.Query.Operator != "" {
				multiMatch = multiMatch.Operator(opts.Query.Operator)
			}
			boolQuery = boolQuery.Must(multiMatch)
		} else {
			// Simple query string across all fields
			queryString := elastic.NewQueryStringQuery(opts.Query.Value)
			if opts.Query.Fuzzy {
				queryString = queryString.FuzzyPrefixLength(2)
			}
			boolQuery = boolQuery.Must(queryString)
		}
	}

	// Add filters
	for _, filter := range opts.Filters {
		filterQuery := s.buildFilterQuery(filter, config)
		if filterQuery != nil {
			boolQuery = boolQuery.Filter(filterQuery)
		}
	}

	// If bool query is empty (no must, should, or filter), match all
	// Note: We can't check if bool query is empty without accessing internals,
	// so we'll return the bool query regardless
	return boolQuery
}

// buildFilterQuery builds an Elasticsearch query for a filter.
func (s *SemanticStore) buildFilterQuery(filter semantic.Filter, config *semantic.ResourceConfig) elastic.Query {
	switch f := filter.(type) {
	case *semantic.PropertyFilter:
		return s.buildPropertyFilterQuery(f)

	case *semantic.RangeFilter:
		return s.buildRangeFilterQuery(f)

	case *semantic.ExistsFilter:
		return elastic.NewExistsQuery(f.FieldName)

	case *semantic.GeoBBoxFilter:
		return s.buildGeoBBoxQuery(f)

	case *semantic.GeoDistanceFilter:
		return s.buildGeoDistanceQuery(f)

	case *semantic.GeoPolygonFilter:
		return s.buildGeoPolygonQuery(f)

	default:
		s.log.Warn().Str("type", filter.Type()).Msg("unsupported filter type")
		return nil
	}
}

// buildPropertyFilterQuery builds a query for PropertyFilter.
func (s *SemanticStore) buildPropertyFilterQuery(filter *semantic.PropertyFilter) elastic.Query {
	field := filter.FieldName

	switch filter.OperatorType {
	case semantic.OpEqual:
		return elastic.NewTermQuery(field+".keyword", filter.Value)

	case semantic.OpNotEqual:
		return elastic.NewBoolQuery().MustNot(
			elastic.NewTermQuery(field+".keyword", filter.Value),
		)

	case semantic.OpIn:
		// Expect value to be a slice
		return elastic.NewTermsQuery(field+".keyword", convertToInterfaceSlice(filter.Value)...)

	case semantic.OpNotIn:
		return elastic.NewBoolQuery().MustNot(
			elastic.NewTermsQuery(field+".keyword", convertToInterfaceSlice(filter.Value)...),
		)

	case semantic.OpContains:
		return elastic.NewMatchQuery(field, filter.Value)

	case semantic.OpStartsWith:
		if str, ok := filter.Value.(string); ok {
			return elastic.NewPrefixQuery(field+".keyword", str)
		}

	case semantic.OpGreaterThan:
		return elastic.NewRangeQuery(field).Gt(filter.Value)

	case semantic.OpGreaterEqual:
		return elastic.NewRangeQuery(field).Gte(filter.Value)

	case semantic.OpLessThan:
		return elastic.NewRangeQuery(field).Lt(filter.Value)

	case semantic.OpLessEqual:
		return elastic.NewRangeQuery(field).Lte(filter.Value)
	}

	return nil
}

// buildRangeFilterQuery builds a query for RangeFilter.
func (s *SemanticStore) buildRangeFilterQuery(filter *semantic.RangeFilter) elastic.Query {
	rangeQuery := elastic.NewRangeQuery(filter.FieldName)

	// RangeFilter uses Min/Max, and operator determines inclusivity
	switch filter.OperatorType {
	case semantic.OpGreaterThan:
		if filter.Min != nil {
			rangeQuery = rangeQuery.Gt(filter.Min)
		}
	case semantic.OpGreaterEqual:
		if filter.Min != nil {
			rangeQuery = rangeQuery.Gte(filter.Min)
		}
	case semantic.OpLessThan:
		if filter.Max != nil {
			rangeQuery = rangeQuery.Lt(filter.Max)
		}
	case semantic.OpLessEqual:
		if filter.Max != nil {
			rangeQuery = rangeQuery.Lte(filter.Max)
		}
	default:
		// "between" - use both min and max
		if filter.Min != nil {
			rangeQuery = rangeQuery.Gte(filter.Min)
		}
		if filter.Max != nil {
			rangeQuery = rangeQuery.Lte(filter.Max)
		}
	}

	return rangeQuery
}

// buildGeoBBoxQuery builds a query for GeoBBoxFilter.
func (s *SemanticStore) buildGeoBBoxQuery(filter *semantic.GeoBBoxFilter) elastic.Query {
	return elastic.NewGeoBoundingBoxQuery(filter.FieldName).
		TopLeft(filter.Bounds.North, filter.Bounds.West).
		BottomRight(filter.Bounds.South, filter.Bounds.East)
}

// buildGeoDistanceQuery builds a query for GeoDistanceFilter.
func (s *SemanticStore) buildGeoDistanceQuery(filter *semantic.GeoDistanceFilter) elastic.Query {
	return elastic.NewGeoDistanceQuery(filter.FieldName).
		Lat(filter.Point.Lat).
		Lon(filter.Point.Lon).
		Distance(filter.Distance)
}

// buildGeoPolygonQuery builds a query for GeoPolygonFilter.
func (s *SemanticStore) buildGeoPolygonQuery(filter *semantic.GeoPolygonFilter) elastic.Query {
	// GeoPolygon uses GeoJSON format with coordinates
	if filter.Polygon == nil || len(filter.Polygon.Coordinates) == 0 {
		return nil
	}

	// Extract the first ring (outer boundary)
	ring := filter.Polygon.Coordinates[0]

	query := elastic.NewGeoPolygonQuery(filter.FieldName)

	for _, coord := range ring {
		if len(coord) >= 2 {
			// GeoJSON uses [lon, lat] order
			// AddPoint takes (lat, lon)
			query = query.AddPoint(coord[1], coord[0])
		}
	}

	return query
}

// convertToInterfaceSlice converts a value to []interface{} for terms query.
func convertToInterfaceSlice(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	case []string:
		result := make([]any, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result
	case []int:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	default:
		return []any{value}
	}
}
