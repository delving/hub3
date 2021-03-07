// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/delving/hub3/ikuzo/service/x/revision"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/spf13/viper"
)

type Option interface {
	AddOptions(cfg *Config) error
}

type Config struct {
	// default orgID when none is given
	OrgID             string `json:"orgID"`
	DataNodeURL       string `json:"dataNodeURL"`
	ElasticSearch     `json:"elasticSearch"`
	HTTP              `json:"http"`
	TimeRevisionStore `json:"timeRevisionStore"`
	Logging           `json:"logging"`
	Nats              `json:"nats"`
	EAD               `json:"ead"`
	DB                `json:"db"`
	ImageProxy        `json:"imageProxy"`
	NameSpace         `json:"nameSpace"`
	PostHooks         []PostHook `json:"posthooks"`
	options           []ikuzo.Option
	logger            logger.CustomLogger
	oto               *otohttp.Server
	is                *index.Service
	trs               *revision.Service
	orgs              *organization.Service
	Organization      `json:"organization"`
	Org               map[string]domain.OrganizationConfig
	OAIPMH            `json:"oaipmh"`
}

func (cfg *Config) IsDataNode() bool {
	return cfg.DataNodeURL == ""
}

func (cfg *Config) Options(cfgOptions ...Option) ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())
	cfg.logger.Info().Str("configPath", viper.ConfigFileUsed()).Msg("starting with config file (ikuzo)")

	if len(cfgOptions) == 0 {
		cfgOptions = []Option{
			&cfg.ElasticSearch, // elastic first because others could depend on the client
			&cfg.Organization,
			&cfg.HTTP,
			&cfg.TimeRevisionStore,
			&cfg.EAD,
			&cfg.ImageProxy,
			&cfg.Logging,
			&cfg.OAIPMH,
			&cfg.NameSpace,
		}
	}

	for _, option := range cfgOptions {
		if err := option.AddOptions(cfg); err != nil {
			return cfg.options, err
		}
	}

	if err := cfg.defaultOptions(); err != nil {
		return nil, err
	}

	cfg.options = append(cfg.options, ikuzo.SetLogger(&cfg.logger))

	if cfg.oto != nil {
		cfg.options = append(cfg.options, ikuzo.RegisterOtoServer(cfg.oto))
	}

	cfg.logger.Info().Str("configPath", viper.ConfigFileUsed()).Msg("starting with config file")

	return cfg.options, nil
}

func SetViperDefaults() {
	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("TimeRevisionStore.dataPath", "/tmp/trs")
}

func (cfg *Config) GetIndexService() (*index.Service, error) {
	if cfg.is != nil {
		return cfg.is, nil
	}

	var (
		ncfg *index.NatsConfig
		err  error
	)

	if cfg.Nats.Enabled {
		ncfg, err = cfg.Nats.GetConfig()
		if err != nil {
			return nil, err
		}
	}

	is, err := cfg.ElasticSearch.IndexService(cfg, ncfg)
	if err != nil {
		return nil, err
	}

	cfg.is = is

	return cfg.is, nil
}

func (cfg *Config) GetRevisionService() (*revision.Service, error) {
	if !cfg.TimeRevisionStore.Enabled {
		return nil, fmt.Errorf("revision.Service is not enabled in the configuration")
	}

	if cfg.trs != nil {
		return cfg.trs, nil
	}

	trs, err := revision.NewService(cfg.TimeRevisionStore.DataPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create revision.Service; %w", err)
	}

	cfg.trs = trs

	return cfg.trs, nil
}

func (cfg *Config) defaultOptions() error {
	return nil
}

func (cfg *Config) getOto() *otohttp.Server {
	if cfg.oto == nil {
		cfg.oto = otohttp.NewServer()
	}

	return cfg.oto
}
