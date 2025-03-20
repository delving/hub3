package index

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestBindComprehensive(t *testing.T) {
	is := is.New(t)

	var (
		viaRemote bool
		rawData   []byte
		err       error
	)

	viaRemote = true

	// Get test graph
	switch viaRemote {
	case true:
		rawData, err = os.ReadFile("./testdata/rdf_brocade_index_graph.json")
		is.NoErr(err)
	default:
		rawGraph, _, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
		is.NoErr(err)
		rawData = []byte(rawGraph)
	}

	var graph *Graph
	marshalErr := json.Unmarshal(rawData, &graph)
	is.NoErr(marshalErr)

	err = graph.Inline()
	is.NoErr(err)

	// Parse from CSV format
	csvLines := []string{
		"TargetField,QueryPath,Source,Notes",
		// "dc_publisher,\"[lrm:F3_Manifestation]lrm:R24i_was_created_through->[lrm:F30_Manifestation_Creation]crm:P14_carried_out_by->[crm:E39_Actor] | [crm:E39_Actor]crm:P1_is_identified_by->[crm:E41_Appellation]crm:P190_has_symbolic_content\",Brocade,\"Publisher\"",
		"dc_title,\"[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h | [crm:E35_Title]crm:P190_has_symbolic_content\",Brocade,\"Main title\"",
		"dc_date,\"[lrm:F3_Manifestation]lrm:R24i_was_created_through->[lrm:F30_Manifestation_Creation]crm:P4_has_time-span->[crm:E52_Time-Span] | [crm:E52_Time-Span]crm:P82a_begin_of_the_begin\",Brocade,\"Date\"",
		"dc_creator,\"[lrm:F3_Manifestation]crm:P94i_was_created_by->[lrm:F28_Expression_Creation]crm:P01i_is_domain_of->[crm:PC14_carried_out_by] | [crm:PC14_carried_out_by]crm:P14.1_in_the_role_of=https://data.antwerp.be/id/term/brocade-catalog/fu/com | [crm:PC14_carried_out_by]crm:P02_has_range->[crm:E39_Actor]crm:P1_is_identified_by->[crm:E41_Appellation]crm:P190_has_symbolic_content\",Brocade,\"Contributor\"",
		// "dc_type,\"[lrm:F3_Manifestation]crm:P2_has_type->[crm:E55_Type] | [crm:E55_Type]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-authorities/ty/pt | [crm:E55_Type]\",\"\",\"\"",
		"dc_language,\"[lrm:F3_Manifestation]crm:P106_is_composed_of->[crm:E33_Linguistic_Object]crm:P72_has_language\",\"\",\"\"",
		// "dc_identifier,\"[lrm:F3_Manifestation]crm:P1_is_identified_by->[crm:E42_Identifier] | [crm:E42_Identifier]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/cloi | [crm:E42_Identifier]->crm:P190_has_symbolic_content\",\"\",\"\"",
		"dc_publisher,\"[lrm:F3_Manifestation]crm:P128i_is_carried_by->[crm:E22_Human-Made_Object]crm:P50_has_current_keeper\",\"\",\"\"",
		"dc_format,\"[lrm:F3_Manifestation]crm:P43_has_dimension->[crm:E54_Dimension]crm:P129i_is_subject_of->[crm:E33_Linguistic_Object] | [crm:E33_Linguistic_Object]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/co-pg | [crm:E33_Linguistic_Object]crm:P190_has_symbolic_content\",Brocade,\"Format\"",
	}

	pathMap, err := ParseCsvPathMap(csvLines)
	is.NoErr(err)

	// Expected values based on the XML data
	expectations := map[string][]string{
		"dc_title":   {"Keur van proza- en dichtstukken"},
		"dc_date":    {"1895"},
		"dc_creator": {"sr.", "Jan", "Beers, van"},
		// // "dc_type":  {"https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1"},
		"dc_language": {"https://data.antwerp.be/id/term/brocade-catalog/lg/dut"},
		// // "dc_identifier": {"c:lvd:538499"},
		"dc_publisher": {"https://data.antwerp.be/id/term/brocade-catalog/library/lh"},
		"dc_format":    {"1 v."},
	}

	// Execute the Bind function
	got, err := graph.Bind(pathMap)
	is.NoErr(err)

	// Verify results against expectations
	for field, expectedValues := range expectations {
		entries, exists := got[field]
		t.Logf("Field: %s, Entries: %v", field, entries)
		is.True(exists) // The field should exist in results

		actualValues := make([]string, len(entries))
		for i, entry := range entries {
			actualValues[i] = entry.Value
			if entry.EntryType != Literal {
				actualValues[i] = entry.ID
			}
		}

		// For contributor, the order might vary, so check differently
		if field == "dc_contributor" {
			// Verify we have all expected values (order doesn't matter)
			for _, expected := range expectedValues {
				found := false
				for _, actual := range actualValues {
					if actual == expected {
						found = true
						break
					}
				}
				is.True(found) // Should find each expected value
			}
		} else {
			// For other fields, verify exact match including order
			is.Equal(len(actualValues), len(expectedValues))
			for i, expectedValue := range expectedValues {
				if i < len(actualValues) {
					is.Equal(actualValues[i], expectedValue)
				}
			}
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No change needed",
			input:    "[class]pred->[class]pred | [filter]path=value",
			expected: "[class]pred->[class]pred | [filter]path=value",
		},
		{
			name:     "Single > to ->",
			input:    "[class]pred>[class]pred",
			expected: "[class]pred->[class]pred",
		},
		{
			name:     "Mixed > and -> usage",
			input:    "[class]pred>[class]pred->[class]pred",
			expected: "[class]pred->[class]pred->[class]pred",
		},
		{
			name:     "Multiple parts with >",
			input:    "[class]pred>[class]pred | [filter]path=value | [value]pred>[class]pred",
			expected: "[class]pred->[class]pred | [filter]path=value | [value]pred->[class]pred",
		},
		{
			name:     "Complex example from test data",
			input:    "[lrm:F3_Manifestation]crm:P102_has_title>[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=value",
			expected: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			got := normalizePath(tt.input)
			is.Equal(tt.expected, got)
		})
	}
}

func TestBindSimple(t *testing.T) {
	is := is.New(t)

	// Get test graph
	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	// Ensure graph is inlined
	err = graph.Inline()
	is.NoErr(err)

	// Simple path map with just one field
	pathMap := PathMap{
		"title": "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
	}

	// Execute the Bind function
	got, err := graph.Bind(pathMap)
	is.NoErr(err)

	// Verify results
	entries, exists := got["title"]
	is.True(exists)
	is.True(len(entries) > 0)
	is.Equal("Keur van proza- en dichtstukken", entries[0].Value)
}

func TestBindEmpty(t *testing.T) {
	is := is.New(t)

	// Get test graph
	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	// Empty path map
	pathMap := PathMap{}

	// Execute the Bind function
	got, err := graph.Bind(pathMap)
	is.NoErr(err)

	// Should return empty map
	is.Equal(0, len(got))
}

func TestBindNonExistentPaths(t *testing.T) {
	is := is.New(t)

	// Get test graph
	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	// Path map with non-existent path
	pathMap := PathMap{
		"nonexistent": "[lrm:F3_Manifestation]crm:NonExistentPredicate->[crm:NonExistentClass]",
	}

	// Execute the Bind function
	got, err := graph.Bind(pathMap)
	is.NoErr(err)

	// Should not have the non-existent field
	_, exists := got["nonexistent"]
	is.Equal(false, exists)
}
