package semantic

import "testing"

func TestClassInfo_Basic(t *testing.T) {
	ci := ClassInfo{
		URI:   "http://www.europeana.eu/schemas/edm/ProvidedCHO",
		Label: "edm:ProvidedCHO",
		Count: 45000,
	}

	if ci.Label != "edm:ProvidedCHO" {
		t.Errorf("Label = %v, want edm:ProvidedCHO", ci.Label)
	}
}

func TestPropertyInfo_HasResolvedLabels(t *testing.T) {
	pi := PropertyInfo{
		Field:             "dc_creator",
		Predicate:         "http://purl.org/dc/elements/1.1/creator",
		Label:             "dc:creator",
		ValueTypes:        []string{"Literal", "Resource"},
		Count:             42000,
		Languages:         []string{"nl", "en"},
		HasResolvedLabels: true,
		Paths:             []string{"dc_creator", "dc_creator/foaf_name"},
	}

	if !pi.HasResolvedLabels {
		t.Error("Expected HasResolvedLabels to be true")
	}

	if len(pi.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(pi.Paths))
	}
}

func TestIntrospectionResult_WithScope(t *testing.T) {
	ir := IntrospectionResult{
		Scope: IntrospectionScope{
			Type:           "query",
			ContextID:      "ctx_a7f3",
			TotalDocuments: 12847,
		},
	}

	if ir.Scope.Type != "query" {
		t.Errorf("Scope.Type = %v, want query", ir.Scope.Type)
	}
}
