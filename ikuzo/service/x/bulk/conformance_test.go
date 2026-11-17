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
		"lrm":    {"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation"},
		"nkc":    {"https://wo2.collectienederland.nl/nk/terms/CHO"},
	}
)

// knownFormatDiffs maps a case number, or a case range (first two digits), to the finding in
// testdata/conformance/FINDINGS.md that explains why another serialisation
// gives a different document. A formats subtest in that range is skipped only
// when it differs; remove the entry once the finding is fixed.
var knownFormatDiffs = map[string]string{
	"05":   "F11: level and context follow triple order",
	"06":   "F11: level and context follow triple order",
	"0006": "F24: context Label follows triple order",
	"0101": "F24: context Label follows triple order",
	"0107": "F24: context Label follows triple order",
	"0203": "F24: context Label follows triple order",
}

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
			got := runConformanceCase(t, num, input)
			checkGolden(t, num, got)
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
					alt := runConformanceCase(t, num, f)
					if alt.kind != "" {
						t.Fatalf("refused as %q", alt.kind)
					}
					if diff := cmp.Diff(want, normaliseForFormats(t, alt.doc)); diff != "" {
						f, ok := knownFormatDiffs[num]
						if !ok {
							f, ok = knownFormatDiffs[num[:2]]
						}
						if ok {
							t.Skipf("known difference %s; diff:\n%s", f, diff)
						}
						t.Errorf("formats/%s.%s differs from %s.json (order, bnode labels and array order ignored):\n%s", num, ext, num, diff)
					}
				})
			}
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
		ds:            &models.DataSet{OrgID: req.OrgID, Spec: req.DatasetID, Revision: req.Revision},
		stats:         &Stats{},
		bi:            ci,
		indexTypes:    []string{"v2"},
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
