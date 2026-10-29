package fragments

import (
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
)

func TestAddToBlankNodeSubject(t *testing.T) {
	fr := &FragmentResource{ID: "_:b1", Types: []string{"http://example.org/T"}}

	g := rdf.NewGraph()
	if err := fr.AddTo(g); err != nil {
		t.Fatal(err)
	}

	got := g.Triples()[0].Subject.String()
	if got != "_:b1" {
		t.Errorf("blank node subject serialized as %q, want _:b1", got)
	}
}
