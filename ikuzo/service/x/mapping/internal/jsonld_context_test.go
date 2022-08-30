package internal

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestParseJSONLD(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/jsonld_schema.jsonld")
		is.NoErr(err)

		schema, err := ParseJSONLD(f)
		is.NoErr(err)
		is.Equal(len(schema.Context), 180)
		is.Equal(len(schema.Predicates), 115)
		is.Equal(len(schema.Resources), 65)

		rsc, ok := schema.Resources["ZelfstandigeExpressie"]
		is.True(ok)

		t.Logf("predicates: %#v", rsc)

		pred, ok := rsc.Predicates["creatie"]
		is.True(ok)
		is.Equal(pred.Label, "creatie")
		is.Equal(pred.Container, "@set")
		is.Equal(pred.Resource, "ZelfstandigeExpressie")
		is.Equal(pred.ID, "https://www.iflastandards.info/fr/frbr/frbroo#R17i")
		is.Equal(pred.Type, "@id")
	})

	t.Run("generate elems", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open("./testdata/jsonld_schema.jsonld")
		is.NoErr(err)

		schema, err := ParseJSONLD(f)
		is.NoErr(err)

		schema.RootResource = "MensgemaaktObject"

		elems, err := schema.Elems()
		is.NoErr(err)
		is.Equal(len(elems), 1)
		is.Equal(len(elems[0].Celem), 65)

		var buf bytes.Buffer
		enc := xml.NewEncoder(&buf)
		enc.Indent("", "    ")
		err = enc.Encode(elems)
		is.NoErr(err)

		t.Logf("elems: \n %s", buf.String())
		// is.True(false)
	})
}
