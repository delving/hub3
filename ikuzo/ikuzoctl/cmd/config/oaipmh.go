package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
)

type OAIPMH struct {
	service *oaipmh.Service
}

func (o *OAIPMH) NewService(cfg *Config) (*oaipmh.Service, error) {
	if o.service != nil {
		return o.service, nil
	}

	// TODO(kiivihal): make stores switchable
	store, err := cfg.client.NewOAIPMHStore()
	if err != nil {
		return nil, err
	}

	svc, err := oaipmh.NewService(
		oaipmh.SetStore(store),
		oaipmh.SetRequireSetSpec(cfg.Harvest.RequireSetSpec),
		oaipmh.SetTagFilters(cfg.Harvest.TagFilters),
	)
	if err != nil {
		return nil, err
	}

	o.service = svc

	return svc, nil
}

func (o *OAIPMH) AddOptions(cfg *Config) error {
	svc, err := o.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}
