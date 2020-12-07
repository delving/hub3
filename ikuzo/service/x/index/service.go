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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
	proto "google.golang.org/protobuf/proto"
)

type BulkIndex interface {
	Publish(ctx context.Context, message ...*domainpb.IndexMessage) error
}

type Metrics struct {
	started time.Time
	Nats    struct {
		Published uint64
		Consumed  uint64
		Failed    uint64
	}
	Index struct {
		Successful uint64
		Failed     uint64
	}
}

type Service struct {
	bi         esutil.BulkIndexer
	stan       *NatsConfig
	direct     bool
	MsgHandler func(ctx context.Context, m *domainpb.IndexMessage) error
	workers    []stan.Subscription // this is for getting statistics
	m          Metrics
	orphanWait int
	postHooks  map[string][]domain.PostHookService
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		m:          Metrics{started: time.Now()},
		orphanWait: 15,
		postHooks:  map[string][]domain.PostHookService{},
	}

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

func (s *Service) Publish(ctx context.Context, messages ...*domainpb.IndexMessage) error {
	for _, msg := range messages {
		// if direct submit msg directly to BulkIndexer
		if s.direct {
			if submitErr := s.submitBulkMsg(ctx, msg); submitErr != nil {
				return fmt.Errorf("unable to index message; %w", submitErr)
			}

			continue
		}

		b, err := proto.Marshal(msg)
		if err != nil {
			atomic.AddUint64(&s.m.Nats.Failed, 1)
			return fmt.Errorf("unable to marshal index message; %w", err)
		}

		if err = s.stan.Conn.Publish(s.stan.SubjectID, b); err != nil {
			atomic.AddUint64(&s.m.Nats.Failed, 1)
			log.Error().Err(err).Msgf("stan config: %+v", s.stan)

			return fmt.Errorf("unable to publish to queue; %w", err)
		}

		atomic.AddUint64(&s.m.Nats.Published, 1)
	}

	return nil
}

func (s *Service) Metrics() Metrics {
	return s.m
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// added to implement ikuzo service interface
}

func (s *Service) Shutdown(ctx context.Context) error {
	// stop all the workers before closing channels
	for _, w := range s.workers {
		w.Close()
	}

	if s.stan != nil {
		if err := s.stan.Conn.Close(); err != nil {
			return err
		}
	}

	if s.bi != nil {
		s.bi.Stats()

		if err := s.bi.Close(ctx); err != nil {
			return err
		}
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
	atomic.AddUint64(&s.m.Nats.Consumed, 1)

	var msg domainpb.IndexMessage
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		log.Error().Err(err).Msg("unable to unmarshal indexmessage in index consumer")
		return
	}

	if s.MsgHandler != nil {
		if err := s.MsgHandler(context.Background(), &msg); err != nil {
			log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
			return
		}
	}

	// TODO(kiivihal): propagate the context
	if s.bi != nil {
		if err := s.submitBulkMsg(context.Background(), &msg); err != nil {
			log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
			return
		}
	}
}

func (s *Service) dropOrphanGroup(orgID, datasetID string, revision *domainpb.Revision) {
	// TODO(kiivihal): implement me
	log.Debug().Msgf("dropping grouped orphans")
}

// dropOrphans is a background function to remove orphans from the index when the timer is expired
func (s *Service) dropOrphans(orgID, datasetID string, revision *domainpb.Revision) {
	go func() {
		// block for orphanWait seconds to allow cluster to be in sync
		timer := time.NewTimer(time.Second * time.Duration(s.orphanWait))
		<-timer.C

		if revision.GetSHA() != "" && revision.GetPath() != "" {
			s.dropOrphanGroup(orgID, datasetID, revision)
			return
		}

		ds, err := models.GetDataSet(orgID, datasetID)
		if err != nil {
			log.Error().
				Err(err).
				Str("datasetID", datasetID).
				Msg("unable to retrieve dataset")

			return
		}

		if ds.Revision != int(revision.GetNumber()) {
			log.Warn().
				Int32("message_revision", revision.GetNumber()).
				Int("dataset_revision", ds.Revision).
				Msg("message revision is older so not dropping orphans")

			return
		}

		if _, err := ds.DropOrphans(context.Background(), nil, nil); err != nil {
			log.Error().
				Err(err).
				Msg("unable to drop orphans")
		}

		if len(s.postHooks) != 0 {
			applyHooks, ok := s.postHooks[orgID]
			if ok {
				go func(revision int) {
					posthookTimer := time.NewTimer(5 * time.Second)
					<-posthookTimer.C

					for _, hook := range applyHooks {
						resp, err := hook.DropDataset(datasetID, revision)
						if err != nil {
							log.Error().Err(err).Str("datasetID", datasetID).Msg("unable to drop posthook dataset")
							continue
						}

						if resp.StatusCode > 299 {
							defer resp.Body.Close()
							body, readErr := ioutil.ReadAll(resp.Body)

							if readErr != nil {
								log.Error().Err(err).Str("datasetID", datasetID).
									Msg("unable to read posthook body")
							}

							log.Error().Err(err).
								Str("body", string(body)).
								Int("revision", revision).
								Int("status_code", resp.StatusCode).
								Str("datasetID", datasetID).
								Msg("unable to drop posthook dataset")
						}

						log.Info().Str("datasetID", datasetID).Str("posthook", hook.Name()).Int("revision", revision).Msg("dropped posthook orphans")
					}
				}(int(revision.GetNumber()))
			}
		}
	}()
}

func (s *Service) submitBulkMsg(ctx context.Context, m *domainpb.IndexMessage) error {
	if s.MsgHandler != nil {
		return s.MsgHandler(ctx, m)
	}

	if m.GetActionType() == domainpb.ActionType_DROP_ORPHANS {
		s.dropOrphans(m.GetOrganisationID(), m.GetDatasetID(), m.GetRevision())

		return nil
	}

	action := "index"

	if m.GetDeleted() {
		action = "delete"
	}

	bulkMsg := esutil.BulkIndexerItem{
		// Action field configures the operation to perform (index, create, delete, update)
		Action: action,

		// Index is the target index
		Index: m.GetIndexName(),

		// DocumentID is the (optional) document ID
		DocumentID: m.GetRecordID(),

		// OnSuccess is called for each successful operation
		OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
			atomic.AddUint64(&s.m.Index.Successful, 1)
		},

		// OnFailure is called for each failed operation
		OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
			atomic.AddUint64(&s.m.Index.Failed, 1)
			if err != nil {
				log.Error().Err(err).Msg("bulk index msg error")
			} else if res.Status != http.StatusNotFound {
				body, _ := ioutil.ReadAll(item.Body)
				log.Error().
					Str("reason", res.Error.Reason).
					Str("type", res.Error.Type).
					Str("hubID", res.DocumentID).
					Str("index", res.Index).
					Int("status", res.Status).
					Str("result", res.Result).
					Bytes("body", body).
					Str("reason", res.Error.Reason).
					Bytes("item", body).
					Msg("bulk index msg error")
			}
		},
	}

	if m.GetSource() != nil {
		// Body is an `io.Reader` with the payload
		bulkMsg.Body = bytes.NewReader(m.GetSource())
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		log.Info().Msg("stop publishing")
		return ctx.Err()
	}

	return s.bi.Add(
		ctx,
		bulkMsg,
	)
}

func (s *Service) BulkIndexStats() esutil.BulkIndexerStats {
	if s.bi == nil {
		return esutil.BulkIndexerStats{}
	}

	return s.bi.Stats()
}

// AddPostHook adds posthook to the indexing service
func (s *Service) AddPostHook(hook domain.PostHookService) error {
	s.postHooks[hook.OrgID()] = append(s.postHooks[hook.OrgID()], hook)
	return nil
}
