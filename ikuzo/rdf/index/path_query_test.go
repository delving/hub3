package index

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestParsePathStringExpanded(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected *PathQuery
		wantErr  bool
	}{
		// Original test cases
		// ...

		// Additional test cases
		{
			name: "Path with wildcard class",
			path: "[]crm:P102_has_title->[crm:E35_Title]",
			expected: &PathQuery{
				Resource: "[]crm:P102_has_title->[crm:E35_Title]",
			},
			wantErr: false,
		},
		{
			name: "Path with wildcard class and filter",
			path: "[]crm:P102_has_title | [crm:E35_Title]crm:P2_has_type=someValue",
			expected: &PathQuery{
				Resource: "[]crm:P102_has_title",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"someValue"},
				},
			},
			wantErr: false,
		},
		{
			name: "Path with consecutive arrows",
			path: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
				Value:    "[crm:E35_Title]crm:P190_has_symbolic_content",
			},
			wantErr: false,
		},
		{
			name:     "Path with unclosed bracket",
			path:     "[lrm:F3_Manifestationcrm:P102_has_title",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Path with different filter paths",
			path:     "[class1]pred1 | [filter1]path1=value1 | [filter2]path2=value2",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Path with multiple value paths",
			path:     "[class1]pred1 | [filter]path=value | [value1]path1 | [value2]path2",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Path with special characters in values",
			path: "[class]pred | [filter]path=value with spaces | [filter]path=value/with/slashes",
			expected: &PathQuery{
				Resource: "[class]pred",
				Filter: FilterPath{
					Path:   "[filter]path",
					Values: []string{"value with spaces", "value/with/slashes"},
				},
			},
			wantErr: false,
		},
		{
			name: "Path with pipe characters in filter values",
			path: "[class]pred | [filter]path=value|with|pipes",
			expected: &PathQuery{
				Resource: "[class]pred",
				Filter: FilterPath{
					Path:   "[filter]path",
					Values: []string{"value|with|pipes"},
				},
			},
			wantErr: false,
		},
		{
			name: "Complex path with multiple segments",
			path: "[class1]pred1->[class2]pred2->[class3]pred3 | [filter]path=value | [value]path",
			expected: &PathQuery{
				Resource: "[class1]pred1->[class2]pred2->[class3]pred3",
				Filter: FilterPath{
					Path:   "[filter]path",
					Values: []string{"value"},
				},
				Value: "[value]path",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePathString(tt.path)

			// Check if we expected an error
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePathString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error or nil result, don't check further
			if tt.wantErr || tt.expected == nil {
				if got != nil && tt.expected == nil {
					t.Errorf("Expected nil result, got %+v", got)
				}
				return
			}

			// Check resource path
			if got.Resource != tt.expected.Resource {
				t.Errorf("Resource = %v, want %v", got.Resource, tt.expected.Resource)
			}

			// Check filter path and values
			if got.Filter.Path != tt.expected.Filter.Path {
				t.Errorf("Filter.Path = %v, want %v", got.Filter.Path, tt.expected.Filter.Path)
			}

			if !compareStringSlices(got.Filter.Values, tt.expected.Filter.Values) {
				t.Errorf("Filter.Values = %v, want %v", got.Filter.Values, tt.expected.Filter.Values)
			}

			// Check value path
			if got.Value != tt.expected.Value {
				t.Errorf("Value = %v, want %v", got.Value, tt.expected.Value)
			}
		})
	}
}

func TestPathQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    *PathQuery
		expected string
	}{
		{
			name: "Simple path",
			query: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]",
			},
			expected: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]",
		},
		{
			name: "Path with filter",
			query: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"value1"},
				},
			},
			expected: "[lrm:F3_Manifestation]crm:P102_has_title | [crm:E35_Title]crm:P2_has_type=value1",
		},
		{
			name: "Path with multiple filter values",
			query: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"value1", "value2"},
				},
			},
			expected: "[lrm:F3_Manifestation]crm:P102_has_title | [crm:E35_Title]crm:P2_has_type=value1 | [crm:E35_Title]crm:P2_has_type=value2",
		},
		{
			name: "Path with filter and value",
			query: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"value1"},
				},
				Value: "[crm:E35_Title]crm:P190_has_symbolic_content",
			},
			expected: "[lrm:F3_Manifestation]crm:P102_has_title | [crm:E35_Title]crm:P2_has_type=value1 | [crm:E35_Title]crm:P190_has_symbolic_content",
		},
		{
			name:     "Nil query",
			query:    nil,
			expected: "",
		},
		{
			name: "Empty resource",
			query: &PathQuery{
				Resource: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.query.String()
			if got != tt.expected {
				t.Errorf("PathQuery.String() = %v, want %v", got, tt.expected)
			}

			// Roundtrip test - parse the string back to a PathQuery if not nil or empty
			if tt.query != nil && tt.query.Resource != "" {
				parsed, err := ParsePathString(got)
				if err != nil {
					t.Errorf("Failed to parse generated string: %v", err)
					return
				}

				// Compare the original query with the parsed one
				if parsed.Resource != tt.query.Resource {
					t.Errorf("Roundtrip Resource = %v, want %v", parsed.Resource, tt.query.Resource)
				}

				if parsed.Filter.Path != tt.query.Filter.Path {
					t.Errorf("Roundtrip Filter.Path = %v, want %v", parsed.Filter.Path, tt.query.Filter.Path)
				}

				if !compareStringSlices(parsed.Filter.Values, tt.query.Filter.Values) {
					t.Errorf("Roundtrip Filter.Values = %v, want %v", parsed.Filter.Values, tt.query.Filter.Values)
				}

				if parsed.Value != tt.query.Value {
					t.Errorf("Roundtrip Value = %v, want %v", parsed.Value, tt.query.Value)
				}
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []PathEntry
		wantErr  bool
	}{
		{
			name: "Simple path with class and predicate",
			path: "[crm:E35_Title]crm:P190_has_symbolic_content",
			expected: []PathEntry{
				{
					Class:     "crm:E35_Title",
					Predicate: "crm:P190_has_symbolic_content",
					IsLeaf:    true,
				},
			},
			wantErr: false,
		},
		{
			name: "Path with multiple segments",
			path: "[class1]pred1->[class2]pred2",
			expected: []PathEntry{
				{
					Class:     "class1",
					Predicate: "pred1",
					IsLeaf:    false,
				},
				{
					Class:     "class2",
					Predicate: "pred2",
					IsLeaf:    true,
				},
			},
			wantErr: false,
		},
		{
			name: "Path with empty class (wildcard)",
			path: "[]pred1",
			expected: []PathEntry{
				{
					Class:     "",
					Predicate: "pred1",
					IsLeaf:    true,
				},
			},
			wantErr: false,
		},
		{
			name:     "Path with unclosed bracket",
			path:     "[class1pred1",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePath(tt.path, false)

			// Check if we expected an error
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected nil, don't check further
			if tt.expected == nil {
				if got != nil {
					t.Errorf("Expected nil result, got %+v", got)
				}
				return
			}

			// Check the number of entries
			if len(got) != len(tt.expected) {
				t.Errorf("Got %d entries, expected %d", len(got), len(tt.expected))
				return
			}

			// Check each entry
			for i, entry := range got {
				expectedEntry := tt.expected[i]
				if entry.Class != expectedEntry.Class {
					t.Errorf("Entry %d Class = %v, want %v", i, entry.Class, expectedEntry.Class)
				}
				if entry.Predicate != expectedEntry.Predicate {
					t.Errorf("Entry %d Predicate = %v, want %v", i, entry.Predicate, expectedEntry.Predicate)
				}
				if entry.IsLeaf != expectedEntry.IsLeaf {
					t.Errorf("Entry [%s]%s IsLeaf = %v, want %v", entry.Class, entry.Predicate, entry.IsLeaf, expectedEntry.IsLeaf)
				}
			}
		})
	}
}

func TestParsePathString(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected *PathQuery
		wantErr  bool
	}{
		{
			name: "Simple path",
			path: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
				Value:    "[crm:E35_Title]crm:P190_has_symbolic_content",
			},
			wantErr: false,
		},
		{
			name: "Path with filter",
			path: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h"},
				},
			},
			wantErr: false,
		},
		{
			name: "Path with filter and value",
			path: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h | [crm:E35_Title]crm:P190_has_symbolic_content",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]",
				Filter: FilterPath{
					Path:   "[crm:E35_Title]crm:P2_has_type",
					Values: []string{"https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h"},
				},
				Value: "[crm:E35_Title]crm:P190_has_symbolic_content",
			},
			wantErr: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Path with multiple values for same filter",
			path: "[lrm:F3_Manifestation]crm:P129_is_about | [rdfs:Resource]rdfs:label=Value1 | [rdfs:Resource]rdfs:label=Value2",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P129_is_about",
				Filter: FilterPath{
					Path:   "[rdfs:Resource]rdfs:label",
					Values: []string{"Value1", "Value2"},
				},
			},
			wantErr: false,
		},
		{
			name: "Path with multiple values for same filter",
			path: "[lrm:F3_Manifestation]crm:P129_is_about | [rdfs:Resource]rdfs:label=Value1 | [rdfs:Resource]rdfs:label=Value2 | [crm:E35_Title]crm:P190_has_symbolic_content",
			expected: &PathQuery{
				Resource: "[lrm:F3_Manifestation]crm:P129_is_about",
				Filter: FilterPath{
					Path:   "[rdfs:Resource]rdfs:label",
					Values: []string{"Value1", "Value2"},
				},
				Value: "[crm:E35_Title]crm:P190_has_symbolic_content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePathString(tt.path)

			// Check if we expected an error
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePathString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error or nil result, don't check further
			if tt.wantErr || tt.expected == nil {
				if got != nil {
					t.Errorf("Expected nil result, got %+v", got)
				}
				return
			}

			// Check resource path
			if got.Resource != tt.expected.Resource {
				t.Errorf("Resource = %v, want %v", got.Resource, tt.expected.Resource)
			}

			// Check filter path and values
			if got.Filter.Path != tt.expected.Filter.Path {
				t.Errorf("Filter.Path = %v, want %v", got.Filter.Path, tt.expected.Filter.Path)
			}

			if !compareStringSlices(got.Filter.Values, tt.expected.Filter.Values) {
				t.Errorf("Filter.Values = %v, want %v", got.Filter.Values, tt.expected.Filter.Values)
			}

			// Check value path
			if got.Value != tt.expected.Value {
				t.Errorf("Value = %v, want %v", got.Value, tt.expected.Value)
			}
		})
	}
}

// compareStringSlices compares two string slices for equality regardless of order
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences of each string
	aCount := make(map[string]int)
	bCount := make(map[string]int)

	// Count occurrences in first slice
	for _, s := range a {
		aCount[s]++
	}

	// Count occurrences in second slice
	for _, s := range b {
		bCount[s]++
	}

	// Compare the counts
	for s, count := range aCount {
		if bCount[s] != count {
			return false
		}
	}

	// Check if there are any strings in b that aren't in a
	for s, count := range bCount {
		if aCount[s] != count {
			return false
		}
	}

	return true
}

func TestParseCsvPathMap(t *testing.T) {
	is := is.New(t)

	// Test data from the CSV example
	csvLines := []string{
		"TargetField,QueryPath,Source,Notes",
		"dc_title,\"[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h | [crm:E35_Title]crm:P190_has_symbolic_content\",Brocade,\"Main title of the resource\"",
		"dc_publisher,\"[lrm:F3_Manifestation]lrm:R24i_was_created_through->[lrm:F30_Manifestation_Creation]crm:P14_carried_out_by->[crm:E39_Actor] | [crm:E39_Actor]crm:P1_is_identified_by->[crm:E41_Appellation]crm:P190_has_symbolic_content\",Brocade,\"Publisher information\"",
	}

	sourceMap, err := ParseCsvPathMap(csvLines)
	is.NoErr(err)

	pathMap := map[string]*PathQuery{}
	for k, entry := range sourceMap {
		pq, err := ParsePathString(entry)
		is.NoErr(err)
		pathMap[k] = pq
	}

	// Verify we got the expected number of entries
	is.Equal(len(sourceMap), 2)

	// Check the first entry
	titlePath, exists := pathMap["dc_title"]
	is.True(exists)
	is.Equal(titlePath.Resource, "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]")
	is.Equal(titlePath.Filter.Path, "[crm:E35_Title]crm:P2_has_type")
	is.Equal(len(titlePath.Filter.Values), 1)
	is.Equal(titlePath.Filter.Values[0], "https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h")
	is.Equal(titlePath.Value, "[crm:E35_Title]crm:P190_has_symbolic_content")

	// Check the second entry
	publisherPath, exists := pathMap["dc_publisher"]
	is.True(exists)
	is.Equal(publisherPath.Resource, "[lrm:F3_Manifestation]lrm:R24i_was_created_through->[lrm:F30_Manifestation_Creation]crm:P14_carried_out_by->[crm:E39_Actor]")
	if diff := cmp.Diff(publisherPath.Value, "[crm:E39_Actor]crm:P1_is_identified_by->[crm:E41_Appellation]crm:P190_has_symbolic_content"); diff != "" {
		t.Errorf("golden file %s mismatch (-want +got):\n%s", "publisher", diff)
	}
}
