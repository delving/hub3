package bulk

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kiivihal/rdf2go"
	"github.com/matryer/is"
)

func TestSerializeTurtle(t *testing.T) {
	is := is.New(t)

	g := rdf2go.NewGraph("")
	g.AddTriple(
		rdf2go.NewResource("urn:subject"),
		rdf2go.NewResource("http://purl.org/dc/elements/1.1/subject"),
		rdf2go.NewLiteral("hello"),
	)
	g.AddTriple(
		rdf2go.NewResource("urn:subject"),
		rdf2go.NewResource("http://www.europeana.eu/schemas/edm/hasView"),
		rdf2go.NewResource("urn:private/123"),
	)
	g.AddTriple(
		rdf2go.NewResource("urn:private/123"),
		rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		rdf2go.NewLiteral("http://www.europeana.eu/schemas/edm/WebResource"),
	)

	is.Equal(g.Len(), 3)

	var buf bytes.Buffer
	err := serializeTurtle(g, &buf)
	is.NoErr(err)

	rdf := buf.String()
	is.True(!strings.Contains(rdf, "urn:private/")) // serialized rdf should not contain urn:private
	t.Logf("rdf: %s", rdf)
	is.True(strings.HasSuffix(rdf, " .\n"))
}
