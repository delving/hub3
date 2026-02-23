package elasticsearch8

import (
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// AggregationMode determines how facet fields are queried.
type AggregationMode int

const (
	// AggFlat uses direct field aggregations (fields.*.keyword pattern).
	AggFlat AggregationMode = iota
	// AggNested uses nested aggregations (resources.entries searchLabel pattern).
	AggNested
)

// defaultFacetSize is used when FacetRequest.Limit is zero.
const defaultFacetSize = 20

// AggregationBuilder constructs Elasticsearch aggregation queries from
// semantic FacetRequest values.
type AggregationBuilder struct {
	Mode AggregationMode
}

// BuildAggregations returns the "aggs" portion of an ES query for the given
// facet requests. Each facet produces one top-level aggregation key.
func (ab *AggregationBuilder) BuildAggregations(facets []semantic.FacetRequest) map[string]any {
	aggs := make(map[string]any, len(facets))

	for _, f := range facets {
		size := f.Limit
		if size == 0 {
			size = defaultFacetSize
		}

		key := translateFieldName(f.Field)

		switch ab.Mode {
		case AggNested:
			aggs[key] = ab.buildNested(f.Field, size, f.Sort)
		default:
			aggs[key] = ab.buildFlat(key, size, f.Sort)
		}
	}

	return aggs
}

// buildFlat creates a simple terms aggregation on fields.<name>.keyword.
func (ab *AggregationBuilder) buildFlat(fieldKey string, size int, sort string) map[string]any {
	terms := map[string]any{
		"field": "fields." + fieldKey + ".keyword",
		"size":  size,
	}

	if order := sortOrder(sort); order != nil {
		terms["order"] = order
	}

	return map[string]any{
		"terms": terms,
	}
}

// buildNested creates a nested aggregation on resources.entries, filtered by searchLabel.
func (ab *AggregationBuilder) buildNested(field string, size int, sort string) map[string]any {
	terms := map[string]any{
		"field": "resources.entries.@value.keyword",
		"size":  size,
	}

	if order := sortOrder(sort); order != nil {
		terms["order"] = order
	}

	return map[string]any{
		"nested": map[string]any{
			"path": "resources.entries",
		},
		"aggs": map[string]any{
			"filtered": map[string]any{
				"filter": map[string]any{
					"term": map[string]any{
						"resources.entries.searchLabel": field,
					},
				},
				"aggs": map[string]any{
					"values": map[string]any{
						"terms": terms,
					},
				},
			},
		},
	}
}

// sortOrder converts a semantic sort string to an ES order clause.
// Returns nil when no explicit order is requested.
func sortOrder(sort string) map[string]any {
	switch strings.ToLower(sort) {
	case "count":
		return map[string]any{"_count": "desc"}
	case "index", "value":
		return map[string]any{"_key": "asc"}
	default:
		return nil
	}
}

// translateFieldName replaces colons with underscores so that field names
// are safe for the flat fields.* path.
func translateFieldName(name string) string {
	return strings.ReplaceAll(name, ":", "_")
}
