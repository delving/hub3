package elasticsearch8

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/google/go-cmp/cmp"
)

func TestAggregationBuilder_BuildAggregations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mode   AggregationMode
		facets []semantic.FacetRequest
		want   map[string]any
	}{
		{
			name: "single flat facet with default limit",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator"},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_creator.keyword",
						"size":  20,
					},
				},
			},
		},
		{
			name: "multiple flat facets",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator"},
				{Field: "dc_subject"},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_creator.keyword",
						"size":  20,
					},
				},
				"dc_subject": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_subject.keyword",
						"size":  20,
					},
				},
			},
		},
		{
			name: "nested mode with default limit",
			mode: AggNested,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator"},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"nested": map[string]any{
						"path": "resources.entries",
					},
					"aggs": map[string]any{
						"filtered": map[string]any{
							"filter": map[string]any{
								"term": map[string]any{
									"resources.entries.searchLabel": "dc_creator",
								},
							},
							"aggs": map[string]any{
								"values": map[string]any{
									"terms": map[string]any{
										"field": "resources.entries.@value.keyword",
										"size":  20,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "sort index produces _key asc",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_type", Sort: "index"},
			},
			want: map[string]any{
				"dc_type": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_type.keyword",
						"size":  20,
						"order": map[string]any{"_key": "asc"},
					},
				},
			},
		},
		{
			name: "sort count produces _count desc",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_type", Sort: "count"},
			},
			want: map[string]any{
				"dc_type": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_type.keyword",
						"size":  20,
						"order": map[string]any{"_count": "desc"},
					},
				},
			},
		},
		{
			name: "custom limit",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator", Limit: 50},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_creator.keyword",
						"size":  50,
					},
				},
			},
		},
		{
			name: "zero limit defaults to 20",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator", Limit: 0},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_creator.keyword",
						"size":  20,
					},
				},
			},
		},
		{
			name: "field name colon translation in flat mode",
			mode: AggFlat,
			facets: []semantic.FacetRequest{
				{Field: "dc:creator"},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"terms": map[string]any{
						"field": "fields.dc_creator.keyword",
						"size":  20,
					},
				},
			},
		},
		{
			name: "nested mode sort count",
			mode: AggNested,
			facets: []semantic.FacetRequest{
				{Field: "dc_creator", Sort: "count", Limit: 10},
			},
			want: map[string]any{
				"dc_creator": map[string]any{
					"nested": map[string]any{
						"path": "resources.entries",
					},
					"aggs": map[string]any{
						"filtered": map[string]any{
							"filter": map[string]any{
								"term": map[string]any{
									"resources.entries.searchLabel": "dc_creator",
								},
							},
							"aggs": map[string]any{
								"values": map[string]any{
									"terms": map[string]any{
										"field": "resources.entries.@value.keyword",
										"size":  10,
										"order": map[string]any{"_count": "desc"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty facets returns empty map",
			mode: AggFlat,
			facets: []semantic.FacetRequest{},
			want:   map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ab := &AggregationBuilder{Mode: tt.mode}
			got := ab.BuildAggregations(tt.facets)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BuildAggregations() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
