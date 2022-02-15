package sparql

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

//nolint: gocritic
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

// func TestMappingXML(t *testing.T) {
// is := is.New(t)

// jsonPath := "/home/kiivihal/projects/01_active/eb/k3-harvest/records"

// output, err := os.Create(filepath.Join(jsonPath, "output.xml"))
// is.NoErr(err)
// fmt.Fprintln(output, "<wrapped>")

// files, err := os.ReadDir(jsonPath)
// is.NoErr(err)

// for _, fname := range files {
// if !strings.HasSuffix(fname.Name(), ".json") {
// continue
// }

// t.Logf("fname: %s", fname.Name())

// f, err := os.Open(filepath.Join(jsonPath, fname.Name()))
// is.NoErr(err)

// resp, err := NewResponse(f)
// is.NoErr(err)

// subj, err := rdf.NewIRI(
// fmt.Sprintf(
// "http://gebouwen.brabantcloud.nl/entity/%s",
// strings.TrimSuffix(fname.Name(), ".json"),
// ),
// )
// is.NoErr(err)
// t.Logf("subject %s; triples %d", subj, len(resp.Results.Bindings))

// f.Close()

// xml, err := resp.MappingXML(rdf.Subject(subj))
// is.NoErr(err)
// fmt.Fprintln(output, xml)
// }

// fmt.Fprintln(output, "</wrapped>")

// is.True(false)

// // loop over all files
// // create file
// // write wrapped
// // construct iri
// // open file
// // create response
// // createMapping xml
// // write to file
// // close with </wrapped>
// // upload to narthex as file
// }
