package index

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
	proto "google.golang.org/protobuf/proto"
)

type Consumer struct {
	cfg           *Config
	started       bool
	workers       []stan.Subscription // this is for getting statistics
	totalReceived int
	m             sync.Mutex
	MsgHandler    func(m *domainpb.IndexMessage) error
}

func NewConsumer(cfg *Config) (*Consumer, error) {
	c := &Consumer{
		cfg: cfg,
	}

	return c, nil
}

func (c *Consumer) Start(ctx context.Context, workers int) error {
	if c.started {
		return fmt.Errorf("consumer is already started")
	}

	c.started = true

	for i := 0; i < workers; i++ {
		// create consumer
		qsub, err := c.cfg.Queue.QueueSubscribe(
			c.cfg.SubjectID,
			c.cfg.DurableQueue,
			c.handleMessage,
			stan.DurableName(c.cfg.DurableName),
		)
		if err != nil {
			return err
		}

		// add worker for statistics
		c.workers = append(c.workers, qsub)
	}

	return nil
}

func (c *Consumer) handleMessage(m *stan.Msg) {
	c.m.Lock()
	defer c.m.Unlock()
	c.totalReceived++

	if c.MsgHandler != nil {
		var msg domainpb.IndexMessage
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			log.Error().Err(err).Msg("unable to unmarshal indexmessage in index consumer")
			return
		}

		if err := c.MsgHandler(&msg); err != nil {
			log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
			return
		}
	}
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	if err := c.cfg.Queue.Close(); !errors.Is(err, stan.ErrConnectionClosed) {
		return err
	}

	// stop all the workers
	for _, w := range c.workers {
		w.Close()
	}

	// stop bulk indexer
	if err := c.cfg.Indexer.Close(ctx); err != nil {
		return err
	}

	return nil
}
