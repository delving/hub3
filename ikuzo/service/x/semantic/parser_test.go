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
				if opts.Facets[0].Limit != 10 {
					t.Errorf("First facet limit = %v, want 10", opts.Facets[0].Limit)
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
