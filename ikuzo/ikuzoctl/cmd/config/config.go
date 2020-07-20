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
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/spf13/viper"
)

type ConfigOption interface {
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
	PostHooks         []PostHook `json:"posthooks"`
	options           []ikuzo.Option
	logger            logger.CustomLogger
}

func (cfg *Config) IsDataNode() bool {
	return cfg.DataNodeURL == ""
}

func (cfg *Config) Options(cfgOptions ...ConfigOption) ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())

	if len(cfgOptions) == 0 {
		cfgOptions = []ConfigOption{
			&cfg.ElasticSearch, // elastic first because others could depend on the client
			&cfg.HTTP,
			&cfg.TimeRevisionStore,
			&cfg.EAD,
			&cfg.ImageProxy,
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

	cfg.logger.Info().Str("configPath", viper.ConfigFileUsed()).Msg("starting with config file")

	return cfg.options, nil
}

func SetViperDefaults() {
	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("TimeRevisionStore.dataPath", "/tmp/trs")
}

func (cfg *Config) GetIndexService() (*index.Service, error) {
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

	is, err := cfg.ElasticSearch.IndexService(&cfg.logger, ncfg)
	if err != nil {
		return nil, err
	}

	return is, nil
}

func (cfg *Config) defaultOptions() error {
	// db, err := cfg.DB.getDB()
	// if err != nil {
	// return err
	// }

	// // Organization
	// orgStore, err := gorm.NewOrganizationStore(db)
	// if err != nil {
	// return err
	// }

	// org, err := organization.NewService(orgStore)
	// if err != nil {
	// return err
	// }

	// cfg.options = append(cfg.options, ikuzo.SetOrganisationService(org))
	// cfg.logger.Debug().Msg("is this called")

	return nil
}
