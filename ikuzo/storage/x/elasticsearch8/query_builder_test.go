package elasticsearch8

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestQueryBuilder_BuildQuery(t *testing.T) {
	const orgID = "test-org"

	tests := []struct {
		name    string
		opts    *semantic.QueryOptions
		check   func(t *testing.T, result map[string]interface{})
		wantErr bool
	}{
		{
			name: "nil options produces match_all-style bool with filters",
			opts: nil,
			check: func(t *testing.T, result map[string]interface{}) {
				assertTrackTotalHits(t, result)
				assertBaseFilters(t, result, orgID)

				// Should not have must clause
				boolQ := getBoolQuery(t, result)
				if _, ok := boolQ["must"]; ok {
					t.Error("expected no must clause for nil options")
				}
			},
		},
		{
			name: "empty options produces match_all-style bool with filters",
			opts: &semantic.QueryOptions{},
			check: func(t *testing.T, result map[string]interface{}) {
				assertTrackTotalHits(t, result)
				assertBaseFilters(t, result, orgID)

				boolQ := getBoolQuery(t, result)
				if _, ok := boolQ["must"]; ok {
					t.Error("expected no must clause for empty options")
				}
			},
		},
		{
			name: "text query without fields produces query_string on full_text",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{Value: "rembrandt"},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				assertBaseFilters(t, result, orgID)

				boolQ := getBoolQuery(t, result)
				must, ok := boolQ["must"].(map[string]interface{})
				if !ok {
					t.Fatal("expected must clause")
				}

				qs, ok := must["query_string"].(map[string]interface{})
				if !ok {
					t.Fatal("expected query_string in must")
				}

				if qs["query"] != "rembrandt" {
					t.Errorf("expected query 'rembrandt', got %v", qs["query"])
				}

				if qs["default_field"] != "full_text" {
					t.Errorf("expected default_field 'full_text', got %v", qs["default_field"])
				}
			},
		},
		{
			name: "text query with fields produces multi_match",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:  "night watch",
					Fields: []string{"dc:title", "dc:description"},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				must, ok := boolQ["must"].(map[string]interface{})
				if !ok {
					t.Fatal("expected must clause")
				}

				mm, ok := must["multi_match"].(map[string]interface{})
				if !ok {
					t.Fatal("expected multi_match in must")
				}

				if mm["query"] != "night watch" {
					t.Errorf("expected query 'night watch', got %v", mm["query"])
				}

				fields, ok := mm["fields"].([]interface{})
				if !ok {
					t.Fatal("expected fields array")
				}

				if len(fields) != 2 {
					t.Fatalf("expected 2 fields, got %d", len(fields))
				}

				if fields[0] != "dc_title" || fields[1] != "dc_description" {
					t.Errorf("expected translated fields [dc_title, dc_description], got %v", fields)
				}
			},
		},
		{
			name: "fuzzy text query includes fuzziness",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{Value: "rembrant", Fuzzy: true},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				must := boolQ["must"].(map[string]interface{})
				qs := must["query_string"].(map[string]interface{})

				if qs["fuzziness"] != "AUTO" {
					t.Errorf("expected fuzziness 'AUTO', got %v", qs["fuzziness"])
				}
			},
		},
		{
			name: "fuzzy multi_match includes fuzziness",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:  "rembrant",
					Fuzzy:  true,
					Fields: []string{"dc:title"},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				must := boolQ["must"].(map[string]interface{})
				mm := must["multi_match"].(map[string]interface{})

				if mm["fuzziness"] != "AUTO" {
					t.Errorf("expected fuzziness 'AUTO', got %v", mm["fuzziness"])
				}
			},
		},
		{
			name: "AND operator in query_string",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{Value: "night watch", Operator: "AND"},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				must := boolQ["must"].(map[string]interface{})
				qs := must["query_string"].(map[string]interface{})

				if qs["default_operator"] != "AND" {
					t.Errorf("expected default_operator 'AND', got %v", qs["default_operator"])
				}
			},
		},
		{
			name: "pagination page 2 size 25",
			opts: &semantic.QueryOptions{
				Pagination: &semantic.Pagination{Page: 2, Size: 25},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				from, ok := result["from"].(float64)
				if !ok {
					t.Fatal("expected from field")
				}

				if from != 25 {
					t.Errorf("expected from=25 for page 2 size 25, got %v", from)
				}

				size, ok := result["size"].(float64)
				if !ok {
					t.Fatal("expected size field")
				}

				if size != 25 {
					t.Errorf("expected size=25, got %v", size)
				}
			},
		},
		{
			name: "sorting desc",
			opts: &semantic.QueryOptions{
				Sort: []semantic.SortField{
					{Field: "dc:date", Direction: semantic.SortDesc},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				sortArr, ok := result["sort"].([]interface{})
				if !ok {
					t.Fatal("expected sort array")
				}

				if len(sortArr) != 1 {
					t.Fatalf("expected 1 sort field, got %d", len(sortArr))
				}

				entry := sortArr[0].(map[string]interface{})
				fieldSort, ok := entry["dc_date"].(map[string]interface{})

				if !ok {
					t.Fatal("expected dc_date sort entry")
				}

				if fieldSort["order"] != "desc" {
					t.Errorf("expected order 'desc', got %v", fieldSort["order"])
				}
			},
		},
		{
			name: "PropertyFilter OpEqual produces term query",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "dc:creator",
						OperatorType: semantic.OpEqual,
						Value:        "Rembrandt",
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				filters := boolQ["filter"].([]interface{})

				// 2 base filters + 1 property filter
				if len(filters) != 3 {
					t.Fatalf("expected 3 filters, got %d", len(filters))
				}

				last := filters[2].(map[string]interface{})
				term, ok := last["term"].(map[string]interface{})

				if !ok {
					t.Fatal("expected term query")
				}

				if term["dc_creator"] != "Rembrandt" {
					t.Errorf("expected dc_creator=Rembrandt, got %v", term["dc_creator"])
				}
			},
		},
		{
			name: "PropertyFilter OpNotEqual goes to must_not",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "dc:type",
						OperatorType: semantic.OpNotEqual,
						Value:        "painting",
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				mustNot, ok := boolQ["must_not"].([]interface{})

				if !ok {
					t.Fatal("expected must_not clause")
				}

				if len(mustNot) != 1 {
					t.Fatalf("expected 1 must_not clause, got %d", len(mustNot))
				}

				term := mustNot[0].(map[string]interface{})["term"].(map[string]interface{})
				if term["dc_type"] != "painting" {
					t.Errorf("expected dc_type=painting in must_not, got %v", term["dc_type"])
				}

				// Should still have only 2 base filters (not 3)
				filters := boolQ["filter"].([]interface{})
				if len(filters) != 2 {
					t.Errorf("expected 2 base filters, got %d", len(filters))
				}
			},
		},
		{
			name: "RangeFilter produces range query",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.RangeFilter{
						FieldName: "dc:date",
						Min:       1600,
						Max:       1700,
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				filters := boolQ["filter"].([]interface{})

				if len(filters) != 3 {
					t.Fatalf("expected 3 filters, got %d", len(filters))
				}

				last := filters[2].(map[string]interface{})
				rangeQ, ok := last["range"].(map[string]interface{})

				if !ok {
					t.Fatal("expected range query")
				}

				dateRange, ok := rangeQ["dc_date"].(map[string]interface{})
				if !ok {
					t.Fatal("expected dc_date range")
				}

				if dateRange["gte"] != float64(1600) {
					t.Errorf("expected gte=1600, got %v", dateRange["gte"])
				}

				if dateRange["lte"] != float64(1700) {
					t.Errorf("expected lte=1700, got %v", dateRange["lte"])
				}
			},
		},
		{
			name: "ExistsFilter produces exists query",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.ExistsFilter{FieldName: "nave:thumbnail"},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				filters := boolQ["filter"].([]interface{})

				if len(filters) != 3 {
					t.Fatalf("expected 3 filters, got %d", len(filters))
				}

				last := filters[2].(map[string]interface{})
				exists, ok := last["exists"].(map[string]interface{})

				if !ok {
					t.Fatal("expected exists query")
				}

				if exists["field"] != "nave_thumbnail" {
					t.Errorf("expected field 'nave_thumbnail', got %v", exists["field"])
				}
			},
		},
		{
			name: "GeoBBoxFilter produces geo_bounding_box query",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.GeoBBoxFilter{
						FieldName: "location",
						Bounds: &semantic.GeoBounds{
							North: 53.0,
							South: 51.0,
							East:  7.0,
							West:  3.0,
						},
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				filters := boolQ["filter"].([]interface{})

				last := filters[len(filters)-1].(map[string]interface{})
				gbb, ok := last["geo_bounding_box"].(map[string]interface{})

				if !ok {
					t.Fatal("expected geo_bounding_box query")
				}

				loc, ok := gbb["location"].(map[string]interface{})
				if !ok {
					t.Fatal("expected location field in geo_bounding_box")
				}

				tl := loc["top_left"].(map[string]interface{})
				if tl["lat"] != 53.0 || tl["lon"] != 3.0 {
					t.Errorf("unexpected top_left: %v", tl)
				}

				br := loc["bottom_right"].(map[string]interface{})
				if br["lat"] != 51.0 || br["lon"] != 7.0 {
					t.Errorf("unexpected bottom_right: %v", br)
				}
			},
		},
		{
			name: "GeoDistanceFilter produces geo_distance query",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.GeoDistanceFilter{
						FieldName: "location",
						Point:     &semantic.GeoPoint{Lat: 52.37, Lon: 4.89},
						Distance:  "10km",
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				boolQ := getBoolQuery(t, result)
				filters := boolQ["filter"].([]interface{})

				last := filters[len(filters)-1].(map[string]interface{})
				gd, ok := last["geo_distance"].(map[string]interface{})

				if !ok {
					t.Fatal("expected geo_distance query")
				}

				if gd["distance"] != "10km" {
					t.Errorf("expected distance '10km', got %v", gd["distance"])
				}

				loc := gd["location"].(map[string]interface{})
				if loc["lat"] != 52.37 || loc["lon"] != 4.89 {
					t.Errorf("unexpected location: %v", loc)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(orgID)

			data, err := qb.BuildQuery(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildQuery() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			tt.check(t, result)
		})
	}
}

func TestTranslateField(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"dc:creator", "dc_creator"},
		{"dc:title", "dc_title"},
		{"nave:thumbnail", "nave_thumbnail"},
		{"plain_field", "plain_field"},
		{"a:b:c", "a_b_c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := translateField(tt.input)
			if got != tt.want {
				t.Errorf("translateField(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Helper functions for test assertions.

func getBoolQuery(t *testing.T, result map[string]interface{}) map[string]interface{} {
	t.Helper()

	query, ok := result["query"].(map[string]interface{})
	if !ok {
		t.Fatal("expected query object")
	}

	boolQ, ok := query["bool"].(map[string]interface{})
	if !ok {
		t.Fatal("expected bool query")
	}

	return boolQ
}

func assertTrackTotalHits(t *testing.T, result map[string]interface{}) {
	t.Helper()

	tth, ok := result["track_total_hits"].(bool)
	if !ok || !tth {
		t.Error("expected track_total_hits to be true")
	}
}

func assertBaseFilters(t *testing.T, result map[string]interface{}, orgID string) {
	t.Helper()

	boolQ := getBoolQuery(t, result)

	filters, ok := boolQ["filter"].([]interface{})
	if !ok {
		t.Fatal("expected filter array")
	}

	if len(filters) < 2 {
		t.Fatalf("expected at least 2 base filters, got %d", len(filters))
	}

	// Check docType filter.
	docTypeFilter := filters[0].(map[string]interface{})
	term, ok := docTypeFilter["term"].(map[string]interface{})

	if !ok {
		t.Fatal("expected term query for docType")
	}

	if term["meta.docType"] != "fragmentGraph" {
		t.Errorf("expected meta.docType=fragmentGraph, got %v", term["meta.docType"])
	}

	// Check orgID filter.
	orgFilter := filters[1].(map[string]interface{})
	term, ok = orgFilter["term"].(map[string]interface{})

	if !ok {
		t.Fatal("expected term query for orgID")
	}

	if term["meta.orgID"] != orgID {
		t.Errorf("expected meta.orgID=%s, got %v", orgID, term["meta.orgID"])
	}
}
