package v2adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

// V2SearchAdapter implements semantic.SearchStore using v2 search infrastructure.
// It acts as a wrapper around the v2 API, translating semantic queries to v2 format
// and extracting results from the v2 semantic format output.
//
// The adapter extracts the orgID from the request context for each operation,
// making it suitable for multi-tenant environments.
type V2SearchAdapter struct {
	client     *elastic.Client
	index      string
	translator *QueryTranslator
	resultTx   *ResultTranslator
	log        zerolog.Logger

	// In-memory context cache with mutex for thread safety
	contextMu    sync.RWMutex
	contextCache map[string]*semantic.SearchContext
}

// NewV2SearchAdapter creates a new v2 search adapter.
// The orgID is extracted from the request context for each operation.
func NewV2SearchAdapter(
	client *elastic.Client,
	index string,
	logger zerolog.Logger,
) *V2SearchAdapter {
	return &V2SearchAdapter{
		client:       client,
		index:        index,
		translator:   NewQueryTranslator(""), // orgID will be set per-request
		resultTx:     NewResultTranslator(),
		log:          logger.With().Str("component", "v2_adapter").Logger(),
		contextCache: make(map[string]*semantic.SearchContext),
	}
}

// Search implements semantic.SearchStore.Search by translating the query to v2 format,
// executing it via v2 infrastructure, and translating results back to semantic format.
func (a *V2SearchAdapter) Search(
	ctx context.Context,
	opts *semantic.QueryOptions,
	config *semantic.ResourceConfig,
) (*semantic.SearchResult, error) {
	// Extract orgID from context (set by middleware in multi-tenant system)
	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok || org.ID == "" {
		return nil, fmt.Errorf("organization ID not found in context")
	}
	orgID := org.ID.String()

	a.log.Debug().
		Str("org_id", orgID).
		Interface("query", opts.Query).
		Int("filters", len(opts.Filters)).
		Int("facets", len(opts.Facets)).
		Msg("executing search via v2 adapter")

	// 1. Translate semantic query to v2 query params (set orgID for this request)
	a.translator.orgID = orgID
	queryParams, err := a.translator.TranslateToV2Query(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to translate query: %w", err)
	}

	a.log.Debug().
		Interface("v2_params", queryParams).
		Msg("translated to v2 query params")

	// 2. Create v2 SearchRequest
	searchRequest, err := fragments.NewSearchRequest(orgID, queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create v2 search request: %w", err)
	}

	// 3. Execute v2 search with parallel aggregations
	startTime := time.Now()
	esResult, err := searchRequest.ExecuteWithParallelAggregations(a.client, ctx)
	if err != nil {
		return nil, fmt.Errorf("v2 search execution failed: %w", err)
	}

	a.log.Debug().
		Int64("total_hits", esResult.TotalHits()).
		Int("returned", len(esResult.Hits.Hits)).
		Dur("took_ms", time.Since(startTime)).
		Msg("v2 search completed")

	// 4. Decode v2 search results to ScrollResultV4
	scrollResult, err := a.decodeV2Results(esResult, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode v2 results: %w", err)
	}

	// 5. Translate v2 results to semantic format
	result, err := a.resultTx.TranslateSearchResult(scrollResult, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to translate search results: %w", err)
	}

	// Update took time from actual execution
	result.Took = time.Since(startTime)

	a.log.Info().
		Int64("total", result.Total).
		Int("results", len(result.Results)).
		Int("facets", len(result.Facets)).
		Dur("took", result.Took).
		Msg("search completed via v2 adapter")

	return result, nil
}

// GetByID implements semantic.SearchStore.GetByID by querying for a specific document ID.
func (a *V2SearchAdapter) GetByID(
	ctx context.Context,
	id string,
	config *semantic.ResourceConfig,
) (map[string]any, error) {
	a.log.Debug().
		Str("id", id).
		Msg("fetching document by ID via v2 adapter")

	// Use Elasticsearch Get API directly for single document retrieval
	// This is more efficient than going through v2 search
	getResult, err := a.client.Get().
		Index(a.index).
		Id(id).
		Do(ctx)

	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, fmt.Errorf("resource not found: %s", id)
		}
		return nil, fmt.Errorf("elasticsearch get failed: %w", err)
	}

	if !getResult.Found {
		return nil, fmt.Errorf("resource not found: %s", id)
	}

	// Decode the source document
	var doc fragments.FragmentGraph
	if err := json.Unmarshal(getResult.Source, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	// Generate semantic view if not already present
	if doc.Semantic == nil {
		doc.Semantic = doc.GenerateSemantic()
	}

	// Marshal semantic view to map[string]any
	data, err := json.Marshal(doc.Semantic)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal semantic view: %w", err)
	}

	var item map[string]any
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal semantic item: %w", err)
	}

	a.log.Info().
		Str("id", id).
		Msg("document retrieved via v2 adapter")

	return item, nil
}

// Aggregate implements semantic.SearchStore.Aggregate by executing facet-only queries.
func (a *V2SearchAdapter) Aggregate(
	ctx context.Context,
	facets []semantic.FacetRequest,
	filters []semantic.Filter,
	config *semantic.ResourceConfig,
) ([]semantic.FacetResult, error) {
	// Extract orgID from context
	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok || org.ID == "" {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	a.log.Debug().
		Str("org_id", org.ID.String()).
		Int("facets", len(facets)).
		Int("filters", len(filters)).
		Msg("executing aggregation via v2 adapter")

	// Build query options with only facets and filters
	opts := &semantic.QueryOptions{
		Facets:  facets,
		Filters: filters,
		Pagination: &semantic.Pagination{
			Page: 1,
			Size: 0, // We only want facets, no results
		},
	}

	// Execute search with size=0 to get only facets
	result, err := a.Search(ctx, opts, config)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}

	return result.Facets, nil
}

// SaveSearchContext implements semantic.SearchStore.SaveSearchContext.
// Uses in-memory cache for simplicity. In production, consider Redis or similar.
func (a *V2SearchAdapter) SaveSearchContext(
	ctx context.Context,
	searchCtx *semantic.SearchContext,
) error {
	a.contextMu.Lock()
	defer a.contextMu.Unlock()

	a.contextCache[searchCtx.Token] = searchCtx

	a.log.Debug().
		Str("token", searchCtx.Token).
		Int("result_ids", len(searchCtx.ResultIDs)).
		Msg("search context saved")

	return nil
}

// GetSearchContext implements semantic.SearchStore.GetSearchContext.
func (a *V2SearchAdapter) GetSearchContext(
	ctx context.Context,
	token string,
) (*semantic.SearchContext, error) {
	a.contextMu.RLock()
	defer a.contextMu.RUnlock()

	searchCtx, ok := a.contextCache[token]
	if !ok {
		return nil, semantic.ErrContextNotFound
	}

	a.log.Debug().
		Str("token", token).
		Msg("search context retrieved")

	return searchCtx, nil
}

// DeleteSearchContext implements semantic.SearchStore.DeleteSearchContext.
func (a *V2SearchAdapter) DeleteSearchContext(
	ctx context.Context,
	token string,
) error {
	a.contextMu.Lock()
	defer a.contextMu.Unlock()

	delete(a.contextCache, token)

	a.log.Debug().
		Str("token", token).
		Msg("search context deleted")

	return nil
}

// Health implements semantic.SearchStore.Health by checking Elasticsearch cluster health.
func (a *V2SearchAdapter) Health(ctx context.Context) error {
	_, err := a.client.ClusterHealth().Do(ctx)
	if err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}
	return nil
}

// FindSimilar implements semantic.SimilarStore using v2 MLT parameters.
func (a *V2SearchAdapter) FindSimilar(
	ctx context.Context,
	id string,
	opts *semantic.SimilarOptions,
	config *semantic.ResourceConfig,
) (*semantic.SearchResult, error) {
	if opts == nil {
		opts = semantic.DefaultSimilarOptions()
	}

	// Extract orgID from context
	org, ok := domain.GetOrganizationFromCtx(ctx)
	if !ok || org.ID == "" {
		return nil, fmt.Errorf("organization ID not found in context")
	}
	orgID := org.ID.String()

	a.log.Debug().
		Str("id", id).
		Str("org_id", orgID).
		Int("count", opts.Count).
		Strs("fields", opts.Fields).
		Msg("executing FindSimilar via v2 adapter")

	params := a.translator.TranslateMLTQuery(id, opts)

	// Execute via v2 search infrastructure
	searchRequest, err := fragments.NewSearchRequest(orgID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create v2 search request for MLT: %w", err)
	}

	startTime := time.Now()

	esResult, err := searchRequest.ExecuteWithParallelAggregations(a.client, ctx)
	if err != nil {
		return nil, fmt.Errorf("v2 MLT search execution failed: %w", err)
	}

	a.log.Debug().
		Int64("total_hits", esResult.TotalHits()).
		Int("returned", len(esResult.Hits.Hits)).
		Dur("took_ms", time.Since(startTime)).
		Msg("v2 MLT search completed")

	scrollResult, err := a.decodeV2Results(esResult, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode v2 MLT results: %w", err)
	}

	result, err := a.resultTx.TranslateSearchResult(scrollResult, &semantic.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to translate MLT search results: %w", err)
	}

	result.Took = time.Since(startTime)

	a.log.Info().
		Str("id", id).
		Int64("total", result.Total).
		Int("results", len(result.Results)).
		Dur("took", result.Took).
		Msg("FindSimilar completed via v2 adapter")

	return result, nil
}

// decodeV2Results converts Elasticsearch search result to v2 ScrollResultV4 format.
// This mirrors the logic in hub3/server/http/handlers/search.go::ProcessSearchRequest.
func (a *V2SearchAdapter) decodeV2Results(
	esResult *elastic.SearchResult,
	searchRequest *fragments.SearchRequest,
) (*fragments.ScrollResultV4, error) {
	result := &fragments.ScrollResultV4{
		Query: &fragments.Query{
			Numfound: int32(esResult.TotalHits()),
			Terms:    searchRequest.GetQuery(),
		},
		Items: make([]*fragments.FragmentGraph, 0, len(esResult.Hits.Hits)),
	}

	// Decode items from search hits
	for _, hit := range esResult.Hits.Hits {
		var fg fragments.FragmentGraph
		if err := json.Unmarshal(hit.Source, &fg); err != nil {
			a.log.Warn().
				Err(err).
				Str("id", hit.Id).
				Msg("failed to unmarshal fragment graph")
			continue
		}

		// Generate semantic view if itemFormat=semantic was used
		// The v2 API should have already set this, but we ensure it's present
		if fg.Semantic == nil {
			fg.Semantic = fg.GenerateSemantic()
		}

		result.Items = append(result.Items, &fg)
	}

	// Decode facets
	if len(searchRequest.FacetField) > 0 {
		facets, err := searchRequest.DecodeFacets(esResult, nil)
		if err != nil {
			a.log.Warn().
				Err(err).
				Msg("failed to decode facets")
			// Don't fail the whole request for facet errors
		} else {
			result.Facets = facets
		}
	}

	return result, nil
}
