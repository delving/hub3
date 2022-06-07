package ntriples

import (
	"io"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/matryer/is"
)

func getReader(name string) (io.ReadCloser, error) {
	f, err := os.Open("./testdata/" + name)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// nolint:gocritic
func TestParse(t *testing.T) {
	t.Run("parse ntriples with graph", func(t *testing.T) {
		is := is.New(t)

		g := rdf.NewGraph()
		is.Equal(g.Len(), 0)
		r, err := getReader("rdf.nt")
		is.NoErr(err)
		defer r.Close()
		returnedGraph, err := Parse(r, g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)
	})

	t.Run("parse ntriples without graph", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("rdf.nt")
		is.NoErr(err)
		defer r.Close()
		returnedGraph, err := Parse(r, nil)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 47)
	})
}
