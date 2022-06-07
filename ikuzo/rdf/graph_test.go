package rdf_test

import (
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestGraph(t *testing.T) {
	t.Run("NewGraph", func(t *testing.T) {
		// nolint:gocritic
		is := is.New(t)

		g := rdf.NewGraph()
		g.UseIndex = true
		is.Equal(g.Len(), 0)

		// build triple
		s, err := rdf.NewIRI("urn:s/123")
		is.NoErr(err)

		p, err := rdf.DC.IRI("subject")
		is.NoErr(err)

		o, err := rdf.NewLiteralWithLang("some text", "en")
		is.NoErr(err)

		triple := rdf.NewTriple(s, p, o)

		g.Add(triple)
		is.Equal(g.Len(), 1)

		// same triple should not be added to the graph
		g.Add(triple)
		is.Equal(g.Len(), 1)

		triples, err := g.TriplesOnce()
		is.NoErr(err)
		is.Equal(len(triples), 1)
		is.Equal(g.Len(), len(g.Triples()))

		g.Add(triple)
		triples, err = g.TriplesOnce()
		is.True(err != nil) // should throw error
		is.Equal(len(triples), 0)

		// test stats
		stats := g.Stats()
		is.True(stats != nil)
		is.Equal(stats.Triples, uint64(1))
		is.Equal(stats.Languages, 1)
	})

	t.Run("test with index", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./formats/ntriples/testdata/rdf.nt")
		is.NoErr(err)

		g := rdf.NewGraph()
		g.UseIndex = true

		returnedGraph, err := ntriples.Parse(f, g)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 47)
		expected := &rdf.GraphStats{
			Languages: 1, ObjectIRIs: 14, Predicates: 41, Resources: 5, Triples: 47,
			Namespaces: 9,
		}
		if diff := cmp.Diff(expected, g.Stats()); diff != "" {
			t.Errorf("graphStats = mismatch (-want +got):\n%s", diff)
		}

		namespaces, err := g.Namespaces()
		is.NoErr(err)
		t.Logf("namespaces: %v", namespaces)
		is.Equal(len(namespaces), 9)
	})
}
