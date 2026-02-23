package semantic

import (
	"testing"
)

func TestNewFilterConfig(t *testing.T) {
	fc := NewFilterConfig("title", "schema:name", "Title")

	if fc.Field != "title" {
		t.Errorf("Field = %v, want title", fc.Field)
	}

	if fc.PropertyURI != "schema:name" {
		t.Errorf("PropertyURI = %v, want schema:name", fc.PropertyURI)
	}

	if fc.Label != "Title" {
		t.Errorf("Label = %v, want Title", fc.Label)
	}

	if len(fc.AllowedOperators) != 1 || fc.AllowedOperators[0] != OpEqual {
		t.Errorf("AllowedOperators = %v, want [OpEqual]", fc.AllowedOperators)
	}

	if !fc.AllowedInQuery || !fc.AllowedInPost {
		t.Error("Expected filter to be allowed in both query and post")
	}

	if fc.ValueType != "string" {
		t.Errorf("ValueType = %v, want string", fc.ValueType)
	}
}

func TestFilterConfig_WithMethods(t *testing.T) {
	fc := NewFilterConfig("title", "schema:name", "Title").
		WithOperators(OpEqual, OpContains, OpStartsWith).
		WithDescription("Filter by title").
		WithValueType("string").
		WithMultiValue(true).
		WithQueryAccess(false).
		WithExamples("rembrandt", "vermeer")

	if len(fc.AllowedOperators) != 3 {
		t.Errorf("Expected 3 operators, got %d", len(fc.AllowedOperators))
	}

	if fc.Description != "Filter by title" {
		t.Errorf("Description = %v, want 'Filter by title'", fc.Description)
	}

	if !fc.MultiValue {
		t.Error("Expected MultiValue to be true")
	}

	if fc.AllowedInQuery {
		t.Error("Expected AllowedInQuery to be false")
	}

	if len(fc.Examples) != 2 {
		t.Errorf("Expected 2 examples, got %d", len(fc.Examples))
	}
}

func TestFilterConfig_IsOperatorAllowed(t *testing.T) {
	fc := NewFilterConfig("title", "schema:name", "Title").
		WithOperators(OpEqual, OpContains)

	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"allowed eq", OpEqual, true},
		{"allowed contains", OpContains, true},
		{"not allowed gt", OpGreaterThan, false},
		{"not allowed in", OpIn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fc.IsOperatorAllowed(tt.operator); got != tt.want {
				t.Errorf("IsOperatorAllowed(%v) = %v, want %v", tt.operator, got, tt.want)
			}
		})
	}
}

func TestFilterConfig_Validate(t *testing.T) {
	fc := NewFilterConfig("title", "schema:name", "Title").
		WithOperators(OpEqual, OpContains)

	tests := []struct {
		name    string
		filter  Filter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: false,
		},
		{
			name: "wrong field",
			filter: &PropertyFilter{
				FieldName:    "creator",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpGreaterThan,
				Value:        "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fc.Validate(tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFacetConfig(t *testing.T) {
	fc := NewFacetConfig("type", "rdf:type", "Type", "enum")

	if fc.Field != "type" {
		t.Errorf("Field = %v, want type", fc.Field)
	}

	if fc.FacetType != "enum" {
		t.Errorf("FacetType = %v, want enum", fc.FacetType)
	}

	if fc.Size != 10 {
		t.Errorf("Size = %v, want 10", fc.Size)
	}

	if fc.SortBy != "count" || fc.SortOrder != "desc" {
		t.Errorf("Sort = %v %v, want count desc", fc.SortBy, fc.SortOrder)
	}
}

func TestFacetConfig_WithMethods(t *testing.T) {
	ranges := []FacetRange{
		{Key: "0-100", From: 0, To: 100, Label: "0-100"},
		{Key: "100+", From: 100, Label: "100+"},
	}

	intervals := []DateInterval{
		{Interval: "year", Format: "yyyy"},
	}

	fc := NewFacetConfig("year", "schema:dateCreated", "Year", "range").
		WithDescription("Filter by year").
		WithSize(20).
		WithMinDocCount(5).
		WithSort("term", "asc").
		WithRanges(ranges...).
		WithDateIntervals(intervals...).
		WithMissing(true).
		WithMetadata("order", 1)

	if fc.Description != "Filter by year" {
		t.Errorf("Description = %v, want 'Filter by year'", fc.Description)
	}

	if fc.Size != 20 {
		t.Errorf("Size = %v, want 20", fc.Size)
	}

	if fc.MinDocCount != 5 {
		t.Errorf("MinDocCount = %v, want 5", fc.MinDocCount)
	}

	if fc.SortBy != "term" || fc.SortOrder != "asc" {
		t.Errorf("Sort = %v %v, want term asc", fc.SortBy, fc.SortOrder)
	}

	if len(fc.Ranges) != 2 {
		t.Errorf("Expected 2 ranges, got %d", len(fc.Ranges))
	}

	if len(fc.DateIntervals) != 1 {
		t.Errorf("Expected 1 date interval, got %d", len(fc.DateIntervals))
	}

	if !fc.Missing {
		t.Error("Expected Missing to be true")
	}

	if fc.Metadata["order"] != 1 {
		t.Errorf("Metadata[order] = %v, want 1", fc.Metadata["order"])
	}
}

func TestNewResourceConfig(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")

	if rc.ResourceType != "schema:CreativeWork" {
		t.Errorf("ResourceType = %v, want schema:CreativeWork", rc.ResourceType)
	}

	if rc.DefaultSize != 20 {
		t.Errorf("DefaultSize = %v, want 20", rc.DefaultSize)
	}

	if rc.MaxSize != 100 {
		t.Errorf("MaxSize = %v, want 100", rc.MaxSize)
	}

	if !rc.AllowFullText {
		t.Error("Expected AllowFullText to be true")
	}
}

func TestResourceConfig_AddFilterAndFacet(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")

	filter := NewFilterConfig("title", "schema:name", "Title")
	facet := NewFacetConfig("type", "rdf:type", "Type", "enum")

	rc.AddFilter(filter).AddFacet(facet)

	if len(rc.Filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(rc.Filters))
	}

	if len(rc.Facets) != 1 {
		t.Errorf("Expected 1 facet, got %d", len(rc.Facets))
	}
}

func TestResourceConfig_GetFilterConfig(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	filter := NewFilterConfig("title", "schema:name", "Title")
	rc.AddFilter(filter)

	// Test existing filter
	got, ok := rc.GetFilterConfig("title")
	if !ok {
		t.Error("Expected to find filter config for 'title'")
	}
	if got.Field != "title" {
		t.Errorf("Field = %v, want title", got.Field)
	}

	// Test non-existent filter
	_, ok = rc.GetFilterConfig("nonexistent")
	if ok {
		t.Error("Expected not to find filter config for 'nonexistent'")
	}
}

func TestResourceConfig_IsFieldFilterable(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	filter := NewFilterConfig("title", "schema:name", "Title")
	rc.AddFilter(filter)

	if !rc.IsFieldFilterable("title") {
		t.Error("Expected 'title' to be filterable")
	}

	if rc.IsFieldFilterable("nonexistent") {
		t.Error("Expected 'nonexistent' not to be filterable")
	}
}

func TestResourceConfig_ValidateFilter(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	filter := NewFilterConfig("title", "schema:name", "Title").
		WithOperators(OpEqual, OpContains)
	rc.AddFilter(filter)

	tests := []struct {
		name    string
		filter  Filter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: false,
		},
		{
			name: "non-filterable field",
			filter: &PropertyFilter{
				FieldName:    "nonexistent",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpGreaterThan,
				Value:        "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rc.ValidateFilter(tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceConfig_ValidateQueryOptions(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work").
		WithMaxSize(50)

	titleFilter := NewFilterConfig("title", "schema:name", "Title").
		WithOperators(OpEqual)
	rc.AddFilter(titleFilter)

	typeFacet := NewFacetConfig("type", "rdf:type", "Type", "enum")
	rc.AddFacet(typeFacet)

	tests := []struct {
		name    string
		opts    *QueryOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: &QueryOptions{
				Filters: []Filter{
					&PropertyFilter{
						FieldName:    "title",
						OperatorType: OpEqual,
						Value:        "test",
					},
				},
				Facets: []FacetRequest{
					{Field: "type", Limit: 10},
				},
				Pagination: &Pagination{
					Page: 1,
					Size: 20,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid filter field",
			opts: &QueryOptions{
				Filters: []Filter{
					&PropertyFilter{
						FieldName:    "nonexistent",
						OperatorType: OpEqual,
						Value:        "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid facet field",
			opts: &QueryOptions{
				Facets: []FacetRequest{
					{Field: "nonexistent", Limit: 10},
				},
			},
			wantErr: true,
		},
		{
			name: "page size too large",
			opts: &QueryOptions{
				Pagination: &Pagination{
					Page: 1,
					Size: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "page size too small",
			opts: &QueryOptions{
				Pagination: &Pagination{
					Page: 1,
					Size: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rc.ValidateQueryOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateQueryOptions_Collapse(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")

	tests := []struct {
		name    string
		opts    *QueryOptions
		wantErr bool
	}{
		{
			name:    "nil collapse is valid",
			opts:    &QueryOptions{},
			wantErr: false,
		},
		{
			name: "valid collapse",
			opts: &QueryOptions{
				Collapse: &CollapseOptions{Field: "edm:dataProvider"},
			},
			wantErr: false,
		},
		{
			name: "empty collapse field is invalid",
			opts: &QueryOptions{
				Collapse: &CollapseOptions{Field: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rc.ValidateQueryOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateQueryOptions_FacetBoolType(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")

	tests := []struct {
		name    string
		fbt     FacetBoolType
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"or is valid", FacetBoolOr, false},
		{"and is valid", FacetBoolAnd, false},
		{"xor is invalid", FacetBoolType("xor"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{FacetBoolType: tt.fbt}
			err := rc.ValidateQueryOptions(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigRegistry(t *testing.T) {
	registry := NewConfigRegistry()

	// Test registration
	config1 := NewResourceConfig("schema:CreativeWork", "Creative Work")
	config2 := NewResourceConfig("schema:Person", "Person")

	registry.Register(config1)
	registry.Register(config2)

	// Test retrieval
	got, ok := registry.Get("schema:CreativeWork")
	if !ok {
		t.Error("Expected to find CreativeWork config")
	}
	if got.ResourceType != "schema:CreativeWork" {
		t.Errorf("ResourceType = %v, want schema:CreativeWork", got.ResourceType)
	}

	// Test non-existent
	_, ok = registry.Get("schema:NonExistent")
	if ok {
		t.Error("Expected not to find NonExistent config")
	}

	// Test list resource types
	types := registry.ListResourceTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 resource types, got %d", len(types))
	}

	// Test get default
	defaultConfig := registry.GetDefault()
	if defaultConfig == nil {
		t.Error("Expected to get a default config")
	}
}
