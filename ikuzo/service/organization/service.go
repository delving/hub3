package organization

import (
	"context"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/storage/memory"
)

// Store is the storage interface for the organization.Service.
type Store interface {
	Delete(ctx context.Context, id domain.OrganizationID) error
	Get(ctx context.Context, id domain.OrganizationID) (domain.Organization, error)
	List(ctx context.Context) ([]domain.Organization, error)
	Put(ctx context.Context, org domain.Organization) error
	Shutdown(ctx context.Context) error
}

// Service manages all interactions with domain.Organization Store
type Service struct {
	store Store
}

// NewService creates an organization.Service.
// The organization.Store implementation is the storage backend for the service.
func NewService(store Store) *Service {
	if store == nil {
		return &Service{store: memory.NewOrganizationStore()}
	}

	return &Service{store: store}
}

// Delete removes the domain.Organization from the Organization Store.
func (s *Service) Delete(ctx context.Context, id domain.OrganizationID) error {
	return s.store.Delete(ctx, id)
}

// Get returns an domain.Organization and returns  ErrOrgNotFound when the Organization
// is not found.
func (s *Service) Get(ctx context.Context, id domain.OrganizationID) (domain.Organization, error) {
	return s.store.Get(ctx, id)
}

// List returns a list of domain.Organization from the Organization Store
func (s *Service) List(ctx context.Context) ([]domain.Organization, error) {
	return s.store.List(ctx)
}

// Put stores an Organization in the Service Store.
func (s *Service) Put(ctx context.Context, org domain.Organization) error {
	if err := org.ID.Valid(); err != nil {
		return err
	}

	return s.store.Put(ctx, org)
}

// Shutdown gracefully shutsdown the organization.Service store.
// The ctx should have a timeout that cancels when the deadline is exceeded.
func (s *Service) Shutdown(ctx context.Context) error {
	return s.store.Shutdown(ctx)
}
