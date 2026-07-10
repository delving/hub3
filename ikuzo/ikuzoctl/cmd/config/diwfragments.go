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
	"github.com/delving/hub3/ikuzo/service/x/diwfragments"
)

// DiwFragments toggles the /api/ui/v1 pre-rendered fragment service.
//
// It follows the same enable/disable pattern as the other optional
// ikuzoctl services (see Sitemap in this package): the service is only
// wired up when both DiwFragments.Enabled and ElasticSearch.Enabled are
// true, since the fragment store is backed by Elasticsearch.
type DiwFragments struct {
	Enabled bool `json:"enabled"`
}

// NewService constructs the diwfragments.Service, wiring it to an
// Elasticsearch-backed fragment store obtained from the shared
// ElasticSearch client configuration.
//
// It returns an error when the Elasticsearch client cannot be created or
// when the underlying diwfragments.NewService construction fails.
func (d *DiwFragments) NewService(cfg *Config) (*diwfragments.Service, error) {
	client, err := cfg.ElasticSearch.NewCustomClient(cfg.log)
	if err != nil {
		return nil, err
	}

	store := client.NewDiwFragmentStore()

	svc, err := diwfragments.NewService(
		diwfragments.SetStore(store),
	)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// AddOptions registers the diwfragments service with the ikuzo application
// when enabled. It is a no-op when DiwFragments is disabled or when
// Elasticsearch is not enabled, since the fragment store requires an
// Elasticsearch client to function.
func (d *DiwFragments) AddOptions(cfg *Config) error {
	if !d.Enabled || !cfg.ElasticSearch.Enabled {
		return nil
	}

	svc, err := d.NewService(cfg)
	if err != nil {
		return err
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(svc),
	)

	return nil
}
