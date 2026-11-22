package bulk

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
)

// TestDumpV2 runs a bulk NDJSON file through the index path and writes every
// v2 document (or refusal kind) as NDJSON. Skipped unless HUB3_DUMP_IN and
// HUB3_DUMP_OUT are set; HUB3_DUMP_RDFTAG optionally names the rdftag config.
func TestDumpV2(t *testing.T) {
	in, out := os.Getenv("HUB3_DUMP_IN"), os.Getenv("HUB3_DUMP_OUT")
	if in == "" || out == "" {
		t.Skip("set HUB3_DUMP_IN and HUB3_DUMP_OUT")
	}
	c.InitConfig()
	prevRDF, prevES, prevTags := c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap
	prevRDFTag, prevRDFTagMap := c.Config.RDFTag, c.Config.RDFTagMap
	c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled = false, false
	c.Config.DatasetTagMap = c.NewDataSetTagMap(&c.RawConfig{})
	t.Cleanup(func() {
		c.Config.RDF.RDFStoreEnabled, c.Config.ElasticSearch.Enabled, c.Config.DatasetTagMap = prevRDF, prevES, prevTags
		c.Config.RDFTag, c.Config.RDFTagMap = prevRDFTag, prevRDFTagMap
	})
	tag := os.Getenv("HUB3_DUMP_RDFTAG")
	if tag == "" {
		tag = filepath.Join(confDir, "rdftag.toml")
	}
	loadRDFTag(t, tag)

	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	bw := bufio.NewWriter(w)

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 256<<20)
	n := 0
	for sc.Scan() {
		req := &Request{}
		if err := json.Unmarshal(sc.Bytes(), req); err != nil {
			t.Fatalf("line %d: %v", n+1, err)
		}
		ci := &captureIndex{}
		p := &Parser{
			// The dump pins revision 1 for a stable comparison (process() overwrites req.Revision with ds.Revision).
			ds:            &models.DataSet{OrgID: req.OrgID, Spec: req.DatasetID, Revision: 1},
			stats:         &Stats{},
			bi:            ci,
			indexTypes:    []string{"v2"},
			recDef:        RecDefResolver{m: conformanceRecDefs, defaultRecDefID: "schema"},
			s:             &Service{},
			maxNestedDocs: defaultMaxNestedDocs,
		}
		p.once.Do(func() {})
		line := map[string]any{"hubID": req.HubID}
		if err := p.process(context.Background(), req); err != nil {
			var be *BulkError
			line["kind"] = "untyped"
			if errors.As(err, &be) {
				line["kind"] = be.Stage
			}
		} else {
			switch len(ci.msgs) {
			case 1:
				src := modifiedRe.ReplaceAll(ci.msgs[0].Source, []byte(`"modified":0`))
				line["doc"] = json.RawMessage(src)
			case 0:
				line["kind"] = "no_message"
				t.Errorf("record %s: want 1 index message, got %d", req.HubID, len(ci.msgs))
			default:
				line["kind"] = "multiple_messages"
				t.Errorf("record %s: want 1 index message, got %d", req.HubID, len(ci.msgs))
			}
		}
		b, err := json.Marshal(line)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := bw.Write(append(b, '\n')); err != nil {
			t.Fatal(err)
		}
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d records", n)
}
