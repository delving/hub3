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

		is.Equal(returnedGraph.Len(), 48)
		expected := &rdf.GraphStats{
			Languages: 1, ObjectIRIs: 14, Predicates: 42, Resources: 5, Triples: 48,
			Namespaces: 10,
		}
		if diff := cmp.Diff(expected, g.Stats()); diff != "" {
			t.Errorf("graphStats = mismatch (-want +got):\n%s", diff)
		}

		namespaces, err := g.Namespaces()
		is.NoErr(err)
		t.Logf("namespaces: %v", namespaces)
		is.Equal(len(namespaces), 10)
	})
}

func TestGetAboutURI(t *testing.T) {
	const (
		aggregation = "http://www.openarchives.org/ore/terms/Aggregation"
		museum      = "http://example.org/ace/Museum"
		archive     = "http://example.org/ace/Archive"
		library     = "http://example.org/ace/Library"
	)

	mustIRI := func(t *testing.T, s string) rdf.IRI {
		t.Helper()
		iri, err := rdf.NewIRI(s)
		if err != nil {
			t.Fatalf("NewIRI(%q): %v", s, err)
		}
		return iri
	}

	addType := func(t *testing.T, g *rdf.Graph, subject, typeURI string) {
		t.Helper()
		g.AddTriple(mustIRI(t, subject), rdf.IsA, mustIRI(t, typeURI))
	}

	addLink := func(t *testing.T, g *rdf.Graph, subject, predicate, object string) {
		t.Helper()
		g.AddTriple(mustIRI(t, subject), rdf.Predicate(mustIRI(t, predicate)), mustIRI(t, object))
	}

	t.Run("returns first matching configured type", func(t *testing.T) {
		g := rdf.NewGraph()
		addType(t, g, "http://example.org/record/1", aggregation)

		got, err := g.GetAboutURI([]string{aggregation})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://example.org/record/1" {
			t.Errorf("got %q, want record/1", got)
		}
	})

	t.Run("ace recDef picks first declared type when several match", func(t *testing.T) {
		// Simulate an 'ace' record where the same graph happens to contain
		// both an Archive and a Museum subject. Declaration order in the
		// configured slice should win.
		g := rdf.NewGraph()
		addType(t, g, "http://example.org/ace/m1", museum)
		addType(t, g, "http://example.org/ace/a1", archive)

		got, err := g.GetAboutURI([]string{archive, museum, library})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://example.org/ace/a1" {
			t.Errorf("got %q, want ace/a1 (archive declared first)", got)
		}
	})

	t.Run("structural fallback when no configured type matches", func(t *testing.T) {
		// Root has an outgoing link to the child; child has no outgoing
		// link back. Root therefore has in-degree 0 and should win even
		// though no rdf:type matches the configured aboutType.
		g := rdf.NewGraph()
		addLink(t, g, "http://example.org/ace/root", "http://example.org/ace/hasPart", "http://example.org/ace/child")
		addType(t, g, "http://example.org/ace/root", "http://example.org/ace/Unknown")
		addType(t, g, "http://example.org/ace/child", "http://example.org/ace/Unknown")

		got, err := g.GetAboutURI([]string{aggregation})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://example.org/ace/root" {
			t.Errorf("got %q, want ace/root via structural fallback", got)
		}
	})

	t.Run("structural fallback prefers higher out-degree on tie", func(t *testing.T) {
		// Two zero-in-degree subjects. The one with more outgoing triples
		// is the more likely top-level resource.
		g := rdf.NewGraph()
		addType(t, g, "http://example.org/sparse", "http://example.org/T")
		addType(t, g, "http://example.org/rich", "http://example.org/T")
		addLink(t, g, "http://example.org/rich", "http://example.org/p1", "http://example.org/x")
		addLink(t, g, "http://example.org/rich", "http://example.org/p2", "http://example.org/y")

		got, err := g.GetAboutURI([]string{aggregation})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://example.org/rich" {
			t.Errorf("got %q, want rich (highest out-degree)", got)
		}
	})

	t.Run("returns error on empty graph", func(t *testing.T) {
		g := rdf.NewGraph()
		if _, err := g.GetAboutURI([]string{aggregation}); err == nil {
			t.Fatalf("expected error for empty graph, got nil")
		}
	})
}
