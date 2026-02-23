package semantic

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// buildCollection builds a Hydra Collection from search results.
func (s *Service) buildCollection(
	r *http.Request,
	result *semantic.SearchResult,
	opts *semantic.QueryOptions,
	config *semantic.ResourceConfig,
) *semantic.Collection {
	// Build base URL for this request
	baseURL := s.buildBaseURL(r)

	// Create collection
	collection := semantic.NewCollection(baseURL)
	collection.Context = semantic.DefaultContext()
	collection.TotalItems = result.Total

	// Convert []map[string]any to []any
	members := make([]any, len(result.Results))
	for i, r := range result.Results {
		members[i] = r
	}
	collection.Member = members

	// Build pagination view
	if opts.Pagination != nil {
		collection.View = s.buildPartialCollectionView(r, result, opts)
	}

	// Build search template
	collection.Search = semantic.BuildTemplateFromConfigs(
		baseURL,
		config.Filters,
		config.Facets,
	)

	// Build facets
	if len(result.Facets) > 0 {
		collection.Facets = s.buildFacets(r, result.Facets, opts)
	}

	// Build active filters breadcrumbs
	if len(opts.Filters) > 0 {
		collection.ActiveFilters = s.buildBreadcrumbs(r, opts)
	}

	// Add timing information
	collection.Timing = &semantic.Timing{
		Took: int64(result.Took / time.Millisecond),
		Unit: "ms",
	}

	// Add debug info if debug mode is enabled
	if opts.Debug != "" && result.Metadata != nil {
		collection.Debug = result.Metadata
	}

	return collection
}

// buildPartialCollectionView builds pagination links.
func (s *Service) buildPartialCollectionView(
	r *http.Request,
	result *semantic.SearchResult,
	opts *semantic.QueryOptions,
) *semantic.PartialCollectionView {
	page := opts.Pagination.Page
	size := opts.Pagination.Size

	totalPages := (result.Total + int64(size) - 1) / int64(size)

	view := &semantic.PartialCollectionView{
		TypeValue: "hydra:PartialCollectionView",
		ID:        buildURLWithPage(r, page),
	}

	// First page
	if totalPages > 0 {
		view.First = buildURLWithPage(r, 1)
	}

	// Previous page
	if page > 1 {
		view.Previous = buildURLWithPage(r, page-1)
	}

	// Next page
	if page < int(totalPages) {
		view.Next = buildURLWithPage(r, page+1)
	}

	// Note: Last link is intentionally omitted per v1 design contract.
	// Context-based pagination does not include a last link.

	return view
}

// buildFacets converts store facet results to API facets.
func (s *Service) buildFacets(
	r *http.Request,
	results []semantic.FacetResult,
	opts *semantic.QueryOptions,
) []semantic.Facet {
	facets := make([]semantic.Facet, len(results))

	// Get default config for facet metadata
	config := s.registry.GetDefault()

	for i, result := range results {
		facet := semantic.Facet{
			TypeValue:   "hub3:Facet",
			Field:       result.Field,
			FacetType:   result.FacetType,
			TotalValues: result.TotalValues,
			Values:      make([]semantic.FacetValue, len(result.Values)),
		}

		// Get facet config for label and property URI
		if config != nil {
			if facetConfig, ok := config.GetFacetConfig(result.Field); ok {
				facet.Label = facetConfig.Label
				facet.PropertyURI = facetConfig.PropertyURI
			}
		}

		for j, val := range result.Values {
			urlField := toURLFieldName(result.Field)
			facet.Values[j] = semantic.FacetValue{
				TypeValue: "hub3:FacetValue",
				Value:     val.Value,
				Label:     val.Label,
				Count:     val.Count,
				Selected:  val.Selected,
				Filter:    fmt.Sprintf("filter[%s][eq]=%s", urlField, url.QueryEscape(val.Value)),
				ApplyURL:  s.buildFacetApplyURL(r, result.Field, val.Value, opts),
				RemoveURL: s.buildFacetRemoveURL(r, result.Field, val.Value, opts),
			}
		}

		if result.NextCursor != "" {
			facet.NextCursor = result.NextCursor
		}

		facets[i] = facet
	}

	return facets
}

// buildBreadcrumbs builds active filter breadcrumbs.
func (s *Service) buildBreadcrumbs(
	r *http.Request,
	opts *semantic.QueryOptions,
) *semantic.BreadcrumbList {
	var items []semantic.BreadcrumbListItem

	for _, filter := range opts.Filters {
		// Skip hidden filters — they should not appear in breadcrumbs.
		if pf, ok := filter.(*semantic.PropertyFilter); ok && pf.Hidden {
			continue
		}

		items = append(items, semantic.BreadcrumbListItem{
			Position:  len(items) + 1,
			Name:      s.buildFilterLabel(filter),
			RemoveURL: s.buildFilterRemoveURL(r, filter, opts),
		})
	}

	return &semantic.BreadcrumbList{
		ItemListElement: items,
	}
}

// buildFilterLabel creates a human-readable label for a filter.
func (s *Service) buildFilterLabel(filter semantic.Filter) string {
	// Simple implementation for now
	return fmt.Sprintf("%s %s %v", filter.Field(), filter.Operator(), "...")
}

// buildFacetApplyURL builds a URL to apply a facet filter.
func (s *Service) buildFacetApplyURL(
	r *http.Request,
	field, value string,
	opts *semantic.QueryOptions,
) string {
	// Clone query params and add facet filter
	query := r.URL.Query()
	// Convert to URL format: dc:creator -> dc_creator
	urlField := toURLFieldName(field)
	query.Set(fmt.Sprintf("filter[%s][eq]", urlField), value)
	return s.buildBaseURL(r) + "?" + query.Encode()
}

// buildFacetRemoveURL builds a URL to remove a facet filter.
func (s *Service) buildFacetRemoveURL(
	r *http.Request,
	field, value string,
	opts *semantic.QueryOptions,
) string {
	// Clone query params and remove facet filter
	query := r.URL.Query()
	// Convert to URL format: dc:creator -> dc_creator
	urlField := toURLFieldName(field)
	query.Del(fmt.Sprintf("filter[%s][eq]", urlField))
	return s.buildBaseURL(r) + "?" + query.Encode()
}

// buildFilterRemoveURL builds a URL to remove a specific filter.
func (s *Service) buildFilterRemoveURL(
	r *http.Request,
	filter semantic.Filter,
	opts *semantic.QueryOptions,
) string {
	// Clone query params and remove this filter
	query := r.URL.Query()
	// Convert to URL format: dc:creator -> dc_creator
	urlField := toURLFieldName(filter.Field())
	key := fmt.Sprintf("filter[%s][%s]", urlField, filter.Operator())
	query.Del(key)
	return s.buildBaseURL(r) + "?" + query.Encode()
}

// buildBaseURL constructs the base URL for the current request.
func (s *Service) buildBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, r.URL.Path)
}

// buildURLWithPage builds a URL with a specific page number.
func buildURLWithPage(r *http.Request, page int) string {
	query := r.URL.Query()
	query.Set("page", fmt.Sprintf("%d", page))

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s%s?%s", scheme, host, r.URL.Path, query.Encode())
}

// respondJSON writes a JSON response.
func (s *Service) respondJSON(w http.ResponseWriter, r *http.Request, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error().Err(err).Msg("failed to encode JSON response")
	}
}

// respondError writes an error response in Hydra format.
func (s *Service) respondError(w http.ResponseWriter, r *http.Request, err *semantic.HydraError) {
	s.log.Warn().
		Str("title", err.Title).
		Str("description", err.Description).
		Int("status", err.StatusCode).
		Msg("request error")

	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(err.StatusCode)

	// Add context to error
	response := map[string]any{
		"@context":          semantic.DefaultContext(),
		"@type":             err.Type(),
		"hydra:title":       err.Title,
		"hydra:description": err.Description,
	}

	if err.Details != nil {
		response["hub3:details"] = err.Details
	}

	if err.Documentation != "" {
		response["hub3:documentation"] = err.Documentation
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		s.log.Error().Err(encodeErr).Msg("failed to encode error response")
	}
}

// createSearchContext creates a search context for detail-level navigation.
func (s *Service) createSearchContext(opts *semantic.QueryOptions, result *semantic.SearchResult) *semantic.SearchContext {
	return &semantic.SearchContext{
		ID:           fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
		Token:        generateToken(),
		Query:        opts,
		ResultIDs:    result.ResultIDs,
		TotalResults: result.Total,
		ExpiresAt:    time.Now().Add(15 * time.Minute).Format(time.RFC3339),
	}
}

// buildNavigationContext builds navigation links from a search context.
func (s *Service) buildNavigationContext(r *http.Request, currentID string, searchCtx *semantic.SearchContext) *semantic.NavigationContext {
	// Find current position in result list
	position := -1
	for i, id := range searchCtx.ResultIDs {
		if id == currentID {
			position = i
			break
		}
	}

	if position == -1 {
		// Current ID not in search results
		return nil
	}

	navCtx := &semantic.NavigationContext{
		Position:     position + 1, // 1-based for display
		TotalResults: searchCtx.TotalResults,
		HasNext:      position < len(searchCtx.ResultIDs)-1,
		HasPrevious:  position > 0,
	}

	// Build URLs with context token
	baseURL := fmt.Sprintf("%s://%s%s", getScheme(r), r.Host, "/api/semantic/v1/resource")

	// First
	if len(searchCtx.ResultIDs) > 0 {
		navCtx.First = fmt.Sprintf("%s/%s?context=%s", baseURL, searchCtx.ResultIDs[0], searchCtx.Token)
	}

	// Previous
	if navCtx.HasPrevious {
		navCtx.Previous = fmt.Sprintf("%s/%s?context=%s", baseURL, searchCtx.ResultIDs[position-1], searchCtx.Token)
	}

	// Next
	if navCtx.HasNext {
		navCtx.Next = fmt.Sprintf("%s/%s?context=%s", baseURL, searchCtx.ResultIDs[position+1], searchCtx.Token)
	}

	// Last
	if len(searchCtx.ResultIDs) > 0 {
		navCtx.Last = fmt.Sprintf("%s/%s?context=%s", baseURL, searchCtx.ResultIDs[len(searchCtx.ResultIDs)-1], searchCtx.Token)
	}

	// Back to search
	if searchCtx.Query != nil {
		navCtx.BackToSearch = s.buildSearchURL(r, searchCtx)
	}

	return navCtx
}

// buildSearchURL rebuilds the search URL from a search context.
func (s *Service) buildSearchURL(r *http.Request, searchCtx *semantic.SearchContext) string {
	baseURL := fmt.Sprintf("%s://%s%s", getScheme(r), r.Host, "/api/semantic/v1/search")

	// Build query string from search context query
	query := url.Values{}

	if searchCtx.Query.Query != nil {
		query.Set("query", searchCtx.Query.Query.Value)
	}

	if searchCtx.Query.Pagination != nil {
		query.Set("page", fmt.Sprintf("%d", searchCtx.Query.Pagination.Page))
		query.Set("size", fmt.Sprintf("%d", searchCtx.Query.Pagination.Size))
	}

	// Add filters
	for _, filter := range searchCtx.Query.Filters {
		if pf, ok := filter.(*semantic.PropertyFilter); ok {
			key := fmt.Sprintf("filter[%s][%s]", pf.FieldName, pf.OperatorType)
			query.Set(key, fmt.Sprintf("%v", pf.Value))
		}
	}

	if len(query) == 0 {
		return baseURL
	}

	return baseURL + "?" + query.Encode()
}

// getScheme returns the URL scheme (http or https).
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// generateToken generates a random token for search context.
func generateToken() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
}
