package config

import (
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/go-chi/chi"
)

type Logging struct {
	DevMode       bool   `json:"devmode"`
	Level         string `json:"level"`
	WithCaller    bool   `json:"withCaller"`
	ConsoleLogger bool   `json:"consoleLogger"`
}

func (l *Logging) AddOptions(cfg *Config) error {
	if l.DevMode {
		cfg.options = append(
			cfg.options,
			ikuzo.SetRouters(func(r chi.Router) {
				r.Delete("/introspect/reset", cfg.ElasticSearch.ResetAll)
			}),
		)
	}

	return nil
}

func (l *Logging) GetConfig() logger.Config {
	return logger.Config{
		LogLevel:            logger.ParseLogLevel(l.Level),
		WithCaller:          l.WithCaller,
		EnableConsoleLogger: l.ConsoleLogger,
	}
}
