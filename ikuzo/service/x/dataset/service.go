package dataset

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/service/x/revision"
	"github.com/delving/hub3/ikuzo/service/x/search/es"
	"github.com/delving/hub3/ikuzo/service/x/sparql"
	"github.com/rs/zerolog"
)

var (
	ErrDataSetNotFound = errors.New("dataset not found")
)

type Store interface {
	// Delete a DataSet
	Delete(ctx context.Context, ds *DataSet) error
	// List all datasets
	List(ctx context.Context) ([]*DataSet, error)
	// Get one dataset
	One(ctx context.Context, orgID, datasetID string) (*DataSet, error)
	// delete all records
	Reset(ctx context.Context) error
	// Save a DataSet
	Save(ctx context.Context, ds *DataSet) error
	// TODO(kiivihal): decide on update field
	Shutdown(ctx context.Context) error
}

type Option func(*Service) error

type Service struct {
	store    Store
	org      *organization.Service
	log      zerolog.Logger
	search   *es.Service
	sparql   *sparql.Service
	revision *revision.Service
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		log: zerolog.Nop(),
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.log.GetLevel() != zerolog.Disabled {
		s.log = s.log.With().
			Str("component", "hub3").
			Str("svc", "dataset").
			Logger()
	}

	return s, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.store != nil {
		if err := s.store.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) CreateDataSetStats(ctx context.Context, orgID, datasetID string) error {

	return nil
}

// GetDataSet returns a DataSet from the store.
// An ErrDataSetNotFound error is returned when no DataSet is found.
func (s *Service) GetDataSet(ctx context.Context, orgID, datasetID string) (*DataSet, error) {
	ds, err := s.store.One(ctx, orgID, datasetID)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// GetOrCreateDataSet returns a DataSet or creates one if it is not found in the store.
func (s *Service) GetOrCreateDataSet(ctx context.Context, orgID, datasetID string) (*DataSet, bool, error) {
	ds, err := s.GetDataSet(ctx, orgID, datasetID)
	if err != nil {
		return s.createDataSet(ctx, orgID, datasetID)
	}

	return ds, false, nil
}

// createDataSet creates and saves a DataSet to the DataSet storage
func (s *Service) createDataSet(ctx context.Context, orgID, datasetID string) (*DataSet, bool, error) {
	id, err := domain.NewOrganizationID(orgID)
	if err != nil {
		return nil, false, err
	}

	org, err := s.org.Get(ctx, id)
	if err != nil {
		return nil, false, err
	}

	ds := NewDataset(*org, datasetID)
	if err := s.store.Save(ctx, &ds); err != nil {
		return nil, false, err
	}

	return &ds, true, nil
}

// Delete removes the DataSet from the DataSet storage
//
// TODO(kiivihal): make sure delete also removes the dataset from all underlying storage
func (s *Service) Delete(ctx context.Context, ds *DataSet) error {
	s.log.Info().
		Str("component", "hub3").
		Str("datasetID", ds.Spec).
		Str("svc", "dataset").
		Msg("deleting dataset")

	return s.store.Delete(ctx, ds)
}

// List returns a DataSet list
//
// TODO(kiivihal): add filter and pagination options later
func (s *Service) List(ctx context.Context) ([]*DataSet, error) {
	return s.store.List(ctx)
}

// Save stored the DataSet in the storage.
//
// The underlying storage returns an error when the DataSet cannot be saved.
func (s *Service) Save(ctx context.Context, ds *DataSet) error {
	ds.Modified = time.Now()
	return s.store.Save(ctx, ds)
}

func (s *Service) DropResources(ctx context.Context, datasetID string) error {
	// TODO(kiivihal): implement me call each registered resource store and drop resources
	return nil

}
