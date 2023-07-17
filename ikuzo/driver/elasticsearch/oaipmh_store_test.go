package elasticsearch

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func ParseFragmentGraph(t *testing.T) {
	is := is.New(t)

	f, err := os.ReadFile("./testdata/sample_graph.json")
	is.NoErr(err)

	store := OAIPMHStore{}

	record, err := store.getOAIPMHRecord(recordWrapper{
		HubID: "123",
		Data:  f,
	},
		"oai_dc",
		false)

	is.NoErr(err)
	is.True(len(record.Metadata.Body) > 0)
}
