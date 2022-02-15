package ntriples

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestSerialize(t *testing.T) {
	is := is.New(t)

	b, err := os.ReadFile("./testdata/rdf.nt")
	is.NoErr(err)

	r := bytes.NewReader(b)

	g, err := Parse(r, nil)
	is.NoErr(err)

	var buf bytes.Buffer
	err = Serialize(g, &buf)
	is.NoErr(err)

	if diff := cmp.Diff(string(b), buf.String()); diff != "" {
		t.Errorf("serialize = mismatch (-want +got):\n%s", diff)
	}
}

func TestSerializeFiltered(t *testing.T) {
	is := is.New(t)

	b := rdf.Builder{}

	g := rdf.NewGraph()
	g.AddTriple(
		b.IRI("urn:subject"),
		b.IRI("http://purl.org/dc/elements/1.1/subject"),
		b.Literal("hello"),
	)
	g.AddTriple(
		b.IRI("urn:subject"),
		b.IRI("http://www.europeana.eu/schemas/edm/hasView"),
		b.IRI("urn:private/123"),
	)
	g.AddTriple(
		b.IRI("urn:private/123"),
		b.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		b.IRI("http://www.europeana.eu/schemas/edm/WebResource"),
	)

	is.Equal(g.Len(), 3)

	var buf bytes.Buffer
	err := SerializeFiltered(g, &buf, "<urn:private")
	is.NoErr(err)

	rdf := buf.String()
	is.True(!strings.Contains(rdf, "urn:private/")) // serialized rdf should not contain urn:private
	t.Logf("rdf: %s", rdf)
	is.True(strings.HasSuffix(rdf, " .\n"))
}
