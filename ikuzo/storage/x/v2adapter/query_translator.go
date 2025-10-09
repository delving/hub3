package v2adapter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// QueryTranslator converts semantic API queries to v2 search API format.
type QueryTranslator struct {
	orgID string
}

// NewQueryTranslator creates a new query translator.
func NewQueryTranslator(orgID string) *QueryTranslator {
	return &QueryTranslator{
		orgID: orgID,
	}
}

// TranslateToV2Query converts semantic QueryOptions to v2 URL query parameters.
// CRITICAL: Always sets itemFormat=semantic to get proper JSON-LD output.
func (qt *QueryTranslator) TranslateToV2Query(opts *semantic.QueryOptions) (url.Values, error) {
	params := url.Values{}

	// CRITICAL: Always use semantic format for proper JSON-LD with EDM context
	params.Set("itemFormat", "semantic")

	// Add text query
	if opts.Query != nil && opts.Query.Value != "" {
		queryStr, err := qt.translateTextQuery(opts.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to translate text query: %w", err)
		}
		params.Set("query", queryStr)

		// Add search fields if specified
		if len(opts.Query.Fields) > 0 {
			params.Set("searchFields", strings.Join(opts.Query.Fields, ","))
		}
	}

	// Add filters
	if len(opts.Filters) > 0 {
		if err := qt.translateFilters(opts.Filters, params); err != nil {
			return nil, fmt.Errorf("failed to translate filters: %w", err)
		}
	}

	// Add facets
	if len(opts.Facets) > 0 {
		qt.translateFacets(opts.Facets, params)
	}

	// Add pagination
	if opts.Pagination != nil {
		qt.translatePagination(opts.Pagination, params)
	}

	// Add sorting
	if len(opts.Sort) > 0 {
		qt.translateSort(opts.Sort, params)
	}

	return params, nil
}

// translateTextQuery converts semantic TextQuery to v2 query string.
func (qt *QueryTranslator) translateTextQuery(q *semantic.TextQuery) (string, error) {
	query := q.Value

	// Handle fuzzy search
	if q.Fuzzy {
		// Add fuzzy operator to each term
		terms := strings.Fields(query)
		for i, term := range terms {
			if !strings.HasSuffix(term, "~") {
				terms[i] = term + "~"
			}
		}
		query = strings.Join(terms, " ")
	}

	// Handle operator (AND/OR)
	// v2 uses default OR, so we only need to handle AND explicitly
	if q.Operator == "AND" {
		terms := strings.Fields(query)
		query = strings.Join(terms, " AND ")
	}

	return query, nil
}

// translateFilters converts semantic filters to v2 qf (query filter) parameters.
func (qt *QueryTranslator) translateFilters(filters []semantic.Filter, params url.Values) error {
	for _, filter := range filters {
		filterStr, err := qt.translateFilter(filter)
		if err != nil {
			return err
		}

		if filterStr != "" {
			params.Add("qf", filterStr)
		}
	}

	return nil
}

// translateFilter converts a single filter to v2 qf syntax.
func (qt *QueryTranslator) translateFilter(filter semantic.Filter) (string, error) {
	switch f := filter.(type) {
	case *semantic.PropertyFilter:
		return qt.translatePropertyFilter(f)

	case *semantic.RangeFilter:
		return qt.translateRangeFilter(f)

	case *semantic.ExistsFilter:
		return qt.translateExistsFilter(f)

	case *semantic.GeoBBoxFilter:
		return "", qt.translateGeoBBoxFilter(f)

	case *semantic.GeoDistanceFilter:
		return "", qt.translateGeoDistanceFilter(f)

	case *semantic.GeoPolygonFilter:
		return "", qt.translateGeoPolygonFilter(f)

	default:
		return "", fmt.Errorf("unsupported filter type: %s", filter.Type())
	}
}

// translatePropertyFilter converts PropertyFilter to v2 syntax.
func (qt *QueryTranslator) translatePropertyFilter(f *semantic.PropertyFilter) (string, error) {
	field := f.FieldName

	switch f.OperatorType {
	case semantic.OpEqual:
		// field:value
		return fmt.Sprintf("%s:%s", field, qt.escapeValue(f.Value)), nil

	case semantic.OpNotEqual:
		// NOT field:value
		return fmt.Sprintf("NOT %s:%s", field, qt.escapeValue(f.Value)), nil

	case semantic.OpIn:
		// field:(val1 OR val2 OR val3)
		values, ok := f.Value.([]string)
		if !ok {
			// Try to convert single value or other types
			values = []string{fmt.Sprintf("%v", f.Value)}
		}

		escapedValues := make([]string, len(values))
		for i, v := range values {
			escapedValues[i] = qt.escapeValue(v)
		}
		return fmt.Sprintf("%s:(%s)", field, strings.Join(escapedValues, " OR ")), nil

	case semantic.OpNotIn:
		// NOT field:(val1 OR val2 OR val3)
		values, ok := f.Value.([]string)
		if !ok {
			values = []string{fmt.Sprintf("%v", f.Value)}
		}

		escapedValues := make([]string, len(values))
		for i, v := range values {
			escapedValues[i] = qt.escapeValue(v)
		}
		return fmt.Sprintf("NOT %s:(%s)", field, strings.Join(escapedValues, " OR ")), nil

	case semantic.OpContains:
		// field:*value*
		return fmt.Sprintf("%s:*%s*", field, qt.escapeValue(f.Value)), nil

	case semantic.OpStartsWith:
		// field:value*
		return fmt.Sprintf("%s:%s*", field, qt.escapeValue(f.Value)), nil

	case semantic.OpGreaterThan:
		// field:{min TO *}
		return fmt.Sprintf("%s:{%v TO *}", field, f.Value), nil

	case semantic.OpGreaterEqual:
		// field:[min TO *]
		return fmt.Sprintf("%s:[%v TO *]", field, f.Value), nil

	case semantic.OpLessThan:
		// field:{* TO max}
		return fmt.Sprintf("%s:{* TO %v}", field, f.Value), nil

	case semantic.OpLessEqual:
		// field:[* TO max]
		return fmt.Sprintf("%s:[* TO %v]", field, f.Value), nil

	default:
		return "", fmt.Errorf("unsupported operator: %s", f.OperatorType)
	}
}

// translateRangeFilter converts RangeFilter to v2 range syntax.
func (qt *QueryTranslator) translateRangeFilter(f *semantic.RangeFilter) (string, error) {
	field := f.FieldName

	// Determine inclusivity based on operator
	leftBracket := "["
	rightBracket := "]"

	switch f.OperatorType {
	case semantic.OpGreaterThan:
		leftBracket = "{"
	case semantic.OpLessThan:
		rightBracket = "}"
	}

	minVal := "*"
	if f.Min != nil {
		minVal = fmt.Sprintf("%v", f.Min)
	}

	maxVal := "*"
	if f.Max != nil {
		maxVal = fmt.Sprintf("%v", f.Max)
	}

	return fmt.Sprintf("%s:%s%s TO %s%s", field, leftBracket, minVal, maxVal, rightBracket), nil
}

// translateExistsFilter converts ExistsFilter to v2 exists syntax.
func (qt *QueryTranslator) translateExistsFilter(f *semantic.ExistsFilter) (string, error) {
	// v2 uses _exists_:field or field:[* TO *]
	return fmt.Sprintf("_exists_:%s", f.FieldName), nil
}

// translateGeoBBoxFilter converts GeoBBoxFilter to v2 geo parameters.
func (qt *QueryTranslator) translateGeoBBoxFilter(f *semantic.GeoBBoxFilter) error {
	// v2 uses separate query parameters for geo bounding box
	// This will be handled separately in the params, not in qf
	// We'll need to add these to the params in the main translation function
	return fmt.Errorf("geo filters require separate parameter handling")
}

// translateGeoDistanceFilter converts GeoDistanceFilter to v2 geo parameters.
func (qt *QueryTranslator) translateGeoDistanceFilter(f *semantic.GeoDistanceFilter) error {
	return fmt.Errorf("geo filters require separate parameter handling")
}

// translateGeoPolygonFilter converts GeoPolygonFilter to v2 geo parameters.
func (qt *QueryTranslator) translateGeoPolygonFilter(f *semantic.GeoPolygonFilter) error {
	return fmt.Errorf("geo filters require separate parameter handling")
}

// translateFacets converts semantic facet requests to v2 facet.field parameters.
func (qt *QueryTranslator) translateFacets(facets []semantic.FacetRequest, params url.Values) {
	for _, facet := range facets {
		// Add facet field
		params.Add("facet.field", facet.Field)

		// Add facet limit if specified
		if facet.Limit > 0 {
			params.Set("facet.limit", fmt.Sprintf("%d", facet.Limit))
		}

		// Add facet sort if specified
		if facet.Sort != "" {
			// v2 uses "count" or "index" (alphabetical)
			params.Set("facet.sort", facet.Sort)
		}

		// Add facet filter if specified
		if facet.Filter != "" {
			params.Add("facet.filter", facet.Filter)
		}
	}
}

// translatePagination converts semantic pagination to v2 start/rows parameters.
func (qt *QueryTranslator) translatePagination(p *semantic.Pagination, params url.Values) {
	// Calculate start position
	start := (p.Page - 1) * p.Size
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("rows", fmt.Sprintf("%d", p.Size))

	// Store page for reference
	params.Set("page", fmt.Sprintf("%d", p.Page))
}

// translateSort converts semantic sort options to v2 sortBy parameter.
func (qt *QueryTranslator) translateSort(sorts []semantic.SortField, params url.Values) {
	if len(sorts) == 0 {
		return
	}

	// v2 typically supports single sort field
	// Use the first sort option
	sort := sorts[0]

	params.Set("sortBy", sort.Field)

	// Set sort direction
	if sort.Direction == semantic.SortDesc {
		params.Set("sortAsc", "false")
	} else {
		params.Set("sortAsc", "true")
	}
}

// escapeValue escapes or quotes query values as needed for v2 API.
// For values with spaces or special characters, we wrap in quotes.
// Otherwise return the value as-is.
func (qt *QueryTranslator) escapeValue(value any) string {
	str := fmt.Sprintf("%v", value)

	// If value contains spaces or special characters, wrap in quotes
	needsQuoting := strings.ContainsAny(str, " :()[]{}+-*/\\\"")

	if needsQuoting {
		// Escape internal quotes
		str = strings.ReplaceAll(str, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", str)
	}

	return str
}
