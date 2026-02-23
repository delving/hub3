package elasticsearch8

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// QueryBuilder constructs Elasticsearch query JSON from semantic.QueryOptions.
type QueryBuilder struct {
	orgID string
}

// NewQueryBuilder creates a new QueryBuilder for the given organisation.
func NewQueryBuilder(orgID string) *QueryBuilder {
	return &QueryBuilder{orgID: orgID}
}

// BuildQuery converts semantic.QueryOptions into an Elasticsearch search request body.
// It returns the JSON-encoded request body suitable for the ES _search API.
func (qb *QueryBuilder) BuildQuery(opts *semantic.QueryOptions) ([]byte, error) {
	body := map[string]interface{}{
		"track_total_hits": true,
	}

	boolQuery := map[string]interface{}{}

	// Always filter on docType and orgID.
	filters := []interface{}{
		map[string]interface{}{"term": map[string]interface{}{"meta.docType": "fragmentGraph"}},
		map[string]interface{}{"term": map[string]interface{}{"meta.orgID": qb.orgID}},
	}

	if opts == nil {
		boolQuery["filter"] = filters
		body["query"] = map[string]interface{}{"bool": boolQuery}

		return json.Marshal(body)
	}

	// Text query.
	if opts.Query != nil && opts.Query.Value != "" {
		must, err := qb.buildTextQuery(opts.Query)
		if err != nil {
			return nil, fmt.Errorf("building text query: %w", err)
		}

		boolQuery["must"] = must
	}

	// Filters from QueryOptions.
	mustNot := []interface{}{}

	for _, f := range opts.Filters {
		clause, negated, err := qb.buildFilter(f)
		if err != nil {
			return nil, fmt.Errorf("building filter for field %q: %w", f.Field(), err)
		}

		if negated {
			mustNot = append(mustNot, clause)
		} else {
			filters = append(filters, clause)
		}
	}

	boolQuery["filter"] = filters

	if len(mustNot) > 0 {
		boolQuery["must_not"] = mustNot
	}

	body["query"] = map[string]interface{}{"bool": boolQuery}

	// Pagination.
	if opts.Pagination != nil {
		body["from"] = opts.Pagination.GetOffset()
		body["size"] = opts.Pagination.GetSize()
	}

	// Sorting.
	if len(opts.Sort) > 0 {
		body["sort"] = qb.buildSort(opts.Sort)
	}

	return json.Marshal(body)
}

// buildTextQuery builds the must clause for a text query.
func (qb *QueryBuilder) buildTextQuery(tq *semantic.TextQuery) (interface{}, error) {
	if len(tq.Fields) > 0 {
		return qb.buildMultiMatch(tq), nil
	}

	return qb.buildQueryString(tq), nil
}

// buildQueryString builds a query_string query targeting the full_text field.
func (qb *QueryBuilder) buildQueryString(tq *semantic.TextQuery) interface{} {
	qs := map[string]interface{}{
		"query":            tq.Value,
		"default_field":    "full_text",
		"default_operator": qb.resolveOperator(tq.Operator),
	}

	if tq.Fuzzy {
		qs["fuzziness"] = "AUTO"
	}

	return map[string]interface{}{"query_string": qs}
}

// buildMultiMatch builds a multi_match query across specified fields.
func (qb *QueryBuilder) buildMultiMatch(tq *semantic.TextQuery) interface{} {
	translated := make([]string, len(tq.Fields))
	for i, f := range tq.Fields {
		translated[i] = translateField(f)
	}

	mm := map[string]interface{}{
		"query":    tq.Value,
		"fields":   translated,
		"operator": strings.ToLower(qb.resolveOperator(tq.Operator)),
	}

	if tq.Fuzzy {
		mm["fuzziness"] = "AUTO"
	}

	return map[string]interface{}{"multi_match": mm}
}

// resolveOperator returns the ES operator string, defaulting to OR.
func (qb *QueryBuilder) resolveOperator(op string) string {
	upper := strings.ToUpper(op)
	if upper == "AND" {
		return "AND"
	}

	return "OR"
}

// buildFilter converts a semantic.Filter to an ES query clause.
// The second return value indicates whether the clause should go into must_not.
func (qb *QueryBuilder) buildFilter(f semantic.Filter) (interface{}, bool, error) {
	switch ft := f.(type) {
	case *semantic.PropertyFilter:
		return qb.buildPropertyFilter(ft)
	case *semantic.RangeFilter:
		clause := qb.buildRangeFilter(ft)
		return clause, false, nil
	case *semantic.ExistsFilter:
		clause := qb.buildExistsFilter(ft)
		return clause, false, nil
	case *semantic.GeoBBoxFilter:
		clause := qb.buildGeoBBoxFilter(ft)
		return clause, false, nil
	case *semantic.GeoDistanceFilter:
		clause := qb.buildGeoDistanceFilter(ft)
		return clause, false, nil
	case *semantic.GeoPolygonFilter:
		clause := qb.buildGeoPolygonFilter(ft)
		return clause, false, nil
	default:
		return nil, false, fmt.Errorf("unsupported filter type: %T", f)
	}
}

// buildPropertyFilter translates a PropertyFilter into an ES clause.
func (qb *QueryBuilder) buildPropertyFilter(pf *semantic.PropertyFilter) (interface{}, bool, error) {
	field := translateField(pf.FieldName)

	switch pf.OperatorType {
	case semantic.OpEqual:
		return map[string]interface{}{"term": map[string]interface{}{field: pf.Value}}, false, nil

	case semantic.OpNotEqual:
		return map[string]interface{}{"term": map[string]interface{}{field: pf.Value}}, true, nil

	case semantic.OpIn:
		return map[string]interface{}{"terms": map[string]interface{}{field: pf.Value}}, false, nil

	case semantic.OpNotIn:
		return map[string]interface{}{"terms": map[string]interface{}{field: pf.Value}}, true, nil

	case semantic.OpContains:
		return map[string]interface{}{"match": map[string]interface{}{field: pf.Value}}, false, nil

	case semantic.OpStartsWith:
		return map[string]interface{}{"prefix": map[string]interface{}{field: pf.Value}}, false, nil

	case semantic.OpGreaterThan:
		return map[string]interface{}{"range": map[string]interface{}{field: map[string]interface{}{"gt": pf.Value}}}, false, nil

	case semantic.OpGreaterEqual:
		return map[string]interface{}{"range": map[string]interface{}{field: map[string]interface{}{"gte": pf.Value}}}, false, nil

	case semantic.OpLessThan:
		return map[string]interface{}{"range": map[string]interface{}{field: map[string]interface{}{"lt": pf.Value}}}, false, nil

	case semantic.OpLessEqual:
		return map[string]interface{}{"range": map[string]interface{}{field: map[string]interface{}{"lte": pf.Value}}}, false, nil

	default:
		return nil, false, fmt.Errorf("unsupported property filter operator: %s", pf.OperatorType)
	}
}

// buildRangeFilter translates a RangeFilter into an ES range clause.
func (qb *QueryBuilder) buildRangeFilter(rf *semantic.RangeFilter) interface{} {
	field := translateField(rf.FieldName)
	rangeCond := map[string]interface{}{}

	if rf.Min != nil {
		rangeCond["gte"] = rf.Min
	}

	if rf.Max != nil {
		rangeCond["lte"] = rf.Max
	}

	return map[string]interface{}{"range": map[string]interface{}{field: rangeCond}}
}

// buildExistsFilter translates an ExistsFilter into an ES exists clause.
func (qb *QueryBuilder) buildExistsFilter(ef *semantic.ExistsFilter) interface{} {
	return map[string]interface{}{"exists": map[string]interface{}{"field": translateField(ef.FieldName)}}
}

// buildGeoBBoxFilter translates a GeoBBoxFilter into an ES geo_bounding_box clause.
func (qb *QueryBuilder) buildGeoBBoxFilter(gbf *semantic.GeoBBoxFilter) interface{} {
	field := translateField(gbf.FieldName)

	return map[string]interface{}{
		"geo_bounding_box": map[string]interface{}{
			field: map[string]interface{}{
				"top_left": map[string]interface{}{
					"lat": gbf.Bounds.North,
					"lon": gbf.Bounds.West,
				},
				"bottom_right": map[string]interface{}{
					"lat": gbf.Bounds.South,
					"lon": gbf.Bounds.East,
				},
			},
		},
	}
}

// buildGeoDistanceFilter translates a GeoDistanceFilter into an ES geo_distance clause.
func (qb *QueryBuilder) buildGeoDistanceFilter(gdf *semantic.GeoDistanceFilter) interface{} {
	field := translateField(gdf.FieldName)

	return map[string]interface{}{
		"geo_distance": map[string]interface{}{
			"distance": gdf.Distance,
			field: map[string]interface{}{
				"lat": gdf.Point.Lat,
				"lon": gdf.Point.Lon,
			},
		},
	}
}

// buildGeoPolygonFilter translates a GeoPolygonFilter into an ES geo_polygon clause.
func (qb *QueryBuilder) buildGeoPolygonFilter(gpf *semantic.GeoPolygonFilter) interface{} {
	field := translateField(gpf.FieldName)

	points := make([]map[string]interface{}, 0)

	if len(gpf.Polygon.Coordinates) > 0 {
		for _, coord := range gpf.Polygon.Coordinates[0] {
			if len(coord) >= 2 {
				points = append(points, map[string]interface{}{
					"lat": coord[1],
					"lon": coord[0],
				})
			}
		}
	}

	return map[string]interface{}{
		"geo_polygon": map[string]interface{}{
			field: map[string]interface{}{
				"points": points,
			},
		},
	}
}

// buildSort converts sort fields to ES sort clauses.
func (qb *QueryBuilder) buildSort(fields []semantic.SortField) []interface{} {
	sort := make([]interface{}, len(fields))

	for i, sf := range fields {
		dir := "asc"
		if !sf.IsAscending() {
			dir = "desc"
		}

		sort[i] = map[string]interface{}{
			translateField(sf.Field): map[string]interface{}{
				"order": dir,
			},
		}
	}

	return sort
}

// translateField converts semantic field names to Elasticsearch field names.
// Colons are replaced with underscores (e.g., dc:creator becomes dc_creator).
func translateField(field string) string {
	return strings.ReplaceAll(field, ":", "_")
}
