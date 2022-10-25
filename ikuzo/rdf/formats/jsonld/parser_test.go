package jsonld

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/matryer/is"
)

func getReader(testname string) (r io.Reader, err error) {
	path := "./testdata/" + testname + ".jsonld"

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// nolint:gocritic
func TestParse(t *testing.T) {
	t.Run("parse jsonld with graph", func(t *testing.T) {
		is := is.New(t)

		g := rdf.NewGraph()
		is.Equal(g.Len(), 0)
		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)
	})

	t.Run("parse jsonld without graph", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, nil)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 47)
	})
}

func TestParseWithContext(t *testing.T) {
	t.Run("parse with external context", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("with_context")
		is.NoErr(err)

		returnedGraph, err := ParseWithContext(r, nil)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 85)
	})
}

func TestParseWithCollections(t *testing.T) {
	t.Run("collections with blank nodes", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("with_collections")
		is.NoErr(err)

		returnedGraph, err := Parse(r, nil)
		is.NoErr(err)

		stats := returnedGraph.Stats()
		is.Equal(stats.Predicates, 25)

		var b bytes.Buffer
		err = ntriples.Serialize(returnedGraph, &b)
		is.NoErr(err)

		t.Logf("triples: \n %s", b.String())
		is.Equal(returnedGraph.Len(), 47)
	})
}
