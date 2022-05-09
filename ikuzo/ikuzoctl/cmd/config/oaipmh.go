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

	svc, err := oaipmh.NewService()
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
