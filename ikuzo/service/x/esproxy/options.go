package esproxy

import "github.com/delving/hub3/ikuzo/driver/elasticsearch"

type Option func(*Service) error

func SetElasticClient(es *elasticsearch.Client) Option {
	return func(s *Service) error {
		s.es = es
		return nil
	}
}
