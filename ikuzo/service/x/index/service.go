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
	"sync"
	"sync/atomic"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/nats-io/stan.go"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
		Identical  uint64
		Command    uint64 // e.g. drop orphans, increment revision
		AddRef     uint64
		StoreRef   uint64
		FlushRef   uint64
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
	store      *store
	queue      chan shaRef
	ctx        context.Context
	cancel     context.CancelFunc
	group      *errgroup.Group
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

	store, err := newStore()
	if err != nil {
		return nil, err
	}

	s.store = store

	workers := 4
	batchSize := 1000
	s.queue = make(chan shaRef, workers*batchSize)

	if err := s.startBatchWriter(workers, batchSize); err != nil {
		return s, err
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
	// cancel context
	s.cancel()

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

	// closing add queue
	close(s.queue)

	if err := s.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
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
			s.handleMessage(ctx),
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

func (s *Service) handleMessage(ctx context.Context) func(m *stan.Msg) {
	return func(m *stan.Msg) {
		atomic.AddUint64(&s.m.Nats.Consumed, 1)

		var msg domainpb.IndexMessage
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			log.Error().Err(err).Msg("unable to unmarshal indexmessage in index consumer")
			return
		}

		if s.MsgHandler != nil {
			if err := s.MsgHandler(ctx, &msg); err != nil {
				log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
				return
			}
		}

		if s.bi != nil {
			if err := s.submitBulkMsg(ctx, &msg); err != nil {
				log.Error().Err(err).Msg("unable to process *domain.IndexMessage")
				return
			}
		}
	}
}

func (s *Service) dropOrphanGroup(orgID, datasetID string, revision *domainpb.Revision) error {
	tags := elastic.NewBoolQuery()
	for _, tag := range []string{"findingAid", "mets"} {
		tags = tags.Should(elastic.NewTermQuery("meta.tags", tag))
	}

	v2 := elastic.NewBoolQuery()
	if revision.GetSHA() != "" {
		v2 = v2.MustNot(elastic.NewMatchQuery("meta.sourceID", revision.GetSHA()))
		v2 = v2.Must(elastic.NewMatchQuery("meta.groupID", revision.GetGroupID()))
	} else {
		// drop all for sourcepath
		v2 = v2.Must(elastic.NewMatchQuery("meta.sourcePath", revision.GetPath()))
	}

	v2 = v2.Must(tags)
	v2 = v2.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, datasetID))
	v2 = v2.Must(elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, orgID))

	res, err := index.ESClient().DeleteByQuery().
		Index(c.Config.ElasticSearch.GetIndexName()).
		Query(v2).
		Conflicts("proceed"). // default is abort
		Do(context.Background())
	if err != nil {
		log.Warn().Msgf("Unable to delete orphaned dataset records from index: %s.", err)
		return err
	}

	if res == nil {
		unexpectedResponseMsg := "expected response != nil; got: %v"
		log.Warn().Msgf(unexpectedResponseMsg, res)

		return fmt.Errorf(unexpectedResponseMsg, res)
	}

	log.Info().Msgf(
		"Removed %d records for spec %s for inventory %s",
		res.Deleted,
		datasetID,
		revision.GetGroupID(),
	)

	return nil
}

// dropOrphans is a background function to remove orphans from the index when the timer is expired
func (s *Service) dropOrphans(orgID, datasetID string, revision *domainpb.Revision) {
	go func() {
		// block for orphanWait seconds to allow cluster to be in sync
		timer := time.NewTimer(time.Second * time.Duration(s.orphanWait))
		<-timer.C

		if revision.GetSHA() != "" && revision.GetPath() != "" {
			if err := s.dropOrphanGroup(orgID, datasetID, revision); err != nil {
				log.Error().Err(err).Msg("unable to drop orphans")
			}

			if err := s.dropOrphanGroup(orgID, datasetID, revision); err != nil {
				log.Error().
					Err(err).
					Str("datasetID", datasetID).
					Msg("unable to drop orphan group")
			}

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

func (s *Service) add(ctx context.Context, ref shaRef) error {
	atomic.AddUint64(&s.m.Index.AddRef, 1)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.queue <- ref:
	}

	return nil
}

func (s *Service) submitBulkMsg(ctx context.Context, m *domainpb.IndexMessage) error {
	if s.MsgHandler != nil {
		return s.MsgHandler(ctx, m)
	}

	if m.GetActionType() == domainpb.ActionType_DROP_ORPHANS {
		s.dropOrphans(m.GetOrganisationID(), m.GetDatasetID(), m.GetRevision())
		atomic.AddUint64(&s.m.Index.Command, 1)

		return nil
	}

	action := "index"

	if m.GetDeleted() {
		action = "delete"

		if err := s.store.Delete(m.GetRecordID()); err != nil {
			return err
		}
	} else {
		ok, err := s.store.HashIsEqual(m.GetRecordID(), m.GetRevision().GetSHA())
		if err != nil {
			return err
		}

		if ok {
			atomic.AddUint64(&s.m.Index.Identical, 1)
			// equal with index so we are done
			return nil
		}

		// TODO(kiivihal): must speed up the storing of records in batches of 1000
		ref := shaRef{HubID: m.GetRecordID(), Sha: m.GetRevision().GetSHA()}
		if err := s.add(ctx, ref); err != nil {
			return err
		}
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

type batchWriter struct {
	refs      []shaRef
	s         *Service
	batchSize int
	ticker    *time.Ticker
	rw        sync.Mutex
}

func newBatchWriter(s *Service, batchSize int) batchWriter {
	ticker := time.NewTicker(5 * time.Second)
	return batchWriter{
		refs:      []shaRef{},
		batchSize: batchSize,
		s:         s,
		ticker:    ticker,
	}
}

func (b *batchWriter) flush() error {
	b.rw.Lock()
	defer b.rw.Unlock()
	if len(b.refs) == 0 {
		// nothing to do
		return nil
	}

	atomic.AddUint64(&b.s.m.Index.StoreRef, uint64(len(b.refs)))
	atomic.AddUint64(&b.s.m.Index.FlushRef, 1)

	if err := b.s.store.Put(b.refs...); err != nil {
		log.Error().Err(err).Msg("unable to store shaRefs")
	}

	b.refs = []shaRef{}

	return nil
}

func (b *batchWriter) run(ctx context.Context) error {

	for ref := range b.s.queue {
		if len(b.refs) >= b.batchSize {
			if err := b.flush(); err != nil {
				log.Error().Err(err).Msg("unable to flush shaRefs")
			}
		}

		b.rw.Lock()
		b.refs = append(b.refs, ref)
		b.rw.Unlock()

		select {
		case <-ctx.Done():
			if err := b.flush(); err != nil {
				log.Error().Err(err).Msg("unable to flush shaRefs")
			}

			return ctx.Err()
		default:
		}
	}

	return nil
}

func (s *Service) startBatchWriter(workers, batchSize int) error {
	// create errgroup and add cancel to service
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	s.cancel = cancel
	s.group = g
	s.ctx = gctx

	batchWorkers := []*batchWriter{}

	for i := 0; i < workers; i++ {
		w := newBatchWriter(s, batchSize)
		batchWorkers = append(batchWorkers, &w)

		g.Go(func() error {
			return w.run(gctx)
		})
	}

	ticker := time.NewTicker(5 * time.Second)

	g.Go(func() error {
		for {
			select {
			case <-gctx.Done():
				return ctx.Err()
			case <-ticker.C:
				log.Debug().Msg("timed flush")
				for _, w := range batchWorkers {
					if err := w.flush(); err != nil {
						log.Error().Err(err).Msg("unable to flush shaRefs")
					}
				}
			}
		}
	})

	return nil
}
