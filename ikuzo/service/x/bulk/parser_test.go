package bulk

import (
	"bytes"
	"encoding/json"
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
	g.AddTriple(
		rdf2go.NewBlankNode("b0"),
		rdf2go.NewResource("http://purl.org/dc/elements/1.1/title"),
		rdf2go.NewLiteral("hello"),
	)

	is.Equal(g.Len(), 4)

	var buf bytes.Buffer
	err := serializeNTriples(g, &buf)
	is.NoErr(err)

	rdf := buf.String()
	is.True(!strings.Contains(rdf, "urn:private/")) // serialized rdf should not contain urn:private
	t.Logf("rdf: %s", rdf)
	is.True(strings.HasSuffix(rdf, " .\n"))
	is.True(!strings.Contains(rdf, "_:b0"))
	is.True(strings.Contains(rdf, "urn:bnode:b0-"))
}

type LogMessage struct {
	Svc string `json:"svc"`
}

func testAddLogger(is *is.I, datasetID string, svc string) {
	bytesBuffer := bytes.Buffer{}

	logger := addLogger(datasetID)
	logger = logger.Output(&bytesBuffer)
	logger.Info().Msg("")

	logMessage := LogMessage{}

	err := json.Unmarshal(bytesBuffer.Bytes(), &logMessage)
	is.NoErr(err)
	is.Equal(logMessage.Svc, svc)
}

func TestAddLogger(t *testing.T) {
	is := is.New(t)

	//	ntfoto
	testAddLogger(is, "somestring-ntfoto", "ntfoto")

	//	nt
	testAddLogger(is, "nt00250-somestring", "nt")

	//	default
	testAddLogger(is, "somestring", "")
}
