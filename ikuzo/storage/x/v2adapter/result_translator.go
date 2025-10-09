package v2adapter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// ResultTranslator converts v2 search results to semantic API format.
type ResultTranslator struct{}

// NewResultTranslator creates a new result translator.
func NewResultTranslator() *ResultTranslator {
	return &ResultTranslator{}
}

// TranslateSearchResult converts v2 ScrollResultV4 to semantic SearchResult.
// When itemFormat=semantic is used, FragmentGraph.Semantic contains the properly formatted output.
func (rt *ResultTranslator) TranslateSearchResult(
	scrollResult *fragments.ScrollResultV4,
	opts *semantic.QueryOptions,
) (*semantic.SearchResult, error) {
	if scrollResult == nil {
		return nil, fmt.Errorf("scroll result is nil")
	}

	result := &semantic.SearchResult{
		Total:     int64(scrollResult.Query.Numfound),
		Results:   make([]map[string]any, 0, len(scrollResult.Items)),
		ResultIDs: make([]string, 0, len(scrollResult.Items)),
		Facets:    make([]semantic.FacetResult, 0),
		Took:      0, // v2 doesn't provide timing info in ScrollResultV4
		Metadata:  make(map[string]any),
	}

	// Extract items from FragmentGraph records
	items, ids, err := rt.translateItems(scrollResult.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to translate items: %w", err)
	}

	result.Results = items
	result.ResultIDs = ids

	// Translate facets if present
	if len(scrollResult.Facets) > 0 && opts != nil && len(opts.Facets) > 0 {
		result.Facets = rt.translateFacets(scrollResult.Facets, opts.Facets)
	}

	// Store v2 metadata
	if scrollResult.Query != nil {
		result.Metadata["v2_query_terms"] = scrollResult.Query.Terms
	}

	return result, nil
}

// translateItems extracts semantic items from v2 FragmentGraph records.
// Each record has a Semantic field (when itemFormat=semantic is used) with:
// - @context: reference to EDM context URL or inline context map
// - @id: resource identifier
// - @type: resource types
// - Fields: all properties as SemanticValue instances
func (rt *ResultTranslator) translateItems(records []*fragments.FragmentGraph) ([]map[string]any, []string, error) {
	items := make([]map[string]any, 0, len(records))
	ids := make([]string, 0, len(records))

	for _, rec := range records {
		if rec == nil {
			continue
		}

		// Extract ID for navigation
		if rec.Meta != nil && rec.Meta.HubID != "" {
			ids = append(ids, rec.Meta.HubID)
		}

		// Extract semantic view
		if rec.Semantic != nil {
			// Marshal SemanticView to map[string]any
			data, err := json.Marshal(rec.Semantic)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal semantic view: %w", err)
			}

			var item map[string]any
			if err := json.Unmarshal(data, &item); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal semantic item: %w", err)
			}

			items = append(items, item)
		} else {
			// Fallback: if Semantic is not present, create basic item from Meta
			item := make(map[string]any)
			if rec.Meta != nil {
				item["@id"] = rec.Meta.HubID
				item["_warning"] = "itemFormat=semantic not used, falling back to basic metadata"
			}
			items = append(items, item)
		}
	}

	return items, ids, nil
}

// translateFacets converts v2 facets to semantic FacetResults.
func (rt *ResultTranslator) translateFacets(
	v2Facets []*fragments.QueryFacet,
	facetReqs []semantic.FacetRequest,
) []semantic.FacetResult {
	results := make([]semantic.FacetResult, 0, len(v2Facets))

	// Create a map of requested facets for quick lookup
	reqMap := make(map[string]semantic.FacetRequest)
	for _, req := range facetReqs {
		reqMap[req.Field] = req
	}

	for _, v2Facet := range v2Facets {
		if v2Facet == nil {
			continue
		}

		// Determine facet type from v2 facet
		facetType := rt.determineFacetType(v2Facet)

		// Translate facet values
		values := make([]semantic.FacetValueResult, 0, len(v2Facet.Links))
		for _, link := range v2Facet.Links {
			if link == nil {
				continue
			}

			values = append(values, semantic.FacetValueResult{
				Value:    link.Value,
				Label:    link.DisplayString,
				Count:    link.Count,
				Selected: link.IsSelected,
			})
		}

		result := semantic.FacetResult{
			Field:       v2Facet.Field,
			FacetType:   facetType,
			TotalValues: v2Facet.Total,
			Values:      values,
			SumOther:    v2Facet.OtherDocs,
			Missing:     v2Facet.MissingDocs,
		}

		results = append(results, result)
	}

	return results
}

// determineFacetType determines the semantic facet type from v2 QueryFacet.
func (rt *ResultTranslator) determineFacetType(v2Facet *fragments.QueryFacet) string {
	// Check v2 facet type field
	if v2Facet.Type != "" {
		switch v2Facet.Type {
		case "histogram", "date_histogram":
			return "date"
		case "range":
			return "range"
		}
	}

	// Check if it has date interval
	if v2Facet.DateInterval != "" {
		return "date"
	}

	// Check if it has min/max (range facet)
	if v2Facet.Min != "" || v2Facet.Max != "" {
		return "range"
	}

	// Check if it's a boolean facet (only true/false values)
	if len(v2Facet.Links) == 2 {
		hasTrue := false
		hasFalse := false
		for _, link := range v2Facet.Links {
			if link.Value == "true" {
				hasTrue = true
			}
			if link.Value == "false" {
				hasFalse = true
			}
		}
		if hasTrue && hasFalse {
			return "boolean"
		}
	}

	// Default to enum (terms facet)
	return "enum"
}

// extractResultIDs extracts document IDs for navigation context.
func (rt *ResultTranslator) extractResultIDs(records []*fragments.FragmentGraph) []string {
	ids := make([]string, 0, len(records))

	for _, rec := range records {
		if rec != nil && rec.Meta != nil && rec.Meta.HubID != "" {
			ids = append(ids, rec.Meta.HubID)
		}
	}

	return ids
}

// TranslateTimestamp converts v2 timestamp to time.Time if available.
func (rt *ResultTranslator) TranslateTimestamp(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(0, ts*int64(time.Millisecond))
}
