package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/sitemap"
)

type Sitemap struct {
	Enabled bool `json:"enabled"`
}

func (s *Sitemap) NewService(cfg *Config) (*sitemap.Service, error) {
	client, err := cfg.ElasticSearch.NewCustomClient(cfg.log)
	if err != nil {
		return nil, err
	}

	store := client.NewSitemapStore()

	svc, err := sitemap.NewService(
		sitemap.SetStore(store),
	)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *Sitemap) AddOptions(cfg *Config) error {
	if !s.Enabled || !cfg.ElasticSearch.Enabled {
		return nil
	}

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
