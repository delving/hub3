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
	sc, err := stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(n.URL))
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
