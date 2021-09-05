package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
)

type OAIPMH struct {
	Enabled        bool     `json:"enabled"`
	AdminEmails    []string `json:"adminEmails"`
	RepositoryName string   `json:"repositoryName"`
	HarvestDelay   int      `json:"harvestDelay"`
	EadHarvestURL  string   `json:"eadHarvestUrl"`
	MetsHarvestURL string   `json:"metsHarvestUrl"`
	HarvestPath    string   `json:"harvestPath"`
	service        *oaipmh.Service
}

func (o *OAIPMH) NewService(cfg *Config) (*oaipmh.Service, error) {
	if o.service != nil {
		return o.service, nil
	}

	svc, err := oaipmh.NewService(
		oaipmh.SetDelay(o.HarvestDelay),
	)
	if err != nil {
		return nil, err
	}

	if err := svc.StartHarvestSync(); err != nil {
		return nil, fmt.Errorf("unable to start OAIPMH harvester; %w", err)
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
