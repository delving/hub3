package sru

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestReadSRU(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/sru.xml")
	is.NoErr(err)

	resp, err := newResponse(f)
	is.NoErr(err)
	is.True(resp != nil)
	is.Equal(resp.NumberOfRecords__srw, "200")

	echo := resp.EchoedSearchRetrieveRequest__srw
	is.True(echo != nil)
	is.Equal(echo.MaximumRecords__srw, "10")
	is.Equal(echo.StartRecord__srw, "1")

	records := resp.Records__srw
	is.True(records != nil)
	is.Equal(len(records.Record__srw), 10)
	first := records.Record__srw[0]
	is.True(first != nil)
	is.Equal(first.RecordPosition__srw, "1")
	is.True(first.RecordData__srw != nil)
}
