package semantic

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// Last page
	if totalPages > 0 {
		view.Last = buildURLWithPage(r, int(totalPages))
	}

	return view
}

// buildFacets converts store facet results to API facets.
func (s *Service) buildFacets(
	r *http.Request,
	results []semantic.FacetResult,
	opts *semantic.QueryOptions,
) []semantic.Facet {
	facets := make([]semantic.Facet, len(results))

	for i, result := range results {
		facet := semantic.Facet{
			Field:       result.Field,
			FacetType:   result.FacetType,
			TotalValues: result.TotalValues,
			Values:      make([]semantic.FacetValue, len(result.Values)),
		}

		for j, val := range result.Values {
			facet.Values[j] = semantic.FacetValue{
				Value:     val.Value,
				Label:     val.Label,
				Count:     val.Count,
				Selected:  val.Selected,
				ApplyURL:  s.buildFacetApplyURL(r, result.Field, val.Value, opts),
				RemoveURL: s.buildFacetRemoveURL(r, result.Field, val.Value, opts),
			}
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
	breadcrumbs := &semantic.BreadcrumbList{
		ItemListElement: make([]semantic.BreadcrumbListItem, len(opts.Filters)),
	}

	for i, filter := range opts.Filters {
		breadcrumbs.ItemListElement[i] = semantic.BreadcrumbListItem{
			Position:  i + 1,
			Name:      s.buildFilterLabel(filter),
			RemoveURL: s.buildFilterRemoveURL(r, filter, opts),
		}
	}

	return breadcrumbs
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
	query.Set(fmt.Sprintf("filter[%s][eq]", field), value)
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
	query.Del(fmt.Sprintf("filter[%s][eq]", field))
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
	key := fmt.Sprintf("filter[%s][%s]", filter.Field(), filter.Operator())
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
