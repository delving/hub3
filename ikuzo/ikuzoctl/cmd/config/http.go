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

import "github.com/delving/hub3/ikuzo"

type HTTP struct {
	Port        int    `json:"port" mapstructure:"port"`
	MetricsPort int    `json:"metricsPort"`
	CertFile    string `json:"certFile"`
	KeyFile     string `json:"keyFile"`
}

func (http *HTTP) AddOptions(cfg *Config) error {
	cfg.options = append(
		cfg.options,
		ikuzo.SetPort(http.Port),
		ikuzo.SetTLS(http.CertFile, http.KeyFile),
	)

	if http.MetricsPort != 0 {
		cfg.options = append(cfg.options, ikuzo.SetMetricsPort(http.MetricsPort))
	}

	return nil
}
