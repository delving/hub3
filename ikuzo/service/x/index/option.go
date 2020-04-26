package index

import "github.com/elastic/go-elasticsearch/v8/esutil"

type Option func(*Service) error

func SetBulkIndexer(bi esutil.BulkIndexer) Option {
	return func(s *Service) error {
		s.bi = bi
		return nil
	}
}

func SetNatsConfiguration(ncfg *NatsConfig) Option {
	return func(s *Service) error {
		s.stan = ncfg
		return nil
	}
}
