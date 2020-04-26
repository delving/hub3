package index

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
	proto "google.golang.org/protobuf/proto"
)

type Metrics struct {
	nats struct {
		published uint64
		consumed  uint64
		failed    uint64
	}
	index struct {
		successful uint64
		failed     uint64
	}
}

type Service struct {
	bi         esutil.BulkIndexer
	stan       *NatsConfig
	direct     bool
	MsgHandler func(m *domainpb.IndexMessage) error
	workers    []stan.Subscription // this is for getting statistics
	m          Metrics
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.stan == nil {
		s.direct = true
		if s.bi == nil {
			return s, fmt.Errorf("in direct mode an esutil.BulkIndexer must be set")
		}
	}

	if !s.direct && (s.stan == nil || s.stan.Conn.NatsConn() == nil) {
		return s, fmt.Errorf("stan.Conn must be established before nats queue can be used")
	}

	return s, nil
}

func (s *Service) Publish(messages ...*domainpb.IndexMessage) error {
	for _, msg := range messages {
		// if direct submit msg directly to BulkIndexer
		if s.direct {
			if submitErr := s.submitBulkMsg(msg); submitErr != nil {
				return fmt.Errorf("unable to index message; %w", submitErr)
			}

			continue
		}

		b, err := proto.Marshal(msg)
		if err != nil {
			atomic.AddUint64(&s.m.nats.failed, 1)
			return fmt.Errorf("unable to marshal index message; %w", err)
		}

		if err = s.stan.Conn.Publish(s.stan.SubjectID, b); err != nil {
			atomic.AddUint64(&s.m.nats.failed, 1)
			return fmt.Errorf("unable to publish to queue; %w", err)
		}

		atomic.AddUint64(&s.m.nats.published, 1)
	}

	return nil
}

func (s *Service) Metrics() Metrics {
	return s.m
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.stan != nil {
		if err := s.stan.Conn.Close(); err != nil {
			return err
		}
	}

	if s.bi != nil {
		if err := s.bi.Close(ctx); err != nil {
			return err
		}
	}

	// stop all the workers
	for _, w := range s.workers {
		w.Close()
	}

	// remove workers so the service could restart
	s.workers = nil

	return nil
}
func (s *Service) Start(ctx context.Context, workers int) error {
	if len(s.workers) != 0 {
		return fmt.Errorf("consumer is already started")
	}

	for i := 0; i < workers; i++ {
		// create consumer
		qsub, err := s.stan.Conn.QueueSubscribe(
			s.stan.SubjectID,
			s.stan.DurableQueue,
			s.handleMessage,
			stan.DurableName(s.stan.DurableName),
		)
		if err != nil {
			return err
		}

		// add worker for statistics
		s.workers = append(s.workers, qsub)
	}

	return nil
}

func (s *Service) handleMessage(m *stan.Msg) {
	atomic.AddUint64(&s.m.nats.consumed, 1)

	if s.MsgHandler != nil {
		var msg domainpb.IndexMessage
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			log.Error().Err(err).Msg("unable to unmarshal indexmessage in index consumer")
			return
		}

		if err := s.MsgHandler(&msg); err != nil {
			log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
			return
		}
	}
}

func (s *Service) submitBulkMsg(m *domainpb.IndexMessage) error {
	if s.MsgHandler != nil {
		return s.MsgHandler(m)
	}

	action := "index"

	if m.GetDeleted() {
		action = "delete"
	}

	return s.bi.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			// Action field configures the operation to perform (index, create, delete, update)
			Action: action,

			// Index is the target index
			Index: m.GetIndexName(),

			// DocumentID is the (optional) document ID
			DocumentID: m.GetRecordID(),

			// Body is an `io.Reader` with the payload
			Body: bytes.NewReader(m.GetSource()),

			// OnSuccess is called for each successful operation
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				atomic.AddUint64(&s.m.index.successful, 1)
			},

			// OnFailure is called for each failed operation
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				atomic.AddUint64(&s.m.index.failed, 1)
				if err != nil {
					log.Error().Err(err).Msg("bulk index error")
				} else {
					log.Error().
						Str("type", res.Error.Type).
						Str("reason", res.Error.Reason).
						Msg("bulk index error")
				}
			},
		},
	)
}
