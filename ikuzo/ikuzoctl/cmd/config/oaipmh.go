package config

import (
	"fmt"
	"time"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/kiivihal/goharvest/oai"
)

type OAIPMH struct {
	Enabled        bool     `json:"enabled"`
	AdminEmails    []string `json:"adminEmails"`
	RepositoryName string   `json:"repositoryName"`
	HarvestDelay   int      `json:"harvestDelay"`
	EadHarvestUrl  string   `json:"eadHarvestUrl"`
	MetsHarvestUrl string   `json:"metsHarvestUrl"`
	Service        *oaipmh.Service
}

func (o *OAIPMH) NewService(cfg *Config) (*oaipmh.Service, error) {
	svc, err := oaipmh.NewService(
		oaipmh.SetDelay(o.HarvestDelay),
	)
	if err != nil {
		return nil, err
	}

	if err := svc.StartHarvestSync(); err != nil {
		return nil, fmt.Errorf("unable to start OAIPMH harvester %w", err)
	}

	o.Service = svc

	return svc, nil
}

func (o *OAIPMH) AddOptions(cfg *Config) error {
	svc, err := o.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.SetOAIPMHServerOption(svc),
		ikuzo.SetShutdownHook("oaipmh-service", svc),
	)

	return nil
}

func (o *OAIPMH) AddEadHarvestTask(orgID string, fn oaipmh.HarvestCallback) {
	if o.EadHarvestUrl == "" {
		return
	}
	t := oaipmh.HarvestTask{
		OrgID:      orgID,
		Name:       "ead-harvester",
		CheckEvery: time.Hour * 1,
		Request: oai.Request{
			BaseURL:        o.EadHarvestUrl,
			MetadataPrefix: "oai_ead",
			Verb:           oaipmh.ListRecords,
		},
		CallbackFn: fn,
	}
	oaipmh.AddTask(t)(o.Service)
}

func (o *OAIPMH) AddMetsHarvestTask(orgID string, fn oaipmh.HarvestCallback) {
	if o.MetsHarvestUrl == "" {
		return
	}
	t := oaipmh.HarvestTask{
		OrgID:      orgID,
		Name:       "mets-harvester",
		CheckEvery: time.Minute * 5,
		Request: oai.Request{
			BaseURL:        o.MetsHarvestUrl,
			MetadataPrefix: "oai_mets",
			Verb:           oaipmh.ListIdentifiers,
		},
		CallbackFn: fn,
	}
	oaipmh.AddTask(t)(o.Service)
}
