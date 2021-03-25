package sparql

import (
	"context"
	"log"
)

// DeleteAllGraphsBySpec issues an SPARQL Update query to delete all graphs for a DataSet from the triple store
func (s *Service) DeleteAllGraphsBySpec(ctx context.Context, spec string) (bool, error) {
	query, err := s.bank.Prepare("deleteAllGraphsBySpec", struct{ Spec string }{spec})
	if err != nil {
		log.Printf("Unable to build deleteAllGraphsBySpec query: %s", err)
		return false, err
	}

	err = s.store.SparqlUpdate(ctx, query)
	if err != nil {
		log.Printf("Unable query endpoint: %s", err)
		return false, err
	}

	return true, nil
}

// DeleteGraphsOrphansBySpec issues an SPARQL Update query to delete all orphaned graphs
// for a DataSet from the triple store.
func (s *Service) DeleteGraphsOrphansBySpec(ctx context.Context, spec string, revision int) (bool, error) {
	query, err := s.bank.Prepare("deleteOrphanGraphsBySpec", struct {
		Spec           string
		RevisionNumber int
	}{spec, revision})
	if err != nil {
		log.Printf("sparql: unable to build deleteOrphanGraphsBySpec query: %s", err)
		return false, err
	}

	err = s.store.SparqlUpdate(ctx, query)
	if err != nil {
		log.Printf("sparql: unable query endpoint: %s", err)
		return false, err
	}

	return true, nil
}
