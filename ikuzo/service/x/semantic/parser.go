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

	// Parse collapse
	if err := parseCollapseFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse collapse: %w", err)
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

	// Parse facet bool type
	if fbt := query.Get("facetBool"); fbt != "" {
		opts.FacetBoolType = semantic.FacetBoolType(fbt)
		if !opts.FacetBoolType.IsValid() {
			return nil, fmt.Errorf("invalid facetBool value: %q (must be 'and' or 'or')", fbt)
		}
	}

	// Parse debug mode
	opts.Debug = query.Get("debug")

	// Parse peek mode
	if peekFields := query.Get("peek"); peekFields != "" {
		opts.Peek = true
		for _, urlField := range strings.Split(peekFields, ",") {
			field := fromURLFieldName(strings.TrimSpace(urlField))
			if field != "" {
				opts.Facets = append(opts.Facets, semantic.FacetRequest{Field: field})
			}
		}
		if opts.Pagination == nil {
			opts.Pagination = &semantic.Pagination{}
		}
		opts.Pagination.Size = 0
	}

	return opts, nil
}

// parseCollapseFromQuery parses collapse parameters from URL query.
func parseCollapseFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	collapseField := query.Get("collapse")
	if collapseField == "" {
		return nil
	}

	opts.Collapse = &semantic.CollapseOptions{
		Field: fromURLFieldName(collapseField),
	}

	if sizeStr := query.Get("collapse.size"); sizeStr != "" {
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("invalid collapse.size: %w", err)
		}
		opts.Collapse.Size = size
	}

	if sortParam := query.Get("collapse.sort"); sortParam != "" {
		direction := semantic.SortAsc
		field := sortParam
		if strings.HasPrefix(sortParam, "-") {
			field = strings.TrimPrefix(sortParam, "-")
			direction = semantic.SortDesc
		}
		opts.Collapse.Sort = []semantic.SortField{{
			Field:     fromURLFieldName(field),
			Direction: direction,
		}}
	}

	return nil
}

// parseFiltersFromQuery parses filter and hfilter parameters from URL query.
// Format: filter[field][operator]=value or hfilter[field][operator]=value
func parseFiltersFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	for key, values := range query {
		var prefix string
		var hidden bool

		if strings.HasPrefix(key, "filter[") {
			prefix = "filter["
		} else if strings.HasPrefix(key, "hfilter[") {
			prefix = "hfilter["
			hidden = true
		} else {
			continue
		}

		// Extract field and operator from filter[field][operator]
		parts := strings.Split(key, "][")
		if len(parts) != 2 {
			continue
		}

		urlField := strings.TrimPrefix(parts[0], prefix)
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
				Hidden:       hidden,
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
// Format: facet[field]=limit, facet[field].cursor=abc, facet[field].sort=count
func parseFacetsFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	facetMap := make(map[string]*semantic.FacetRequest)

	for key, values := range query {
		if !strings.HasPrefix(key, "facet[") {
			continue
		}

		fullKey := strings.TrimPrefix(key, "facet[")

		// Check for sub-parameters: facet[field].cursor, facet[field].sort
		if idx := strings.Index(fullKey, "]."); idx >= 0 {
			urlField := fullKey[:idx]
			subParam := fullKey[idx+2:]
			field := fromURLFieldName(urlField)

			fr, ok := facetMap[field]
			if !ok {
				fr = &semantic.FacetRequest{Field: field}
				facetMap[field] = fr
			}

			switch subParam {
			case "cursor":
				if len(values) > 0 {
					fr.Cursor = values[0]
				}
			case "sort":
				if len(values) > 0 {
					fr.Sort = values[0]
				}
			}

			continue
		}

		// Main facet parameter: facet[field]=limit
		urlField := strings.TrimSuffix(fullKey, "]")
		field := fromURLFieldName(urlField)

		if len(values) == 0 {
			continue
		}

		fr, ok := facetMap[field]
		if !ok {
			fr = &semantic.FacetRequest{Field: field}
			facetMap[field] = fr
		}

		if limit, err := strconv.Atoi(values[0]); err == nil && limit > 0 {
			fr.Limit = limit
		}
	}

	for _, fr := range facetMap {
		opts.Facets = append(opts.Facets, *fr)
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
