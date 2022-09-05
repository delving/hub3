package config

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/nde"
)

// NDE is a place-holder struct for configurations
type NDE struct{}

type NDECfg struct {
	URLPrefix        string   `json:"urlPrefix"`
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

func (n *NDE) createConfig(cfg *Config) ([]*nde.RegisterConfig, error) {
	var cfgs []*nde.RegisterConfig

	for name, ndeCfg := range cfg.NDE {
		if ndeCfg.Name == "" {
			return nil, fmt.Errorf("NDE config must be set for register")
		}

		config := &nde.RegisterConfig{
			URLPrefix:        name,
			DataPath:         cfg.EAD.CacheDir,
			DatasetFmt:       ndeCfg.DatasetFmt,
			RDFBaseURL:       cfg.RDF.BaseURL,
			Description:      ndeCfg.Description,
			Name:             ndeCfg.Name,
			DefaultLicense:   ndeCfg.DefaultLicense,
			DefaultLanguages: ndeCfg.DefaultLanguages,
			Distributions:    ndeCfg.Distribution,
			Publisher: struct {
				Name    string
				AltName string
				URL     string
			}{
				Name:    ndeCfg.Publisher.Name,
				AltName: ndeCfg.Publisher.AltName,
				URL:     ndeCfg.Publisher.URL,
			},
		}

		cfgs = append(cfgs, config)
	}

	return cfgs, nil
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
