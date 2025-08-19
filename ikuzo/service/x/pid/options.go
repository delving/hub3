package pid

type Option func(*Service) error

func DataPath(path string) Option {
	return func(s *Service) error {
		s.dataPath = path
		return nil
	}
}
