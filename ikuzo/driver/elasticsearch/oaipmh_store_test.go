package elasticsearch

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"
)

func ParseFragmentGraph(t *testing.T) {
	is := is.New(t)

	f, err := os.ReadFile("./testdata/sample_graph.json")
	is.NoErr(err)

	fg, err := decodeFragmentGraph(json.RawMessage(f))
	is.NoErr(err)
	_ = fg

	store := OAIPMHStore{}
	var buf bytes.Buffer
	serializeErr := store.serialize("rdf-xml", fg, &buf)
	is.NoErr(serializeErr)

	is.True(buf.Len() > 0)
}
