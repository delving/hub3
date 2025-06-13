package fragments

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDetectSearchFields(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected FieldInfo
	}{
		{
			name:  "simple fielded query",
			query: "maria AND fields.dc_creator:lanting",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"fields.dc_creator"},
				Count:     1,
			},
		},
		{
			name:  "no fields",
			query: "just plain text search",
			expected: FieldInfo{
				HasFields: false,
				Fields:    []string{},
				Count:     0,
			},
		},
		{
			name:  "multiple different fields",
			query: "title:something AND author:someone",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"title", "author"},
				Count:     2,
			},
		},
		{
			name:  "complex field names",
			query: "fields.dc_creator:author AND meta.spec:value",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"fields.dc_creator", "meta.spec"},
				Count:     2,
			},
		},
		{
			name:  "repeated field names",
			query: "field1:value1 OR field2:value2 AND field1:value3",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"field1", "field2"},
				Count:     3,
			},
		},
		{
			name:  "colon without field pattern",
			query: "text with colon: but no field",
			expected: FieldInfo{
				HasFields: false,
				Fields:    []string{},
				Count:     0,
			},
		},
		{
			name:  "colon at end of text",
			query: "some text:",
			expected: FieldInfo{
				HasFields: false,
				Fields:    []string{},
				Count:     0,
			},
		},
		{
			name:  "field with numbers and underscores",
			query: "field_1:value AND field2_name:another",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"field_1", "field2_name"},
				Count:     2,
			},
		},
		{
			name:  "field starting with underscore",
			query: "_private_field:secret",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"_private_field"},
				Count:     1,
			},
		},
		{
			name:  "field with dots and complex structure",
			query: "nested.field.name:value AND simple:test",
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"nested.field.name", "simple"},
				Count:     2,
			},
		},
		{
			name:  "empty query",
			query: "",
			expected: FieldInfo{
				HasFields: false,
				Fields:    []string{},
				Count:     0,
			},
		},
		{
			name:  "query with quotes around fielded term",
			query: `title:"quoted value" AND author:unquoted`,
			expected: FieldInfo{
				HasFields: true,
				Fields:    []string{"title", "author"},
				Count:     2,
			},
		},
		{
			name:  "field starting with number should not match",
			query: "123field:value",
			expected: FieldInfo{
				HasFields: false,
				Fields:    []string{},
				Count:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectSearchFields(tt.query)

			// Use google/cmp for comparison with slice sorting option
			if diff := cmp.Diff(tt.expected, result, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("detectSearchFields() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHasSearchFields(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"has fields", "maria AND fields.dc_creator:lanting", true},
		{"no fields", "just plain text search", false},
		{"multiple fields", "title:something AND author:someone", true},
		{"colon without field", "text with colon: but no field", false},
		{"empty query", "", false},
		{"field with underscore", "field_name:value", true},
		{"field starting with number", "123field:value", false},
		{"complex nested field", "meta.nested.field:value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSearchFields(tt.query)
			if result != tt.expected {
				t.Errorf("hasSearchFields(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestGetSearchFields(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "single field",
			query:    "maria AND fields.dc_creator:lanting",
			expected: []string{"fields.dc_creator"},
		},
		{
			name:     "no fields",
			query:    "just plain text search",
			expected: []string{},
		},
		{
			name:     "multiple fields",
			query:    "title:something AND author:someone",
			expected: []string{"title", "author"},
		},
		{
			name:     "repeated fields",
			query:    "field1:value1 AND field1:value2",
			expected: []string{"field1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSearchFields(tt.query)

			// Use google/cmp with slice sorting for comparison
			if diff := cmp.Diff(tt.expected, result, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("getSearchFields() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestShouldRemoveFieldsParam(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"query with fields should remove param", "maria AND fields.dc_creator:lanting", true},
		{"query without fields should keep param", "just plain text search", false},
		{"mixed query should remove param", "text AND field:value", true},
		{"empty query should keep param", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRemoveFieldsParam(tt.query)
			if result != tt.expected {
				t.Errorf("shouldRemoveFieldsParam(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkDetectSearchFields(b *testing.B) {
	query := "maria AND fields.dc_creator:lanting AND title:something OR author:someone"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectSearchFields(query)
	}
}

func BenchmarkHasSearchFields(b *testing.B) {
	query := "maria AND fields.dc_creator:lanting AND title:something OR author:someone"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasSearchFields(query)
	}
}
