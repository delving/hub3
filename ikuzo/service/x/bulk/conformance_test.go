package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/testutil"
)

// The conformance suite: one bulk request in, the v2 document out.
// See testdata/conformance/README.md.
const confDir = "testdata/conformance"

var (
	caseFile   = regexp.MustCompile(`^(\d{4})\.(nt|ttl|rdf|jsonld)$`)
	modifiedRe = regexp.MustCompile(`"modified":\d+`)
	mimeByExt  = map[string]string{
		".nt":     "application/n-triples",
		".ttl":    "text/turtle",
		".rdf":    "application/rdf+xml",
		".jsonld": "application/ld+json",
	}
	// conformanceRecDefs maps a recDefId to the root class hub3 looks for.
	conformanceRecDefs = map[string][]string{
		"schema": {"https://schema.org/CreativeWork"},
		"edm":    {"http://www.openarchives.org/ore/terms/Aggregation"},
		"crm":    {"http://www.cidoc-crm.org/cidoc-crm/E22_Human-Made_Object"},
	}
)

type caseResult struct {
	doc  []byte // indented v2 document, when indexed
	kind string // BulkError stage or "untyped", when refused
}

type captureIndex struct{ msgs []*domainpb.IndexMessage }

func (ci *captureIndex) Publish(_ context.Context, m ...*domainpb.IndexMessage) error {
	ci.msgs = append(ci.msgs, m...)
	return nil
}

func TestConformance(t *testing.T) {
	c.InitConfig()
	prevRDF, prevES, prevTags := c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap
	c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = false, false
	c.Config.DatasetTagMap = c.NewDataSetTagMap(&c.RawConfig{})
	t.Cleanup(func() {
		c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap = prevRDF, prevES, prevTags
	})

	entries, err := os.ReadDir(confDir) // sorted by name, so by number
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		m := caseFile.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		num, input := m[1], filepath.Join(confDir, e.Name())
		t.Run(num, func(t *testing.T) {
			checkGolden(t, num, runConformanceCase(t, num, input))
		})
	}
}

// runConformanceCase runs one input file through the bulk index action.
func runConformanceCase(t *testing.T, num, inputPath string) caseResult {
	t.Helper()
	req := loadEnvelope(t, num)
	graph, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	req.Graph = string(graph)
	req.GraphMimeType = mimeByExt[filepath.Ext(inputPath)]

	ci := &captureIndex{}
	p := &Parser{
		ds:         &models.DataSet{OrgID: req.OrgID, Spec: req.DatasetID, Revision: req.Revision},
		stats:      &Stats{},
		bi:         ci,
		indexTypes: []string{"v2"},
		recDef:     RecDefResolver{m: conformanceRecDefs, defaultRecDefID: "schema"},
		s:          &Service{}, // no task service: PID tasks are logged and skipped
	}

	// What Parser.setDataSet does for a parser's first request, and the
	// aboutTypeURI resolution in Parser.process; setDataSet needs the dataset store.
	req.setTags()
	aboutURI, err := p.recDef.Resolve(req.RecDefID)
	if err != nil {
		return caseResult{kind: "recdef"}
	}
	req.aboutTypeURI = aboutURI

	if err := p.Publish(context.Background(), req); err != nil {
		var be *BulkError
		if errors.As(err, &be) {
			return caseResult{kind: be.Stage}
		}
		return caseResult{kind: "untyped"}
	}
	if len(ci.msgs) != 1 {
		t.Fatalf("want 1 index message, got %d", len(ci.msgs))
	}

	src := modifiedRe.ReplaceAll(ci.msgs[0].Source, []byte(`"modified":0`))
	var buf bytes.Buffer
	if err := json.Indent(&buf, src, "", "  "); err != nil {
		t.Fatal(err)
	}
	buf.WriteByte('\n')
	return caseResult{doc: buf.Bytes()}
}

// loadEnvelope reads envelope.json with NNNN replaced, then NNNN.bulk.json over it.
func loadEnvelope(t *testing.T, num string) *Request {
	t.Helper()
	base, err := os.ReadFile(filepath.Join(confDir, "envelope.json"))
	if err != nil {
		t.Fatal(err)
	}
	req := &Request{}
	if err := json.Unmarshal(bytes.ReplaceAll(base, []byte("NNNN"), []byte(num)), req); err != nil {
		t.Fatal(err)
	}
	over, err := os.ReadFile(filepath.Join(confDir, num+".bulk.json"))
	switch {
	case err == nil:
		// Unmarshalling onto the same struct merges key by key.
		if err := json.Unmarshal(over, req); err != nil {
			t.Fatal(err)
		}
	case !os.IsNotExist(err):
		t.Fatal(err)
	}
	return req
}

// checkGolden compares with NNNN.json or NNNN.error.json; -update rewrites them.
func checkGolden(t *testing.T, num string, got caseResult) {
	t.Helper()
	jsonPath := filepath.Join(confDir, num+".json")
	errPath := filepath.Join(confDir, num+".error.json")

	if *testutil.Update {
		_ = os.Remove(jsonPath)
		_ = os.Remove(errPath)
		path, data := jsonPath, got.doc
		if got.kind != "" {
			path, data = errPath, []byte(fmt.Sprintf("{\"kind\":%q}\n", got.kind))
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	if want, err := os.ReadFile(jsonPath); err == nil {
		if got.kind != "" {
			t.Fatalf("want a document, got refusal %q", got.kind)
		}
		if diff := cmp.Diff(string(want), string(got.doc)); diff != "" {
			t.Errorf("%s.json mismatch (-want +got):\n%s", num, diff)
		}
		return
	}
	wantErr, err := os.ReadFile(errPath)
	if err != nil {
		t.Fatalf("no %s.json or %s.error.json; got kind=%q doc:\n%s", num, num, got.kind, got.doc)
	}
	var we struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(wantErr, &we); err != nil {
		t.Fatal(err)
	}
	if got.kind != we.Kind {
		t.Errorf("want refusal %q, got kind=%q doc:\n%s", we.Kind, got.kind, got.doc)
	}
}
