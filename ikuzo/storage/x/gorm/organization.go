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

package gorm

import (
	"context"
	"fmt"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/jinzhu/gorm"
)

// compile time check to see if full interface is implemented
var _ organization.Store = (*OrganizationStore)(nil)

type OrganizationStore struct {
	db *gorm.DB
}

func NewOrganizationStore(db *gorm.DB) (*OrganizationStore, error) {
	if db == nil {
		return nil, fmt.Errorf("*gorm.DB cannot be nil")
	}

	db.AutoMigrate(domain.Organization{})

	return &OrganizationStore{db: db}, nil
}

func (o *OrganizationStore) Delete(ctx context.Context, id domain.OrganizationID) error {
	return o.db.Delete(domain.Organization{}, "ID = ?", id).Error
}

func (o *OrganizationStore) Get(ctx context.Context, id domain.OrganizationID) (domain.Organization, error) {
	var org domain.Organization

	if err := o.db.Where("ID = ?", id).First(&org).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return org, domain.ErrOrgNotFound
		}

		return org, err
	}

	return org, nil
}

func (o *OrganizationStore) getFilter(filter ...domain.OrganizationFilter) *gorm.DB {
	q := o.db

	if len(filter) != 0 {
		f := filter[0]
		if f.OffSet != 0 {
			q = q.Offset(f.OffSet)
		}

		if f.Limit > 0 {
			q = q.Limit(f.Limit)
		}

		if (f.Org != domain.Organization{}) {
			q = q.Where(&f.Org)
		}

		o.db.Where(&filter[0])
	}

	return q
}

func (o *OrganizationStore) Filter(ctx context.Context, filter ...domain.OrganizationFilter) ([]domain.Organization, error) {
	var orgs []domain.Organization

	q := o.getFilter(filter...)

	if err := q.Find(&orgs).Error; err != nil {
		return nil, err
	}

	return orgs, nil
}

func (o *OrganizationStore) Count(ctx context.Context, filter ...domain.OrganizationFilter) (int, error) {
	var (
		count int
	)

	q := o.getFilter(filter...)

	if err := q.Model(&domain.Organization{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (o *OrganizationStore) Put(ctx context.Context, org domain.Organization) error {
	return o.db.Save(&org).Error
}

func (o *OrganizationStore) Shutdown(ctx context.Context) error {
	if o.db != nil {
		o.db.Close()
	}

	return nil
}
