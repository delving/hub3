package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
)

type Organization struct {
	// domain is a list of all valid domains (including subdomains) for an domain.Organization
	// the domain ID will be injected in each request by the organization middleware.
	Store string
}

func (o *Organization) AddOptions(cfg *Config) error {
	svc, err := cfg.getOrganisationService(o.Store)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.SetOrganisationService(svc),
	)

	return nil
}

func (cfg *Config) getOrganisationService(storeType string) (*organization.Service, error) {
	if cfg.orgs != nil {
		return cfg.orgs, nil
	}

	// TODO(kiivihal): create func to create the different stores
	store := memory.NewOrganizationStore()

	svc, err := organization.NewService(store)
	if err != nil {
		return nil, fmt.Errorf("unable to configure organization service; %w", err)
	}

	if err := svc.AddOrgs(cfg.Org); err != nil {
		return nil, err
	}

	return svc, nil
}
