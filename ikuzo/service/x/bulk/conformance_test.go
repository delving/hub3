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
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/testutil"
	"github.com/spf13/viper"
)

// The conformance suite: one bulk request in, the v2 and v1 documents out.
// See testdata/conformance/README.md.
const confDir = "testdata/conformance"

// confDirV1 holds the v1 goldens for the same inputs.
var confDirV1 = filepath.Join(confDir, "v1")

var (
	caseFile   = regexp.MustCompile(`^(\d{4})\.(nt|ttl|rdf|jsonld)$`)
	modifiedRe = regexp.MustCompile(`"modified":\d+`)
	// v1's system block stamps the indexing time on every document.
	v1TimeRe = regexp.MustCompile(`"(created_at|modified_at)":"[^"]*"`)
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
		"lrm":    {"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation"},
		"nkc":    {"https://wo2.collectienederland.nl/nk/terms/CHO"},
	}
)

type caseResult struct {
	doc  []byte // indented v2 document, when indexed
	kind string // BulkError stage or "untyped", when refused
}

// loadRDFTag replaces the rdftag config with the [rdftag] section of path (the
// NDE config), so goldens show what NDE indexes rather than hub3's defaults.
func loadRDFTag(t *testing.T, path string) {
	t.Helper()
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	var tag c.RDFTag
	if err := v.UnmarshalKey("rdftag", &tag); err != nil {
		t.Fatal(err)
	}
	c.Config.RDFTag = tag
	c.Config.RDFTagMap = c.NewRDFTagMap(&c.Config)
}

type captureIndex struct{ msgs []*domainpb.IndexMessage }

func (ci *captureIndex) Publish(_ context.Context, m ...*domainpb.IndexMessage) error {
	ci.msgs = append(ci.msgs, m...)
	return nil
}

func TestConformance(t *testing.T) {
	c.InitConfig()
	prevRDF, prevES, prevTags := c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap
	prevRDFTag, prevRDFTagMap := c.Config.RDFTag, c.Config.RDFTagMap
	c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = false, false
	c.Config.DatasetTagMap = c.NewDataSetTagMap(&c.RawConfig{})
	loadRDFTag(t, filepath.Join(confDir, "rdftag.toml"))
	t.Cleanup(func() {
		c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap = prevRDF, prevES, prevTags
		c.Config.RDFTag, c.Config.RDFTagMap = prevRDFTag, prevRDFTagMap
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
			t.Run("v1", func(t *testing.T) {
				checkGolden(t, confDirV1, num, runConformanceCase(t, num, input, "v1"))
			})
			got := runConformanceCase(t, num, input, "v2")
			checkGolden(t, confDir, num, got)
			if got.kind != "" || *testutil.Update {
				return
			}
			want := normaliseForFormats(t, got.doc)
			for _, ext := range []string{"nt", "ttl", "rdf", "jsonld"} {
				f := filepath.Join(confDir, "formats", num+"."+ext)
				if _, err := os.Stat(f); err != nil {
					continue
				}
				t.Run("formats/"+ext, func(t *testing.T) {
					alt := runConformanceCase(t, num, f, "v2")
					if alt.kind != "" {
						t.Fatalf("refused as %q", alt.kind)
					}
					if diff := cmp.Diff(want, normaliseForFormats(t, alt.doc)); diff != "" {
						t.Errorf("formats/%s.%s differs from %s.json (order, bnode labels and array order ignored):\n%s", num, ext, num, diff)
					}
				})
			}
		})
	}
}

// runConformanceCase runs one input file through the bulk index action for one
// index type ("v1" or "v2").
func runConformanceCase(t *testing.T, num, inputPath, indexType string) caseResult {
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
		ds:            &models.DataSet{OrgID: req.OrgID, Spec: req.DatasetID, Revision: req.Revision},
		stats:         &Stats{},
		bi:            ci,
		indexTypes:    []string{indexType},
		recDef:        RecDefResolver{m: conformanceRecDefs, defaultRecDefID: "schema"},
		s:             &Service{}, // no task service: PID tasks are logged and skipped
		maxNestedDocs: defaultMaxNestedDocs,
	}

	// setDataSet needs the dataset store: skip it, the dataset is preset above.
	p.once.Do(func() {})

	if err := p.process(context.Background(), req); err != nil {
		var be *BulkError
		if errors.As(err, &be) {
			return caseResult{kind: be.Stage}
		}
		return caseResult{kind: "untyped"}
	}
	if len(ci.msgs) != 1 {
		t.Fatalf("want 1 index message, got %d", len(ci.msgs))
	}

	src := stableSource(ci.msgs[0].Source)
	var buf bytes.Buffer
	if err := json.Indent(&buf, src, "", "  "); err != nil {
		t.Fatal(err)
	}
	buf.WriteByte('\n')
	return caseResult{doc: buf.Bytes()}
}

// stableSource replaces the indexing time in a v2 (meta.modified) or v1
// (system.created_at/modified_at) document with a fixed value, and puts v1's
// system.source_graph in canonical order (see canonicalJSONLD).
func stableSource(src []byte) []byte {
	src = modifiedRe.ReplaceAll(src, []byte(`"modified":0`))
	src = v1TimeRe.ReplaceAll(src, []byte(`"$1":"1970-01-01T00:00:00Z"`))
	var doc map[string]json.RawMessage
	if json.Unmarshal(src, &doc) != nil || doc["system"] == nil {
		return src
	}
	var sys map[string]any
	if json.Unmarshal(doc["system"], &sys) != nil {
		return src
	}
	sg, ok := sys["source_graph"].(string)
	if !ok {
		return src
	}
	sys["source_graph"] = canonicalJSONLD(sg)
	b, err := json.Marshal(sys)
	if err != nil {
		return src
	}
	doc["system"] = b
	out, err := json.Marshal(doc) // map keys come out sorted, as hub3 writes them
	if err != nil {
		return src
	}
	return out
}

// canonicalJSONLD sorts the nodes of a flat JSON-LD array by their JSON and
// every value array inside them. hub3 builds source_graph from maps, so node
// order and value order change from run to run; the content does not.
func canonicalJSONLD(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, err := json.Marshal(sortArrays(v))
	if err != nil {
		return s
	}
	return string(b)
}

func sortArrays(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			x[k] = sortArrays(val)
		}
	case []any:
		keys := make([]string, len(x))
		for i := range x {
			x[i] = sortArrays(x[i])
			b, _ := json.Marshal(x[i])
			keys[i] = string(b)
		}
		sort.Sort(byKey{x, keys})
	}
	return v
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

// checkGolden compares with dir/NNNN.json or dir/NNNN.error.json; -update rewrites them.
func checkGolden(t *testing.T, dir, num string, got caseResult) {
	t.Helper()
	jsonPath := filepath.Join(dir, num+".json")
	errPath := filepath.Join(dir, num+".error.json")

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

func TestNormaliseForFormats(t *testing.T) {
	a := []byte(`{"resources":[{"id":"_:b7","entries":[{"order":2,"@value":"y"},{"order":1,"@value":"x"}]}]}`)
	b := []byte(`{"resources":[{"id":"_:n1","entries":[{"order":1,"@value":"x"},{"order":5,"@value":"y"}]}]}`)
	if normaliseForFormats(t, a) != normaliseForFormats(t, b) {
		t.Fatalf("order, bnode label and array order should not matter:\n%s\n%s",
			normaliseForFormats(t, a), normaliseForFormats(t, b))
	}
	cDoc := []byte(`{"resources":[{"id":"_:b7","entries":[{"order":1,"@value":"z"}]}]}`)
	if normaliseForFormats(t, a) == normaliseForFormats(t, cDoc) {
		t.Fatal("a changed value must still differ")
	}
}

// bnodePrefix is how a blank node appears in a v2 document.
const bnodePrefix = "_:"

// normaliseForFormats makes two documents from different serialisations of one
// graph comparable: drops every "order", gives every blank node one label and
// sorts arrays. Lost or changed values still differ.
func normaliseForFormats(t *testing.T, doc []byte) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(doc, &v); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(normaliseValue(v)) // map keys come out sorted
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func normaliseValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		delete(x, "order")
		for k, val := range x {
			x[k] = normaliseValue(val)
		}
		return x
	case []any:
		keys := make([]string, len(x))
		for i := range x {
			x[i] = normaliseValue(x[i])
			b, _ := json.Marshal(x[i])
			keys[i] = string(b)
		}
		sort.Sort(byKey{x, keys})
		return x
	case string:
		if strings.HasPrefix(x, bnodePrefix) {
			return bnodePrefix + "b"
		}
		return x
	}
	return v
}

type byKey struct {
	v    []any
	keys []string
}

func (s byKey) Len() int           { return len(s.v) }
func (s byKey) Less(i, j int) bool { return s.keys[i] < s.keys[j] }
func (s byKey) Swap(i, j int) {
	s.v[i], s.v[j] = s.v[j], s.v[i]
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
}
