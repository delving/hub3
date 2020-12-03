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
	"fmt"

	"github.com/delving/hub3/ikuzo"
)

type TimeRevisionStore struct {
	Enabled  bool   `json:"enabled"`
	DataPath string `json:"dataPath"`
}

func (trs *TimeRevisionStore) AddOptions(cfg *Config) error {
	if trs.Enabled && trs.DataPath != "" {
		svc, err := cfg.GetRevisionService()
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
