package fragments_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/matryer/is"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/fragments/internal"
)

func TestFragmentGraph(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/nk_fg.json")
	is.NoErr(err)
	defer f.Close()

	b, err := io.ReadAll(f)
	is.NoErr(err)

	var fg fragments.FragmentGraph
	err = json.Unmarshal(b, &fg)
	is.NoErr(err)

	t.Run("NewGrouped", func(t *testing.T) {
		is := is.New(t)

		rsc, err := fg.NewGrouped()
		is.NoErr(err)

		entries := rsc.GetByResourcesBySearchLabel("nk_timeline")
		is.Equal(len(entries), 10)
	})

	t.Run("Marshal", func(t *testing.T) {
		is := is.New(t)

		rsc, err := fg.NewGrouped()
		is.NoErr(err)
		_ = rsc

		var record internal.BaseRecord

		err = rsc.UnmarshalRDF(&record)
		is.NoErr(err)
		is.Equal(record.BaseID, "NK1066")
		is.Equal(record.CleanID, "NK1066")
		is.Equal(record.DcTitle, "Vloerkleed, engels")
		is.Equal(record.ID, "https://wo2.collectienederland.nl/id/nk/NK1066")
		is.Equal(record.Type, []string{"https://wo2.collectienederland.nl/nk/terms/NKRecord"})
		is.Equal(len(record.EdmHasView), 1)
		t.Logf("edmHasView: %#v", record.EdmHasView)
		is.Equal(record.EdmHasView[0].ID, "https://images.memorix.nl/rce/thumb/fullsize/aba1980c-18c9-4ec4-3163-29efff75be8f.jpg")
	})
}
