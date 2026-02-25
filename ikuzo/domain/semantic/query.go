package semantic

import (
	"encoding/json"
	"fmt"
)

// QueryOptions represents unified query options for both GET and POST requests.
// This is the internal representation that both URL parameters and JSON-LD queries
// are parsed into.
type QueryOptions struct {
	// Query is the text search query
	Query *TextQuery

	// Filters are the applied filters
	Filters []Filter

	// Facets are the requested facet aggregations
	Facets []FacetRequest

	// Sort specifies the sort order
	Sort []SortField

	// Pagination controls result pagination
	Pagination *Pagination

	// Languages specifies language preference for multilingual content
	Languages []string

	// Expand specifies which relations to expand inline
	Expand []string

	// DetailLevel specifies the level of detail in responses
	DetailLevel string

	// Fields specifies which fields to include (field selection)
	Fields []string

	// Collapse groups results by a field value (e.g., dedup by data provider).
	Collapse *CollapseOptions

	// FacetBoolType controls AND/OR logic for facet selections.
	FacetBoolType FacetBoolType

	// Peek when true returns only facets with zero items (size=0).
	Peek bool

	// Debug when set returns diagnostic information (e.g., "query" shows ES query).
	Debug string

	// ContextIndices specifies additional organization IDs to search across.
	ContextIndices []string
}

// FacetBoolType controls how multiple facet selections combine.
type FacetBoolType string

const (
	FacetBoolOr  FacetBoolType = "or"  // Selected values broaden results (default)
	FacetBoolAnd FacetBoolType = "and" // Selected values narrow results
)

// IsValid returns true if the facet bool type is valid.
func (fbt FacetBoolType) IsValid() bool {
	switch fbt {
	case FacetBoolOr, FacetBoolAnd, "":
		return true
	default:
		return false
	}
}

// CollapseOptions configures result collapsing (grouping/deduplication).
type CollapseOptions struct {
	// Field is the field to collapse on (required).
	Field string `json:"field"`
	// Size is the number of inner hits per group (default: 1).
	Size int `json:"size,omitempty"`
	// Sort specifies sort order for inner hits.
	Sort []SortField `json:"sort,omitempty"`
}

// Validate checks if the collapse options are valid.
func (co *CollapseOptions) Validate() error {
	if co == nil {
		return nil
	}
	if co.Field == "" {
		return fmt.Errorf("collapse field cannot be empty")
	}
	return nil
}

// TextQuery represents a full-text search query.
type TextQuery struct {
	TypeValue string   `json:"@type,omitempty"`
	Value     string   `json:"value"`
	Fields    []string `json:"fields,omitempty"`    // Fields to search in
	Operator  string   `json:"operator,omitempty"`  // AND, OR
	Fuzzy     bool     `json:"fuzzy,omitempty"`     // Enable fuzzy matching
	Boost     float64  `json:"boost,omitempty"`     // Boost factor
}

// Type returns the @type for JSON-LD.
func (tq *TextQuery) Type() string {
	if tq.TypeValue == "" {
		return "hub3:TextQuery"
	}
	return tq.TypeValue
}

// MarshalJSON implements custom JSON marshaling.
func (tq *TextQuery) MarshalJSON() ([]byte, error) {
	type Alias TextQuery
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  tq.Type(),
		Alias: (*Alias)(tq),
	})
}

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}

// SortDirection constants.
const (
	SortAsc  = "ASC"
	SortDesc = "DESC"
)

// IsAscending returns true if the sort direction is ascending.
func (sf *SortField) IsAscending() bool {
	return sf.Direction == "" || sf.Direction == SortAsc
}

// FacetRequest represents a request for facet aggregation.
type FacetRequest struct {
	TypeValue string `json:"@type,omitempty"`
	Field     string `json:"field"`
	Limit     int    `json:"limit,omitempty"`
	Filter    string `json:"filter,omitempty"` // Filter facet values (e.g., "rem*")
	Sort      string `json:"sort,omitempty"`   // count, value, index
	Cursor    string `json:"cursor,omitempty"` // Opaque cursor for paginating facet values
}

// Type returns the @type for JSON-LD.
func (fr *FacetRequest) Type() string {
	if fr.TypeValue == "" {
		return "hub3:FacetRequest"
	}
	return fr.TypeValue
}

// MarshalJSON implements custom JSON marshaling.
func (fr *FacetRequest) MarshalJSON() ([]byte, error) {
	type Alias FacetRequest
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  fr.Type(),
		Alias: (*Alias)(fr),
	})
}

// GeoClusterFacetRequest represents a request for geographic clustering.
type GeoClusterFacetRequest struct {
	FacetRequest
	Precision int        `json:"precision"` // Geohash precision (1-12)
	Bounds    *GeoBounds `json:"bounds,omitempty"`
}

// Type returns the @type for JSON-LD.
func (gcfr *GeoClusterFacetRequest) Type() string {
	return "hub3:GeoClusterFacet"
}

// SearchQuery represents a complete search query in JSON-LD format.
// This is the structure for POST /search requests.
type SearchQuery struct {
	Context        interface{}              `json:"@context,omitempty"`
	TypeValue      string                   `json:"@type"`
	Query          *TextQuery               `json:"query,omitempty"`
	Filters        []json.RawMessage        `json:"filters,omitempty"`
	Facets         []json.RawMessage        `json:"facets,omitempty"`
	Sort           []SortField              `json:"sort,omitempty"`
	Pagination     *Pagination              `json:"pagination,omitempty"`
	ResponseOptions *ResponseOptions        `json:"responseOptions,omitempty"`
}

// Type returns the @type for JSON-LD.
func (sq *SearchQuery) Type() string {
	if sq.TypeValue == "" {
		return "hub3:SearchQuery"
	}
	return sq.TypeValue
}

// ResponseOptions specifies options for response formatting.
type ResponseOptions struct {
	DetailLevel string   `json:"detailLevel,omitempty"` // minimal, standard, full
	Expand      []string `json:"expand,omitempty"`      // Relations to expand
	Fields      []string `json:"fields,omitempty"`      // Field selection
	Languages   []string `json:"languages,omitempty"`   // Language preference
}

// ToQueryOptions converts a SearchQuery to QueryOptions.
// The filters and facets need to be parsed from RawMessage separately.
func (sq *SearchQuery) ToQueryOptions() *QueryOptions {
	opts := &QueryOptions{
		Query:      sq.Query,
		Sort:       sq.Sort,
		Pagination: sq.Pagination,
	}

	if sq.ResponseOptions != nil {
		opts.DetailLevel = sq.ResponseOptions.DetailLevel
		opts.Expand = sq.ResponseOptions.Expand
		opts.Fields = sq.ResponseOptions.Fields
		opts.Languages = sq.ResponseOptions.Languages
	}

	return opts
}

// ParseFilters parses the raw filter messages into Filter instances.
// This needs to inspect the @type field to determine which filter type to use.
func ParseFilters(rawFilters []json.RawMessage) ([]Filter, error) {
	filters := make([]Filter, 0, len(rawFilters))

	for _, raw := range rawFilters {
		// First, peek at the @type to determine which filter type
		var typeCheck struct {
			Type string `json:"@type"`
		}
		if err := json.Unmarshal(raw, &typeCheck); err != nil {
			return nil, err
		}

		var filter Filter
		switch typeCheck.Type {
		case "hub3:PropertyFilter", "PropertyFilter":
			var pf PropertyFilter
			if err := json.Unmarshal(raw, &pf); err != nil {
				return nil, err
			}
			filter = &pf

		case "hub3:RangeFilter", "RangeFilter":
			var rf RangeFilter
			if err := json.Unmarshal(raw, &rf); err != nil {
				return nil, err
			}
			filter = &rf

		case "hub3:ExistsFilter", "ExistsFilter":
			var ef ExistsFilter
			if err := json.Unmarshal(raw, &ef); err != nil {
				return nil, err
			}
			filter = &ef

		case "hub3:GeoBBoxFilter", "GeoBBoxFilter":
			var gbf GeoBBoxFilter
			if err := json.Unmarshal(raw, &gbf); err != nil {
				return nil, err
			}
			filter = &gbf

		case "hub3:GeoDistanceFilter", "GeoDistanceFilter":
			var gdf GeoDistanceFilter
			if err := json.Unmarshal(raw, &gdf); err != nil {
				return nil, err
			}
			filter = &gdf

		case "hub3:GeoPolygonFilter", "GeoPolygonFilter":
			var gpf GeoPolygonFilter
			if err := json.Unmarshal(raw, &gpf); err != nil {
				return nil, err
			}
			filter = &gpf

		default:
			// Unknown filter type - try PropertyFilter as default
			var pf PropertyFilter
			if err := json.Unmarshal(raw, &pf); err != nil {
				return nil, err
			}
			filter = &pf
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

// ParseFacets parses the raw facet messages into FacetRequest instances.
func ParseFacets(rawFacets []json.RawMessage) ([]FacetRequest, error) {
	facets := make([]FacetRequest, 0, len(rawFacets))

	for _, raw := range rawFacets {
		// Check if it's a simple string (field name only)
		var simpleString string
		if err := json.Unmarshal(raw, &simpleString); err == nil {
			facets = append(facets, FacetRequest{
				Field: simpleString,
			})
			continue
		}

		// Otherwise, parse as FacetRequest object
		var fr FacetRequest
		if err := json.Unmarshal(raw, &fr); err != nil {
			return nil, err
		}
		facets = append(facets, fr)
	}

	return facets, nil
}
