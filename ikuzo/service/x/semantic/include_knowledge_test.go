package semantic

import (
	"context"
	"testing"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestKnowledgeContextProvider_Name(t *testing.T) {
	p := NewKnowledgeContextProvider()
	if p.Name() != "knowledgeContext" {
		t.Errorf("Name() = %q, want %q", p.Name(), "knowledgeContext")
	}
}

func TestKnowledgeContextProvider_NilDoc(t *testing.T) {
	p := NewKnowledgeContextProvider()
	result, err := p.Provide(context.Background(), &domainSemantic.IncludeRequest{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil doc, got %v", result)
	}
}

func TestKnowledgeContextProvider_EmptyDoc(t *testing.T) {
	p := NewKnowledgeContextProvider()
	result, err := p.Provide(context.Background(), &domainSemantic.IncludeRequest{
		Document: map[string]any{"@id": "test", "@type": "edm:ProvidedCHO"},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	// No entities found, should return nil
	if result != nil {
		t.Errorf("expected nil for doc without entities, got %v", result)
	}
}

func TestKnowledgeContextProvider_ExtractsAgents(t *testing.T) {
	p := NewKnowledgeContextProvider()

	doc := map[string]any{
		"@id":   "test-doc",
		"@type": "edm:ProvidedCHO",
		"dc:creator": map[string]any{
			"@id":            "http://example.org/agent/rembrandt",
			"@type":          "edm:Agent",
			"skos:prefLabel": "Rembrandt van Rijn",
		},
	}

	result, err := p.Provide(context.Background(), &domainSemantic.IncludeRequest{
		DocumentID: "test-doc",
		Document:   doc,
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	panel, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result is %T, want map[string]any", result)
	}

	if panel["@type"] != "hub3:KnowledgePanel" {
		t.Errorf("@type = %v, want hub3:KnowledgePanel", panel["@type"])
	}

	agents, ok := panel["hub3:agents"]
	if !ok {
		t.Fatal("expected hub3:agents in panel")
	}

	agentList, ok := agents.([]map[string]any)
	if !ok {
		t.Fatalf("agents is %T, want []map[string]any", agents)
	}

	if len(agentList) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agentList))
	}

	if agentList[0]["skos:prefLabel"] != "Rembrandt van Rijn" {
		t.Errorf("label = %v, want 'Rembrandt van Rijn'", agentList[0]["skos:prefLabel"])
	}
}

func TestKnowledgeContextProvider_MultipleTypes(t *testing.T) {
	p := NewKnowledgeContextProvider()

	doc := map[string]any{
		"@id":   "test-doc",
		"@type": "edm:ProvidedCHO",
		"dc:creator": map[string]any{
			"@id":            "http://example.org/agent/rembrandt",
			"@type":          "edm:Agent",
			"skos:prefLabel": "Rembrandt",
		},
		"dc:spatial": map[string]any{
			"@id":            "http://example.org/place/amsterdam",
			"@type":          "edm:Place",
			"skos:prefLabel": "Amsterdam",
		},
		"dc:subject": []any{
			map[string]any{
				"@id":            "http://example.org/concept/painting",
				"@type":          "skos:Concept",
				"skos:prefLabel": "Oil painting",
			},
		},
	}

	result, err := p.Provide(context.Background(), &domainSemantic.IncludeRequest{
		DocumentID: "test-doc",
		Document:   doc,
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	panel, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result is %T", result)
	}

	// Should have agents, places, and concepts
	if _, ok := panel["hub3:agents"]; !ok {
		t.Error("expected hub3:agents")
	}
	if _, ok := panel["hub3:places"]; !ok {
		t.Error("expected hub3:places")
	}
	if _, ok := panel["hub3:concepts"]; !ok {
		t.Error("expected hub3:concepts")
	}
	// No timespans
	if _, ok := panel["hub3:timespans"]; ok {
		t.Error("unexpected hub3:timespans")
	}
}
