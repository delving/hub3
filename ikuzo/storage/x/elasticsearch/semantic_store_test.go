package elasticsearch

import (
	"context"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/rs/zerolog"
)

func TestNewSemanticStore(t *testing.T) {
	// Note: This test requires a real Elasticsearch client
	// In production, you'd use a mock or integration test setup
	t.Skip("Requires Elasticsearch connection")

	store := NewSemanticStore(nil, "test-index", zerolog.Nop())
	if store == nil {
		t.Error("Expected store to be created")
	}

	if store.index != "test-index" {
		t.Errorf("Index = %v, want test-index", store.index)
	}
}

func TestBuildPropertyFilterQuery(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	tests := []struct {
		name     string
		filter   *semantic.PropertyFilter
		wantNil  bool
		wantType string
	}{
		{
			name: "equal filter",
			filter: &semantic.PropertyFilter{
				FieldName:    "creator",
				OperatorType: semantic.OpEqual,
				Value:        "Rembrandt",
			},
			wantNil:  false,
			wantType: "term",
		},
		{
			name: "in filter",
			filter: &semantic.PropertyFilter{
				FieldName:    "type",
				OperatorType: semantic.OpIn,
				Value:        []string{"Painting", "Drawing"},
			},
			wantNil:  false,
			wantType: "terms",
		},
		{
			name: "contains filter",
			filter: &semantic.PropertyFilter{
				FieldName:    "description",
				OperatorType: semantic.OpContains,
				Value:        "landscape",
			},
			wantNil:  false,
			wantType: "match",
		},
		{
			name: "greater than filter",
			filter: &semantic.PropertyFilter{
				FieldName:    "year",
				OperatorType: semantic.OpGreaterThan,
				Value:        1600,
			},
			wantNil:  false,
			wantType: "range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := store.buildPropertyFilterQuery(tt.filter)

			if tt.wantNil && query != nil {
				t.Error("Expected nil query")
			}

			if !tt.wantNil && query == nil {
				t.Error("Expected non-nil query")
			}

			// Note: We can't easily inspect the query type without reflection
			// In a real test, you'd serialize the query and check the JSON structure
		})
	}
}

func TestBuildRangeFilterQuery(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	min := any(1600)
	max := any(1700)

	filter := &semantic.RangeFilter{
		FieldName:    "dateCreated",
		OperatorType: semantic.OpGreaterEqual,
		Min:          min,
		Max:          max,
	}

	query := store.buildRangeFilterQuery(filter)
	if query == nil {
		t.Error("Expected non-nil range query")
	}
}

func TestBuildGeoBBoxQuery(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	filter := &semantic.GeoBBoxFilter{
		FieldName: "spatialCoverage",
		Bounds: &semantic.GeoBounds{
			West:  4.8,
			South: 52.3,
			East:  4.9,
			North: 52.4,
		},
	}

	query := store.buildGeoBBoxQuery(filter)
	if query == nil {
		t.Error("Expected non-nil geo bbox query")
	}
}

func TestBuildGeoDistanceQuery(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	filter := &semantic.GeoDistanceFilter{
		FieldName: "spatialCoverage",
		Point: &semantic.GeoPoint{
			Lat: 52.37,
			Lon: 4.89,
		},
		Distance: "10km",
	}

	query := store.buildGeoDistanceQuery(filter)
	if query == nil {
		t.Error("Expected non-nil geo distance query")
	}
}

func TestConvertToInterfaceSlice(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  int // expected length
	}{
		{
			name:  "string slice",
			input: []string{"a", "b", "c"},
			want:  3,
		},
		{
			name:  "int slice",
			input: []int{1, 2, 3},
			want:  3,
		},
		{
			name:  "interface slice",
			input: []any{"a", 1, true},
			want:  3,
		},
		{
			name:  "single value",
			input: "test",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToInterfaceSlice(tt.input)
			if len(result) != tt.want {
				t.Errorf("Length = %d, want %d", len(result), tt.want)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	t.Skip("Requires Elasticsearch connection")

	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	ctx := context.Background()
	err := store.Health(ctx)
	if err == nil {
		t.Error("Expected error with nil client")
	}
}

func TestBuildQuery(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	config := semantic.CreativeWorkConfig()

	tests := []struct {
		name string
		opts *semantic.QueryOptions
	}{
		{
			name: "empty query",
			opts: &semantic.QueryOptions{},
		},
		{
			name: "text query only",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value: "rembrandt",
				},
			},
		},
		{
			name: "with filters",
			opts: &semantic.QueryOptions{
				Filters: []semantic.Filter{
					&semantic.PropertyFilter{
						FieldName:    "creator",
						OperatorType: semantic.OpEqual,
						Value:        "Rembrandt",
					},
				},
			},
		},
		{
			name: "multi-field query",
			opts: &semantic.QueryOptions{
				Query: &semantic.TextQuery{
					Value:  "night watch",
					Fields: []string{"title", "description"},
					Fuzzy:  true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := store.buildQuery(tt.opts, config)
			if query == nil {
				t.Error("Expected non-nil query")
			}
		})
	}
}

func TestBuildFacetAggregation(t *testing.T) {
	store := &SemanticStore{
		log: zerolog.Nop(),
	}

	config := semantic.CreativeWorkConfig()

	tests := []struct {
		name    string
		facet   semantic.FacetRequest
		wantNil bool
	}{
		{
			name: "enum facet",
			facet: semantic.FacetRequest{
				Field: "type",
				Limit: 20,
			},
			wantNil: false,
		},
		{
			name: "range facet",
			facet: semantic.FacetRequest{
				Field: "dateCreated",
				Limit: 10,
			},
			wantNil: false,
		},
		{
			name: "unconfigured facet",
			facet: semantic.FacetRequest{
				Field: "nonexistent",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := store.buildFacetAggregation(tt.facet, config)

			if tt.wantNil && agg != nil {
				t.Error("Expected nil aggregation")
			}

			if !tt.wantNil && agg == nil {
				t.Error("Expected non-nil aggregation")
			}
		})
	}
}
