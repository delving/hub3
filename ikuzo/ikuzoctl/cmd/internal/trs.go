package internal

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/revision"
)

type TimeRevisionStore struct {
	Enabled  bool   `json:"enabled"`
	DataPath string `json:"dataPath"`
}

func (trs *TimeRevisionStore) AddOptions(cfg *Config) error {
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
