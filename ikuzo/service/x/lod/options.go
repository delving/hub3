package lod

type Option func(*Service) error

func SetResolver(name string, r Resolver, isDefault bool) Option {
	return func(s *Service) error {
		s.stores[name] = r

		if isDefault {
			s.defaultStore = name
		}

		return nil
	}
}
