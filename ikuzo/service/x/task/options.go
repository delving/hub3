package task

import "github.com/delving/hub3/ikuzo/storage/x/redis"

type Option func(*Service) error

func SetRedisConfig(cfg redis.Config) Option {
	return func(s *Service) error {
		s.redisCfg = &cfg
		return nil
	}
}

func SetNrWorkers(i int) Option {
	return func(s *Service) error {
		s.nrWorkers = i
		return nil
	}
}
