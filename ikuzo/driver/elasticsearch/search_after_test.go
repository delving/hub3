package elasticsearch

import (
	"testing"

	"github.com/matryer/is"
)

func Test_SearchAfterEncoding(t *testing.T) {
	searchAfter := []interface{}{"NL-HaNA_2-24-01-02ntfoto_ad8297b2-d0b4-102d-bcf8-003048976d84", 35469}
	encoded := "C38CAQL_gAABEAAAV_-AAAIGc3RyaW5nDD8APU5MLUhhTkFfMi0yNC0wMS0wMm50Zm90b19hZDgyOTdiMi1kMGI0LTEwMmQtYmNmOC0wMDMwNDg5NzZkODQDaW50BAUA_QEVGg=="

	t.Run("encode search after", func(t *testing.T) {
		is := is.New(t)
		s, err := encodeSearchAfter(searchAfter)
		is.NoErr(err)

		is.Equal(s, encoded)
	})

	t.Run("decode search after", func(t *testing.T) {
		is := is.New(t)
		got, err := decodeSearchAfter(encoded)
		is.NoErr(err)

		is.Equal(got, searchAfter)
	})
}
