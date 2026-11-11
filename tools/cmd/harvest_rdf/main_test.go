package main

import (
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
)

func TestWriteRecord(t *testing.T) {
	record := func(t *testing.T) *rdf.Graph {
		t.Helper()
		s, _ := rdf.NewIRI("http://example.org/doc/1")
		p, _ := rdf.NewIRI("http://example.org/p")
		b, _ := rdf.NewBlankNode("b0")
		lit, _ := rdf.NewLiteral("x")
		priv, _ := rdf.NewIRI("urn:private/1")

		g := rdf.NewGraph()
		g.AddTriple(s, p, b)
		g.AddTriple(b, p, lit)
		g.AddTriple(s, p, priv)
		return g
	}

	var sb strings.Builder
	if err := writeRecord(&sb, record(t), "hub_1", "<urn:hub_1/graph>"); err != nil {
		t.Fatal(err)
	}
	if err := writeRecord(&sb, record(t), "hub_2", ""); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(sb.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("want 4 lines (private triples dropped), got %d:\n%s", len(lines), sb.String())
	}

	if f := strings.Fields(lines[0]); len(f) != 5 || f[3] != "<urn:hub_1/graph>" {
		t.Errorf("quad line missing graph name: %s", lines[0])
	}
	if f := strings.Fields(lines[2]); len(f) != 4 || f[3] != "." {
		t.Errorf("triple line should have no graph name: %s", lines[2])
	}

	bnode := func(line string) string { return strings.Fields(line)[2] }
	if b1, b2 := bnode(lines[0]), bnode(lines[2]); b1 == b2 || !strings.HasPrefix(b1, "_:r") {
		t.Errorf("blank nodes of different records must differ: %s vs %s", b1, b2)
	}
	if bnode(lines[0]) != strings.Fields(lines[1])[0] {
		t.Errorf("blank node must stay the same within a record: %s / %s", lines[0], lines[1])
	}
}
