package index

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
	"github.com/tidwall/sjson"
)

func TestGraph_Marshal(t *testing.T) {
	header := Header{
		OrgID:    "org1",
		Spec:     "spec1",
		HubID:    "hub1",
		EntryURI: "urn:123",
	}

	graph, err := NewGraph(header)
	if err != nil {
		t.Fatalf("failed to create new graph: %v", err)
	}

	resource1 := &Resource{
		ID:    "resource1",
		order: 1,
		Entries: []*Entry{
			{
				SearchLabel: "label1",
			},
		},
	}

	resource2 := &Resource{
		ID:    "resource2",
		order: 2,
		Entries: []*Entry{
			{
				SearchLabel: "label2",
			},
		},
	}

	resource3 := &Resource{
		ID:    "resource3",
		order: 3,
		Entries: []*Entry{
			{
				SearchLabel: "label3",
			},
		},
	}

	graph.Resources = append(graph.Resources, resource1, resource2, resource3)

	got, err := graph.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal graph: %v", err)
	}

	want, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("failed to marshal graph using json.Marshal: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
	}
}

func getIndexGraph() (string, error) {
	f, err := os.Open("../formats/ntriples/testdata/rdf.nt")
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

func formatJSON(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func TestGraph_MarshalFromGraph(t *testing.T) {
	is := is.New(t)

	want, err := getIndexGraph()
	is.NoErr(err)

	got, err := getIndexGraph()
	is.NoErr(err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
	}
}

func TestGraphInline(t *testing.T) {
	is := is.New(t)

	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	is.Equal(len(graph.roots), 0)

	err = graph.Inline()
	is.NoErr(err)
	is.Equal(len(graph.roots), 1)
	is.Equal(graph.roots[0].Types, []string{"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation"})
}

func TestGraph_GenerateFields(t *testing.T) {
	is := is.New(t)

	// Create a test graph with resources that have duplicate values
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject",
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create a resource
	resource1 := &Resource{
		ID:    "urn:test:subject",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, resource1)

	// Add entries with same predicate and some duplicate values
	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title",
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title", // Duplicate value
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Another Title", // Different value, same predicate
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/type",
		SearchLabel: "dc_type",
		Value:       "", // Empty value should be skipped
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/relation",
		SearchLabel: "dc_relation",
		EntryType:   ResourceType, // Non-literal entry should be skipped
		ID:          "urn:test:related",
	})

	// Initially, Fields should be nil or empty
	is.True(g.Fields == nil || len(g.Fields) == 0)

	// Generate fields
	g.GenerateFields()

	// Verify Fields is populated correctly
	is.True(g.Fields != nil)
	is.Equal(len(g.Fields), 2) // Should have dc_title and dc_date

	// Check that dc_title has both unique values and no duplicates
	is.Equal(len(g.Fields["dc_title"]), 2)
	titleValues := g.Fields["dc_title"]
	// Sort for consistent comparison
	sort.Strings(titleValues)
	is.Equal(titleValues[0], "Another Title")
	is.Equal(titleValues[1], "Test Title")

	// Check that dc_date has one value
	is.Equal(len(g.Fields["dc_date"]), 1)
	is.Equal(g.Fields["dc_date"][0], "2020-01-01")

	// Empty value predicate should be skipped
	_, hasEmptyValue := g.Fields["dc_type"]
	is.Equal(hasEmptyValue, false)

	// Non-literal entry should be skipped
	_, hasRelation := g.Fields["dc_relation"]
	is.Equal(hasRelation, false)

	// Run GenerateFields again to ensure it doesn't duplicate values
	g.GenerateFields()
	is.Equal(len(g.Fields["dc_title"]), 2) // Still just 2 values
}

func TestGraph_MarshalWithFields(t *testing.T) {
	is := is.New(t)

	// Create a test graph with some entries
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject",
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create a resource
	resource1 := &Resource{
		ID:    "urn:test:subject",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, resource1)

	// Add entries
	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title",
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	// Generate fields
	g.GenerateFields()

	// Marshal the graph
	data, err := g.Marshal()
	is.NoErr(err)

	// Unmarshal to verify Fields is included
	var graphData map[string]interface{}
	err = json.Unmarshal(data, &graphData)
	is.NoErr(err)

	// Check that fields exist in the marshalled JSON
	fields, ok := graphData["fields"]
	is.True(ok) // Fields should exist in the JSON

	// Convert to map for inspection
	fieldsMap, ok := fields.(map[string]interface{})
	is.True(ok)
	
	// Check that both searchLabels are included as keys (not predicates)
	titleValues, ok := fieldsMap["dc_title"].([]interface{})
	is.True(ok)
	is.Equal(len(titleValues), 1)
	is.Equal(titleValues[0], "Test Title")

	dateValues, ok := fieldsMap["dc_date"].([]interface{})
	is.True(ok)
	is.Equal(len(dateValues), 1)
	is.Equal(dateValues[0], "2020-01-01")
}
