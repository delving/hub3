package semantic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// namespacePattern matches namespace prefixes like dc_, edm_, schema_, etc.
var namespacePattern = regexp.MustCompile(`^([a-z]+)_`)

// fromURLFieldName converts URL field format to internal format.
// Example: dc_creator -> dc:creator, edm_dataProvider -> edm:dataProvider
func fromURLFieldName(urlField string) string {
	// Convert underscore to colon for namespace prefixes
	// dc_creator -> dc:creator
	// edm_dataProvider -> edm:dataProvider
	// schema_addressCountry -> schema:addressCountry
	// But keep underscores that are part of the field name
	if namespacePattern.MatchString(urlField) {
		// Replace only the first underscore (between namespace and field name)
		parts := strings.SplitN(urlField, "_", 2)
		if len(parts) == 2 {
			return parts[0] + ":" + parts[1]
		}
	}
	return urlField
}

// toURLFieldName converts internal field format to URL format.
// Example: dc:creator -> dc_creator, edm:dataProvider -> edm_dataProvider
func toURLFieldName(internalField string) string {
	// Convert colon to underscore for URL safety
	return strings.ReplaceAll(internalField, ":", "_")
}

// parseQueryParams parses URL query parameters into QueryOptions.
func parseQueryParams(r *http.Request) (*semantic.QueryOptions, error) {
	query := r.URL.Query()

	opts := &semantic.QueryOptions{
		Filters: []semantic.Filter{},
		Facets:  []semantic.FacetRequest{},
		Sort:    []semantic.SortField{},
	}

	// Parse text query
	if q := query.Get("query"); q != "" {
		opts.Query = &semantic.TextQuery{
			Value: q,
		}
	}

	// Parse filters: filter[field][operator]=value
	if err := parseFiltersFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse filters: %w", err)
	}

	// Parse facets: facet[field]=true or facet[field]=limit
	if err := parseFacetsFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse facets: %w", err)
	}

	// Parse pagination
	if err := parsePaginationFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse pagination: %w", err)
	}

	// Parse sort
	if err := parseSortFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse sort: %w", err)
	}

	// Parse other options
	if languages := query.Get("languages"); languages != "" {
		opts.Languages = strings.Split(languages, ",")
	}

	if expand := query.Get("expand"); expand != "" {
		opts.Expand = strings.Split(expand, ",")
	}

	if fields := query.Get("fields"); fields != "" {
		opts.Fields = strings.Split(fields, ",")
	}

	opts.DetailLevel = query.Get("detailLevel")

	return opts, nil
}

// parseFiltersFromQuery parses filter parameters from URL query.
// Format: filter[field][operator]=value
func parseFiltersFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	for key, values := range query {
		if !strings.HasPrefix(key, "filter[") {
			continue
		}

		// Extract field and operator from filter[field][operator]
		parts := strings.Split(key, "][")
		if len(parts) != 2 {
			continue
		}

		urlField := strings.TrimPrefix(parts[0], "filter[")
		operator := strings.TrimSuffix(parts[1], "]")

		// Convert from URL format (dc_creator) to internal format (dc:creator)
		field := fromURLFieldName(urlField)

		op := semantic.Operator(operator)
		if !op.IsValid() {
			return fmt.Errorf("invalid operator '%s' for field '%s'", operator, field)
		}

		// Handle multiple values for operators like 'in'
		var value any
		if len(values) == 1 {
			value = values[0]
		} else {
			value = values
		}

		// Create filter based on operator type
		if op.IsGeospatial() {
			// Parse geospatial filter
			geoFilter, err := parseGeoFilter(field, op, values)
			if err != nil {
				return fmt.Errorf("failed to parse geo filter: %w", err)
			}
			opts.Filters = append(opts.Filters, geoFilter)
		} else {
			// Regular property filter
			opts.Filters = append(opts.Filters, &semantic.PropertyFilter{
				FieldName:    field,
				OperatorType: op,
				Value:        value,
			})
		}
	}

	return nil
}

// parseGeoFilter parses geospatial filter from query values.
func parseGeoFilter(field string, op semantic.Operator, values []string) (semantic.Filter, error) {
	switch op {
	case semantic.OpBBox:
		// Format: filter[spatialCoverage][bbox]=west,south,east,north
		if len(values) != 1 {
			return nil, fmt.Errorf("bbox filter requires exactly one value")
		}

		coords := strings.Split(values[0], ",")
		if len(coords) != 4 {
			return nil, fmt.Errorf("bbox requires 4 coordinates (west,south,east,north)")
		}

		west, err := strconv.ParseFloat(coords[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid west coordinate: %w", err)
		}

		south, err := strconv.ParseFloat(coords[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid south coordinate: %w", err)
		}

		east, err := strconv.ParseFloat(coords[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid east coordinate: %w", err)
		}

		north, err := strconv.ParseFloat(coords[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid north coordinate: %w", err)
		}

		return &semantic.GeoBBoxFilter{
			FieldName: field,
			Bounds: &semantic.GeoBounds{
				West:  west,
				South: south,
				East:  east,
				North: north,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported geospatial operator: %s", op)
	}
}

// parseFacetsFromQuery parses facet parameters from URL query.
// Format: facet[field]=true or facet[field]=limit
func parseFacetsFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	for key, values := range query {
		if !strings.HasPrefix(key, "facet[") {
			continue
		}

		// Extract field from facet[field]
		urlField := strings.TrimSuffix(strings.TrimPrefix(key, "facet["), "]")

		// Convert from URL format (dc_creator) to internal format (dc:creator)
		field := fromURLFieldName(urlField)

		if len(values) == 0 {
			continue
		}

		facetReq := semantic.FacetRequest{
			Field: field,
		}

		// Try to parse as limit
		if limit, err := strconv.Atoi(values[0]); err == nil && limit > 0 {
			facetReq.Limit = limit
		}

		opts.Facets = append(opts.Facets, facetReq)
	}

	return nil
}

// parsePaginationFromQuery parses pagination parameters from URL query.
func parsePaginationFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	pagination := &semantic.Pagination{}

	if page := query.Get("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			return fmt.Errorf("invalid page number: %w", err)
		}
		pagination.Page = p
	} else {
		pagination.Page = 1 // Default to page 1
	}

	if size := query.Get("size"); size != "" {
		s, err := strconv.Atoi(size)
		if err != nil {
			return fmt.Errorf("invalid page size: %w", err)
		}
		pagination.Size = s
	} else {
		pagination.Size = 20 // Default size
	}

	if cursor := query.Get("cursor"); cursor != "" {
		pagination.Cursor = cursor
	}

	opts.Pagination = pagination
	return nil
}

// parseSortFromQuery parses sort parameters from URL query.
// Format: sort=field or sort=-field (default asc, - for desc)
// Note: + gets decoded as space in URL params, so we treat leading space as +
func parseSortFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	sortParams := query["sort"]

	for _, sortParam := range sortParams {
		sortParam = strings.TrimSpace(sortParam)
		if sortParam == "" {
			continue
		}

		direction := semantic.SortAsc
		field := sortParam

		// Check if it starts with - for descending
		if strings.HasPrefix(sortParam, "-") {
			field = strings.TrimPrefix(sortParam, "-")
			direction = semantic.SortDesc
		}

		opts.Sort = append(opts.Sort, semantic.SortField{
			Field:     field,
			Direction: direction,
		})
	}

	return nil
}

// parseJSONBody parses JSON request body into the given value.
func parseJSONBody(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}
