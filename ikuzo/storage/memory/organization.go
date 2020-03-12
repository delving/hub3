package memory

import (
	"context"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
)

// compile time check to see if full interface is implemented
// var _ Store = (*OrganizationStore)(nil)

type OrganizationStore struct {
	shutdownCalled bool
	sync.RWMutex
	organizations map[domain.OrganizationID]domain.Organization
}

func NewOrganizationStore() *OrganizationStore {
	return &OrganizationStore{
		organizations: map[domain.OrganizationID]domain.Organization{},
	}
}

func (ms *OrganizationStore) Delete(ctx context.Context, id domain.OrganizationID) error {
	delete(ms.organizations, id)
	return nil
}

func (ms *OrganizationStore) Get(ctx context.Context, id domain.OrganizationID) (domain.Organization, error) {
	org, ok := ms.organizations[id]
	if !ok {
		return domain.Organization{}, domain.ErrOrgNotFound
	}

	return org, nil
}

func (ms *OrganizationStore) List(ctx context.Context) ([]domain.Organization, error) {
	organizations := []domain.Organization{}
	for _, org := range ms.organizations {
		organizations = append(organizations, org)
	}

	return organizations, nil
}

func (ms *OrganizationStore) Put(ctx context.Context, org domain.Organization) error {
	ms.Lock()
	defer ms.Unlock()
	ms.organizations[org.ID] = org

	return nil
}

func (ms *OrganizationStore) Shutdown(ctx context.Context) error {
	ms.shutdownCalled = true
	return nil
}
