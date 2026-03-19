package turtle_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/turtle"
)

func buildTestGraph(t *testing.T) *rdf.Graph {
	t.Helper()
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	// Subject
	s, err := rdf.NewIRI("https://example.net/actor/E21")
	if err != nil {
		t.Fatal(err)
	}
	// rdf:type
	rdfType, _ := rdf.RDF.IRI("type")
	crmE21, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/E21_Person")
	g.AddTriple(s, rdfType, crmE21)

	// Property
	p1, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/P1_is_identified_by")
	o1, _ := rdf.NewIRI("https://example.net/conceptual_object/4_1")
	g.AddTriple(s, p1, o1)

	// Second resource
	crmE33E41, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/E33_E41_Linguistic_Appellation")
	g.AddTriple(o1, rdfType, crmE33E41)

	p190, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/P190_has_symbolic_content")
	lit, _ := rdf.NewLiteral("value")
	g.AddTriple(o1, p190, lit)

	return g
}

func TestSerialize_BasicGraph(t *testing.T) {
	g := buildTestGraph(t)

	var buf bytes.Buffer
	err := turtle.Serialize(g, &buf, turtle.WithPrefixes(map[string]string{
		"crm": "http://www.cidoc-crm.org/cidoc-crm/",
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	}))
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := buf.String()

	// Should contain prefix declarations
	if !strings.Contains(output, "@prefix") {
		t.Error("output should contain @prefix declarations")
	}

	// Should contain the subject
	if !strings.Contains(output, "example.net/actor/E21") {
		t.Error("output should contain the subject URI")
	}

	// Should contain rdf:type shorthand 'a'
	if !strings.Contains(output, " a ") {
		t.Error("output should use 'a' shorthand for rdf:type")
	}

	// Should contain semicolons for same-subject predicates
	if !strings.Contains(output, ";") {
		t.Error("output should use semicolons for same-subject predicates")
	}

	// Should end statements with period
	if !strings.Contains(output, " .") {
		t.Error("output should end resource blocks with period")
	}

	// Should contain literal value
	if !strings.Contains(output, `"value"`) {
		t.Error("output should contain literal value")
	}
}

func TestSerialize_EmptyGraph(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	var buf bytes.Buffer
	err := turtle.Serialize(g, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("empty graph should produce empty output, got %q", buf.String())
	}
}

func TestSerialize_WithCustomPrefixes(t *testing.T) {
	g := buildTestGraph(t)

	var buf bytes.Buffer
	err := turtle.Serialize(g, &buf, turtle.WithPrefixes(map[string]string{
		"crm": "http://www.cidoc-crm.org/cidoc-crm/",
	}))
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := buf.String()

	// Should use the custom prefix
	if !strings.Contains(output, "@prefix crm: <http://www.cidoc-crm.org/cidoc-crm/> .") {
		t.Error("output should contain custom crm prefix declaration")
	}

	// Should use prefixed names in body
	if !strings.Contains(output, "crm:E21_Person") {
		t.Errorf("output should use prefixed crm:E21_Person, got:\n%s", output)
	}

	if !strings.Contains(output, "crm:P1_is_identified_by") {
		t.Errorf("output should use prefixed crm:P1_is_identified_by, got:\n%s", output)
	}
}

func TestSerialize_PrefixesSorted(t *testing.T) {
	g := buildTestGraph(t)

	// Add a triple with rdfs namespace
	s, _ := rdf.NewIRI("https://example.net/conceptual_object/4_1")
	rdfsLabel, _ := rdf.RDFS.IRI("label")
	lit, _ := rdf.NewLiteralWithLang("Name", "en")
	g.AddTriple(s, rdfsLabel, lit)

	var buf bytes.Buffer
	err := turtle.Serialize(g, &buf, turtle.WithPrefixes(map[string]string{
		"crm":  "http://www.cidoc-crm.org/cidoc-crm/",
		"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
		"rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	}))
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Find prefix lines and verify they are sorted
	var prefixLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "@prefix") {
			prefixLines = append(prefixLines, line)
		}
	}

	if len(prefixLines) < 2 {
		t.Fatalf("expected at least 2 prefix lines, got %d", len(prefixLines))
	}

	for i := 1; i < len(prefixLines); i++ {
		if prefixLines[i-1] > prefixLines[i] {
			t.Errorf("prefixes not sorted: %q > %q", prefixLines[i-1], prefixLines[i])
		}
	}
}

func TestSerialize_LanguageTaggedLiteral(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("https://example.net/actor/E21")
	p, _ := rdf.RDFS.IRI("label")
	o, _ := rdf.NewLiteralWithLang("Person", "en")
	g.AddTriple(s, p, o)

	var buf bytes.Buffer
	err := turtle.Serialize(g, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), `"Person"@en`) {
		t.Errorf("expected language-tagged literal, got:\n%s", buf.String())
	}
}
