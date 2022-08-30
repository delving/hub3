package record

import (
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
)

func (s *Service) Add(record *rdf.Record) error {
	return nil
}

func (s *Service) Get(hubID domain.HubID) *rdf.Record {
	return nil
}

// AddTemp adds record to temporary storage to resolve it with main records later
func (s *Service) AddTemp(record *rdf.Record) error {
	return nil
}
