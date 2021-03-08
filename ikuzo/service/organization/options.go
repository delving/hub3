package organization

type Option func(*Service) error

func SetDomainRoutes() Option {
	return func(s *Service) error {
		// s.index = is
		return nil
	}
}
