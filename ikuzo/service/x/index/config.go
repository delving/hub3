package index

import (
	"github.com/nats-io/stan.go"
)

const (
	clientID     = "hub3-pub"
	clusterID    = "hub3-nats"
	durableName  = "hub3-worker"
	durableQueue = "hub3-queue"
	subjectID    = "hub3-bulk-index"
)

type NatsConfig struct {
	Conn         stan.Conn
	SubjectID    string
	ClusterID    string
	ClientID     string
	DurableName  string
	DurableQueue string
}

func (c *NatsConfig) setDefaults() {
	if c.ClusterID == "" {
		c.ClusterID = clusterID
	}

	if c.ClientID == "" {
		c.ClientID = clientID
	}

	if c.DurableName == "" {
		c.DurableName = durableName
	}

	if c.DurableQueue == "" {
		c.DurableQueue = durableQueue
	}

	if c.SubjectID == "" {
		c.SubjectID = subjectID
	}
}
