package namespace

import (
	"fmt"

	"github.com/pkg/errors"
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
//   service, err := elastic.NewService(
//     namespace.WithDefaults(),
//	 )
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
		for _, nsMap := range []map[string]string{defaultNS, customNS} {
			for prefix, base := range nsMap {
				if err := s.Add(prefix, base); err != nil {
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

// WithDefaults enables the namespace.Store to be initialise with default namespaces
func WithDefaults() ServiceOptionFunc {
	return func(s *Service) error {
		s.loadDefaults = true
		return nil
	}
}

// checkStore sets the default store when no store is set.
// This makes the default useful when the struct is directly initialised.
// The prefered way to initialise Service is by using NewService()
func (s *Service) checkStore() {
	if s.store == nil {
		s.store = newMemoryStore()
	}
}

// Len returns the number of namespaces in the Service
func (s *Service) Len() int {
	s.checkStore()
	return s.store.Len()
}

// Set sets the default prefix and base-URI for a namespace.
// When the namespace is already present it will be overwritten.
// When the NameSpace contains an unknown prefix and base-URI pair but one of them
// is found in the NameSpace service, the current default is stored in PrefixAlt
// or BaseAlt and the new default set.
func (s *Service) Set(ns *NameSpace) error {
	s.checkStore()
	return s.store.Set(ns)
}

// Add adds the prefix and base-URI to the namespace service.
// When either the prefix or the base-URI is already present in the service the
// unknown is stored as an alternative. If neither is present a new NameSpace
// is created.
func (s *Service) Add(prefix, base string) error {
	s.checkStore()
	_, err := s.store.Add(prefix, base)
	return err
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
	base, label := SplitURI(uri)
	ns, err := s.store.GetWithBase(base)
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve namespace for %s", base)
	}
	return fmt.Sprintf("%s_%s", ns.Prefix, label), nil
}

//type Service interface {

//// Set sets the default prefix and base-URI for a namespace.
//// When the namespace is already present it will be overwritten.
////
//// When you want to add alternative base-URIs use Add(prefix, base string)
//Set(prefix, base string) error

//// Delete removes the prefix and/or base combination from the Store.
////
//// The either the prefix or the base can be empty. When you specify the
//// prefix or when the base is the primary NameSpace
//// base-URI the whole NameSpace is removed. When the base is an alternative it is removed
//// from the alternative list.
////
//// If you want to update the prefix and the base use Set
//Delete(prefix, base string) error

//// Add adds a NameSpace to the Service.
//// When the prefix is already present the base is stored as an alternative.
//Add(prefix, base string) error

//// Len returns the number of namespaces in the Service
//Len() int

//// GetBase returns the NameSpace for a given prefix.
//// When the prefix is not found it returns false for ok
//GetBase(prefix string) (ns *NameSpace, ok bool)

//// GetPrefix returns the NameSpace for a given base-URI.
//// When the base-URI is not found it returns false for ok
//GetPrefix(base string) (ns *NameSpace, ok bool)
//}
