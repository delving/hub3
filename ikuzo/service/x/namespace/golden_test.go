package namespace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/delving/hub3/ikuzo/testutil"
	"github.com/google/go-cmp/cmp"
)

// TestDefaultNamespacesGolden writes the default prefix table as data, so
// other writers of v2 documents can derive the same searchLabel values.
func TestDefaultNamespacesGolden(t *testing.T) {
	svc, err := NewService(WithDefaults())
	if err != nil {
		t.Fatal(err)
	}
	all, err := svc.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		Prefix string `json:"prefix"`
		Base   string `json:"base"`
		Weight int    `json:"weight"`
	}
	rows := make([]row, 0, len(all))
	for _, ns := range all {
		// The entry GetSearchLabel resolves for this base wins.
		got, err := svc.GetWithBase(ns.URI)
		if err != nil || got.Prefix != ns.Prefix {
			continue
		}
		rows = append(rows, row{Prefix: ns.Prefix, Base: ns.URI, Weight: ns.Weight})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Base < rows[j].Base })
	got, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	path := filepath.Join("testdata", "namespaces.json")
	if *testutil.Update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update)", err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("namespaces.json mismatch (-want +got):\n%s", diff)
	}
}
