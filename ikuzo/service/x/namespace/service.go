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
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
)

const (
	ebuCoreURN = "urn:ebu:metadata-schema:ebuCore_2014"
)

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
				if _, err := s.Put(e.Prefix, e.BaseURI, e.Weight); err != nil {
					return nil, err
				}
			}
		}
	}

	return s, nil
}

func (s *Service) GetWithPrefix(prefix string) (*domain.Namespace, error) {
	s.checkStore()

	ns, err := s.store.List(&ListOptions{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	if len(ns) == 0 {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns[0], nil
}

func (s *Service) GetWithBase(baseURI string) (*domain.Namespace, error) {
	s.checkStore()

	ns, err := s.List(&ListOptions{URI: baseURI})
	if err != nil {
		return nil, err
	}

	if len(ns) == 0 {
		return nil, domain.ErrNamespaceNotFound
	}

	return ns[0], nil
}

// checkStore sets the default store when no store is set.
// This makes the default useful when the struct is directly initialized.
// The preferred way to initialize Service is by using NewService()
func (s *Service) checkStore() {
	if s.store == nil {
		s.store = NewNamespaceStore()
	}
}

// Put adds the prefix and base-URI to the namespace service.
func (s *Service) Put(prefix, base string, weight int) (*domain.Namespace, error) {
	s.checkStore()

	if base == "" || prefix == "" {
		return nil, domain.ErrNamespaceNotValid
	}

	ns := &domain.Namespace{
		Prefix: prefix,
		URI:    base,
		Weight: weight,
	}

	err := s.store.Put(ns)
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
func (s *Service) List(opts *ListOptions) ([]*domain.Namespace, error) {
	return s.store.List(opts)
}

// GetSearchLabel returns the URI in a short namespaced form.
// The string is formatted as namespace prefix
// and label joined with an underscore, e.g. "dc_title".
//
// The underscore is used instead of the more common colon because it mainly
// used as the search field in Lucene-based search engine, where it would
// conflict with the separator between the query-field and value.
func (s *Service) GetSearchLabel(uri string) (string, error) {
	s.checkStore()

	if strings.HasPrefix(uri, ebuCoreURN) {
		uri = strings.TrimPrefix(uri, ebuCoreURN)
		uri = strings.TrimLeft(uri, "/")
		uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
	}

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
