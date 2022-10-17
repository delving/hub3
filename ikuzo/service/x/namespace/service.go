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

package namespace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
)

// Store provides functionality to query and persist namespaces.
type Store interface {
	// Put persists the NameSpace object.
	//
	// When the object already exists it is overwritten.
	Put(ns *domain.Namespace) error

	// Delete removes the NameSpace from the store.
	//
	// Delete matches by the Prefix of the Namespace.
	Delete(ID string) error

	// Len returns the number of stored namespaces
	Len() int

	// Search returns a filtered list Namespaces
	// Search() ([]*domain.Namespace, error)

	// Get returns a namespace by its ID
	Get(id string) (ns *domain.Namespace, err error)

	// GetWithPrefix returns the NameSpace for a given prefix.
	// When the prefix is not found, an ErrNameSpaceNotFound error is returned.
	GetWithPrefix(prefix string) (ns *domain.Namespace, err error)

	// GetWithBase returns the NameSpace for a given base-URI.
	// When the base-URI is not found, an ErrNameSpaceNotFound error is returned.
	GetWithBase(base string) (ns *domain.Namespace, err error)

	// List returns a list of all the NameSpaces
	List() ([]*domain.Namespace, error)
}

// ServiceOptionFunc is a function that configures a Service.
// It is used in NewService.
type ServiceOptionFunc func(*Service) error

// Service provides functionality to query and persist namespaces.
type Service struct {
	// store is the backend where namespaces are stored
	// It defaults to the memoryStore.
	// Other implementation can be set as an ServiceOptFunc
	store Store

	// loadDefaults determines if the defaults are loaded into the store
	// when it is empty.
	loadDefaults bool
}

// NewService creates a new client to work with namespaces.
//
// NewService, by default, is meant to be long-lived and shared across
// your application.
//
// The caller can configure the new service by passing configuration options
// to the func.
//
// Example:
//
//	  service, err := namespace.NewService(
//	    namespace.WithDefaults(),
//		 )
//
// If no Store is configured, Service uses a in-memory store by default.
//
// An error is also returned when some configuration option is invalid.
func NewService(options ...ServiceOptionFunc) (*Service, error) {
	s := &Service{}

	// Run the options on it
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.loadDefaults {
		for _, nsMap := range [][]nsEntry{defaultNS, customNS} {
			for _, e := range nsMap {
				if _, err := s.Put(e.Prefix, e.BaseURI); err != nil {
					return nil, err
				}
			}
		}
	}

	return s, nil
}

// SetStore sets the persistence store for the namespace.Service.
func SetStore(store Store) ServiceOptionFunc {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}

// WithDefaults enables the namespace.Store to be initialize with default namespaces
func WithDefaults() ServiceOptionFunc {
	return func(s *Service) error {
		s.loadDefaults = true
		return nil
	}
}

// checkStore sets the default store when no store is set.
// This makes the default useful when the struct is directly initialized.
// The preferred way to initialize Service is by using NewService()
func (s *Service) checkStore() {
	if s.store == nil {
		s.store = memory.NewNamespaceStore()
	}
}

// Put adds the prefix and base-URI to the namespace service.
// When either the prefix or the base-URI is already present in the service the
// unknown is stored as an alternative. If neither is present a new NameSpace
// is created.
func (s *Service) Put(prefix, base string) (*domain.Namespace, error) {
	s.checkStore()

	if base == "" {
		return nil, domain.ErrNamespaceNotValid
	}

	if prefix == "" {
		ns := &domain.Namespace{
			Base:      base,
			Temporary: true,
		}
		ns.Prefix = ns.GetID()

		err := s.store.Put(ns)
		if err != nil {
			return nil, err
		}

		return ns, nil
	}

	ns, err := s.store.GetWithPrefix(prefix)
	if err != nil {
		if err != domain.ErrNamespaceNotFound {
			return nil, err
		}
	}

	if ns != nil {
		if base != ns.Base {
			// base is not linked to the NameSpace
			// so creating a new temporary NameSpace
			ns = &domain.Namespace{
				Base:      base,
				PrefixAlt: []string{prefix},
				Temporary: true,
			}
			ns.Prefix = ns.GetID()

			err = s.store.Put(ns)
			if err != nil {
				return nil, err
			}
		}

		return ns, nil
	}

	ns, err = s.GetWithBase(base)
	if err != nil {
		if err != domain.ErrNamespaceNotFound {
			return nil, err
		}
	}

	if ns != nil {
		err = ns.AddPrefix(prefix)
		if err != nil {
			return nil, err
		}

		err = s.store.Put(ns)
		if err != nil {
			return nil, err
		}

		return ns, nil
	}

	ns = &domain.Namespace{
		Prefix: prefix,
		Base:   base,
	}

	err = s.store.Put(ns)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

// Get returns a Namespace by its identifier.
// When the it is not found it returns domain.ErrNameSpaceNotFound
func (s *Service) Get(id string) (*domain.Namespace, error) {
	return s.store.Get(id)
}

// Delete removes a namespace from the store
func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}

// Len returns the number of namespaces in the Service
func (s *Service) Len() int {
	s.checkStore()
	return s.store.Len()
}

// List returns a list of all stored NameSpace objects.
// An error is returned when the underlying storage can't be accessed.
func (s *Service) List() ([]*domain.Namespace, error) {
	return s.store.List()
}

// SearchLabel returns the URI in a short namespaced form.
// The string is formatted as namespace prefix
// and label joined with an underscore, e.g. "dc_title".
//
// The underscore is used instead of the more common colon because it mainly
// used as the search field in Lucene-based search engine, where it would
// conflict with the separator between the query-field and value.
func (s *Service) SearchLabel(uri string) (string, error) {
	s.checkStore()

	base, label := domain.SplitURI(uri)

	ns, err := s.GetWithBase(base)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve namespace for %s; %w", base, err)
	}

	return fmt.Sprintf("%s_%s", ns.Prefix, label), nil
}

// Set sets the default prefix and base-URI for a namespace.
// When the namespace is already present it will be overwritten.
// When the NameSpace contains an unknown prefix and base-URI pair but one of them
// is found in the NameSpace service, the current default is stored in PrefixAlt
// or BaseAlt and the new default set.
func (s *Service) Set(ns *domain.Namespace) error {
	s.checkStore()
	return s.store.Put(ns)
}

func (s *Service) GetWithPrefix(prefix string) (*domain.Namespace, error) {
	s.checkStore()
	return s.store.GetWithPrefix(prefix)
}

func (s *Service) GetWithBase(baseURI string) (*domain.Namespace, error) {
	s.checkStore()

	ns, err := s.store.GetWithBase(baseURI)
	if err != nil {
		if errors.Is(err, domain.ErrNamespaceNotFound) && strings.HasPrefix(baseURI, "https://") {
			baseURI = strings.ReplaceAll(baseURI, "https:", "http:")
			return s.store.GetWithBase(baseURI)
		}
	}

	return ns, err
}
