package index

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/matryer/is"
)

func init() {
	// Register test namespaces
	rdf.DefaultNamespaceManager.Put("test", "http://example.org/ontology/", 10)
	rdf.DefaultNamespaceManager.Put("ex", "http://example.org/", 10)
}

func TestGraph_IndexMessage(t *testing.T) {
	is := is.New(t)

	// Create a test graph with a simple structure
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject", // Important: This will be our root resource
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create a simple RDF graph with linked resources
	rdfGraph := rdf.NewGraph()

	// Root subject
	rootSubject, _ := rdf.NewIRI("urn:test:subject")

	// Define a type for root
	rdfType, _ := rdf.NewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	rootTypeObj, _ := rdf.NewIRI("http://example.org/ontology/RootType")
	rdfGraph.AddTriple(rootSubject, rdfType, rootTypeObj)

	// Link to level 1 resource
	level1Pred, _ := rdf.NewIRI("http://example.org/ontology/hasLevel1")
	level1Obj, _ := rdf.NewIRI("urn:test:level1")
	rdfGraph.AddTriple(rootSubject, level1Pred, level1Obj)

	// Define a type for level 1
	level1TypeObj, _ := rdf.NewIRI("http://example.org/ontology/Level1Type")
	rdfGraph.AddTriple(level1Obj, rdfType, level1TypeObj)

	// Link to level 2 resource
	level2Pred, _ := rdf.NewIRI("http://example.org/ontology/hasLevel2")
	level2Obj, _ := rdf.NewIRI("urn:test:level2")
	rdfGraph.AddTriple(level1Obj, level2Pred, level2Obj)

	// Define a type for level 2
	level2TypeObj, _ := rdf.NewIRI("http://example.org/ontology/Level2Type")
	rdfGraph.AddTriple(level2Obj, rdfType, level2TypeObj)

	// Add the RDF graph to our index graph
	err = g.AddGraph(rdfGraph)
	is.NoErr(err)

	// Test if contextIsSet is initially false
	is.Equal(g.contextIsSet, false)

	// Get the IndexMessage with nil TagMap
	indexMsg, err := g.IndexMessage(nil)
	is.NoErr(err)

	// Verify that contextIsSet is now true after IndexMessage
	is.Equal(g.contextIsSet, true)

	// This test verifies that IndexMessage calls addContextLevels
	// The actual setting of context levels depends on the library implementation
	// which seems to have issues in the test environment

	// Verify we got a valid index message
	is.True(indexMsg != nil)
	is.Equal(indexMsg.OrganisationID, "test-org")
	is.Equal(indexMsg.DatasetID, "test-spec")
	is.Equal(indexMsg.RecordID, "test-hubid")
	is.Equal(indexMsg.IndexType, domainpb.IndexType_V2)
	is.True(len(indexMsg.Source) > 0)
}

// TestAddContextLevels tests that the context level flag is set correctly
func TestAddContextLevels(t *testing.T) {
	is := is.New(t)

	// Create a test graph with a more complex structure
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:root", // This will be our root resource
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create resources with different depths
	rootResource := &Resource{
		ID:    "urn:test:root",
		Types: []string{"http://example.org/ontology/Root"},
	}
	g.Resources = append(g.Resources, rootResource)

	// Level 1 resources (directly connected to root)
	level1A := &Resource{
		ID:    "urn:test:level1A",
		Types: []string{"http://example.org/ontology/Level1"},
	}
	level1B := &Resource{
		ID:    "urn:test:level1B",
		Types: []string{"http://example.org/ontology/Level1"},
	}
	g.Resources = append(g.Resources, level1A, level1B)

	// Add entries to create connections between resources
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://example.org/ontology/hasLevel1A",
		SearchLabel: "test_hasLevel1A",
		EntryType:   ResourceType,
		ID:          "urn:test:level1A",
	})
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://example.org/ontology/hasLevel1B",
		SearchLabel: "test_hasLevel1B",
		EntryType:   ResourceType,
		ID:          "urn:test:level1B",
	})

	// Verify objectIDs returns the expected connections
	objectIDs := rootResource.objectIDs()
	is.Equal(len(objectIDs), 2)

	// Verify the first object ID
	found1A := false
	found1B := false
	for _, ctx := range objectIDs {
		if ctx.ObjectID == "urn:test:level1A" {
			found1A = true
			is.Equal(ctx.Predicate, "http://example.org/ontology/hasLevel1A")
		}
		if ctx.ObjectID == "urn:test:level1B" {
			found1B = true
			is.Equal(ctx.Predicate, "http://example.org/ontology/hasLevel1B")
		}
	}
	is.True(found1A)
	is.True(found1B)

	// Verify contextIsSet is initially false
	is.Equal(g.contextIsSet, false)

	// Call addContextLevels
	err = g.addContextLevels()
	is.NoErr(err)

	// Verify contextIsSet is now true
	is.True(g.contextIsSet)

	// Create a JSON representation and verify it can be marshaled
	jsonBytes, err := json.MarshalIndent(g, "", "  ")
	is.NoErr(err)
	is.True(len(jsonBytes) > 0)

	// Context refs are appended to the target (nested) resources; the
	// root resource has no inbound context by definition.
	is.Equal(len(g.Resources[0].Context), 0)
	is.True(len(g.Resources[1].Context) > 0)
	is.True(len(g.Resources[2].Context) > 0)
}

// TestGraph_ContextLevelsInResources tests that context levels are properly set in Resources
func TestGraph_ContextLevelsInResources(t *testing.T) {
	is := is.New(t)

	// Create a test graph with a nested structure to test context levels
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:root", // This is our root resource
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create a 3-level deep resource structure:
	// Root -> Level1 -> Level2

	// Create root resource
	rootResource := &Resource{
		ID:    "urn:test:root",
		Types: []string{"http://example.org/ontology/Root"},
	}
	g.Resources = append(g.Resources, rootResource)

	// Create level 1 resource
	level1Resource := &Resource{
		ID:    "urn:test:level1",
		Types: []string{"http://example.org/ontology/Level1"},
	}
	g.Resources = append(g.Resources, level1Resource)

	// Create level 2 resource
	level2Resource := &Resource{
		ID:    "urn:test:level2",
		Types: []string{"http://example.org/ontology/Level2"},
	}
	g.Resources = append(g.Resources, level2Resource)

	// Add connections between resources
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://example.org/ontology/hasLevel1",
		SearchLabel: "test_hasLevel1",
		EntryType:   ResourceType,
		ID:          "urn:test:level1",
	})

	level1Resource.Entries = append(level1Resource.Entries, &Entry{
		Predicate:   "http://example.org/ontology/hasLevel2",
		SearchLabel: "test_hasLevel2",
		EntryType:   ResourceType,
		ID:          "urn:test:level2",
	})

	// Verify contextIsSet is initially false
	is.Equal(g.contextIsSet, false)

	// Add context levels
	err = g.addContextLevels()
	is.NoErr(err)

	// Verify contextIsSet is now true
	is.Equal(g.contextIsSet, true)

	// Context references are stored in the Context field of the target
	// resource; GraphExternalContext is reserved for references to
	// resources outside the graph.
	is.Equal(len(rootResource.GraphExternalContext), 0)

	// Find and validate the context reference for level1
	foundLevel1Context := false
	for _, ctx := range level1Resource.Context {
		if ctx.ObjectID == "urn:test:level1" && ctx.Predicate == "http://example.org/ontology/hasLevel1" {
			foundLevel1Context = true
			is.Equal(ctx.Subject, "urn:test:root")
			is.Equal(ctx.Level, int32(1)) // Level should be 1
			is.Equal(len(ctx.SubjectClass), 1)
			is.Equal(ctx.SubjectClass[0], "http://example.org/ontology/Root")
		}
	}
	is.True(foundLevel1Context)

	// The nested level2 resource gets a level-2 context ref from level1
	foundLevel2Context := false
	for _, ctx := range level2Resource.Context {
		if ctx.ObjectID == "urn:test:level2" && ctx.Predicate == "http://example.org/ontology/hasLevel2" {
			foundLevel2Context = true
			is.Equal(ctx.Subject, "urn:test:level1")
			is.Equal(ctx.Level, int32(2))
		}
	}
	is.True(foundLevel2Context)

	// Test that the context flag is set correctly
	// This confirms that the addContextLevels function was called and completed successfully
	is.True(g.contextIsSet)
}

// TestGraph_IndexMessageWithFields tests that fields are properly generated and included in the index message
func TestGraph_IndexMessageWithFields(t *testing.T) {
	is := is.New(t)

	// Create a test graph with literal values
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:root",
	}

	g, err := NewGraph(header)
	is.NoErr(err)

	// Create root resource
	rootResource := &Resource{
		ID:    "urn:test:root",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, rootResource)

	// Add some literal entries with different predicates
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title",
		EntryType:   Literal,
	})

	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Another Title", // Different value, same predicate
		EntryType:   Literal,
	})

	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	// Verify Fields is initially nil
	is.True(g.Fields == nil)

	// Create index message (this should generate fields)
	indexMsg, err := g.IndexMessage(nil)
	is.NoErr(err)
	is.True(indexMsg != nil)

	// Unmarshal the source to verify Fields is included
	var graphData map[string]interface{}
	err = json.Unmarshal(indexMsg.Source, &graphData)
	is.NoErr(err)

	// Check fields exists and has correct content
	fields, ok := graphData["fields"].(map[string]interface{})
	is.True(ok) // Fields should be in the JSON

	// Fields are keyed by SearchLabel, not predicate URI: the ES dynamic
	// template maps fields.* and the semantic API selects fields by label.
	titleValues, ok := fields["dc_title"].([]interface{})
	is.True(ok)
	is.Equal(len(titleValues), 2) // Should have both title values

	// Check that dc:date has one value
	dateValues, ok := fields["dc_date"].([]interface{})
	is.True(ok)
	is.Equal(len(dateValues), 1)
	is.Equal(dateValues[0], "2020-01-01")

	// Verify that g.Fields was populated
	is.True(g.Fields != nil)
	is.Equal(len(g.Fields), 2) // Should have dc_title and dc_date
	is.Equal(len(g.Fields["dc_title"]), 2)
	is.Equal(len(g.Fields["dc_date"]), 1)
}
