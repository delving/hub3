package semantic

import (
	"net/http/httptest"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*testing.T, *semantic.QueryOptions)
	}{
		{
			name:    "empty query",
			url:     "/search",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if opts.Query != nil {
					t.Error("Expected no query")
				}
				if len(opts.Filters) != 0 {
					t.Error("Expected no filters")
				}
			},
		},
		{
			name:    "text query",
			url:     "/search?query=rembrandt",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if opts.Query == nil {
					t.Fatal("Expected query to be set")
				}
				if opts.Query.Value != "rembrandt" {
					t.Errorf("Query value = %v, want rembrandt", opts.Query.Value)
				}
			},
		},
		{
			name:    "simple filter",
			url:     "/search?filter[creator][eq]=Rembrandt",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Filters) != 1 {
					t.Fatalf("Expected 1 filter, got %d", len(opts.Filters))
				}
				filter := opts.Filters[0]
				if filter.Field() != "creator" {
					t.Errorf("Field = %v, want creator", filter.Field())
				}
				if filter.Operator() != semantic.OpEqual {
					t.Errorf("Operator = %v, want eq", filter.Operator())
				}
			},
		},
		{
			name:    "multiple filters",
			url:     "/search?filter[creator][eq]=Rembrandt&filter[type][in]=Painting",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Filters) != 2 {
					t.Fatalf("Expected 2 filters, got %d", len(opts.Filters))
				}
			},
		},
		{
			name:    "facet request",
			url:     "/search?facet[type]=10&facet[creator]=20",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Facets) != 2 {
					t.Fatalf("Expected 2 facets, got %d", len(opts.Facets))
				}
				// Check both facets exist with correct limits (order may vary with map-based parsing)
				facetLimits := map[string]int{}
				for _, f := range opts.Facets {
					facetLimits[f.Field] = f.Limit
				}
				if facetLimits["type"] != 10 {
					t.Errorf("type facet limit = %v, want 10", facetLimits["type"])
				}
				if facetLimits["creator"] != 20 {
					t.Errorf("creator facet limit = %v, want 20", facetLimits["creator"])
				}
			},
		},
		{
			name:    "pagination",
			url:     "/search?page=2&size=50",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if opts.Pagination == nil {
					t.Fatal("Expected pagination to be set")
				}
				if opts.Pagination.Page != 2 {
					t.Errorf("Page = %v, want 2", opts.Pagination.Page)
				}
				if opts.Pagination.Size != 50 {
					t.Errorf("Size = %v, want 50", opts.Pagination.Size)
				}
			},
		},
		{
			name:    "sort ascending",
			url:     "/search?sort=+title",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Sort) != 1 {
					t.Fatalf("Expected 1 sort field, got %d", len(opts.Sort))
				}
				if opts.Sort[0].Field != "title" {
					t.Errorf("Sort field = %v, want title", opts.Sort[0].Field)
				}
				if opts.Sort[0].Direction != semantic.SortAsc {
					t.Errorf("Sort direction = %v, want ASC", opts.Sort[0].Direction)
				}
			},
		},
		{
			name:    "sort descending",
			url:     "/search?sort=-dateCreated",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Sort) != 1 {
					t.Fatalf("Expected 1 sort field, got %d", len(opts.Sort))
				}
				if opts.Sort[0].Direction != semantic.SortDesc {
					t.Errorf("Sort direction = %v, want DESC", opts.Sort[0].Direction)
				}
			},
		},
		{
			name:    "bbox filter",
			url:     "/search?filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4",
			wantErr: false,
			check: func(t *testing.T, opts *semantic.QueryOptions) {
				if len(opts.Filters) != 1 {
					t.Fatalf("Expected 1 filter, got %d", len(opts.Filters))
				}
				geoFilter, ok := opts.Filters[0].(*semantic.GeoBBoxFilter)
				if !ok {
					t.Fatal("Expected GeoBBoxFilter")
				}
				if geoFilter.Bounds.West != 4.8 {
					t.Errorf("West = %v, want 4.8", geoFilter.Bounds.West)
				}
				if geoFilter.Bounds.North != 52.4 {
					t.Errorf("North = %v, want 52.4", geoFilter.Bounds.North)
				}
			},
		},
		{
			name:    "invalid operator",
			url:     "/search?filter[creator][invalid]=test",
			wantErr: true,
		},
		{
			name:    "invalid page number",
			url:     "/search?page=abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			opts, err := parseQueryParams(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, opts)
			}
		})
	}
}

func TestParsePaginationDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/search", nil)
	opts, err := parseQueryParams(req)

	if err != nil {
		t.Fatalf("parseQueryParams() failed: %v", err)
	}

	if opts.Pagination == nil {
		t.Fatal("Expected pagination to be set")
	}

	// Check defaults
	if opts.Pagination.Page != 1 {
		t.Errorf("Default page = %v, want 1", opts.Pagination.Page)
	}

	if opts.Pagination.Size != 20 {
		t.Errorf("Default size = %v, want 20", opts.Pagination.Size)
	}
}

func TestParseJSONBody(t *testing.T) {
	t.Skip("JSON body parsing tested through integration tests")
	// Note: parseJSONBody is tested implicitly through the service POST handler tests
}

func TestParseGeoFilter(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		op      semantic.Operator
		values  []string
		wantErr bool
	}{
		{
			name:    "valid bbox",
			field:   "spatialCoverage",
			op:      semantic.OpBBox,
			values:  []string{"4.8,52.3,4.9,52.4"},
			wantErr: false,
		},
		{
			name:    "bbox with wrong number of coordinates",
			field:   "spatialCoverage",
			op:      semantic.OpBBox,
			values:  []string{"4.8,52.3"},
			wantErr: true,
		},
		{
			name:    "bbox with invalid coordinate",
			field:   "spatialCoverage",
			op:      semantic.OpBBox,
			values:  []string{"4.8,abc,4.9,52.4"},
			wantErr: true,
		},
		{
			name:    "unsupported geo operator",
			field:   "spatialCoverage",
			op:      semantic.OpWithin,
			values:  []string{"polygon"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGeoFilter(tt.field, tt.op, tt.values)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseGeoFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFromURLFieldName(t *testing.T) {
	tests := []struct {
		name     string
		urlField string
		want     string
	}{
		{
			name:     "dc_creator to dc:creator",
			urlField: "dc_creator",
			want:     "dc:creator",
		},
		{
			name:     "edm_dataProvider to edm:dataProvider",
			urlField: "edm_dataProvider",
			want:     "edm:dataProvider",
		},
		{
			name:     "schema_addressCountry to schema:addressCountry",
			urlField: "schema_addressCountry",
			want:     "schema:addressCountry",
		},
		{
			name:     "skos_prefLabel to skos:prefLabel",
			urlField: "skos_prefLabel",
			want:     "skos:prefLabel",
		},
		{
			name:     "field without namespace stays same",
			urlField: "simpleField",
			want:     "simpleField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fromURLFieldName(tt.urlField)
			if got != tt.want {
				t.Errorf("fromURLFieldName(%q) = %q, want %q", tt.urlField, got, tt.want)
			}
		})
	}
}

func TestToURLFieldName(t *testing.T) {
	tests := []struct {
		name          string
		internalField string
		want          string
	}{
		{
			name:          "dc:creator to dc_creator",
			internalField: "dc:creator",
			want:          "dc_creator",
		},
		{
			name:          "edm:dataProvider to edm_dataProvider",
			internalField: "edm:dataProvider",
			want:          "edm_dataProvider",
		},
		{
			name:          "schema:addressCountry to schema_addressCountry",
			internalField: "schema:addressCountry",
			want:          "schema_addressCountry",
		},
		{
			name:          "field without colon stays same",
			internalField: "simpleField",
			want:          "simpleField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toURLFieldName(tt.internalField)
			if got != tt.want {
				t.Errorf("toURLFieldName(%q) = %q, want %q", tt.internalField, got, tt.want)
			}
		})
	}
}

func TestFieldNameRoundTrip(t *testing.T) {
	tests := []string{
		"dc:creator",
		"edm:dataProvider",
		"schema:addressCountry",
		"skos:prefLabel",
		"ebucore:hasMimeType",
	}

	for _, internal := range tests {
		t.Run(internal, func(t *testing.T) {
			url := toURLFieldName(internal)
			back := fromURLFieldName(url)
			if back != internal {
				t.Errorf("Round trip failed: %q -> %q -> %q", internal, url, back)
			}
		})
	}
}

func TestParseFacetsWithNamespaces(t *testing.T) {
	// Test that facets with underscores in URL are converted to colons
	req := httptest.NewRequest("GET", "/search?facet[dc_creator]=10&facet[edm_dataProvider]=20", nil)
	opts, err := parseQueryParams(req)

	if err != nil {
		t.Fatalf("parseQueryParams() failed: %v", err)
	}

	if len(opts.Facets) != 2 {
		t.Fatalf("Expected 2 facets, got %d", len(opts.Facets))
	}

	// Check that field names were converted from URL format to internal format
	expectedFields := map[string]bool{
		"dc:creator":       false,
		"edm:dataProvider": false,
	}

	for _, facet := range opts.Facets {
		if _, ok := expectedFields[facet.Field]; ok {
			expectedFields[facet.Field] = true
		} else {
			t.Errorf("Unexpected facet field: %s", facet.Field)
		}
	}

	for field, found := range expectedFields {
		if !found {
			t.Errorf("Expected facet field %s not found", field)
		}
	}
}

func TestParseFiltersWithNamespaces(t *testing.T) {
	// Test that filters with underscores in URL are converted to colons
	req := httptest.NewRequest("GET", "/search?filter[dc_creator][eq]=Rembrandt&filter[edm_dataProvider][in]=Museum", nil)
	opts, err := parseQueryParams(req)

	if err != nil {
		t.Fatalf("parseQueryParams() failed: %v", err)
	}

	if len(opts.Filters) != 2 {
		t.Fatalf("Expected 2 filters, got %d", len(opts.Filters))
	}

	// Check that field names were converted from URL format to internal format
	expectedFields := map[string]bool{
		"dc:creator":       false,
		"edm:dataProvider": false,
	}

	for _, filter := range opts.Filters {
		if _, ok := expectedFields[filter.Field()]; ok {
			expectedFields[filter.Field()] = true
		} else {
			t.Errorf("Unexpected filter field: %s", filter.Field())
		}
	}

	for field, found := range expectedFields {
		if !found {
			t.Errorf("Expected filter field %s not found", field)
		}
	}
}

func TestParseQueryParams_Collapse(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantField string
		wantSize  int
		wantNil   bool
	}{
		{"no collapse", "query=test", "", 0, true},
		{"collapse field only", "collapse=edm_dataProvider", "edm:dataProvider", 0, false},
		{"collapse with size", "collapse=edm_dataProvider&collapse.size=3", "edm:dataProvider", 3, false},
		{"collapse with sort", "collapse=edm_dataProvider&collapse.sort=-dc_date", "edm:dataProvider", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/search?"+tt.query, nil)
			opts, err := parseQueryParams(r)
			if err != nil {
				t.Fatalf("parseQueryParams() error = %v", err)
			}
			if tt.wantNil {
				if opts.Collapse != nil {
					t.Error("expected nil Collapse")
				}
				return
			}
			if opts.Collapse == nil {
				t.Fatal("expected non-nil Collapse")
			}
			if opts.Collapse.Field != tt.wantField {
				t.Errorf("Collapse.Field = %q, want %q", opts.Collapse.Field, tt.wantField)
			}
			if tt.wantSize > 0 && opts.Collapse.Size != tt.wantSize {
				t.Errorf("Collapse.Size = %d, want %d", opts.Collapse.Size, tt.wantSize)
			}
		})
	}
}

func TestParseQueryParams_FacetBool(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?facetBool=and", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.FacetBoolType != semantic.FacetBoolAnd {
		t.Errorf("FacetBoolType = %q, want %q", opts.FacetBoolType, semantic.FacetBoolAnd)
	}
}

func TestParseQueryParams_FacetBool_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?facetBool=xor", nil)
	_, err := parseQueryParams(r)
	if err == nil {
		t.Error("expected error for invalid facetBool")
	}
}

func TestParseQueryParams_Peek(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?peek=dc_creator,dc_type", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !opts.Peek {
		t.Error("Peek should be true")
	}
	if len(opts.Facets) != 2 {
		t.Fatalf("expected 2 facets from peek, got %d", len(opts.Facets))
	}
	// Check fields (order may vary due to peek appending)
	fields := map[string]bool{}
	for _, f := range opts.Facets {
		fields[f.Field] = true
	}
	if !fields["dc:creator"] || !fields["dc:type"] {
		t.Errorf("expected dc:creator and dc:type facets, got %v", opts.Facets)
	}
	if opts.Pagination == nil || opts.Pagination.Size != 0 {
		t.Error("Peek should set pagination size to 0")
	}
}

func TestParseQueryParams_Debug(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?debug=query", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.Debug != "query" {
		t.Errorf("Debug = %q, want %q", opts.Debug, "query")
	}
}

func TestParseQueryParams_HiddenFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?hfilter[orgID][eq]=museum-x", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(opts.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(opts.Filters))
	}
	pf, ok := opts.Filters[0].(*semantic.PropertyFilter)
	if !ok {
		t.Fatal("expected PropertyFilter")
	}
	if !pf.Hidden {
		t.Error("filter should be hidden")
	}
	if pf.FieldName != "orgID" {
		t.Errorf("field = %q, want %q", pf.FieldName, "orgID")
	}
}

func TestParseQueryParams_ContextIndex(t *testing.T) {
	t.Run("single context index", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?contextIndex=other-org", nil)
		opts, err := parseQueryParams(r)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(opts.ContextIndices) != 1 {
			t.Fatalf("expected 1 context index, got %d", len(opts.ContextIndices))
		}
		if opts.ContextIndices[0] != "other-org" {
			t.Errorf("expected 'other-org', got %q", opts.ContextIndices[0])
		}
	})

	t.Run("multiple context indices", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?contextIndex=org-a&contextIndex=org-b", nil)
		opts, err := parseQueryParams(r)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(opts.ContextIndices) != 2 {
			t.Fatalf("expected 2 context indices, got %d", len(opts.ContextIndices))
		}
	})

	t.Run("empty context index is ignored", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?contextIndex=org-a&contextIndex=+&contextIndex=org-b", nil)
		opts, err := parseQueryParams(r)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(opts.ContextIndices) != 2 {
			t.Fatalf("expected 2 context indices (empty ignored), got %d", len(opts.ContextIndices))
		}
	})

	t.Run("no context index", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?query=test", nil)
		opts, err := parseQueryParams(r)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(opts.ContextIndices) != 0 {
			t.Errorf("expected 0 context indices, got %d", len(opts.ContextIndices))
		}
	})
}

func TestParseQueryParams_FacetCursor(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?facet[dc_creator]=50&facet[dc_creator].cursor=abc123", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(opts.Facets) == 0 {
		t.Fatal("expected at least 1 facet")
	}
	var found bool
	for _, f := range opts.Facets {
		if f.Field == "dc:creator" {
			found = true
			if f.Cursor != "abc123" {
				t.Errorf("Cursor = %q, want %q", f.Cursor, "abc123")
			}
			if f.Limit != 50 {
				t.Errorf("Limit = %d, want %d", f.Limit, 50)
			}
		}
	}
	if !found {
		t.Error("dc:creator facet not found")
	}
}
