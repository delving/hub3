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
	"github.com/go-chi/chi"
)

type Logging struct {
	DevMode        bool   `json:"devmode"`
	Level          string `json:"level"`
	WithCaller     bool   `json:"withCaller"`
	ConsoleLogger  bool   `json:"consoleLogger"`
	ErrorFieldName string `json:"errorFieldName"`
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
		ErrorFieldName:      l.ErrorFieldName,
	}
}
