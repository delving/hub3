package sparql

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/knakk/sparql"
)

type TripleStore interface {
	CreateDB(ctx context.Context, name string) error
	DropDB(ctx context.Context, name string) error
	SparqlUpdate(ctx context.Context, query string) error
	SparqlQuery(ctx context.Context, query string) (sparql.Results, error)
}

type Service struct {
	bank  sparql.Bank
	store TripleStore
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	f := bytes.NewBufferString(queries)
	s.bank = sparql.LoadBank(f)

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.store == nil {
		return s, fmt.Errorf("sparql: TripleStore interface must have a concrete implementation")
	}

	return s, nil
}

// mergeBank overrides queries on the service query bank when they have the same key.
func (s *Service) mergeBank(customBank sparql.Bank) error {
	for k, v := range customBank {
		s.bank[k] = v
	}

	return nil
}

// implement sparql proxy implementation
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

// should connections be shutdown
func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
