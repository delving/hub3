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

	f, err := os.Open("./testdata/nk_fg_grouped.json")
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
		is.Equal(len(entries), 7)
	})

	t.Run("Marshal", func(t *testing.T) {
		is := is.New(t)

		rsc := fg.Resources[0]

		var record internal.BaseRecord
		err = rsc.UnmarshalRDF(&record)
		is.NoErr(err)
		is.Equal(record.BaseID, "NK1145")
		is.Equal(record.CleanID, "NK1145")
		is.Equal(record.DcTitle, "De Ivan Wiliki te Moskou")
		is.Equal(record.ID, "https://wo2.collectienederland.nl/id/nk/NK1145")
		is.Equal(record.Type, []string{"https://wo2.collectienederland.nl/nk/terms/NKRecord"})
		is.Equal(len(record.EdmHasView), 4)
		// t.Logf("edmHasView: %#v", record.EdmHasView)
		is.Equal(record.EdmHasView[0].ID, "https://images.memorix.nl/rce/thumb/fullsize/c2c7b29a-fdb0-bc53-bcf0-89f31c0bbc18.jpg")
		// is.True(record.Cho[0].Creator[0].RKDArtistLink != nil)
	})

	t.Run("Marshal Resource", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/nk_fg_creator.json")
		is.NoErr(err)
		defer f.Close()

		b, err := io.ReadAll(f)
		is.NoErr(err)

		var rsc fragments.FragmentResource
		err = json.Unmarshal(b, &rsc)
		is.NoErr(err)

		var creator internal.Creator
		err = rsc.UnmarshalRDF(&creator)
		is.NoErr(err)
		is.True(len(creator.RKDArtistLink) != 0)
		is.Equal(creator.RKDArtistLink[0].String(), "https://rkd.nl/explore/artists/5113")

		links := rsc.GetByResourcesBySearchLabel("nk_rkdArtistLink")
		is.Equal(len(links), 1)
	})

	t.Run("Marshal Resource with nested resource", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/nk_fg_cho.json")
		is.NoErr(err)
		defer f.Close()

		b, err := io.ReadAll(f)
		is.NoErr(err)

		var rsc fragments.ResourceEntry
		err = json.Unmarshal(b, &rsc)
		is.NoErr(err)

		var cho internal.CHO
		err = rsc.Inline.UnmarshalRDF(&cho)
		is.NoErr(err)
		is.Equal(len(cho.Creator), 1)
		// creator := cho.Creator[0]
		// is.True(len(creator.RKDArtistLink) != 0)
		// is.Equal(creator.RKDArtistLink[0].String(), "https://rkd.nl/explore/artists/5113")
	})
}

func TestResourceUnmarshal(t *testing.T) {
	t.Run("Marshal Resource with nested resource validation", func(t *testing.T) {
		is := is.New(t)
		f, err := os.Open("./testdata/nk_fg_cho.json")
		is.NoErr(err)
		defer f.Close()

		b, err := io.ReadAll(f)
		is.NoErr(err)

		var rsc fragments.ResourceEntry
		err = json.Unmarshal(b, &rsc)
		is.NoErr(err)

		order := rsc.Inline.FilterEntries("nave_order")
		is.Equal(len(order), 1)
		is.Equal(order[0].Value, "123")

		creator := rsc.Inline.FilterEntries("nk_creator")
		is.Equal(len(creator), 1)
		firstCreator := creator[0].Inline.Entries[0].Inline
		is.Equal(firstCreator.FilterEntries("nk_creatorName")[0].Value, "Bauer, M.A.J.")
		is.Equal(firstCreator.FilterEntries("nk_rkdArtistLink")[0].Value, "https://rkd.nl/explore/artists/5113")

		links := creator[0].Inline.GetByResourcesBySearchLabel("nk_rkdArtistLink")
		is.Equal(len(links), 1)
	})

	t.Run("Marshal Resource with nested resource", func(t *testing.T) {
		is := is.New(t)
		f, err := os.Open("./testdata/nk_fg_cho.json")
		is.NoErr(err)
		defer f.Close()

		b, err := io.ReadAll(f)
		is.NoErr(err)

		var re fragments.ResourceEntry
		err = json.Unmarshal(b, &re)
		is.NoErr(err)

		var cho internal.CHO
		err = re.Inline.UnmarshalRDF(&cho)
		is.NoErr(err)
		is.Equal(cho.Order, 123)

		// production
		is.Equal(len(cho.ProductionDate), 1)
		productionDate := cho.ProductionDate[0]
		is.Equal(productionDate.DateStart[0], "1900")

		// creator
		is.Equal(len(cho.Creator), 1)
		firstCreator := cho.Creator[0]
		t.Logf("first Creator: %#v", firstCreator)
		is.Equal(len(firstCreator.CreatorName), 1)
		is.Equal(firstCreator.CreatorName[0].String(), "Bauer, M.A.J.")
		is.Equal(firstCreator.RKDArtistLink[0].String(), "https://rkd.nl/explore/artists/5113")
	})

	// Test case for issue #2658: Multiple creators where only one has an RKD link
	// Structure: CHO -> nk_creator -> wrapper -> nk_Creator -> [creator1 (with RKD), creator2 (without RKD)]
	// BUG: Currently all creators get assigned ALL RKD links from siblings due to recursive search
	t.Run("Multiple creators with mixed RKD links - issue 2658", func(t *testing.T) {
		is := is.New(t)
		f, err := os.Open("./testdata/nk_fg_cho_multi_creator.json")
		is.NoErr(err)
		defer f.Close()

		b, err := io.ReadAll(f)
		is.NoErr(err)

		var re fragments.ResourceEntry
		err = json.Unmarshal(b, &re)
		is.NoErr(err)

		var cho internal.CHO
		err = re.Inline.UnmarshalRDF(&cho)
		is.NoErr(err)

		t.Logf("Number of creators: %d", len(cho.Creator))
		for i, creator := range cho.Creator {
			t.Logf("Creator %d: %#v", i, creator)
		}

		// This test demonstrates the current behavior vs expected behavior:
		// CURRENT: All creators get the same RKD link due to recursive search
		// EXPECTED: Each creator should only have its own RKD link

		// We expect 2 creators (Rembrandt and onbekend)
		is.Equal(len(cho.Creator), 2) // This will likely fail - showing the bug

		if len(cho.Creator) >= 2 {
			rembrandt := cho.Creator[0]
			onbekend := cho.Creator[1]

			// Rembrandt should have an RKD link
			is.True(len(rembrandt.RKDArtistLink) > 0)
			is.Equal(rembrandt.RKDArtistLink[0].String(), "https://rkd.nl/explore/artists/66219")
			is.Equal(rembrandt.CreatorName[0].String(), "Rembrandt van Rijn")

			// onbekend should NOT have an RKD link - this is likely where the bug manifests
			is.Equal(len(onbekend.RKDArtistLink), 0) // BUG: This will fail because onbekend incorrectly gets Rembrandt's RKD link
			is.Equal(onbekend.CreatorName[0].String(), "onbekend")
		}
	})
}
