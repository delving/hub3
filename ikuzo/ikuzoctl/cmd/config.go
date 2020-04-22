package cmd

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/revision"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/spf13/viper"
)

type configOption interface {
	AddOptions(cfg *config) error
}

type config struct {
	// default orgID when none is given
	OrgID             string `json:"orgID"`
	ElasticSearch     `json:"elasticSearch"`
	HTTP              `json:"http"`
	TimeRevisionStore `json:"timeRevisionStore"`
	options           []ikuzo.Option `json:"options"`
}

func (cfg *config) Options() ([]ikuzo.Option, error) {
	cfgOptions := []configOption{
		&cfg.HTTP,
		&cfg.ElasticSearch,
		&cfg.TimeRevisionStore,
	}

	for _, option := range cfgOptions {
		if err := option.AddOptions(cfg); err != nil {
			return cfg.options, err
		}
	}

	return cfg.options, nil
}

type HTTP struct {
	Port int `json:"port" mapstructure:"port"`
}

func (http *HTTP) AddOptions(cfg *config) error {
	cfg.options = append(cfg.options, ikuzo.SetPort(http.Port))
	return nil
}

type ElasticSearch struct {
	// enable elasticsearch client
	Enabled bool
	// urls to connect to elasticsearch cluster
	Urls []string
	// enable elasticsearch caching proxy
	Proxy bool
}

func (e *ElasticSearch) AddOptions(cfg *config) error {
	if e.Enabled && len(e.Urls) != 0 {
		if e.Proxy {
			esProxy, err := elasticsearch.NewProxy(e.Urls...)
			if err != nil {
				return fmt.Errorf("unable to create ES proxy: %w", err)
			}

			cfg.options = append(cfg.options, ikuzo.SetElasticSearchProxy(esProxy))
		}
	}

	return nil
}

type TimeRevisionStore struct {
	Enabled  bool   `json:"enabled"`
	DataPath string `json:"dataPath"`
}

func (trs *TimeRevisionStore) AddOptions(cfg *config) error {
	if trs.Enabled && trs.DataPath != "" {
		svc, err := revision.NewService(trs.DataPath)
		if err != nil {
			return fmt.Errorf("unable to start revision store from config: %w", err)
		}

		cfg.options = append(
			cfg.options,
			ikuzo.SetRevisionService(svc),
		)
	}

	return nil
}

func setDefaults() {
	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("TimeRevisionStore.dataPath", "/tmp/trs")
}
