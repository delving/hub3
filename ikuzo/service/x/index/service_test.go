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
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/organization/organizationtests"
	"github.com/matryer/is"
	"github.com/nats-io/stan.go"
)

func (s *indexSuite) testConfig() (*NatsConfig, error) {
	var (
		err error
		cfg = &NatsConfig{}
	)

	cfg.setDefaults()

	// Connect to Streaming server
	natsURL := fmt.Sprintf("nats://%s:%s", s.ip, s.port.Port())

	cfg.Conn, err = stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(natsURL))
	if err != nil {
		return cfg, fmt.Errorf("can't connect: %w.\nMake sure a NATS Streaming Server is running at: %s", err, natsURL)
	}

	return cfg, nil
}

// nolint:gocritic
func (s *indexSuite) TestProducer_Publish() {
	is := is.New(s.T())

	cfg, err := s.testConfig()
	is.NoErr(err)

	svc, err := NewService(
		SetNatsConfiguration(cfg),
		SetOrganisationService(organizationtests.NewTestOrganizationService()),
	)
	is.NoErr(err)

	messages := []*domainpb.IndexMessage{}

	msgCount := 100

	for i := 0; i < msgCount; i++ {
		msg := &domainpb.IndexMessage{
			OrganisationID: "demo",
			DatasetID:      "spec",
			RecordID:       strconv.Itoa(i),
			Revision: &domainpb.Revision{
				SHA:  "",
				Path: "",
			},
			Source: []byte(fmt.Sprintf("source doc-%d", i)),
		}

		messages = append(messages, msg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)

	err = svc.Start(ctx, 4)
	is.NoErr(err)

	err = svc.Publish(context.Background(), messages...)
	is.NoErr(err)

	published := atomic.LoadUint64(&svc.m.Nats.Published)
	is.Equal(published, uint64(msgCount))

	var consumed uint64

L:
	for {
		select {
		case <-ctx.Done():
			consumed = atomic.LoadUint64(&svc.m.Nats.Consumed)
			s.T().Logf("context expired; messages consumed: %d", svc.m.Nats.Consumed)
			ticker.Stop()
			break L
		case <-ticker.C:
			consumed = atomic.LoadUint64(&svc.m.Nats.Consumed)
			s.T().Logf("ticker; messages consumed: %d", svc.m.Nats.Consumed)

			if int(consumed) >= msgCount {
				break L
			}
		}
	}

	is.Equal(uint64(0), atomic.LoadUint64(&svc.m.Nats.Failed))
	is.Equal(msgCount, int(consumed))

	err = svc.Shutdown(ctx)
	is.NoErr(err)
}
