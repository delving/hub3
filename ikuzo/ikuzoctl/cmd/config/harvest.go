package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh/harvest"
)

type Harvest struct {
	Enabled        bool     `json:"enabled"`
	AdminEmails    []string `json:"adminEmails"`
	RepositoryName string   `json:"repositoryName"`
	HarvestDelay   int      `json:"harvestDelay"`
	EadHarvestURL  string   `json:"eadHarvestUrl"`
	MetsHarvestURL string   `json:"metsHarvestUrl"`
	HarvestPath    string   `json:"harvestPath"`
	RequireSetSpec bool     `json:"requireSetSpec"`
	TagFilters     []string `json:"tagFilters"`
	service        *harvest.Service
}

func (h *Harvest) NewService(cfg *Config) (*harvest.Service, error) {
	if h.service != nil {
		return h.service, nil
	}

	svc, err := harvest.NewService(
		harvest.SetDelay(h.HarvestDelay),
	)
	if err != nil {
		return nil, err
	}

	if err := svc.StartHarvestSync(); err != nil {
		return nil, fmt.Errorf("unable to start OAIPMH harvester; %w", err)
	}

	h.service = svc

	return svc, nil
}

func (h *Harvest) AddOptions(cfg *Config) error {
	svc, err := h.NewService(cfg)
	if err != nil {
		return err
	}

	// TODO(kiivihal): enable again after testing
	// serverStore, err := oaipmh.NewFsRepoStore(o.HarvestPath)
	// if err != nil {
	// return fmt.Errorf("unable to create OAI-PMH server store; %w", err)
	// }

	// server, err := oaipmh.NewServer(oaipmh.SetServerStore(serverStore))
	// if err != nil {
	// return fmt.Errorf("unable to create OAI-PMH server; %w", err)
	// }

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}
