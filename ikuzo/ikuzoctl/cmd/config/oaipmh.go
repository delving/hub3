package config

import (
	"fmt"
	"time"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/kiivihal/goharvest/oai"
)

type OAIPMH struct {
	Service *oaipmh.Service
}

func (o *OAIPMH) NewService(cfg *Config) (*oaipmh.Service, error) {
	svc, err := oaipmh.NewService(
		oaipmh.SetDelay(config.Config.OAIPMH.HarvestDelay),
		oaipmh.AddTask(tasksFromConfig(config.Config)...),
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

func tasksFromConfig(ro config.RawConfig) []oaipmh.HarvestTask {
	tasks := make([]oaipmh.HarvestTask, 0)
	if ro.OAIPMH.EadHarvestUrl != "" {
		t := oaipmh.HarvestTask{
			OrgID:      ro.OrgID,
			Name:       "ead-harvester",
			CheckEvery: time.Hour * 1,
			Request: oai.Request{
				BaseURL:        ro.OAIPMH.EadHarvestUrl,
				MetadataPrefix: "oai_ead",
				Verb:           "ListIdentifiers",
			},
			// TODO(kiivihal): inject from EADHarvester
			// CallbackFn: ead.ProcessEadFromOai,
		}
		tasks = append(tasks, t)
	}

	if ro.OAIPMH.MetsHarvestUrl != "" {
		t := oaipmh.HarvestTask{
			OrgID:      ro.OrgID,
			Name:       "mets-harvester",
			CheckEvery: time.Minute * 5,
			Request: oai.Request{
				BaseURL:        ro.OAIPMH.MetsHarvestUrl,
				MetadataPrefix: "oai_mets",
				Verb:           "ListIdentifiers",
			},
			// TODO(kiivihal): Inject from METSHarvester
			// CallbackFn: ead.ProcessMetsFromOai,
		}
		tasks = append(tasks, t)
	}

	return tasks
}
