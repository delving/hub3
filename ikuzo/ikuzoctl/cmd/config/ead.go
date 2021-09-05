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
	"expvar"
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/service/x/ead"
)

type EAD struct {
	CacheDir       string `json:"cacheDir"`
	Metrics        bool   `json:"metrics"`
	Workers        int    `json:"workers"`
	ProcessDigital bool   `json:"processDigital"`
}

func (e EAD) NewService(cfg *Config) (*ead.Service, error) {
	is, err := cfg.GetIndexService()
	if err != nil {
		return nil, err
	}

	trs, err := cfg.GetRevisionService()
	if err != nil {
		return nil, err
	}

	svc, err := ead.NewService(
		ead.SetIndexService(is),
		// TODO(kiivihal): can be removed later for TRS
		ead.SetDataDir(e.CacheDir),
		ead.SetWorkers(e.Workers),
		ead.SetProcessDigital(e.ProcessDigital),
		ead.SetRevisionService(trs),
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
		ikuzo.RegisterService(svc),
	)

	return nil
}
