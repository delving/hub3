package sparql

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"

	"github.com/delving/hub3/ikuzo/rdf"
)

// nolint: gocritic
func TestReadResponse(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/query.json")
	is.NoErr(err)

	resp, err := newResponse(f)
	is.NoErr(err)
	is.True(resp != nil)
	is.Equal(len(resp.Results.Bindings), 24)

	// test entry
	binding := resp.Results.Bindings[15]
	is.True(binding != nil)
	is.Equal(binding.S1.Type, TypeURI)
	is.Equal(binding.S1.Value, "https://klek.si/208B7R")
	is.Equal(binding.S1.XMLLang, "")
	is.Equal(binding.O1.Type, TypeBnode)
	is.Equal(binding.O1.Value, "b2")
	is.Equal(binding.O2.Value, "Vitrine 8")
	is.Equal(binding.O2.XMLLang, "nl")
	is.Equal(binding.O2.Type, TypeLiteral)
	is.Equal(binding.P2.Value, "https://schema.org/name")

	// test ntriples
	triples, err := resp.NTriples()
	is.NoErr(err)
	// t.Log(triples)

	// should be the same as golden file
	b, err := os.ReadFile("./testdata/query_ntriples.golden.nt")
	is.NoErr(err)
	is.Equal(triples, string(b))

	iri, err := rdf.NewIRI("https://klek.si/208B7R")
	is.NoErr(err)
	xml, err := resp.MappingXML(rdf.Subject(iri), "")
	is.NoErr(err)

	expected, err := os.ReadFile("./testdata/query_xml.golden.xml")
	is.NoErr(err)
	t.Logf(xml)
	if diff := cmp.Diff(string(expected), xml); diff != "" {
		t.Errorf("mappingXML = mismatch (-want +got):\n%s", diff)
	}
}
