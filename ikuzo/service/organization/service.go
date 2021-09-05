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

package organization

import (
	"context"
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

var _ domain.Service = (*Service)(nil)

type FilterOption struct {
	// OffSet is the start of the list where
	OffSet int
	Limit  int
	Org    domain.Organization
}

// Store is the storage interface for the organization.Service.
type Store interface {
	Delete(ctx context.Context, id domain.OrganizationID) error
	Get(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
	Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]*domain.Organization, error)
	Put(ctx context.Context, org *domain.Organization) error
	Shutdown(ctx context.Context) error
}

// Service manages all interactions with domain.Organization Store
type Service struct {
	store Store
	log   zerolog.Logger
}

// NewService creates an organization.Service.
// The organization.Store implementation is the storage backend for the service.
func NewService(store Store) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("organization.Store implementation cannot be nil")
	}

	return &Service{store: store}, nil
}

func (s *Service) AddOrgs(orgs map[string]domain.OrganizationConfig) error {
	for id, orgCfg := range orgs {
		orgCfg.SetOrgID(id)

		if orgCfg.CustomID != "" {
			id = orgCfg.CustomID
		}

		orgID, err := domain.NewOrganizationID(id)
		if err != nil {
			return fmt.Errorf("unable to create domain.OrganizationID %s; %w", id, err)
		}

		org := domain.Organization{
			Config: orgCfg,
			ID:     orgID,
		}

		if err := s.Put(context.TODO(), &org); err != nil {
			return fmt.Errorf("unable to store Organization; %w", err)
		}
	}

	return nil
}

// Delete removes the domain.Organization from the Organization Store.
func (s *Service) Delete(ctx context.Context, id domain.OrganizationID) error {
	return s.store.Delete(ctx, id)
}

// Get returns an domain.Organization and returns  ErrOrgNotFound when the Organization
// is not found.
func (s *Service) Get(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	return s.store.Get(ctx, id)
}

// RetrieveConfig returns a domain.OrganizationConfig.
// When it is not found false is returned
func (s *Service) RetrieveConfig(orgID string) (cfg domain.OrganizationConfig, ok bool) {
	id, err := domain.NewOrganizationID(orgID)
	if err != nil {
		return cfg, false
	}
	org, err := s.Get(context.TODO(), id)
	if err != nil {
		return cfg, false
	}

	return org.Config, true
}

// Filter returns a list of domain.Organization based on the filterOptions.
//
// When the filterOptions are nil, the first 10 are returned.
func (s *Service) Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]*domain.Organization, error) {
	return s.store.Filter(ctx, filter...)
}

// Put stores an Organization in the Service Store.
func (s *Service) Put(ctx context.Context, org *domain.Organization) error {
	if err := org.ID.Valid(); err != nil {
		return err
	}

	return s.store.Put(ctx, org)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// Shutdown gracefully shutsdown the organization.Service store.
// The ctx should have a timeout that cancels when the deadline is exceeded.
func (s *Service) Shutdown(ctx context.Context) error {
	return s.store.Shutdown(ctx)
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "organization").Logger()
}
