package internal

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestModels(t *testing.T) {
	t.Run("parse artwork", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/oslo/input_csv/modellen_artwork.csv")
		is.NoErr(err)
		defer f.Close()

		model, err := ParseModel(f)
		is.NoErr(err)

		is.Equal(len(model.rows), 130)
		is.Equal(len(model.resources), 0)
		is.Equal(model.nextNodeID, 0)

		is.Equal(model.ns.Len(), 2025)

		err = model.inline()
		is.NoErr(err)
		is.Equal(len(model.resources), 28)
		is.Equal(model.nextNodeID, 78)

		rsc, ok := model.resources["MensgemaaktObject"]
		is.True(ok)
		is.Equal(rsc.DomainNode.ID, 0)
		is.Equal(rsc.DomainNode.ClassLabel, "MensgemaaktObject")

		t.Logf("rsc: %#v", rsc)

		pred, ok := rsc.Predicates["Entiteit.beschrijving"]
		is.True(ok)
		t.Logf("pred: %#v", pred)
		is.Equal(pred.TargetNode.ID, 1)
		// is.True(false)
	})

	t.Run("generate elems", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/oslo/input_csv/modellen_artwork.csv")
		is.NoErr(err)
		defer f.Close()

		model, err := ParseModel(f)
		is.NoErr(err)

		err = model.inline()
		is.NoErr(err)

		model.RootResource = "MensgemaaktObject"

		elems, err := model.Elems()
		is.NoErr(err)
		is.Equal(len(elems), 1)
		// is.Equal(len(elems[0].Celem), 65)

		var buf bytes.Buffer
		enc := xml.NewEncoder(&buf)
		enc.Indent("", "    ")
		err = enc.Encode(elems)
		is.NoErr(err)

		t.Logf("elems: \n %s", buf.String())
		is.True(false)
	})
}
