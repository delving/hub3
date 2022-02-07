package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/nde"
)

type NDE struct {
	Enabled          bool     `json:"enabled"`
	Description      string   `json:"description"`
	Name             string   `json:"name"`
	DefaultLanguages []string `json:"defaultLanguages"`
	DefaultLicense   string   `json:"defaultLicense"`
	Publisher        struct {
		Name    string `json:"name"`
		AltName string `json:"altName"`
		URL     string `json:"url"`
	} `json:"publisher"`
	DatasetFmt   string                `json:"datasetFmt"`
	Distribution []nde.DistributionCfg `json:"distribution"`
}

func (n *NDE) createConfig(cfg *Config) (*nde.RegisterConfig, error) {
	if n.Name == "" {
		return nil, fmt.Errorf("NDE config must be set for register")
	}

	config := &nde.RegisterConfig{
		DataPath:         cfg.EAD.CacheDir,
		DatasetFmt:       n.DatasetFmt,
		RDFBaseURL:       cfg.RDF.BaseURL,
		Description:      n.Description,
		Name:             n.Name,
		DefaultLicense:   n.DefaultLicense,
		DefaultLanguages: n.DefaultLanguages,
		Distributions:    n.Distribution,
		Publisher: struct {
			Name    string
			AltName string
			URL     string
		}{
			Name:    n.Publisher.Name,
			AltName: n.Publisher.AltName,
			URL:     n.Publisher.URL,
		},
	}

	return config, nil
}

func (n *NDE) NewService(cfg *Config) (*nde.Service, error) {
	config, err := n.createConfig(cfg)
	if err != nil {
		return nil, err
	}

	svc, err := nde.NewService(
		nde.SetConfig(config),
	)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (n *NDE) AddOptions(cfg *Config) error {
	svc, err := n.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.SetRouters(svc.Routes),
	)

	return nil
}
