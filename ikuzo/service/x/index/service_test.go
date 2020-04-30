package index

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/matryer/is"
	"github.com/nats-io/stan.go"
)

func testConfig() (*NatsConfig, error) {
	var (
		err error
		cfg = &NatsConfig{}
	)

	cfg.setDefaults()

	// Connect to Streaming server
	cfg.Conn, err = stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(stan.DefaultNatsURL))
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

	s, err := NewService(
		SetNatsConfiguration(cfg),
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

	err = s.Publish(context.Background(), messages...)
	is.NoErr(err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)

	err = s.Start(ctx, 4)
	is.NoErr(err)

	var consumed uint64

L:
	for {
		select {
		case <-ctx.Done():
			consumed = atomic.LoadUint64(&s.m.Nats.Consumed)
			t.Logf("messages consumed: %d", s.m.Nats.Consumed)
			ticker.Stop()
			break L
		case <-ticker.C:
			consumed = atomic.LoadUint64(&s.m.Nats.Consumed)
			t.Logf("messages consumed: %d", s.m.Nats.Consumed)

			if int(consumed) >= msgCount {
				break L
			}
		}
	}

	is.Equal(msgCount, int(consumed))

	err = s.Shutdown(ctx)
	is.NoErr(err)
}
