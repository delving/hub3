package bulk

import (
	"fmt"
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
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

func TestParseLongLine(t *testing.T) {
	is := is.New(t)

	p := &Parser{stats: &Stats{}}
	// invalid json longer than the old 5MB scanner limit: skipped, not fatal
	line := "{" + strings.Repeat("x", 6*1024*1024) + "\n"

	err := p.Parse(context.Background(), strings.NewReader(line))
	is.NoErr(err)
	is.Equal(p.stats.JSONErrors, uint64(1))
}

// Every request in one upload resolves its own tags; fragmentsOnly on a later
// line applies to that line only (F5).
func TestProcessResolvesTagsPerRequest(t *testing.T) {
	c.InitConfig()
	prevRDF, prevES := c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled
	c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = false, false
	t.Cleanup(func() { c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = prevRDF, prevES })

	ci := &captureIndex{}
	p := &Parser{
		ds:         &models.DataSet{OrgID: "org1", Spec: "ds1", Revision: 3},
		stats:      &Stats{},
		bi:         ci,
		indexTypes: []string{"v2"},
		recDef:     RecDefResolver{m: map[string][]string{"schema": {"https://schema.org/CreativeWork"}}, defaultRecDefID: "schema"},
		s:          &Service{},
	}
	p.once.Do(func() {}) // as if an earlier request had set up the dataset

	req := func(id, tags string) *Request {
		return &Request{
			Action: "index", OrgID: "org1", DatasetID: "ds1", HubID: "org1_ds1_" + id, LocalID: id,
			NamedGraphURI: "http://example.org/" + id + "/graph", Tags: tags, GraphMimeType: "application/n-triples",
			Graph: "<http://example.org/" + id + "> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/CreativeWork> .\n",
		}
	}

	if err := p.process(context.Background(), req("a", "extra")); err != nil {
		t.Fatal(err)
	}
	if err := p.process(context.Background(), req("b", "fragmentsOnly")); err != nil {
		t.Fatal(err)
	}

	var v2 []*domainpb.IndexMessage
	for _, m := range ci.msgs {
		if m.GetIndexType() == domainpb.IndexType_V2 {
			v2 = append(v2, m)
		}
	}
	if len(v2) != 1 {
		t.Fatalf("want 1 v2 index message (b is fragmentsOnly), got %d", len(v2))
	}
	var doc struct {
		Meta struct {
			Tags     []string `json:"tags"`
			Revision int      `json:"revision"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(v2[0].Source, &doc); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(doc.Meta.Tags, "extra") || doc.Meta.Revision != 3 {
		t.Fatalf("meta tags %v revision %d, want tags with \"extra\" and revision 3", doc.Meta.Tags, doc.Meta.Revision)
	}
}

// A record over maxNestedDocs is skipped and counted; the rest of the upload
// is still indexed.
func TestParseSkipsRecordTooLarge(t *testing.T) {
	c.InitConfig()
	prevRDF, prevES := c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled
	c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = false, false
	t.Cleanup(func() { c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = prevRDF, prevES })

	ci := &captureIndex{}
	p := &Parser{
		ds:            &models.DataSet{OrgID: "org1", Spec: "ds1", Revision: 1},
		stats:         &Stats{},
		bi:            ci,
		indexTypes:    []string{"v2"},
		recDef:        RecDefResolver{m: map[string][]string{"schema": {"https://schema.org/CreativeWork"}}, defaultRecDefID: "schema"},
		s:             &Service{},
		maxNestedDocs: 3,
	}
	p.once.Do(func() {})

	line := func(id string, names int) string {
		s := "http://example.org/" + id
		g := "<" + s + "> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/CreativeWork> .\n"
		for i := 0; i < names; i++ {
			g += fmt.Sprintf("<%s> <https://schema.org/name> \"n%d\" .\n", s, i)
		}
		b, err := json.Marshal(Request{
			Action: "index", OrgID: "org1", DatasetID: "ds1", HubID: "org1_ds1_" + id, LocalID: id,
			NamedGraphURI: s + "/graph", GraphMimeType: "application/n-triples", Graph: g,
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(b) + "\n"
	}

	err := p.Parse(context.Background(), strings.NewReader(line("big", 10)+line("small", 0)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p.stats.RecordsTooLarge != 1 || p.stats.RecordsStored != 1 {
		t.Fatalf("recordsTooLarge %d recordsStored %d, want 1 and 1", p.stats.RecordsTooLarge, p.stats.RecordsStored)
	}
	if len(ci.msgs) != 1 {
		t.Fatalf("want 1 index message (small), got %d", len(ci.msgs))
	}
}
