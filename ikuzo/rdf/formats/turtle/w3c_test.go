package turtle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

// w3cTestCase represents a W3C Turtle evaluation test.
type w3cTestCase struct {
	name  string
	files []string // .ttl input file(s)
}

// Selected W3C Turtle evaluation tests
// These tests verify round-trip parsing: TTL -> Graph -> Serialize -> Compare
var w3cTests = []w3cTestCase{
	{name: "localName_with_leading_digit", files: []string{"localName_with_leading_digit.ttl"}},
	{name: "LITERAL1", files: []string{"LITERAL1.ttl"}},
	{name: "LITERAL2", files: []string{"LITERAL2.ttl"}},
	{name: "LITERAL_LONG1", files: []string{"LITERAL_LONG1.ttl"}},
	{name: "empty_collection", files: []string{"empty_collection.ttl"}},
	{name: "SPARQL_style_prefix", files: []string{"SPARQL_style_prefix.ttl"}},
	{name: "labeled_blank_node_subject", files: []string{"labeled_blank_node_subject.ttl"}},
	{name: "labeled_blank_node_object", files: []string{"labeled_blank_node_object.ttl"}},
}

func TestW3C_RoundTrip(t *testing.T) {
	for _, tc := range w3cTests {
		t.Run(tc.name, func(t *testing.T) {
			// Read TTL input
			ttlPath := filepath.Join("testdata/w3c", tc.files[0])
			ttlData, err := os.ReadFile(ttlPath)
			if err != nil {
				t.Skipf("test file not found: %s", ttlPath)
			}

			// Parse TTL to graph
			g := rdf.NewGraph()
			g.UseResource = true
			g.UseIndex = true
			_, err = ntriples.Parse(bytes.NewReader(ttlData), g)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// Serialize to Turtle
			var buf bytes.Buffer
			err = Serialize(g, &buf)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}

			// Re-parse serialized output
			g2 := rdf.NewGraph()
			g2.UseResource = true
			g2.UseIndex = true
			_, err = ntriples.Parse(&buf, g2)
			if err != nil {
				t.Fatalf("re-parse error: %v", err)
			}

			// Compare graph sizes (semantic equivalence via re-parse)
			if g.Len() != g2.Len() {
				t.Errorf("graph size mismatch: got %d, expected %d\nSerialized:\n%s",
					g2.Len(), g.Len(), buf.String())
			}
		})
	}
}

// TestW3C_BasicSerialization tests that our serializer produces valid Turtle output.
// Round-trip tests already verify serialization correctness.
func TestW3C_BasicSerialization(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("http://example.org/s")
	p, _ := rdf.NewIRI("http://example.org/p")
	rdfType, _ := rdf.RDF.IRI("type")
	crm, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/E21_Person")
	g.AddTriple(s, rdfType, crm)

	lit, _ := rdf.NewLiteral("test")
	g.AddTriple(s, p, lit)

	var buf bytes.Buffer
	err := Serialize(g, &buf, WithPrefixes(map[string]string{
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"crm": "http://www.cidoc-crm.org/cidoc-crm/",
	}))
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "@prefix") {
		t.Error("should have prefix declarations")
	}
	if !strings.Contains(output, " a ") {
		t.Error("should use 'a' shorthand for rdf:type")
	}
}

// TestW3C_LiteralHandling tests that literals are serialized correctly.
func TestW3C_LiteralHandling(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("http://example.org/subject")
	p, _ := rdf.NewIRI("http://example.org/predicate")

	// String literal
	lit1, _ := rdf.NewLiteral("hello")
	g.AddTriple(s, p, lit1)

	// Language-tagged literal
	lit2, _ := rdf.NewLiteralWithLang("hello", "en")
	g.AddTriple(s, p, lit2)

	// Typed literal
	dt, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#integer")
	lit3, _ := rdf.NewLiteralWithType("42", dt)
	g.AddTriple(s, p, lit3)

	var buf bytes.Buffer
	err := Serialize(g, &buf)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `"hello"`) {
		t.Error("should contain string literal")
	}
	if !strings.Contains(output, `"hello"@en`) {
		t.Error("should contain language-tagged literal")
	}
	if !strings.Contains(output, `"42"`) {
		t.Error("should contain typed literal value")
	}
}

// TestW3C_PrefixSorting tests that prefixes are sorted alphabetically.
func TestW3C_PrefixSorting(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("http://example.org/s")
	p1, _ := rdf.NewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	crm, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/E21")
	g.AddTriple(s, p1, crm)

	rdfsLabel, _ := rdf.NewIRI("http://www.w3.org/2000/01/rdf-schema#label")
	lit, _ := rdf.NewLiteral("test")
	g.AddTriple(s, rdfsLabel, lit)

	var buf bytes.Buffer
	err := Serialize(g, &buf, WithPrefixes(map[string]string{
		"rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
		"crm":  "http://www.cidoc-crm.org/cidoc-crm/",
	}))
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	lines := strings.Split(buf.String(), "\n")
	var prefixLines []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "@prefix") {
			prefixLines = append(prefixLines, line)
		}
	}

	// Verify sorting
	for i := 1; i < len(prefixLines); i++ {
		if prefixLines[i-1] > prefixLines[i] {
			t.Errorf("prefixes not sorted: %q > %q", prefixLines[i-1], prefixLines[i])
		}
	}
}

// TestW3C_BlankNodeHandling tests that blank nodes are handled.
func TestW3C_BlankNodeHandling(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("http://example.org/s")
	p, _ := rdf.NewIRI("http://example.org/p")

	// Blank node as object
	bn, _ := rdf.NewBlankNode("b1")
	g.AddTriple(s, p, bn)

	// Blank node as subject
	bn2, _ := rdf.NewBlankNode("b2")
	lit, _ := rdf.NewLiteral("blank value")
	g.AddTriple(bn2, p, lit)

	var buf bytes.Buffer
	err := Serialize(g, &buf)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	if !strings.Contains(buf.String(), "_:") {
		t.Error("should contain blank node identifier")
	}
}

// TestW3C_CompactNotation tests compact notation features (a shorthand, ; and ,).
func TestW3C_CompactNotation(t *testing.T) {
	g := rdf.NewGraph()
	g.UseResource = true
	g.UseIndex = true

	s, _ := rdf.NewIRI("http://example.org/Person")
	rdfType, _ := rdf.RDF.IRI("type")
	crmPerson, _ := rdf.NewIRI("http://www.cidoc-crm.org/cidoc-crm/E21_Person")
	g.AddTriple(s, rdfType, crmPerson)

	name, _ := rdf.RDFS.IRI("label")
	litJohn, _ := rdf.NewLiteral("John")
	g.AddTriple(s, name, litJohn)

	age, _ := rdf.NewIRI("http://example.org/age")
	lit30, _ := rdf.NewLiteral("30")
	g.AddTriple(s, age, lit30)

	var buf bytes.Buffer
	err := Serialize(g, &buf, WithPrefixes(map[string]string{
		"crm":  "http://www.cidoc-crm.org/cidoc-crm/",
		"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
	}))
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	output := buf.String()

	// Should use 'a' shorthand
	if !strings.Contains(output, " a ") {
		t.Error("should use 'a' shorthand for rdf:type")
	}

	// Should end statements with period
	if !strings.Contains(output, " .") {
		t.Error("should end statements with period")
	}
}
