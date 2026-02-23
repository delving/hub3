package semantic

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// Option is a functional option for configuring the Service.
type Option func(*Service) error

// WithStore sets the search store implementation.
func WithStore(store semantic.SearchStore) Option {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}

// WithRegistry sets a custom configuration registry.
func WithRegistry(registry *semantic.ConfigRegistry) Option {
	return func(s *Service) error {
		s.registry = registry
		return nil
	}
}

// WithBaseURL sets the base URL for the service.
func WithBaseURL(baseURL string) Option {
	return func(s *Service) error {
		s.baseURL = baseURL
		return nil
	}
}

// WithIntrospectionStore sets the introspection store for the service.
func WithIntrospectionStore(store semantic.IntrospectionStore) Option {
	return func(s *Service) error {
		s.introspect = store
		return nil
	}
}

// WithIncludeProvider registers an include provider for the detail endpoint.
func WithIncludeProvider(provider semantic.IncludeProvider) Option {
	return func(s *Service) error {
		if provider == nil {
			return fmt.Errorf("include provider cannot be nil")
		}
		s.includes[provider.Name()] = provider
		return nil
	}
}

// WithResourceConfig adds a resource configuration to the registry.
func WithResourceConfig(config *semantic.ResourceConfig) Option {
	return func(s *Service) error {
		if s.registry == nil {
			s.registry = semantic.NewConfigRegistry()
		}
		s.registry.Register(config)
		return nil
	}
}
