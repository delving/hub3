package index

import (
	"bytes"
	"context"
	"sync/atomic"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
)

const (
	clientID     = "hub3-pub"
	clusterID    = "hub3-nats"
	durableName  = "hub3-worker"
	durableQueue = "hub3-queue"
	subjectID    = "hub3-bulk-index"
)

// TODO(kiivihal): create switch for submitting to BulkIndexer directly
type Config struct {
	Indexer         esutil.BulkIndexer
	Client          *elasticsearch.Client
	Queue           stan.Conn
	SubjectID       string
	ClusterID       string
	ClientID        string
	DurableName     string
	DurableQueue    string
	Direct          bool // schedule direct in elasticsearch and not via a queue
	countSuccessful uint64
	countFailure    uint64
}

func (c *Config) setDefaults() {
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

func (c *Config) submitBulkMsg(m *domainpb.IndexMessage) error {
	action := "index"

	if m.GetDeleted() {
		action = "delete"
	}

	return c.Indexer.Add(
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
				atomic.AddUint64(&c.countSuccessful, 1)
			},

			// OnFailure is called for each failed operation
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				atomic.AddUint64(&c.countFailure, 1)
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
