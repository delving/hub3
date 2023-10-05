package nde

func SetConfig(cfgs []*RegisterConfig) Option {
	return func(s *Service) error {
		s.cfgs = cfgs
		return nil
	}
}
