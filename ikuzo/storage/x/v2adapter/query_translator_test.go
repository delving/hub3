package v2adapter

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestQueryTranslator_TranslateToV2Query(t *testing.T) {
	qt := NewQueryTranslator("test-org")

	tests := []struct {
		name        string
		opts        *semantic.QueryOptions
		wantParams  map[string]string
		wantMulti   map[string][]string
		wantErr     bool
		description string
	}{
		{
			name: "simple text query",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value: "test query",
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"query":      "test query",
			},
			description: "Basic text query should set itemFormat and query param",
		},
		{
			name: "fuzzy text query",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value: "test query",
					Fuzzy: true,
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"query":      "test~ query~",
			},
			description: "Fuzzy query should add ~ to each term",
		},
		{
			name: "AND operator query",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:    "test query",
					Operator: "AND",
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"query":      "test AND query",
			},
			description: "AND operator should join terms with AND",
		},
		{
			name: "query with search fields",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:  "test",
					Fields: []string{"title", "description"},
				},
			},
			wantParams: map[string]string{
				"itemFormat":   "semantic",
				"query":        "test",
				"searchFields": "title,description",
			},
			description: "Search fields should be comma-separated",
		},
		{
			name: "equal filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "creator",
						OperatorType: semantic.OpEqual,
						Value:        "John Doe",
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"creator:\"John Doe\""},
			},
			description: "Equal filter should create field:value syntax",
		},
		{
			name: "not equal filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "type",
						OperatorType: semantic.OpNotEqual,
						Value:        "draft",
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"NOT type:draft"},
			},
			description: "Not equal filter should use NOT prefix",
		},
		{
			name: "in filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "status",
						OperatorType: semantic.OpIn,
						Value:        []string{"published", "archived"},
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"status:(published OR archived)"},
			},
			description: "In filter should create OR syntax",
		},
		{
			name: "not in filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "status",
						OperatorType: semantic.OpNotIn,
						Value:        []string{"draft", "deleted"},
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"NOT status:(draft OR deleted)"},
			},
			description: "Not in filter should use NOT with OR syntax",
		},
		{
			name: "contains filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "description",
						OperatorType: semantic.OpContains,
						Value:        "archive",
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"description:*archive*"},
			},
			description: "Contains filter should use wildcards",
		},
		{
			name: "starts with filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "title",
						OperatorType: semantic.OpStartsWith,
						Value:        "The",
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"title:The*"},
			},
			description: "Starts with filter should use trailing wildcard",
		},
		{
			name: "greater than filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpGreaterThan,
						Value:        2000,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"year:{2000 TO *}"},
			},
			description: "Greater than filter should use exclusive range",
		},
		{
			name: "greater equal filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpGreaterEqual,
						Value:        2000,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"year:[2000 TO *]"},
			},
			description: "Greater equal filter should use inclusive range",
		},
		{
			name: "less than filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpLessThan,
						Value:        2020,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"year:{* TO 2020}"},
			},
			description: "Less than filter should use exclusive range",
		},
		{
			name: "less equal filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpLessEqual,
						Value:        2020,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"year:[* TO 2020]"},
			},
			description: "Less equal filter should use inclusive range",
		},
		{
			name: "range filter with min and max",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.RangeFilter{
						FieldName:    "year",
						OperatorType: semantic.OpGreaterEqual,
						Min:          1900,
						Max:          2000,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"year:[1900 TO 2000]"},
			},
			description: "Range filter should use inclusive brackets",
		},
		{
			name: "exists filter",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.ExistsFilter{
						FieldName: "thumbnail",
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"_exists_:thumbnail"},
			},
			description: "Exists filter should use _exists_ syntax",
		},
		{
			name: "multiple filters",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "type",
						OperatorType: semantic.OpEqual,
						Value:        "image",
					},
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpGreaterEqual,
						Value:        2000,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
			},
			wantMulti: map[string][]string{
				"qf": {"type:image", "year:[2000 TO *]"},
			},
			description: "Multiple filters should each add qf param",
		},
		{
			name: "single facet",
			opts: &semantic.QueryOptions{
				Facets: []semantic.FacetRequest{
					{
						Field: "creator",
						Limit: 10,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat":  "semantic",
				"facet.limit": "10",
			},
			wantMulti: map[string][]string{
				"facet.field": {"creator"},
			},
			description: "Facet should add facet.field and facet.limit",
		},
		{
			name: "multiple facets",
			opts: &semantic.QueryOptions{
				Facets: []semantic.FacetRequest{
					{Field: "creator", Limit: 10},
					{Field: "type", Limit: 5},
					{Field: "year", Limit: 20},
				},
			},
			wantParams: map[string]string{
				"itemFormat":  "semantic",
				"facet.limit": "20",
			},
			wantMulti: map[string][]string{
				"facet.field": {"creator", "type", "year"},
			},
			description: "Multiple facets should add multiple facet.field params",
		},
		{
			name: "pagination first page",
			opts: &semantic.QueryOptions{
				Pagination: &semantic.Pagination{
					Page: 1,
					Size: 20,
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"start":      "0",
				"rows":       "20",
				"page":       "1",
			},
			description: "First page should have start=0",
		},
		{
			name: "pagination second page",
			opts: &semantic.QueryOptions{
				Pagination: &semantic.Pagination{
					Page: 2,
					Size: 20,
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"start":      "20",
				"rows":       "20",
				"page":       "2",
			},
			description: "Second page should have start=20",
		},
		{
			name: "pagination custom size",
			opts: &semantic.QueryOptions{
				Pagination: &semantic.Pagination{
					Page: 3,
					Size: 50,
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"start":      "100",
				"rows":       "50",
				"page":       "3",
			},
			description: "Custom size should calculate start correctly",
		},
		{
			name: "sort ascending",
			opts: &semantic.QueryOptions{
				Sort: []semantic.SortField{
					{
						Field:     "title",
						Direction: semantic.SortAsc,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"sortBy":     "title",
				"sortAsc":    "true",
			},
			description: "Sort ascending should set sortAsc=true",
		},
		{
			name: "sort descending",
			opts: &semantic.QueryOptions{
				Sort: []semantic.SortField{
					{
						Field:     "date",
						Direction: semantic.SortDesc,
					},
				},
			},
			wantParams: map[string]string{
				"itemFormat": "semantic",
				"sortBy":     "date",
				"sortAsc":    "false",
			},
			description: "Sort descending should set sortAsc=false",
		},
		{
			name: "complex query with everything",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:  "cultural heritage",
					Fields: []string{"title", "description"},
				},
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "type",
						OperatorType: semantic.OpEqual,
						Value:        "image",
					},
					&semantic.PropertyFilter{
						FieldName:    "year",
						OperatorType: semantic.OpGreaterEqual,
						Value:        1900,
					},
				},
				Facets: []semantic.FacetRequest{
					{Field: "creator", Limit: 10},
					{Field: "year", Limit: 20},
				},
				Pagination: &semantic.Pagination{
					Page: 2,
					Size: 25,
				},
				Sort: []semantic.SortField{
					{Field: "date", Direction: semantic.SortDesc},
				},
			},
			wantParams: map[string]string{
				"itemFormat":   "semantic",
				"query":        "cultural heritage",
				"searchFields": "title,description",
				"start":        "25",
				"rows":         "25",
				"page":         "2",
				"sortBy":       "date",
				"sortAsc":      "false",
				"facet.limit":  "20",
			},
			wantMulti: map[string][]string{
				"qf":          {"type:image", "year:[1900 TO *]"},
				"facet.field": {"creator", "year"},
			},
			description: "Complex query should combine all elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := qt.TranslateToV2Query(tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("TranslateToV2Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check single-value params
			for key, wantValue := range tt.wantParams {
				gotValue := params.Get(key)
				if gotValue != wantValue {
					t.Errorf("%s: param %s = %q, want %q", tt.description, key, gotValue, wantValue)
				}
			}

			// Check multi-value params
			for key, wantValues := range tt.wantMulti {
				gotValues := params[key]
				if len(gotValues) != len(wantValues) {
					t.Errorf("%s: param %s has %d values, want %d", tt.description, key, len(gotValues), len(wantValues))
					continue
				}
				for i, wantVal := range wantValues {
					if gotValues[i] != wantVal {
						t.Errorf("%s: param %s[%d] = %q, want %q", tt.description, key, i, gotValues[i], wantVal)
					}
				}
			}
		})
	}
}

func TestQueryTranslator_EscapeValue(t *testing.T) {
	qt := NewQueryTranslator("test-org")

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "simple string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "string with space",
			input: "hello world",
			want:  "\"hello world\"",
		},
		{
			name:  "string with colon",
			input: "http://example.com",
			want:  "\"http://example.com\"",
		},
		{
			name:  "string with special chars",
			input: "test+value",
			want:  "\"test+value\"",
		},
		{
			name:  "integer",
			input: 123,
			want:  "123",
		},
		{
			name:  "string with quotes",
			input: "test\"value",
			want:  "\"test\\\"value\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := qt.escapeValue(tt.input)
			if got != tt.want {
				t.Errorf("escapeValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateToV2Query_Collapse(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Collapse: &semantic.CollapseOptions{
			Field: "edm:dataProvider",
			Size:  3,
		},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	if got := params.Get("collapseOn"); got != "edm:dataProvider" {
		t.Errorf("collapseOn = %q, want %q", got, "edm:dataProvider")
	}
	if got := params.Get("collapseSize"); got != "3" {
		t.Errorf("collapseSize = %q, want %q", got, "3")
	}
}

func TestTranslateToV2Query_CollapseWithSort(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Collapse: &semantic.CollapseOptions{
			Field: "edm:dataProvider",
			Sort:  []semantic.SortField{{Field: "dc:date", Direction: semantic.SortDesc}},
		},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("collapseSort"); got != "dc:date" {
		t.Errorf("collapseSort = %q, want %q", got, "dc:date")
	}
}

func TestTranslateToV2Query_FacetBoolType(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		FacetBoolType: semantic.FacetBoolAnd,
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("facetBoolType"); got != "and" {
		t.Errorf("facetBoolType = %q, want %q", got, "and")
	}
}

func TestTranslateToV2Query_FacetBoolType_NotSetWhenEmpty(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("facetBoolType"); got != "" {
		t.Errorf("facetBoolType should be empty, got %q", got)
	}
}

func TestTranslateToV2Query_HiddenFilter(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Filters: []semantic.Filter{
			&semantic.PropertyFilter{
				FieldName:    "orgID",
				OperatorType: semantic.OpEqual,
				Value:        "museum-x",
				Hidden:       true,
			},
			&semantic.PropertyFilter{
				FieldName:    "dc:type",
				OperatorType: semantic.OpEqual,
				Value:        "painting",
			},
		},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	hqfValues := params["hqf"]
	if len(hqfValues) != 1 {
		t.Fatalf("expected 1 hqf, got %d: %v", len(hqfValues), hqfValues)
	}
	if hqfValues[0] != "orgID:\"museum-x\"" {
		t.Errorf("hqf[0] = %q, want %q", hqfValues[0], "orgID:\"museum-x\"")
	}

	qfValues := params["qf"]
	if len(qfValues) != 1 {
		t.Fatalf("expected 1 qf, got %d: %v", len(qfValues), qfValues)
	}
}

func TestTranslateToV2Query_Peek(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Peek: true,
		Facets: []semantic.FacetRequest{
			{Field: "dc:creator"},
			{Field: "dc:type"},
		},
		Pagination: &semantic.Pagination{Size: 0},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("peek"); got != "dc:creator,dc:type" {
		t.Errorf("peek = %q, want %q", got, "dc:creator,dc:type")
	}
}

func TestTranslateToV2Query_Debug(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Debug: "query",
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("echo"); got != "searchResponse" {
		t.Errorf("echo = %q, want %q", got, "searchResponse")
	}
}

func TestTranslateToV2Query_NoDebugWhenEmpty(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := params.Get("echo"); got != "" {
		t.Errorf("echo should be empty, got %q", got)
	}
}

func TestTranslateMLTQuery(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.SimilarOptions{
		Count:       10,
		Fields:      []string{"dc:title", "dc:creator"},
		MinTermFreq: 3,
		MinDocFreq:  5,
	}

	params := qt.TranslateMLTQuery("doc-123", opts)

	if got := params.Get("mlt"); got != "true" {
		t.Errorf("mlt = %q, want %q", got, "true")
	}
	if got := params.Get("mlt.id"); got != "doc-123" {
		t.Errorf("mlt.id = %q, want %q", got, "doc-123")
	}
	if got := params.Get("mlt.count"); got != "10" {
		t.Errorf("mlt.count = %q, want %q", got, "10")
	}
	if got := params.Get("itemFormat"); got != "semantic" {
		t.Errorf("itemFormat = %q, want %q", got, "semantic")
	}

	filterkeys := params["mlt.filterkey"]
	if len(filterkeys) != 2 {
		t.Fatalf("expected 2 mlt.filterkey values, got %d: %v", len(filterkeys), filterkeys)
	}
}

func TestTranslateMLTQuery_Defaults(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := semantic.DefaultSimilarOptions()

	params := qt.TranslateMLTQuery("doc-abc", opts)

	if got := params.Get("mlt"); got != "true" {
		t.Errorf("mlt = %q, want %q", got, "true")
	}
	if got := params.Get("mlt.count"); got != "5" {
		t.Errorf("mlt.count = %q, want %q", got, "5")
	}
	// No filterkeys when Fields is empty
	if filterkeys := params["mlt.filterkey"]; len(filterkeys) > 0 {
		t.Errorf("expected no mlt.filterkey with empty fields, got %v", filterkeys)
	}
}

func TestQueryTranslator_ItemFormatAlwaysSet(t *testing.T) {
	qt := NewQueryTranslator("test-org")

	// Test with minimal options
	params, err := qt.TranslateToV2Query(&semantic.QueryOptions{})
	if err != nil {
		t.Fatalf("TranslateToV2Query() unexpected error: %v", err)
	}

	// CRITICAL: itemFormat must always be "semantic"
	itemFormat := params.Get("itemFormat")
	if itemFormat != "semantic" {
		t.Errorf("itemFormat = %q, want %q - CRITICAL: must always be 'semantic'", itemFormat, "semantic")
	}
}
