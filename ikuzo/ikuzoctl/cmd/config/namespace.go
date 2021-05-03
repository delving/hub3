package config

import (
	"github.com/delving/hub3/ikuzo/service/x/namespace"
)

type NameSpace struct{}

func (ns NameSpace) AddOptions(cfg *Config) error {
	cfg.logger.Debug().Msg("setting up namespaces")
	svc, err := namespace.NewService(
		namespace.WithDefaults(),
	)
	if err != nil {
		return err
	}

	if err := svc.RegisterOtoService(cfg.getOto()); err != nil {
		return err
	}

	return nil
}
