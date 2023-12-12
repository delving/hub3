package elasticsearch

import (
	"testing"

	"github.com/matryer/is"
)

func Test_SearchAfterEncoding(t *testing.T) {
	searchAfter := []interface{}{"NL-HaNA_2-24-01-02ntfoto_ad8297b2-d0b4-102d-bcf8-003048976d84", 35469}
	// encoded := "DP-BAgEC_4IAARAAAFf_ggACBnN0cmluZww_AD1OTC1IYU5BXzItMjQtMDEtMDJudGZvdG9fYWQ4Mjk3YjItZDBiNC0xMDJkLWJjZjgtMDAzMDQ4OTc2ZDg0A2ludAQFAP0BFRo="
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
