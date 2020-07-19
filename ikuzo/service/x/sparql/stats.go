package sparql

import (
	"fmt"
	"log"
	"strconv"

	"github.com/delving/hub3/ikuzo/domain"
)

// CountGraphsBySpec counts all the named graphs for a spec
func (s *Service) CountGraphsBySpec(spec string) (int, error) {
	query, err := s.bank.Prepare("countGraphPerSpec", struct{ Spec string }{spec})
	if err != nil {
		log.Printf("sparql: unable to build CountGraphsBySpec query: %s", err)
		return 0, err
	}

	res, err := s.store.SparqlQuery(query)
	if err != nil {
		log.Printf("sparql: unable query endpoint: %s", err)
		return 0, err
	}

	countStr, ok := res.Bindings()["count"]
	if !ok {
		return 0, fmt.Errorf("sparql: unable to get count from result bindings: %#v", res.Bindings())
	}

	var count int

	count, err = strconv.Atoi(countStr[0].String())
	if err != nil {
		return 0, fmt.Errorf("sparql: unable to convert count %s to integer", countStr)
	}

	return count, err
}

// CountRevisionsBySpec counts each revision available in the spec
func (s *Service) CountRevisionsBySpec(spec string) ([]domain.DataSetRevisions, error) {
	revisions := []domain.DataSetRevisions{}

	query, err := s.bank.Prepare("countRevisionsBySpec", struct{ Spec string }{spec})
	if err != nil {
		log.Printf("Unable to build countRevisionsBySpec query: %s", err)
		return revisions, err
	}

	res, err := s.store.SparqlQuery(query)
	if err != nil {
		log.Printf("Unable query endpoint: %s", err)
		return revisions, err
	}

	for _, v := range res.Solutions() {
		revisionTerm, ok := v["revision"]
		if !ok {
			log.Printf("No revisions found for spec %s", spec)
			return revisions, nil
		}

		revision, err := strconv.Atoi(revisionTerm.String())
		if err != nil {
			return revisions, fmt.Errorf("unable to convert %#v to integer", v["revision"])
		}

		revisionCount, err := strconv.Atoi(v["rCount"].String())
		if err != nil {
			return revisions, fmt.Errorf("unable to convert %#v to integer", v["rCount"])
		}

		revisions = append(revisions, domain.DataSetRevisions{
			Number:      revision,
			RecordCount: revisionCount,
		})
	}

	return revisions, nil
}
