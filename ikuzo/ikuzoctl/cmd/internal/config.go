package internal

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
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
	options           []ikuzo.Option
	logger            logger.CustomLogger
}

func (cfg *Config) Options() ([]ikuzo.Option, error) {
	cfg.logger = logger.NewLogger(cfg.Logging.GetConfig())

	cfgOptions := []configOption{
		&cfg.ElasticSearch, // elastic first because others could depend on the client
		&cfg.HTTP,
		&cfg.TimeRevisionStore,
	}

	for _, option := range cfgOptions {
		if err := option.AddOptions(cfg); err != nil {
			return cfg.options, err
		}
	}

	cfg.options = append(cfg.options, ikuzo.SetLogger(&cfg.logger))

	return cfg.options, nil
}

func SetViperDefaults() {
	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("TimeRevisionStore.dataPath", "/tmp/trs")
}
