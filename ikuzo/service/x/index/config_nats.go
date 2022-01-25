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
