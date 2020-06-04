package config

import (
	"expvar"
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/ead"
)

type EAD struct {
	CacheDir string `json:"cacheDir"`
	Metrics  bool   `json:"metrics"`
}

func (e EAD) NewService(cfg *Config) (*ead.Service, error) {
	is, err := cfg.GetIndexService()
	if err != nil {
		return nil, err
	}

	svc, err := ead.NewService(
		ead.SetIndexService(is),
		ead.SetDataDir(e.CacheDir),
	)
	if err != nil {
		return nil, err
	}

	if err := svc.StartWorkers(); err != nil {
		return nil, fmt.Errorf("unable to start EAD service workers; %w", err)
	}

	if e.Metrics {
		expvar.Publish("hub3-ead-service", expvar.Func(func() interface{} { m := svc.Metrics(); return m }))
	}

	return svc, nil
}

func (e *EAD) AddOptions(cfg *Config) error {
	svc, err := e.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.SetEADService(svc),
		ikuzo.SetShutdownHook("ead-service", svc),
	)

	return nil
}
