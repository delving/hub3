package mapping

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/delving/hub3/ikuzo/testutil"
	"github.com/google/go-cmp/cmp"
)

// TestV2MappingGolden writes the v2 create and update mappings as JSON files.
func TestV2MappingGolden(t *testing.T) {
	for name, src := range map[string]string{
		"v2-mapping.json":        V2ESMapping(1, 0),
		"v2-mapping-update.json": v2MappingUpdate,
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(src), "", "  "); err != nil {
				t.Fatalf("not valid JSON: %v", err)
			}
			buf.WriteByte('\n')
			path := filepath.Join("testdata", name)
			if *testutil.Update {
				_ = os.MkdirAll("testdata", 0o755)
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run with -update)", err)
			}
			if diff := cmp.Diff(string(want), buf.String()); diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}
