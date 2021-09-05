package imageproxy

type Option func(*Service) error

func SetCacheDir(path string) Option {
	return func(s *Service) error {
		s.cacheDir = path
		return nil
	}
}

func SetTimeout(duration int) Option {
	return func(s *Service) error {
		s.timeOut = duration
		return nil
	}
}

func SetProxyReferrer(referrer []string) Option {
	return func(s *Service) error {
		s.referrers = referrer
		return nil
	}
}

func SetBlackList(blacklist []string) Option {
	return func(s *Service) error {
		s.blacklist = blacklist
		return nil
	}
}

func SetProxyPrefix(prefix string) Option {
	return func(s *Service) error {
		s.proxyPrefix = prefix
		return nil
	}
}
