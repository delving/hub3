// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"context"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
)

// compile time check to see if full interface is implemented
var _ organization.Store = (*OrganizationStore)(nil)

type OrganizationStore struct {
	shutdownCalled bool
	rw             sync.RWMutex
	organizations  map[domain.OrganizationID]*domain.Organization
	domains        map[string]*domain.Organization
}

func NewOrganizationStore() *OrganizationStore {
	return &OrganizationStore{
		organizations: map[domain.OrganizationID]*domain.Organization{},
		domains:       map[string]*domain.Organization{},
	}
}

func (ms *OrganizationStore) Delete(ctx context.Context, id domain.OrganizationID) error {
	delete(ms.organizations, id)
	return nil
}

func (ms *OrganizationStore) Get(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	org, ok := ms.organizations[id]
	if !ok {
		return &domain.Organization{}, domain.ErrOrgNotFound
	}

	return org, nil
}

// Filter returns a subset of the available domains. When the filter returns no result domain.ErrOrgNotFound is returned.
func (ms *OrganizationStore) Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]*domain.Organization, error) {
	organizations := []*domain.Organization{}

	// TODO(kiivihal): only accept first filter now
	if len(filter) != 0 {
		filt := filter[0]
		if filt.Domain != "" {
			org, ok := ms.domains[filt.Domain]
			if !ok {
				return organizations, domain.ErrOrgNotFound
			}

			return []*domain.Organization{org}, nil
		}
	}

	for _, org := range ms.organizations {
		organizations = append(organizations, org)
	}

	return organizations, nil
}

func (ms *OrganizationStore) Put(ctx context.Context, org *domain.Organization) error {
	ms.rw.Lock()
	defer ms.rw.Unlock()
	ms.organizations[org.ID] = org

	for _, domain := range org.Config.Domains {
		ms.domains[domain] = org
	}

	return nil
}

func (ms *OrganizationStore) Shutdown(ctx context.Context) error {
	ms.shutdownCalled = true
	return nil
}
