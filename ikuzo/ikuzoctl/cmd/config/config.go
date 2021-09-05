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
	// default orgID when none is given
	OrgID         string `json:"orgID"`
	ElasticSearch `json:"elasticSearch"`
	HTTP          `json:"http"`
	Logging       `json:"logging"`
	Nats          `json:"nats"`
	EAD           `json:"ead"`
	ImageProxy    `json:"imageProxy"`
	NameSpace     `json:"nameSpace"`
	PostHooks     []PostHook `json:"posthooks"`
	options       []ikuzo.Option
	logger        logger.CustomLogger
	is            *index.Service
	orgs          *organization.Service
	Organization  `json:"organization"`
	Org           map[string]domain.OrganizationConfig `json:"org"`
	OAIPMH        `json:"oaipmh"`
	NDE           `json:"nde"`
	RDF           `json:"rdf"`
	Sitemap       `json:"sitemap"`
	oto           *otohttp.Server
}

func (cfg *Config) Options(cfgOptions ...Option) ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())
	cfg.logger.Info().Str("configPath", viper.ConfigFileUsed()).Msg("starting with config file (ikuzo)")

	if len(cfgOptions) == 0 {
		cfgOptions = []Option{
			&cfg.ElasticSearch, // elastic first because others could depend on the client
			&cfg.Organization,
			&cfg.HTTP,
			&cfg.EAD,
			&cfg.ImageProxy,
			&cfg.OAIPMH,
			&cfg.NameSpace,
			&cfg.NDE,
			&cfg.Sitemap,
			&cfg.Logging,
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

func (cfg *Config) defaultOptions() error {
	return nil
}

func (cfg *Config) getOto() *otohttp.Server {
	if cfg.oto == nil {
		cfg.oto = otohttp.NewServer()
	}

	return cfg.oto
}
