package sparql

import (
	"bytes"

	"github.com/knakk/sparql"
)

type Option func(*Service) error

// SetCustomQueries allows custom Sparql Queries to be added to the sparql.Bank
//
// Each query must be preceded by a comment otherwise the query is silently ignored.
func SetCustomQueries(queries string) Option {
	return func(s *Service) error {
		f := bytes.NewBufferString(queries)
		return s.mergeBank(sparql.LoadBank(f))
	}
}
