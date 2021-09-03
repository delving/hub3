package config

import (
	"context"
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/storage/memory"
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

	for id, orgCfg := range cfg.Org {
		if err != nil {
			return nil, fmt.Errorf("unable to configure organization.Service: %w", err)
		}

		orgCfg.SetOrgID(id)

		if orgCfg.CustomID != "" {
			id = orgCfg.CustomID
		}

		orgID, err := domain.NewOrganizationID(id)
		if err != nil {
			return nil, fmt.Errorf("unable to create domain.OrganizationID %s; %w", id, err)
		}

		org := domain.Organization{
			Config: orgCfg,
			ID:     orgID,
		}

		if err := svc.Put(context.TODO(), &org); err != nil {
			return nil, fmt.Errorf("unable to store Organization; %w", err)
		}
	}

	return svc, nil
}
