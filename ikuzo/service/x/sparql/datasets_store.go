package sparql

import (
	"context"
	fmt "fmt"
	"log"
	"strconv"

	"github.com/delving/hub3/ikuzo/domain"
)

// DataSetStore manage domain.DataSet specific options for
// saved rdf.Record instances.
type DataSetStore struct {
	repo      *Repo
	datasetID string
}

func NewDataSetStore(repo *Repo, datasetID string) (*DataSetStore, error) {
	ds := &DataSetStore{
		repo:      repo,
		datasetID: datasetID,
	}
	return ds, nil
}

// DeleteAllGraphsBySpec issues an SPARQL Update query to delete all graphs for a DataSet from the triple store
func (ds *DataSetStore) DeleteAllGraphsBySpec(ctx context.Context) (bool, error) {
	query, err := ds.repo.cfg.Bank.Prepare("deleteAllGraphsBySpec", struct{ Spec string }{ds.datasetID})
	if err != nil {
		log.Printf("Unable to build deleteAllGraphsBySpec query: %s", err)
		return false, err
	}

	err = ds.repo.Update(query)
	if err != nil {
		log.Printf("Unable query endpoint: %s", err)
		return false, err
	}

	return true, nil
}

// DeleteGraphsOrphansBySpec issues an SPARQL Update query to delete all orphaned graphs
// for a DataSet from the triple store.
func (ds *DataSetStore) DeleteGraphsOrphansBySpec(ctx context.Context, revision int) (bool, error) {
	query, err := ds.repo.cfg.Bank.Prepare("deleteOrphanGraphsBySpec", struct {
		Spec           string
		RevisionNumber int
	}{ds.datasetID, revision})
	if err != nil {
		log.Printf("sparql: unable to build deleteOrphanGraphsBySpec query: %s", err)
		return false, err
	}

	err = ds.repo.Update(query)
	if err != nil {
		log.Printf("sparql: unable query endpoint: %s", err)
		return false, err
	}

	return true, nil
}

// CountGraphsBySpec counts all the named graphs for a spec
func (ds *DataSetStore) CountGraphsBySpec() (int, error) {
	query, err := ds.repo.cfg.Bank.Prepare("countGraphPerSpec", struct{ Spec string }{ds.datasetID})
	if err != nil {
		log.Printf("sparql: unable to build CountGraphsBySpec query: %s", err)
		return 0, err
	}

	res, err := ds.repo.Query(query)
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
func (ds *DataSetStore) CountRevisionsBySpec() ([]domain.DataSetRevisions, error) {
	revisions := []domain.DataSetRevisions{}

	query, err := ds.repo.cfg.Bank.Prepare("countRevisionsBySpec", struct{ Spec string }{ds.datasetID})
	if err != nil {
		log.Printf("Unable to build countRevisionsBySpec query: %s", err)
		return revisions, err
	}

	res, err := ds.repo.Query(query)
	if err != nil {
		log.Printf("Unable query endpoint: %s", err)
		return revisions, err
	}

	for _, v := range res.Solutions() {
		revisionTerm, ok := v["revision"]
		if !ok {
			log.Printf("No revisions found for spec %s", ds.datasetID)
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
