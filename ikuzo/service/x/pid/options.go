package pid

type Option func(*Service) error

func DataPath(path string) Option {
	return func(s *Service) error {
		s.dataPath = path
		return nil
	}
}

func FallbackTripleStore(urls []string) Option {
	return func(s *Service) error {
		s.fallbackTripleStores = urls
		return nil
	}
}

func FallbackNaan(mapping map[string]string) Option {
	return func(s *Service) error {
		s.fallbackNaan = mapping
		return nil
	}
}
