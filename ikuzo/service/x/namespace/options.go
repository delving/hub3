package namespace

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
