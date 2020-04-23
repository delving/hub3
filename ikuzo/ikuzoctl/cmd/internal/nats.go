package internal

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/nats-io/stan.go"
)

// Nats are configuration options to access NATS streaming server
type Nats struct {
	Enabled      bool   `json:"enabled"`
	ClusterID    string `json:"clusterID"`
	ClientID     string `json:"clientID"`
	DurableName  string `json:"durableName"`
	DurableQueue string `json:"durableQueue"`
	URL          string `json:"url"`
}

func (n *Nats) AddOptions(cfg *Config) error {
	return nil
}

func (n *Nats) newClient() (stan.Conn, error) {
	if n.URL == "" {
		n.URL = stan.DefaultNatsURL
	}

	// Connect to Streaming server
	sc, err := stan.Connect(n.ClusterID, n.ClientID, stan.NatsURL(n.URL))
	if err != nil {
		return sc, fmt.Errorf("can't connect: %w.\nMake sure a NATS Streaming Server is running at: %s", err, n.URL)
	}

	return sc, nil
}

func (n *Nats) GetConfig() *index.Config {
	cfg := &index.Config{}

	return cfg
}
