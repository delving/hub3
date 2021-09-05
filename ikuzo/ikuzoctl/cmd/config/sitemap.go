package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/ikuzo/service/x/sitemap"
)

type Sitemap struct{}

func (s *Sitemap) NewService(cfg *Config) (*sitemap.Service, error) {
	// TODO(kiivihal): refactor to use new elasticsearch driver client
	store := elasticsearch.NewSitemapStore()

	svc, err := sitemap.NewService(
		sitemap.SetStore(store),
	)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *Sitemap) AddOptions(cfg *Config) error {
	svc, err := s.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}
