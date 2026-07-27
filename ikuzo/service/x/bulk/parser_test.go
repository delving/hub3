package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
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

func TestDropRecordsValidatesHubIDPrefix(t *testing.T) {
	is := is.New(t)
	p := &Parser{
		ds: &models.DataSet{OrgID: "org1", Spec: "ds1"},
	}
	req := &Request{
		Action:    "drop_records",
		OrgID:     "org1",
		DatasetID: "ds1",
		HubIDs:    []string{"org1_ds1_a", "org2_ds1_b"}, // second id cross-org
	}
	err := p.dropRecords(context.Background(), req)
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "does not belong to dataset"))
}

// hubIds are interpolated into SPARQL Update statements downstream, so
// anything outside the [A-Za-z0-9_.-] allowlist must be rejected.
func TestDropRecordsRejectsInjectionCharacters(t *testing.T) {
	is := is.New(t)
	p := &Parser{
		ds: &models.DataSet{OrgID: "org1", Spec: "ds1"},
	}
	for _, hid := range []string{
		"org1_ds1_x/graph>; DROP ALL; #",
		"org1_ds1_a b",
		"org1_ds1_a>",
		"org1_ds1_a\n",
	} {
		req := &Request{
			Action:    "drop_records",
			OrgID:     "org1",
			DatasetID: "ds1",
			HubIDs:    []string{hid},
		}
		err := p.dropRecords(context.Background(), req)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "invalid hubId"))
	}
}

// A drop_records action without hubIds is a malformed client payload (e.g.
// ids under a different field name). It must FAIL loudly — silently
// returning nil let clients confirm drops that never happened (found
// live: Narthex sent `ids`, got 201, marked 170 orphans as dropped).
func TestDropRecordsEmptyListIsAnError(t *testing.T) {
	is := is.New(t)
	p := &Parser{
		ds: &models.DataSet{OrgID: "org1", Spec: "ds1"},
	}
	req := &Request{
		Action:    "drop_records",
		OrgID:     "org1",
		DatasetID: "ds1",
		HubIDs:    nil,
	}
	err := p.dropRecords(context.Background(), req)
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "no hubIds"))
}

func TestProcessRoutesDropRecords(t *testing.T) {
	is := is.New(t)
	p := &Parser{
		ds:    &models.DataSet{OrgID: "org1", Spec: "ds1", Revision: 1},
		stats: &Stats{},
		recDef: RecDefResolver{
			m:               map[string][]string{"": {}},
			defaultRecDefID: "",
		},
	}
	// Preflight p.once so process skips setDataSet and uses the ds we injected.
	p.once.Do(func() {})

	req := &Request{
		Action:    "drop_records",
		OrgID:     "org1",
		DatasetID: "ds1",
		HubIDs:    []string{"org1_ds1_a"},
	}
	c.InitConfig()
	prevRDF := c.Config.RDF.RDFStoreEnabled
	prevES := c.Config.ElasticSearch.Enabled
	c.Config.RDF.RDFStoreEnabled = false
	c.Config.ElasticSearch.Enabled = false
	defer func() {
		c.Config.RDF.RDFStoreEnabled = prevRDF
		c.Config.ElasticSearch.Enabled = prevES
	}()

	err := p.process(context.Background(), req)
	// Validation + real method with storages off → no error.
	is.NoErr(err)
}
