package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/nats-io/stan.go"
	proto "google.golang.org/protobuf/proto"
)

type Producer struct {
	cfg *Config
}

func NewProducer(cfg *Config) (*Producer, error) {
	if cfg == nil {
		return &Producer{}, fmt.Errorf("index.Config cannot be nil when starting producer")
	}

	if !cfg.Direct && (cfg.Queue == nil || cfg.Queue.NatsConn() == nil) {
		return &Producer{}, fmt.Errorf("stan.Conn must be established before started the producer")
	}

	if cfg.Direct && cfg.Indexer == nil {
		return &Producer{}, fmt.Errorf("in direct mode the elastic.BulkIndexer must be set")
	}

	return &Producer{
		cfg: cfg,
	}, nil
}

func (p *Producer) Publish(messages ...*domainpb.IndexMessage) error {
	for _, msg := range messages {
		b, err := proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("unable to marshal index message; %w", err)
		}

		// if direct submit msg directly to BulkIndexer
		if p.cfg.Direct {
			if submitErr := p.cfg.submitBulkMsg(msg); submitErr != nil {
				return fmt.Errorf("unable to index message; %w", submitErr)
			}

			continue
		}

		if err = p.cfg.Queue.Publish(subjectID, b); err != nil {
			return fmt.Errorf("unable to publish to queue; %w", err)
		}
	}

	return nil
}

func (p *Producer) Shutdown(ctx context.Context) error {
	if err := p.cfg.Queue.Close(); !errors.Is(err, stan.ErrConnectionClosed) {
		return err
	}

	if p.cfg.Direct {
		if err := p.cfg.Indexer.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}
