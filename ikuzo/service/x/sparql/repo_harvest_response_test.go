package sparql

import (
	fmt "fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/matryer/is"
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
	is.Equal(binding.S.Type, TypeURI)
	is.Equal(binding.S.Value, "https://klek.si/208B7R")
	is.Equal(binding.S.XMLLang, "")
	is.Equal(binding.O.Type, TypeBnode)
	is.Equal(binding.O.Value, "b2")
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

	// iri, err := rdf.NewIRI("https://klek.si/208B7R")
	// is.NoErr(err)
	// xml, err := resp.MappingXML(rdf.Subject(iri))
	// is.NoErr(err)
	// expected := `<hello>`
	// is.Equal(xml, expected)
}

// //nolint: gocritic
// func TestWikibaseResponse(t *testing.T) {
// is := is.New(t)

// f, err := os.Open("./testdata/wikibase.json")
// is.NoErr(err)

// resp, err := NewResponse(f)
// is.NoErr(err)
// is.True(resp != nil)
// is.Equal(len(resp.Results.Bindings), 387)

// // test entry
// // test ntriples
// triples, err := resp.NTriples()
// is.NoErr(err)
// t.Log(triples)

// // should be the same as golden file
// // b, err := os.ReadFile("./testdata/query_ntriples.golden.nt")
// // is.NoErr(err)
// // is.Equal(triples, string(b))

// // iri, err := rdf.NewIRI("http://gebouwen.brabantcloud.nl/entity/Q2452")
// // is.NoErr(err)
// // xml, err := resp.MappingXML(rdf.Subject(iri))
// // is.NoErr(err)
// // expected := `<hello>`
// // is.Equal(xml, expected)
// }

// TestMappingXML is a specific test for BrabantCloud wikibase data
//
// Once this format is accepted it should be removed and replaced with a
// more general purpose test.
func TestMappingXML(t *testing.T) {
	is := is.New(t)

	jsonPath := "./testdata/wikibase"

	output, err := os.Create(filepath.Join(os.TempDir(), "output.xml"))
	is.NoErr(err)
	fmt.Fprintln(output, "<wrapped>")

	files, err := os.ReadDir(jsonPath)
	is.NoErr(err)

	for _, fname := range files {
		if !strings.HasSuffix(fname.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(fname.Name(), ".json")
		t.Logf("fname: %s ; id %s", fname.Name(), id)

		fmt.Fprintf(output, "<record id=\"%s\">\n", id)

		f, err := os.Open(filepath.Join(jsonPath, fname.Name()))
		is.NoErr(err)

		resp, err := newResponse(f)
		is.NoErr(err)

		subj, err := rdf.NewIRI(
			fmt.Sprintf(
				"http://gebouwen.brabantcloud.nl/entity/%s",
				id,
			),
		)
		is.NoErr(err)
		if len(resp.Results.Bindings) > 200 {
			t.Logf("subject %s; triples %d", subj, len(resp.Results.Bindings))
		}

		f.Close()

		xml, err := resp.MappingXML(
			rdf.Subject(subj),
			"http://gebouwen.brabantcloud.nl/prop/direct/P1",
		)
		is.NoErr(err)
		fmt.Fprintln(output, xml)
		fmt.Fprintln(output, "</record>")
	}

	fmt.Fprintln(output, "</wrapped>")
}
