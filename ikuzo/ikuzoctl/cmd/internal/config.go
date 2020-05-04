package internal

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/spf13/viper"
)

type configOption interface {
	AddOptions(cfg *Config) error
}

type Config struct {
	// default orgID when none is given
	OrgID             string `json:"orgID"`
	ElasticSearch     `json:"elasticSearch"`
	HTTP              `json:"http"`
	TimeRevisionStore `json:"timeRevisionStore"`
	Logging           `json:"logging"`
	Nats              `json:"nats"`
	EAD               `json:"ead"`
	options           []ikuzo.Option
	logger            logger.CustomLogger
}

func (cfg *Config) Options() ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())

	cfgOptions := []configOption{
		&cfg.ElasticSearch, // elastic first because others could depend on the client
		&cfg.HTTP,
		&cfg.TimeRevisionStore,
		&cfg.EAD,
	}

	for _, option := range cfgOptions {
		if err := option.AddOptions(cfg); err != nil {
			return cfg.options, err
		}
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

func (cfg *Config) getIndexService() (*index.Service, error) {
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
