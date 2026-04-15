package elasticsearch8

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// QueryBuilder constructs Elasticsearch query JSON from semantic.QueryOptions.
type QueryBuilder struct {
	orgID         string
	contextOrgIDs []string
}

// NewQueryBuilder creates a new QueryBuilder for the given organisation.
// Optional contextOrgIDs specify additional organisations to include in search (cross-index).
func NewQueryBuilder(orgID string, contextOrgIDs ...string) *QueryBuilder {
	return &QueryBuilder{orgID: orgID, contextOrgIDs: contextOrgIDs}
}

// BuildQuery converts semantic.QueryOptions into an Elasticsearch search request body.
// It returns the JSON-encoded request body suitable for the ES _search API.
func (qb *QueryBuilder) BuildQuery(opts *semantic.QueryOptions) ([]byte, error) {
	body := map[string]interface{}{
		"track_total_hits": true,
	}

	boolQuery := map[string]interface{}{}

	// Always filter on docType and orgID(s).
	orgFilter := qb.buildOrgFilter()

	filters := []interface{}{
		map[string]interface{}{"term": map[string]interface{}{"meta.docType": "fragmentGraph"}},
		orgFilter,
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

	// Collapse: group results by a field value with inner_hits.
	if opts.Collapse != nil {
		body["collapse"] = qb.buildCollapse(opts.Collapse)
	}

	// Peek: when true, return zero items (only aggregations/facets matter).
	if opts.Peek {
		body["size"] = 0
	}

	// Debug: when set, enable ES explain mode for scoring diagnostics.
	if opts.Debug != "" {
		body["explain"] = true
	}

	// FacetBoolType: stored for use by the Store layer.
	// When FacetBoolType == "and", the Store layer will use a post_filter approach
	// so that facet counts reflect the unfiltered result set while the main results
	// are narrowed. The actual post_filter construction happens in the Store, not here.

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

// buildPropertyFilter translates a PropertyFilter into a nested ES query against
// the resources.entries structure. The searchLabel field identifies the predicate
// (e.g., "dc:type") and @value.keyword holds the filterable value.
func (qb *QueryBuilder) buildPropertyFilter(pf *semantic.PropertyFilter) (interface{}, bool, error) {
	label := pf.FieldName // e.g., "dc:type"

	var valueClause map[string]interface{}

	switch pf.OperatorType {
	case semantic.OpEqual:
		valueClause = map[string]interface{}{"term": map[string]interface{}{"resources.entries.@value.keyword": pf.Value}}

	case semantic.OpNotEqual:
		valueClause = map[string]interface{}{"term": map[string]interface{}{"resources.entries.@value.keyword": pf.Value}}
		return nestedEntriesQuery(label, valueClause), true, nil

	case semantic.OpIn:
		valueClause = map[string]interface{}{"terms": map[string]interface{}{"resources.entries.@value.keyword": pf.Value}}

	case semantic.OpNotIn:
		valueClause = map[string]interface{}{"terms": map[string]interface{}{"resources.entries.@value.keyword": pf.Value}}
		return nestedEntriesQuery(label, valueClause), true, nil

	case semantic.OpContains:
		valueClause = map[string]interface{}{"match": map[string]interface{}{"resources.entries.@value": pf.Value}}

	case semantic.OpStartsWith:
		valueClause = map[string]interface{}{"prefix": map[string]interface{}{"resources.entries.@value.keyword": pf.Value}}

	case semantic.OpGreaterThan:
		valueClause = map[string]interface{}{"range": map[string]interface{}{"resources.entries.@value.keyword": map[string]interface{}{"gt": pf.Value}}}

	case semantic.OpGreaterEqual:
		valueClause = map[string]interface{}{"range": map[string]interface{}{"resources.entries.@value.keyword": map[string]interface{}{"gte": pf.Value}}}

	case semantic.OpLessThan:
		valueClause = map[string]interface{}{"range": map[string]interface{}{"resources.entries.@value.keyword": map[string]interface{}{"lt": pf.Value}}}

	case semantic.OpLessEqual:
		valueClause = map[string]interface{}{"range": map[string]interface{}{"resources.entries.@value.keyword": map[string]interface{}{"lte": pf.Value}}}

	default:
		return nil, false, fmt.Errorf("unsupported property filter operator: %s", pf.OperatorType)
	}

	return nestedEntriesQuery(label, valueClause), false, nil
}

// buildRangeFilter translates a RangeFilter into a nested ES range query
// against resources.entries.@value.keyword.
func (qb *QueryBuilder) buildRangeFilter(rf *semantic.RangeFilter) interface{} {
	rangeCond := map[string]interface{}{}

	if rf.Min != nil {
		rangeCond["gte"] = rf.Min
	}

	if rf.Max != nil {
		rangeCond["lte"] = rf.Max
	}

	valueClause := map[string]interface{}{
		"range": map[string]interface{}{"resources.entries.@value.keyword": rangeCond},
	}

	return nestedEntriesQuery(rf.FieldName, valueClause)
}

// buildExistsFilter translates an ExistsFilter into a nested query that checks
// whether any entry exists with the given searchLabel.
func (qb *QueryBuilder) buildExistsFilter(ef *semantic.ExistsFilter) interface{} {
	// An entry exists for this field if any nested entry has the matching searchLabel.
	labelClause := map[string]interface{}{
		"term": map[string]interface{}{"resources.entries.searchLabel": ef.FieldName},
	}

	return map[string]interface{}{
		"nested": map[string]interface{}{
			"path": "resources.entries",
			"query": labelClause,
		},
	}
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

// buildCollapse constructs the ES collapse clause from CollapseOptions.
func (qb *QueryBuilder) buildCollapse(co *semantic.CollapseOptions) map[string]interface{} {
	size := co.Size
	if size <= 0 {
		size = 1
	}

	innerHits := map[string]interface{}{
		"name": "collapse",
		"size": size,
	}

	if len(co.Sort) > 0 {
		innerHits["sort"] = qb.buildSort(co.Sort)
	}

	return map[string]interface{}{
		"field":                          translateField(co.Field),
		"inner_hits":                     innerHits,
		"max_concurrent_group_requests":  4,
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

// buildOrgFilter returns a term or terms filter depending on whether context orgs are set.
func (qb *QueryBuilder) buildOrgFilter() interface{} {
	if len(qb.contextOrgIDs) == 0 {
		return map[string]interface{}{"term": map[string]interface{}{"meta.orgID": qb.orgID}}
	}

	allOrgs := make([]interface{}, 0, 1+len(qb.contextOrgIDs))
	allOrgs = append(allOrgs, qb.orgID)

	for _, id := range qb.contextOrgIDs {
		allOrgs = append(allOrgs, id)
	}

	return map[string]interface{}{"terms": map[string]interface{}{"meta.orgID": allOrgs}}
}

// nestedEntriesQuery builds a nested query on resources.entries that matches both
// the searchLabel (predicate name like "dc:type") and a value condition.
func nestedEntriesQuery(searchLabel string, valueClause map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"nested": map[string]interface{}{
			"path": "resources.entries",
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{"term": map[string]interface{}{"resources.entries.searchLabel": searchLabel}},
						valueClause,
					},
				},
			},
		},
	}
}

// translateField converts semantic field names to Elasticsearch field names.
// Colons are replaced with underscores (e.g., dc:creator becomes dc_creator).
func translateField(field string) string {
	return strings.ReplaceAll(field, ":", "_")
}
