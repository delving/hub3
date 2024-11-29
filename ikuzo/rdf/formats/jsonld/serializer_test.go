package jsonld

import (
	"bytes"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/matryer/is"
)

func TestSerialize(t *testing.T) {
	t.Run("serialize jsonld with graph", func(t *testing.T) {
		is := is.New(t)

		g := rdf.NewGraph()
		is.Equal(g.Len(), 0)
		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)

		var buf bytes.Buffer
		err = Serialize(g, &buf, nil)
		is.NoErr(err)

		is.True(buf.String() != "")
	})
}
