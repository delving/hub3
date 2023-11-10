package bulk

import "github.com/delving/hub3/hub3/fragments"

func (s *Service) getPreviousUpdates(ids []string) ([]*fragments.SparqlUpdate, error) {
	return []*fragments.SparqlUpdate{}, nil
}

func (s *Service) storeUpdatedHashes(diffs []*DiffConfig) error {
	return nil
}

func (s *Service) incrementRevisionForSeen(ids []string) error {
	return nil
}

type sparqlOrphan struct {
	NamedGraphURI string
	OrgID         string
	DatasetID     string
}

func (s *Service) findOrphans(orgID, dataSetID string, revision int) ([]sparqlOrphan, error) {
	return []sparqlOrphan{}, nil
}

func (s *Service) dropOrphans(orphans []sparqlOrphan) error {
	return nil
}
