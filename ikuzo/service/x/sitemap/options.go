package sitemap

type Option func(*Service) error

func SetStore(store Store) Option {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}
