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
