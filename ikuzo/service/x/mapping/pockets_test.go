package mapping

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestPockets(t *testing.T) {
	//nolint:gocritic
	is := is.New(t)

	f, err := os.Open("./testdata/source.xml")
	is.NoErr(err)

	defer f.Close()

	src, err := ParseSource(f)
	is.NoErr(err)

	t.Run("unmarshal", func(t *testing.T) {
		//nolint:gocritic
		// is := is.New(t)

		is.Equal(len(src.Pockets), 2)
		is.Equal(src.OrgID, "test")
		is.Equal(src.DatasetID, "spec")
		is.Equal(src.RecdefName, "edm")
		is.Equal(src.Pockets[1].ID, "2")
	})

	t.Run("marshal", func(t *testing.T) {
		//nolint:gocritic
		// is := is.New(t)

		var buf bytes.Buffer
		err := src.MarshalToXML(&buf)
		is.NoErr(err)

		f, err := os.Open("./testdata/source.xml")
		is.NoErr(err)
		golden, err := io.ReadAll(f)
		is.NoErr(err)

		got := buf.String()
		want := strings.TrimSuffix(string(golden), "\n")

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Source.MarshalToXML() = mismatch (-want +got):\n%s", diff)
		}
	})
}
