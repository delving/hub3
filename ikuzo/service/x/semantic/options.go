package semantic

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// Option is a functional option for configuring the Service.
type Option func(*Service) error

// WithStore sets the primary search store implementation.
// The storeName identifies this backend (e.g., "v2" or "es8").
func WithStore(store semantic.SearchStore) Option {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}

// WithStoreName sets the name for the primary store (e.g., "v2" or "es8").
func WithStoreName(name string) Option {
	return func(s *Service) error {
		s.storeName = name
		return nil
	}
}

// WithAlternateStore registers an alternate backend for internal experiments.
// Runtime backend selection is not part of the public Semantic V1 contract.
func WithAlternateStore(name string, store semantic.SearchStore) Option {
	return func(s *Service) error {
		s.altStore = store
		s.altStoreName = name
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
