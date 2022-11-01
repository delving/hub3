package recdef_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf/schema/recdef"
	"github.com/matryer/is"
)

func TestRecDef(t *testing.T) {
	testFile := "./testdata/edm_5.2.6_record-definition.xml"

	t.Run("parse and write", func(t *testing.T) {
		is := is.New(t)

		f, err := os.Open(testFile)
		is.NoErr(err)

		rd, err := recdef.Parse(f)
		is.NoErr(err)
		is.True(rd != nil) // should not return a nil RecDef
		is.Equal(rd.Attrprefix, "edm")

		is.Equal(len(rd.Root.Elem), 16)

		var buf bytes.Buffer
		err = rd.Write(&buf)
		is.NoErr(err)

		b, err := os.ReadFile(testFile)
		is.NoErr(err)
		is.True(len(b) > 0)

		os.WriteFile("/tmp/rec-def.xml", b, os.ModePerm)

		// if diff := cmp.Diff(string(b), buf.String()+"\n"); diff != "" {
		// t.Errorf("write recdef = mismatch (-want +got):\n%s", diff)
		// }
	})
}
