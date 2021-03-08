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

	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
)

// Nats are configuration options to access NATS streaming server
type Nats struct {
	Enabled      bool   `json:"enabled"`
	ClusterID    string `json:"clusterID"`
	ClientID     string `json:"clientID"`
	DurableName  string `json:"durableName"`
	DurableQueue string `json:"durableQueue"`
	URL          string `json:"url"`
	cfg          *index.NatsConfig
}

func (n *Nats) AddOptions(cfg *Config) error {
	return nil
}

func (n *Nats) newClient(cfg *index.NatsConfig) (stan.Conn, error) {
	if n.URL == "" {
		n.URL = stan.DefaultNatsURL
	}

	// Connect to Streaming server
	sc, err := stan.Connect(
		cfg.ClusterID,
		cfg.ClientID,
		stan.NatsURL(n.URL),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			// TODO(kiivihal): implement reconnect functionality
			// https://github.com/nats-io/stan.go/issues/273
			log.Error().Err(reason).Msg("stan streaming server: connection lost")
		}),
	)
	if err != nil {
		log.Error().Msgf("nats configuration: %+v", cfg)
		log.Error().Msgf("nats struct: %+v", n)

		return sc, fmt.Errorf("can't connect: %w.\nMake sure a NATS Streaming Server is running at: %s", err, n.URL)
	}

	return sc, nil
}

func (n *Nats) GetConfig() (*index.NatsConfig, error) {
	if n.cfg != nil {
		return n.cfg, nil
	}

	cfg := &index.NatsConfig{
		ClusterID:    n.ClusterID,
		ClientID:     n.ClientID,
		DurableName:  n.DurableName,
		DurableQueue: n.DurableQueue,
	}

	conn, err := n.newClient(cfg)
	if err != nil {
		return nil, err
	}

	cfg.Conn = conn

	n.cfg = cfg

	return n.cfg, nil
}
