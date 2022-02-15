package sparql

import (
	"bytes"
	"testing"

	"github.com/knakk/sparql"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestMergeBank(t *testing.T) {
	is := is.New(t)

	s, err := NewService()
	is.NoErr(err)

	queryCount := len(s.bank)
	is.True(queryCount != 0)

	const queries = `
# Comments are ignored, except those tagging a query.

# tag: my-query
SELECT *
WHERE {
  ?s ?p ?o
} LIMIT {{.Limit}} OFFSET {{.Offset}}

# tag: describe
DESCRIBE <{{.URI}}>
`

	f := bytes.NewBufferString(queries)
	bank := sparql.LoadBank(f)

	is.True(len(bank) == 2)

	err = s.mergeBank(bank)
	is.NoErr(err)

	t.Logf("bank: %+v", s.bank)

	is.Equal(queryCount+1, len(s.bank))
}
