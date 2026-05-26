package fragments_test

import (
	"testing"

	"github.com/matryer/is"

	. "github.com/delving/hub3/hub3/fragments"
)

// TestFragmentGraph_NewJSONLD_Framed verifies that NewJSONLD returns a framed
// JSON-LD document anchored on Meta.EntryURI when the FragmentGraph has
// reconstructable resources.
func TestFragmentGraph_NewJSONLD_Framed(t *testing.T) {
	is := is.New(t)

	fb, err := testDataGraph(false)
	is.NoErr(err)

	doc, err := fb.Doc()
	is.NoErr(err)
	is.True(len(doc.Resources) > 0)
	is.True(doc.Meta.EntryURI != "")

	out := doc.NewJSONLD()
	is.Equal(len(out), 1) // single framed root, not a per-resource array

	framed := out[0]
	is.Equal(framed["@id"], doc.Meta.EntryURI)

	// Embedded entries are kept (default frame uses @embed:@always), so the
	// framed root must hold more than just identity metadata.
	if len(framed) < 3 {
		t.Fatalf("expected embedded predicates on framed root, got keys: %v", mapKeys(framed))
	}
}

// TestFragmentGraph_NewJSONLD_FallbackWhenNoResources ensures the legacy flat
// per-resource representation is still emitted when the graph cannot be
// reconstructed (no resources at all).
func TestFragmentGraph_NewJSONLD_FallbackWhenNoResources(t *testing.T) {
	is := is.New(t)

	fg := NewFragmentGraph()
	fg.Meta.OrgID = "test"
	fg.Meta.Spec = "test"
	fg.Meta.EntryURI = "urn:test/1"

	out := fg.NewJSONLD()
	is.Equal(len(out), 0) // no resources → empty payload, no panic
}

// TestFragmentGraph_NewJSONLD_FallbackWhenNoEntryURI ensures the legacy path is
// used when EntryURI is missing, since the graph cannot be anchored.
func TestFragmentGraph_NewJSONLD_FallbackWhenNoEntryURI(t *testing.T) {
	is := is.New(t)

	fb, err := testDataGraph(false)
	is.NoErr(err)

	doc, err := fb.Doc()
	is.NoErr(err)
	doc.Meta.EntryURI = ""

	out := doc.NewJSONLD()
	// Legacy path returns one entry per unique resource ID.
	is.True(len(out) >= 1)
	// None of the entries should carry a single @id matching the (cleared)
	// EntryURI — confirms we are on the flat path.
	for _, entry := range out {
		if id, ok := entry["@id"].(string); ok && id == "" {
			t.Fatalf("flat path entry has empty @id: %v", entry)
		}
	}
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
