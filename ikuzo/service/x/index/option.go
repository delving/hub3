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
		s.stan.setDefaults()
		return nil
	}
}

func WithDefaultMessageHandle() Option {
	return func(s *Service) error {
		s.MsgHandler = s.submitBulkMsg
		return nil
	}
}
