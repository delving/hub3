package resource

import (
	"testing"

	"github.com/matryer/is"
)

func TestGraph(t *testing.T) {
	t.Run("NewGraph", func(t *testing.T) {
		// nolint:gocritic
		is := is.New(t)

		g := NewGraph()
		is.Equal(len(g.triples), 0)
		is.Equal(g.Len(), len(g.triples))

		// build triple
		s, err := NewIRI("urn:s/123")
		is.NoErr(err)

		p, err := DC.IRI("subject")
		is.NoErr(err)

		o, err := NewLiteralWithLang("some text", "en")
		is.NoErr(err)

		triple := NewTriple(s, p, o)

		g.Add(triple)
		is.Equal(g.Len(), 1)
		is.Equal(g.Len(), len(g.Triples()))

		triples, err := g.TriplesOnce()
		is.NoErr(err)
		is.Equal(len(triples), 1)

		g.Add(triple)
		triples, err = g.TriplesOnce()
		is.True(err != nil)
		is.Equal(len(triples), 0)
	})
}
