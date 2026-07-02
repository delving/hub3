package index

import (
	"testing"

	"github.com/matryer/is"
)

// getGraphByFile is assumed to be defined elsewhere in the test package
// It loads an RDF file, parses it, and returns a Graph

func TestGraph_Query(t *testing.T) {
	is := is.New(t)
	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	// Ensure graph is inlined for querying
	err = graph.Inline()
	is.NoErr(err)

	tests := []struct {
		name            string
		path            string
		wantResources   int
		wantEntries     int
		wantErr         bool
		checkResourceID string
		checkEntryValue string
	}{
		{
			name:          "Simple resource path",
			path:          "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]",
			wantResources: 1,
			wantEntries:   0,
			wantErr:       false,
		},
		{
			name:            "Path with value extraction",
			path:            "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content",
			wantResources:   1,
			wantEntries:     1,
			wantErr:         false,
			checkEntryValue: "Keur van proza- en dichtstukken",
		},
		{
			name:            "Path with filter",
			path:            "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h",
			wantResources:   1,
			wantEntries:     0,
			wantErr:         false,
			checkResourceID: "_:b9b5bf0b4a85e3c",
		},
		{
			name:            "Path with filter and value",
			path:            "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h | [crm:E35_Title]crm:P190_has_symbolic_content",
			wantResources:   1,
			wantEntries:     1,
			wantErr:         false,
			checkEntryValue: "Keur van proza- en dichtstukken",
		},
		{
			name:          "Path with wildcard class",
			path:          "[]crm:P102_has_title->[crm:E35_Title]",
			wantResources: 1,
			wantEntries:   0,
			wantErr:       false,
		},
		{
			name:          "Non-existent path",
			path:          "[lrm:F3_Manifestation]crm:Dummy->[]crm:unknown",
			wantResources: 0,
			wantEntries:   0,
			wantErr:       false,
		},
		{
			name:          "Invalid path syntax",
			path:          "[unclosed[bracket]path",
			wantResources: 0,
			wantEntries:   0,
			wantErr:       true,
		},
		{
			name:          "Empty path",
			path:          "",
			wantResources: 0,
			wantEntries:   0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			result, err := graph.Query(tt.path)

			// Check error expectation
			if tt.wantErr {
				is.True(err != nil) // Should have an error
				return
			} else {
				is.NoErr(err) // Should not have an error
			}

			// Check nil result
			if tt.wantResources == 0 && tt.wantEntries == 0 {
				if result != nil {
					is.Equal(0, len(result.Resources)) // Should have no resources
					is.Equal(0, len(result.Entries))   // Should have no entries
				}
				return
			}

			// Check resource and entries count
			is.Equal(tt.wantResources, len(result.Resources)) // Resource count should match
			is.Equal(tt.wantEntries, len(result.Entries))     // Entries count should match

			// Check specific resource ID and entry value if specified
			verifyQueryResults(t, result, tt.checkResourceID, tt.checkEntryValue)
		})
	}
}

// createTestGraph creates a test graph mimicking the structure in the RDF example
// Helper function to verify result contents
func verifyQueryResults(t *testing.T, result *QueryResult, expectedResourceID string, expectedEntryValue string) {
	is := is.New(t)

	// Check if expected resource ID is present
	if expectedResourceID != "" && len(result.Resources) > 0 {
		found := false
		for _, r := range result.Resources {
			if r.ID == expectedResourceID {
				found = true
				break
			}
		}
		is.True(found) // Expected resource ID should be found
	}

	// Check if expected entry value is present
	if expectedEntryValue != "" && len(result.Entries) > 0 {
		found := false
		for _, e := range result.Entries {
			if e.Value == expectedEntryValue {
				found = true
				break
			}
		}
		is.True(found) // Expected entry value should be found
	}
}

