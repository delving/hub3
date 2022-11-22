package hextuples

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestHexTupleFields(t *testing.T) {
	t.Run("date literal", func(t *testing.T) {
		is := is.New(t)
		b := []byte(
			`["https://www.w3.org/People/Berners-Lee/","http://schema.org/birthDate","1955-06-08","http://www.w3.org/2001/XMLSchema#date","",""]`,
		)

		ht, err := New(b)
		is.NoErr(err)
		is.Equal(ht.Subject, "https://www.w3.org/People/Berners-Lee/")
		is.Equal(ht.Predicate, "http://schema.org/birthDate")
		is.Equal(ht.Value, "1955-06-08")
		is.Equal(ht.DataType, "http://www.w3.org/2001/XMLSchema#date")
		is.Equal(ht.Language, "")
		is.Equal(ht.Graph, "")

		htBytes, err := ht.MarshalJSON()
		is.NoErr(err)
		if diff := cmp.Diff(b, htBytes); diff != "" {
			t.Errorf("HexTuple JsonMarshal() = mismatch (-want +got):\n%s", diff)
		}
	})
}
