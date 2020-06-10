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

	"github.com/delving/hub3/ikuzo/domain"
)

type FilterOption struct {
	// OffSet is the start of the list where
	OffSet int
	Limit  int
	Org    domain.Organization
}

// Store is the storage interface for the organization.Service.
type Store interface {
	Delete(ctx context.Context, id domain.OrganizationID) error
	Get(ctx context.Context, id domain.OrganizationID) (domain.Organization, error)
	Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]domain.Organization, error)
	Put(ctx context.Context, org domain.Organization) error
	Shutdown(ctx context.Context) error
}

// Service manages all interactions with domain.Organization Store
type Service struct {
	store Store
}

// NewService creates an organization.Service.
// The organization.Store implementation is the storage backend for the service.
func NewService(store Store) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("organization.Store implementation cannot be nil")
	}

	return &Service{store: store}, nil
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

// Filter returns a list of domain.Organization based on the filterOptions.
//
// When the filterOptions are nil, the first 10 are returned.
func (s *Service) Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]domain.Organization, error) {
	return s.store.Filter(ctx, filter...)
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
