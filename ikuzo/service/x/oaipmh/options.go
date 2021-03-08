package oaipmh

type Option func(*Service) error

func SetDelay(delay int) Option {
	return func(s *Service) error {
		s.defaultDelay = delay
		return nil
	}
}
