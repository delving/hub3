package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

// The Linked Art shapes are only test fixtures, so nothing else notices when one
// stops parsing or refers to a shape that does not exist.
func TestLinkedArtShapes(t *testing.T) {
	files, err := filepath.Glob("testdata/rdf/linkedart/shapes/*.ttl")
	if err != nil || len(files) == 0 {
		t.Fatalf("no shape files found: %v", err)
	}

	// one graph per file: gon3 restarts its blank node labels for every file
	var triples []*rdf.Triple
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		// the ntriples package parses with gon3, which reads Turtle
		g, err := ntriples.Parse(f, nil)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		f.Close()
		triples = append(triples, g.Triples()...)
	}

	const shacl = "http://www.w3.org/ns/shacl#"
	defined := map[string]bool{}
	for _, triple := range triples {
		defined[triple.Subject.RawValue()] = true
	}

	for _, triple := range triples {
		switch strings.TrimPrefix(triple.Predicate.RawValue(), shacl) {
		case "node", "property", "and", "or", "xone", "not", "qualifiedValueShape":
		default:
			continue
		}
		ref, ok := triple.Object.(rdf.IRI)
		if ok && !defined[ref.RawValue()] {
			t.Errorf("%s %s %s: undefined shape", triple.Subject, triple.Predicate, triple.Object)
		}
	}
}
