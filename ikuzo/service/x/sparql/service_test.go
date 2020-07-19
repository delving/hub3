package sparql

import (
	"bytes"
	"context"
	"testing"

	"github.com/knakk/sparql"
	"github.com/matryer/is"
)

// compile time check to see if full interface is implemented
var _ TripleStore = (*mockTripleStore)(nil)

type mockTripleStore struct{}

func (m *mockTripleStore) CreateDB(ctx context.Context, name string) error      { return nil }
func (m *mockTripleStore) DropDB(ctx context.Context, name string) error        { return nil }
func (m *mockTripleStore) SparqlUpdate(ctx context.Context, query string) error { return nil }
func (m *mockTripleStore) SparqlQuery(ctx context.Context, query string) (sparql.Results, error) {
	return sparql.Results{}, nil
}

// nolint:gocritic
func TestMergeBank(t *testing.T) {
	is := is.New(t)

	s, err := NewService(
		SetTripleStore(&mockTripleStore{}),
	)
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
