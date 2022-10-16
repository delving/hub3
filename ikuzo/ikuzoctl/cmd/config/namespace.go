package config

import (
	"github.com/delving/hub3/ikuzo/service/x/namespace"
)

type Namespace struct{}

func (ns Namespace) AddOptions(cfg *Config) error {
	cfg.logger.Debug().Msg("setting up namespaces")
	svc, err := namespace.NewService(
		namespace.WithDefaults(),
	)
	if err != nil {
		return err
	}

	// TODO(kiivihal): register namespace
	_ = svc

	return nil
}
