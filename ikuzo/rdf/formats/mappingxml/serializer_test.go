package mappingxml

import (
	"bytes"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestSerialize(t *testing.T) {
	var g *rdf.Graph
	t.Run("test flat", func(t *testing.T) {
		is := is.New(t)
		g = rdf.NewGraph()
		g.UseIndex = true
		g.UseResource = true

		is.Equal(g.Len(), 0)
		f, err := os.Open("testdata/rdf.nt")
		is.NoErr(err)
		defer f.Close()
		_, err = ntriples.Parse(f, g)
		is.NoErr(err)

		iri, err := rdf.SCHEMA.IRI("CreativeWork")
		is.NoErr(err)

		cfg := FilterConfig{RDFType: iri}
		var buf bytes.Buffer
		err = Serialize(g, &buf, &cfg)
		is.NoErr(err)

		b, err := os.ReadFile("./testdata/rdf.golden.xml")
		is.NoErr(err)
		if diff := cmp.Diff(string(b), buf.String()+"\n"); diff != "" {
			t.Errorf("mapping xml = mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("filterResources", func(t *testing.T) {
		is := is.New(t)

		is.Equal(g.Len(), 28)
		is.Equal(len(g.Resources()), 6)

		iri, err := rdf.SCHEMA.IRI("CreativeWork")
		is.NoErr(err)

		cfg := FilterConfig{RDFType: iri}
		filtered := filterResources(g.Resources(), &cfg)
		is.Equal(len(filtered), 2)

		s, err := rdf.NewIRI("http://klek.si/208B7R")
		is.NoErr(err)
		cfg = FilterConfig{Subject: s}
		filtered = filterResources(g.Resources(), &cfg)
		is.Equal(len(filtered), 1)
	})
}
