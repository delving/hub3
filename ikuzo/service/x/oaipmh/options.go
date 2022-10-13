package oaipmh

type Option func(*Service) error

func SetStore(store Store) Option {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}

// SetRequireSetSpec determines if a set must be provided when harvesting list.
// default is true
func SetRequireSetSpec(allow bool) Option {
	return func(s *Service) error {
		s.requireSetSpecForList = allow
		return nil
	}
}

// SetTagFilters add filters to limit which records and sets are
// available for OAI-PMH harvesting
func SetTagFilters(filters []string) Option {
	return func(s *Service) error {
		s.filters = filters
		return nil
	}
}
