package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/pid"
)

type PID struct {
	DataPath            string            `json:"dataPath"`
	FallbackTripleStore []string          `json:"fallbackTripleStore"`
	FallbackNaan        map[string]string `json:"fallbackNaan"`
}

func (p *PID) AddOptions(cfg *Config) error {
	svc, err := pid.NewService(
		pid.DataPath(p.DataPath),
		pid.FallbackTripleStore(p.FallbackTripleStore),
		pid.FallbackNaan(p.FallbackNaan),
	)
	if err != nil {
		return err
	}
	cfg.options = append(cfg.options, ikuzo.RegisterService(svc))

	return nil
}
