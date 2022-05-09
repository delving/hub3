package file

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestParseNarthexNtriples(t *testing.T) {
	is := is.New(t)

	idx, err := newIndex("test", "testdataset", FormatNTriples)
	is.NoErr(err)

	path := "./testdata/rdf/00000_sha256-8cc5389997ef49980ccc0665bcc0ad2124700a4c7250a90cdd992e77710422c2_v1.nq"

	f, err := os.Open(path)
	is.NoErr(err)

	defer f.Close()

	err = idx.ParseNarthexNtriples(
		f,
		path,
		"sha256-8cc5389997ef49980ccc0665bcc0ad2124700a4c7250a90cdd992e77710422c2",
	)
	is.NoErr(err)

	is.Equal(len(idx.records), len(idx.lookUp))
	is.Equal(len(idx.lookUp), 1284)

	rp, ok := idx.lookUp["NL-HaNA_2-21-284ntfoto_22325057-f020-4d14-b2ab-fc43bd2bed8e"]
	is.True(ok)
	is.Equal(rp.HubID, "NL-HaNA_2-21-284ntfoto_22325057-f020-4d14-b2ab-fc43bd2bed8e")
	is.Equal(rp.Lines, int64(53))

	b, err := idx.Data(rp)
	is.NoErr(err)
	t.Logf("rp %#v", rp)
	is.Equal(rp.Offset, int64(14466300-1))

	testb, err := os.ReadFile("./testdata/rdf/test_record.nq")
	is.NoErr(err)

	if diff := cmp.Diff(string(testb), string(b)); diff != "" {
		t.Errorf("ParseNarthexNtriples() = mismatch (-want +got):\n%s", diff)
	}

	err = idx.Close()
	is.NoErr(err)
}
