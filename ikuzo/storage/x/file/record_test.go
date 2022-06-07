package file

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestRecordSeparator(t *testing.T) {
	is := is.New(t)

	comment := `# !{"hubID":"NL-HaNA_2-21-284ntfoto_7ea6f4cc-2013-1d7f-e053-09f0900a4002","orgID":"NL-HaNA","localID":"7ea6f4cc-2013-1d7f-e053-09f0900a4002","graphURI":"https://archief.nl/doc/Fotocollectie/2.21.284ntfoto/graph","datasetID":"2-21-284ntfoto","contentHash":"2094dc2f4083943eee14ea784ef10f06b9dc4b95"}`

	ok := IsRecordSeparator([]byte(comment))
	is.True(ok)

	want := RecordSeparator{
		OrgID:       "NL-HaNA",
		DatasetID:   "2-21-284ntfoto",
		HubID:       "NL-HaNA_2-21-284ntfoto_7ea6f4cc-2013-1d7f-e053-09f0900a4002",
		LocalID:     "7ea6f4cc-2013-1d7f-e053-09f0900a4002",
		GraphURI:    "https://archief.nl/doc/Fotocollectie/2.21.284ntfoto/graph",
		ContentHash: "2094dc2f4083943eee14ea784ef10f06b9dc4b95",
	}

	sep, err := NewRecordSeparator([]byte(comment))
	is.NoErr(err)
	if diff := cmp.Diff(want, sep); diff != "" {
		t.Errorf("NewRecordSeparator() = mismatch (-want +got):\n%s", diff)
	}
}

func TestFileHash(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/rdf/00000_sha256-8cc5389997ef49980ccc0665bcc0ad2124700a4c7250a90cdd992e77710422c2_v1.nq")
	is.NoErr(err)
	defer f.Close()

	hash, err := fileHash(f)
	is.NoErr(err)
	is.Equal(hash, "sha256-8cc5389997ef49980ccc0665bcc0ad2124700a4c7250a90cdd992e77710422c2")
}
