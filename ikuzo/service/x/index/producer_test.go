package index

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/matryer/is"
	"github.com/nats-io/stan.go"
)

func testConfig() (*Config, error) {
	var (
		err error
		cfg = &Config{}
	)

	cfg.setDefaults()

	// Connect to Streaming server
	cfg.Queue, err = stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(stan.DefaultNatsURL))
	if err != nil {
		return cfg, fmt.Errorf("can't connect: %w.\nMake sure a NATS Streaming Server is running at: %s", err, stan.DefaultNatsURL)
	}

	return cfg, nil
}

// nolint:gocritic
func TestProducer_Publish(t *testing.T) {
	is := is.New(t)

	cfg, err := testConfig()
	is.NoErr(err)

	p, err := NewProducer(cfg)
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

	err = p.Publish(messages...)
	is.NoErr(err)

	// start the consumer
	c, err := NewConsumer(cfg)
	is.NoErr(err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Microsecond)
	defer cancel()

	err = c.Start(ctx, 4)
	is.NoErr(err)

	for {
		if c.totalReceived >= msgCount {
			break
		}
	}

	is.Equal(msgCount, c.totalReceived)

	err = c.Shutdown(ctx)
	is.NoErr(err)

	err = p.Shutdown(context.Background())
	is.NoErr(err)
}
