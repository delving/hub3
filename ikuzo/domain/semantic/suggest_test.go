package semantic

import (
	"strings"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"identical strings", "kitten", "kitten", 0},
		{"single substitution", "kitten", "sitten", 1},
		{"classic example", "kitten", "sitting", 3},
		{"empty first", "", "abc", 3},
		{"empty second", "abc", "", 3},
		{"both empty", "", "", 0},
		{"single char difference", "cat", "car", 1},
		{"transposition", "ab", "ba", 2},
		{"completely different", "abc", "xyz", 3},
		{"insertion", "abc", "abcd", 1},
		{"deletion", "abcd", "abc", 1},
		{"prefix", "dc_title", "dc_tilte", 2},
		{"case sensitive", "Title", "title", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevenshteinDistance(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLevenshteinDistance_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"kitten", "sitting"},
		{"abc", "xyz"},
		{"", "hello"},
		{"dc_title", "dc_tilte"},
	}

	for _, pair := range pairs {
		d1 := LevenshteinDistance(pair[0], pair[1])
		d2 := LevenshteinDistance(pair[1], pair[0])

		if d1 != d2 {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, but reverse = %d", pair[0], pair[1], d1, d2)
		}
	}
}

func TestSuggestFields(t *testing.T) {
	knownFields := []string{"dc_title", "dc_creator", "dc_subject", "dc_date", "dc_type", "edm_isShownAt"}

	tests := []struct {
		name        string
		unknown     string
		maxDistance int
		want        []string
	}{
		{
			name:        "close typo",
			unknown:     "dc_tilte",
			maxDistance:  3,
			want:        []string{"dc_title", "dc_date", "dc_type"},
		},
		{
			name:        "no close matches",
			unknown:     "zzzzzzzzz",
			maxDistance:  3,
			want:        nil,
		},
		{
			name:        "multiple close matches",
			unknown:     "dc_tite",
			maxDistance:  3,
			want:        []string{"dc_title", "dc_date", "dc_type"},
		},
		{
			name:        "exact match excluded",
			unknown:     "dc_title",
			maxDistance:  3,
			want:        []string{"dc_date", "dc_type"},
		},
		{
			name:        "case sensitive",
			unknown:     "DC_Title",
			maxDistance:  3,
			want:        []string{"dc_title"},
		},
		{
			name:        "max distance 1",
			unknown:     "dc_titl",
			maxDistance:  1,
			want:        []string{"dc_title"},
		},
		{
			name:        "max distance 0 returns nil",
			unknown:     "dc_title",
			maxDistance:  0,
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestFields(tt.unknown, knownFields, tt.maxDistance)
			if tt.want == nil {
				if got != nil {
					t.Errorf("SuggestFields(%q) = %v, want nil", tt.unknown, got)
				}

				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("SuggestFields(%q) returned %d results %v, want %d results %v",
					tt.unknown, len(got), got, len(tt.want), tt.want)
				return
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("SuggestFields(%q)[%d] = %q, want %q", tt.unknown, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSuggestFields_SortOrder(t *testing.T) {
	// Fields with same distance should be sorted alphabetically
	knownFields := []string{"zebra", "alpha", "delta"}
	got := SuggestFields("zelta", knownFields, 3)

	if len(got) < 2 {
		t.Fatalf("Expected at least 2 suggestions, got %d: %v", len(got), got)
	}

	// Verify distance ordering: closest first
	for i := 1; i < len(got); i++ {
		d1 := LevenshteinDistance("zelta", got[i-1])
		d2 := LevenshteinDistance("zelta", got[i])

		if d1 > d2 {
			t.Errorf("Results not sorted by distance: %q (d=%d) before %q (d=%d)", got[i-1], d1, got[i], d2)
		}

		// If same distance, should be alphabetical
		if d1 == d2 && got[i-1] > got[i] {
			t.Errorf("Tied results not sorted alphabetically: %q before %q", got[i-1], got[i])
		}
	}
}

func TestFormatSuggestion(t *testing.T) {
	tests := []struct {
		name        string
		unknown     string
		suggestions []string
		want        string
	}{
		{
			name:        "no suggestions",
			unknown:     "foo",
			suggestions: nil,
			want:        "",
		},
		{
			name:        "empty slice",
			unknown:     "foo",
			suggestions: []string{},
			want:        "",
		},
		{
			name:        "single suggestion",
			unknown:     "dc_tilte",
			suggestions: []string{"dc_title"},
			want:        "; did you mean 'dc_title'?",
		},
		{
			name:        "multiple suggestions",
			unknown:     "dc_tite",
			suggestions: []string{"dc_title", "dc_date"},
			want:        "; did you mean one of: 'dc_title', 'dc_date'?",
		},
		{
			name:        "three suggestions",
			unknown:     "field",
			suggestions: []string{"field1", "field2", "field3"},
			want:        "; did you mean one of: 'field1', 'field2', 'field3'?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSuggestion(tt.unknown, tt.suggestions)
			if got != tt.want {
				t.Errorf("FormatSuggestion(%q, %v) = %q, want %q", tt.unknown, tt.suggestions, got, tt.want)
			}
		})
	}
}

func TestFilterableFields(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	rc.AddFilter(NewFilterConfig("title", "schema:name", "Title"))
	rc.AddFilter(NewFilterConfig("creator", "schema:creator", "Creator"))
	rc.AddFilter(NewFilterConfig("date", "schema:dateCreated", "Date"))

	got := rc.FilterableFields()

	if len(got) != 3 {
		t.Fatalf("FilterableFields() returned %d fields, want 3", len(got))
	}

	want := []string{"title", "creator", "date"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("FilterableFields()[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestFacetableFields(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	rc.AddFacet(NewFacetConfig("type", "rdf:type", "Type", "enum"))
	rc.AddFacet(NewFacetConfig("collection", "schema:isPartOf", "Collection", "enum"))

	got := rc.FacetableFields()

	if len(got) != 2 {
		t.Fatalf("FacetableFields() returned %d fields, want 2", len(got))
	}

	want := []string{"type", "collection"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("FacetableFields()[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestFilterableFields_Empty(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	got := rc.FilterableFields()

	if len(got) != 0 {
		t.Errorf("FilterableFields() on empty config returned %d fields, want 0", len(got))
	}
}

func TestFacetableFields_Empty(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	got := rc.FacetableFields()

	if len(got) != 0 {
		t.Errorf("FacetableFields() on empty config returned %d fields, want 0", len(got))
	}
}

func TestValidateFilter_WithSuggestions(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	rc.AddFilter(NewFilterConfig("dc_title", "dc:title", "Title").WithOperators(OpEqual))
	rc.AddFilter(NewFilterConfig("dc_creator", "dc:creator", "Creator").WithOperators(OpEqual))
	rc.AddFilter(NewFilterConfig("dc_subject", "dc:subject", "Subject").WithOperators(OpEqual))

	tests := []struct {
		name          string
		fieldName     string
		wantErr       bool
		wantSuggest   string // substring expected in error message
		noSuggest     bool   // true if no suggestion expected
	}{
		{
			name:        "typo suggests correct field",
			fieldName:   "dc_tilte",
			wantErr:     true,
			wantSuggest: "did you mean 'dc_title'",
		},
		{
			name:        "close typo in creator",
			fieldName:   "dc_cretor",
			wantErr:     true,
			wantSuggest: "did you mean 'dc_creator'",
		},
		{
			name:      "completely unknown field no suggestion",
			fieldName: "zzzzzzzzzzz",
			wantErr:   true,
			noSuggest: true,
		},
		{
			name:      "valid field no error",
			fieldName: "dc_title",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &PropertyFilter{
				FieldName:    tt.fieldName,
				OperatorType: OpEqual,
				Value:        "test",
			}

			err := rc.ValidateFilter(filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantSuggest != "" {
				if !strings.Contains(err.Error(), tt.wantSuggest) {
					t.Errorf("error %q does not contain suggestion %q", err.Error(), tt.wantSuggest)
				}
			}

			if err != nil && tt.noSuggest {
				if strings.Contains(err.Error(), "did you mean") {
					t.Errorf("error %q should not contain suggestion", err.Error())
				}
			}
		})
	}
}

func TestValidateQueryOptions_FacetWithSuggestions(t *testing.T) {
	rc := NewResourceConfig("schema:CreativeWork", "Creative Work")
	rc.AddFacet(NewFacetConfig("dc_type", "dc:type", "Type", "enum"))
	rc.AddFacet(NewFacetConfig("dc_subject", "dc:subject", "Subject", "enum"))

	opts := &QueryOptions{
		Facets: []FacetRequest{
			{Field: "dc_tyep"}, // typo for dc_type
		},
	}

	err := rc.ValidateQueryOptions(opts)
	if err == nil {
		t.Fatal("Expected error for unknown facet field")
	}

	if !strings.Contains(err.Error(), "did you mean 'dc_type'") {
		t.Errorf("error %q does not contain suggestion for 'dc_type'", err.Error())
	}
}
