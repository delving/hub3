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
	"github.com/spf13/viper"
)

type Option interface {
	AddOptions(cfg *Config) error
}

type Config struct {
	ElasticSearch `json:"elasticSearch"`
	HTTP          `json:"http"`
	Logging       `json:"logging"`
	Nats          `json:"nats"`
	EAD           `json:"ead"`
	DB            `json:"db"`
	ImageProxy    `json:"imageProxy"`
	Namespace     `json:"namespace"`
	PostHooks     []PostHook `json:"posthooks"`
	options       []ikuzo.Option
	logger        logger.CustomLogger
	is            *index.Service
	orgs          *organization.Service
	Organization  `json:"organization"`
	Org           map[string]domain.OrganizationConfig `json:"org"`
	Harvest       `json:"harvest"`
	OAIPMH        `json:"oaipmh"`
	NDE           `json:"nde"`
	RDF           `json:"rdf"`
	Sitemap       `json:"sitemap"`
	ns            *namespace.Service
}

func (cfg *Config) Options(cfgOptions ...Option) ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())
	cfg.logger.Info().Str("configPath", viper.ConfigFileUsed()).Msg("starting with config file (ikuzo)")

	if len(cfgOptions) == 0 {
		cfgOptions = []Option{
			&cfg.DB,
			&cfg.ElasticSearch, // elastic first because others could depend on the client
			&cfg.Organization,
			&cfg.HTTP,
			&cfg.EAD,
			&cfg.ImageProxy,
			&cfg.Harvest,
			&cfg.Namespace,
			&cfg.NDE,
			&cfg.Sitemap,
			&cfg.Logging,
			&cfg.OAIPMH,
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

	if !cfg.ElasticSearch.Enabled {
		cfg.logger.Warn().Msg("elasticsearch is disabled, so index service is disabled as well")
		return nil, fmt.Errorf("elasticsearch is not enabled")
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

func (cfg *Config) defaultOptions() error {
	return nil
}
