package index

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/matryer/is"
	"github.com/tidwall/sjson"
)

func getGraphByFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	g1, err := ntriples.Parse(f, nil)
	if err != nil {
		return "", err
	}

	header := Header{
		OrgID:    "org1",
		Spec:     "spec1",
		HubID:    "hub1",
		EntryURI: "urn:123",
	}

	graph, err := NewGraph(header)
	if err != nil {
		return "", err
	}

	if err := graph.AddGraph(g1); err != nil {
		return "", err
	}

	b, err := graph.Marshal()
	if err != nil {
		return "", err
	}

	b, err = sjson.DeleteBytes(b, "meta.modified")
	if err != nil {
		return "", err
	}

	return formatJSON(b)
}

func TestQuery(t *testing.T) {
	// Set up test resources
	resources := []*Resource{
		{
			ID:    "person1",
			Types: []string{"Person"},
			Entries: []*Entry{
				{
					Predicate: "knows",
					ID:        "person2",
					EntryType: ResourceType,
				},
				{
					Predicate: "name",
					Value:     "John",
					EntryType: Literal,
				},
			},
		},
		{
			ID:    "person2",
			Types: []string{"Person"},
			Entries: []*Entry{
				{
					Predicate: "worksAt",
					ID:        "org1",
					EntryType: ResourceType,
				},
				{
					Predicate: "name",
					Value:     "Jane",
					EntryType: Literal,
				},
			},
		},
		{
			ID:    "org1",
			Types: []string{"Organization"},
			Entries: []*Entry{
				{
					Predicate: "name",
					Value:     "TechCorp",
					EntryType: Literal,
				},
				{
					Predicate: "location",
					Value:     "New York",
					EntryType: Literal,
				},
			},
		},
	}

	tests := []struct {
		name     string
		path     []PathFilter
		wantLen  int
		wantPred string // predicate of the entries we expect to find
	}{
		{
			name: "single level - find person names",
			path: []PathFilter{
				{Type: "Person", Predicate: "name"},
			},
			wantLen:  2,
			wantPred: "name",
		},
		{
			name: "two levels - person knows person worksAt",
			path: []PathFilter{
				{Type: "Person", Predicate: "knows"},
				{Type: "Person", Predicate: "worksAt"},
			},
			wantLen:  1,
			wantPred: "worksAt",
		},
		{
			name: "specific person to org",
			path: []PathFilter{
				{Type: "Person", ID: "person2", Predicate: "worksAt"},
				{Type: "Organization", Predicate: "name"},
			},
			wantLen:  1,
			wantPred: "name",
		},
		{
			name: "no matches - wrong type",
			path: []PathFilter{
				{Type: "NonExistent", Predicate: "knows"},
			},
			wantLen: 0,
		},
		{
			name: "no matches - wrong ID",
			path: []PathFilter{
				{Type: "Person", ID: "nonexistent", Predicate: "knows"},
			},
			wantLen: 0,
		},
		{
			name:    "empty path",
			path:    []PathFilter{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Query(resources, tt.path)

			if len(got) != tt.wantLen {
				t.Errorf("Query() got %d entries, want %d", len(got), tt.wantLen)
			}

			// If we expect entries, check they have the correct predicate
			if tt.wantLen > 0 && tt.wantPred != "" {
				for i, entry := range got {
					if entry.Predicate != tt.wantPred {
						t.Errorf("Entry[%d] has predicate %s, want %s",
							i, entry.Predicate, tt.wantPred)
					}
				}
			}
		})
	}
}

// TestQueryWithCycles tests that the Query function handles cyclic relationships
func TestQueryWithCycles(t *testing.T) {
	resources := []*Resource{
		{
			ID:    "person1",
			Types: []string{"Person"},
			Entries: []*Entry{
				{
					Predicate: "knows",
					ID:        "person2",
					EntryType: ResourceType,
				},
			},
		},
		{
			ID:    "person2",
			Types: []string{"Person"},
			Entries: []*Entry{
				{
					Predicate: "knows",
					ID:        "person1", // Creates a cycle
					EntryType: ResourceType,
				},
			},
		},
	}

	path := []PathFilter{
		{Type: "Person", Predicate: "knows"},
		{Type: "Person", Predicate: "knows"},
	}

	got := Query(resources, path)
	wantLen := 2 // Should find both "knows" relationships
	if len(got) != wantLen {
		t.Errorf("Query() with cycles got %d entries, want %d", len(got), wantLen)
	}
}

func TestQueryTitlePath(t *testing.T) {
	// Setup test resources based on the N-triples
	resources := []*Resource{
		{
			ID: "https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:10",
			Types: []string{
				"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
			},
			Entries: []*Entry{
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
					ID:        "_:B6fca6bb5X2Dc5c3X2D4ebeX2D83c0X2D0900345d896f",
					EntryType: ResourceType,
				},
			},
		},
		{
			ID: "_:B6fca6bb5X2Dc5c3X2D4ebeX2D83c0X2D0900345d896f",
			Types: []string{
				"http://www.cidoc-crm.org/cidoc-crm/E35_Title",
			},
			Entries: []*Entry{
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
					Value:     "Fortran: initiation au langage de l'informatique scientifique",
					EntryType: Literal,
				},
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P2_has_type",
					ID:        "https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h",
					EntryType: ResourceType,
				},
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P72_has_language",
					ID:        "https://data.antwerp.be/id/term/brocade-catalog/lg/fre",
					EntryType: ResourceType,
				},
			},
		},
	}

	// Define the path we want to test
	path := []PathFilter{
		{
			Type:      "https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
			Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
		},
		{
			Type:      "http://www.cidoc-crm.org/cidoc-crm/E35_Title",
			Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
		},
	}

	// Execute the query
	got := Query(resources, path)

	// Log the initial state and execution
	t.Logf("Testing path: F3_Manifestation > P102_has_title > E35_Title > P190_has_symbolic_content")
	t.Logf("Starting resources:")
	for _, r := range resources {
		t.Logf("  Resource ID: %s", r.ID)
		t.Logf("  Types: %v", r.Types)
		t.Logf("  Entries:")
		for _, e := range r.Entries {
			t.Logf("    Predicate: %s", e.Predicate)
			t.Logf("    ID: %s", e.ID)
			t.Logf("    Value: %s", e.Value)
			t.Logf("    EntryType: %s", e.EntryType)
		}
		t.Logf("---")
	}

	// Log the results
	t.Logf("Results found: %d", len(got))
	for i, entry := range got {
		t.Logf("Result %d:", i+1)
		t.Logf("  ID: %s", entry.ID)
		t.Logf("  Value: %s", entry.Value)
		t.Logf("  EntryType: %s", entry.EntryType)
		t.Logf("  Predicate: %s", entry.Predicate)
	}

	// Verify the results
	if len(got) != 1 {
		t.Errorf("Expected 1 result, got %d", len(got))
		return
	}

	expected := &Entry{
		Value:     "Fortran: initiation au langage de l'informatique scientifique",
		EntryType: Literal,
		Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
	}

	result := got[0]
	if result.Value != expected.Value {
		t.Errorf("Expected value %q, got %q", expected.Value, result.Value)
	}
	if result.EntryType != expected.EntryType {
		t.Errorf("Expected EntryType %v, got %v", expected.EntryType, result.EntryType)
	}
	if result.Predicate != expected.Predicate {
		t.Errorf("Expected Predicate %s, got %s", expected.Predicate, result.Predicate)
	}
}

func TestQueryPaths(t *testing.T) {
	// Define test paths with expected entries
	pathTests := map[string][]*Entry{
		"[lrm:F3_Manifestation]crm:P102_has_title>[crm:E35_Title]crm:P190_has_symbolic_content": {
			{
				Value:     "Fortran: initiation au langage de l'informatique scientifique",
				EntryType: Literal,
			},
		},
	}

	is := is.New(t)
	graph, err := getGraphByFile("./testdata/rdf.nt")
	is.NoErr(err)
	g, err := GraphFromBytes([]byte(graph))
	is.NoErr(err)
	resources := g.Resources

	// Create test cases from the pathTests map
	var tests []struct {
		name     string
		path     []PathFilter
		wantLen  int
		expected []*Entry
	}

	for pathStr, expectedEntries := range pathTests {
		filters := parsePathString(pathStr)
		t.Logf("Filters: %+v\n", filters)
		for _, f := range filters {
			slog.Info("filters", "filter", f)
		}
		if len(filters) > 0 {
			pathDesc := strings.Join(pathNames(filters), " > ")
			test := struct {
				name     string
				path     []PathFilter
				wantLen  int
				expected []*Entry
			}{
				name:     pathDesc,
				path:     filters,
				wantLen:  len(expectedEntries),
				expected: expectedEntries,
			}
			tests = append(tests, test)
		}
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Query(resources, tt.path)
			is := is.New(t)

			// Log the results
			t.Logf("Path: %s", strings.Join(pathNames(tt.path), " > "))
			t.Logf("Found %d entries", len(got))

			// Check length
			is.Equal(len(got), tt.wantLen)

			// Compare results with expected entries
			for i, entry := range got {
				if i >= len(tt.expected) {
					t.Errorf("Unexpected extra entry: %+v", entry)
					continue
				}

				expected := tt.expected[i]

				// Compare IDs if they exist
				if expected.ID != "" {
					is.Equal(entry.ID, expected.ID)
				}

				// Compare Values if they exist
				if expected.Value != "" {
					is.Equal(entry.Value, expected.Value)
				}

				// Compare EntryType
				is.Equal(entry.EntryType, expected.EntryType)

				// Log the details
				if entry.Value != "" {
					t.Logf("Entry %d Value: %s", i+1, entry.Value)
				}
				if entry.ID != "" {
					t.Logf("Entry %d ID: %s", i+1, entry.ID)
				}
			}
		})
	}
}

// cleanPath removes all spaces and <br> tags from a path string
func cleanPath(path string) string {
	// Remove all spaces
	path = strings.ReplaceAll(path, " ", "")
	// Remove <br> tags
	path = strings.ReplaceAll(path, "<br>", "")
	return path
}

// parsePathString parses a path string into PathFilters using the following format:
// [Type]Predicate>[Type]Predicate>...
func parsePathString(path string) []PathFilter {
	// Remove all spaces and <br> tags
	path = strings.ReplaceAll(path, " ", "")
	path = strings.ReplaceAll(path, "<br>", "")

	if path == "" {
		return nil
	}

	parts := strings.Split(path, ">")
	filters := make([]PathFilter, 0, len(parts)-1)

	for i := 0; i < len(parts)-1; i++ {
		current := parts[i]
		// next := parts[i+1]

		// Skip if this is a Literal target
		if current == "rdfs:Literal" {
			continue
		}

		filter := PathFilter{}

		// Extract type if present
		if strings.HasPrefix(current, "[") {
			if idx := strings.Index(current, "]"); idx != -1 {
				filter.Type = expandPrefix(current[1:idx])
				filter.Predicate = expandPrefix(current[idx+1:])
			}
		} else {
			filter.Predicate = expandPrefix(current)
		}

		// Add the filter if we have at least a predicate
		if filter.Predicate != "" {
			filters = append(filters, filter)
		}
	}

	// Handle the last part if it's not a Literal
	lastPart := parts[len(parts)-1]
	if lastPart != "rdfs:Literal" {
		filter := PathFilter{}
		if strings.HasPrefix(lastPart, "[") {
			if idx := strings.Index(lastPart, "]"); idx != -1 {
				filter.Type = expandPrefix(lastPart[1:idx])
				filter.Predicate = expandPrefix(lastPart[idx+1:])
				filters = append(filters, filter)
			}
		} else {
			filter.Predicate = expandPrefix(lastPart)
			filters = append(filters, filter)
		}
	}

	return filters
}

func TestParsePathString(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []PathFilter
	}{
		{
			name: "Full path with types",
			path: "[lrm:F3_Manifestation]crm:P102_has_title>[crm:E35_Title]crm:P190_has_symbolic_content>rdfs:Literal",
			expected: []PathFilter{
				{
					Type:      "https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
				{
					Type:      "http://www.cidoc-crm.org/cidoc-crm/E35_Title",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
				},
			},
		},
		{
			name: "Path without types",
			path: "crm:P102_has_title>crm:P190_has_symbolic_content>rdfs:Literal",
			expected: []PathFilter{
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
				},
			},
		},
		{
			name: "Mixed path with and without types",
			path: "[lrm:F3_Manifestation]crm:P102_has_title>crm:P190_has_symbolic_content>rdfs:Literal",
			expected: []PathFilter{
				{
					Type:      "https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
				},
			},
		},
		{
			name: "Path with spaces",
			path: "  [lrm:F3_Manifestation]  crm:P102_has_title  >  [crm:E35_Title]  crm:P190_has_symbolic_content  >  rdfs:Literal  ",
			expected: []PathFilter{
				{
					Type:      "https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
				{
					Type:      "http://www.cidoc-crm.org/cidoc-crm/E35_Title",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
				},
			},
		},
		{
			name:     "Empty path",
			path:     "",
			expected: nil,
		},
		{
			name: "Single predicate",
			path: "crm:P102_has_title",
			expected: []PathFilter{
				{
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
			},
		},
		{
			name: "Single typed predicate",
			path: "[lrm:F3_Manifestation]crm:P102_has_title",
			expected: []PathFilter{
				{
					Type:      "https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
					Predicate: "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePathString(tt.path)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("parsePathString() mismatch (-want +got):\n%s", diff)
				t.Logf("Input path: %s", tt.path)
				t.Logf("Got %d filters, expected %d", len(got), len(tt.expected))
			}
		})
	}
}

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal path",
			input:    "lrm:F3_Manifestation=>crm:P102_has_title",
			expected: "lrm:F3_Manifestation=>crm:P102_has_title",
		},
		{
			name:     "Path with spaces",
			input:    "  lrm:F3_Manifestation  =>  crm:P102_has_title  ",
			expected: "lrm:F3_Manifestation=>crm:P102_has_title",
		},
		{
			name:     "Path with <br> tags",
			input:    "lrm:F3_Manifestation<br>=>crm:P102_has_title",
			expected: "lrm:F3_Manifestation=>crm:P102_has_title",
		},
		{
			name:     "Path with mixed spaces and <br>",
			input:    "  lrm:F3_Manifestation  <br>  =>  crm:P102_has_title  ",
			expected: "lrm:F3_Manifestation=>crm:P102_has_title",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Only spaces and <br>",
			input:    "   <br>   <br>   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPath(tt.input)
			if got != tt.expected {
				t.Errorf("cleanPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Helper functions remain the same
func pathNames(filters []PathFilter) []string {
	names := make([]string, len(filters))
	for i, f := range filters {
		typeShort := shortenURI(f.Type)
		predShort := shortenURI(f.Predicate)
		names[i] = fmt.Sprintf("%s:%s", typeShort, predShort)
	}
	return names
}

func shortenURI(uri string) string {
	if idx := strings.LastIndex(uri, "/"); idx != -1 {
		return uri[idx+1:]
	}
	if idx := strings.LastIndex(uri, "#"); idx != -1 {
		return uri[idx+1:]
	}
	return uri
}

func expandPrefix(s string) string {
	prefixMap := map[string]string{
		"crm:":  "http://www.cidoc-crm.org/cidoc-crm/",
		"lrm:":  "https://www.iflastandards.info/ns/lrm/lrmoo#",
		"rdfs:": "http://www.w3.org/2000/01/rdf-schema#",
	}

	for prefix, expansion := range prefixMap {
		if strings.HasPrefix(s, prefix) {
			return expansion + strings.TrimPrefix(s, prefix)
		}
	}
	return s
}

func TestQueryTitlePathWithVerify(t *testing.T) {
	is := is.New(t)

	graph, err := getGraphByFile("./testdata/rdf.nt")
	is.NoErr(err)

	g, err := GraphFromBytes([]byte(graph))
	is.NoErr(err)

	resources := g.Resources

	// Define the expected resources for the title path
	expected := []*Resource{
		{
			ID: "https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:10",
			Types: []string{
				"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation",
			},
			Entries: []*Entry{
				{
					Predicate:   "http://www.cidoc-crm.org/cidoc-crm/P102_has_title",
					ID:          "B6fca6bb5X2Dc5c3X2D4ebeX2D83c0X2D0900345d896f",
					EntryType:   Bnode,
					SearchLabel: "crm_P102_has_title",
				},
			},
			order: 0,
		},
		{
			ID: "B6fca6bb5X2Dc5c3X2D4ebeX2D83c0X2D0900345d896f",
			Types: []string{
				"http://www.cidoc-crm.org/cidoc-crm/E35_Title",
			},
			Entries: []*Entry{
				{
					Predicate:   "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content",
					Value:       "Fortran: initiation au langage de l'informatique scientifique",
					EntryType:   Literal,
					SearchLabel: "crm_P190_has_symbolic_content",
					DataType:    "http://www.w3.org/2001/XMLSchema#string",
				},
			},
			order: 1,
		},
	}

	// Extract and filter the relevant resources from the full list
	var extracted []*Resource
	manifestationID := "https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:10"
	titleID := "B6fca6bb5X2Dc5c3X2D4ebeX2D83c0X2D0900345d896f"

	for _, res := range resources {
		switch res.ID {
		case manifestationID:
			// Filter manifestation to only include title entry
			filteredEntries := make([]*Entry, 0)
			for _, entry := range res.Entries {
				if entry.ID == titleID {
					filteredEntries = append(filteredEntries, entry)
				}
			}

			filteredResource := &Resource{
				ID:      res.ID,
				Types:   res.Types,
				Entries: filteredEntries,
			}
			extracted = append(extracted, filteredResource)
		case titleID:
			t.Logf("title resource: %#v", res)
			// Filter title resource to only include symbolic content entry
			filteredEntries := make([]*Entry, 0)
			for _, entry := range res.Entries {
				if entry.Predicate == "http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content" {
					filteredEntries = append(filteredEntries, entry)
				}
			}

			filteredResource := &Resource{
				ID:      res.ID,
				Types:   res.Types,
				Entries: filteredEntries,
			}
			extracted = append(extracted, filteredResource)
		}
	}

	is.Equal(len(extracted), len(expected)) // it should extract all the rerources

	// Sort the extracted resources by ID to ensure consistent ordering
	sort.Slice(extracted, func(i, j int) bool {
		return extracted[i].order < extracted[j].order
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].order < expected[j].order
	})

	// Setup cmp options to handle unexported fields and mutexes
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(Resource{}, Entry{}, CustomFilterField{}, TypeIndexField{}, sync.RWMutex{}),
	}

	// Compare using cmp with the options
	if diff := cmp.Diff(expected, extracted, opts...); diff != "" {
		t.Errorf("Resource mismatch (-want +got):\n%s", diff)
	}

	// Log the resources for debugging
	t.Log("Expected resources:")
	for _, r := range expected {
		t.Logf("Resource ID: %s", r.ID)
		t.Logf("Types: %v", r.Types)
		for _, e := range r.Entries {
			t.Logf("  Entry: Predicate=%s, ID=%s, Value=%s, Type=%s",
				e.Predicate, e.ID, e.Value, e.EntryType)
		}
	}

	t.Log("\nExtracted resources:")
	for _, r := range extracted {
		t.Logf("Resource ID: %s", r.ID)
		t.Logf("Types: %v", r.Types)
		for _, e := range r.Entries {
			t.Logf("  Entry: Predicate=%s, ID=%s, Value=%s, Type=%s",
				e.Predicate, e.ID, e.Value, e.EntryType)
		}
	}
}
