package alto

import (
	"bytes"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestExtractText(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/NL-AsdNIOD_244_001954_0005_alto.xml")
	is.NoErr(err)

	var page Alto
	err = xml.NewDecoder(f).Decode(&page)
	is.NoErr(err)

	content, err := page.extractStrings()
	is.NoErr(err)
	is.Equal(len(content), 7)
	is.True(strings.HasPrefix(content[0], "duikvliegers"))
	t.Logf("last line: %s", content[6])
	is.True(strings.HasPrefix(content[6], "4.9"))
	is.True(strings.HasSuffix(content[6], "<br> mee tot Venlo. "))

	var buf bytes.Buffer
	n, err := page.WriteTo(&buf)
	is.NoErr(err)
	is.True(n != 0)
	is.True(buf.String() != "")
}
