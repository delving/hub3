package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

type Publisher struct {
	host       string
	dataPath   string
	BulkSize   int
	client     *retryablehttp.Client
	rw         sync.Mutex
	requests   [][]byte
	nrRequests uint64
	nrSubmits  uint64
}

type PublisherStats struct {
	NrRequests uint64
	NrSubmits  uint64
}

func NewPublisher(host, dataPath string) *Publisher {
	client := retryablehttp.NewClient()
	client.RetryMax = 5

	log.Logger = log.With().Caller().Logger()

	return &Publisher{
		host:     strings.TrimSuffix(host, "/"),
		dataPath: dataPath,
		BulkSize: 250,
		client:   client,
	}
}

// Stats returns the number of request and submits by the Publisher
func (p *Publisher) Stats() PublisherStats {
	return PublisherStats{
		NrRequests: atomic.LoadUint64(&p.nrRequests),
		NrSubmits:  atomic.LoadUint64(&p.nrSubmits),
	}
}

func (p *Publisher) MaxRetries(max int) {
	p.client.RetryMax = max
}

// AppendBytes appends a bulk.Request as []byte to p.requests
//
// When p.requests > p.BulkSize it will submit the chunk to the endpoint
func (p *Publisher) AppendBytes(b []byte) error {
	p.requests = append(p.requests, b)
	atomic.AddUint64(&p.nrRequests, 1)

	if len(p.requests) >= p.BulkSize {
		if err := p.send(); err != nil {
			return err
		}
	}

	return nil
}

// Append appends a bulk.Request as []byte to p.requests
//
// When p.requests > p.BulkSize it will submit the chunk to the endpoint
func (p *Publisher) Append(request *Request) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return p.AppendBytes(b)
}

// Do parses the records in the dataPath and submits them in chunks to the Hub3 BulkAPI endpoint
//
// Chunks are defined by p.BulkSize.
//
// The expected directory structure is:
//
// {orgId}
//
//	/{datasetID}
//		/bulk
//			/{hubid}.jsonl
//
// The jsonl file is assumed to be a bulk.Request serialized on a single line.
// Inside the struct newlines can be escaped.
//
// It will call increment revision at the start and on final submit it will
// clear orphans.
func (p *Publisher) Do(ctx context.Context) error {
	orgIDs, err := os.ReadDir(p.dataPath)
	if err != nil {
		return fmt.Errorf("unable to find datapath path: %w", err)
	}

	for _, orgID := range orgIDs {
		if !orgID.IsDir() {
			continue
		}

		datasetIDs, err := os.ReadDir(filepath.Join(p.dataPath, orgID.Name()))
		if err != nil {
			return fmt.Errorf("unable to find datasetID path: %w", err)
		}

		for _, datasetID := range datasetIDs {
			if !datasetID.IsDir() {
				continue
			}

			recordPath := filepath.Join(p.dataPath, orgID.Name(), datasetID.Name(), "bulk")

			requests, err := os.ReadDir(recordPath)
			if err != nil {
				return fmt.Errorf("unable to find path: %w", err)
			}

			if err := p.incrementRevision(orgID.Name(), datasetID.Name()); err != nil {
				return err
			}

			for _, request := range requests {
				requestPath := filepath.Join(recordPath, request.Name())

				if !strings.HasSuffix(request.Name(), ".jsonl") {
					continue
				}

				// log.Info().Str("path", requestPath).Msg("file being processed")

				b, err := os.ReadFile(requestPath)
				if err != nil {
					return fmt.Errorf("unable to read jsonl file; %w", err)
				}

				if err := p.AppendBytes(b); err != nil {
					return err
				}

				select {
				case <-ctx.Done():
					log.Info().Msg("context canceled bulk loading")
					return ctx.Err()
				default:
				}
			}

			if err := p.dropOrphans(orgID.Name(), datasetID.Name()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Publisher) endpoint() string {
	return fmt.Sprintf("%s/api/index/bulk", p.host)
}

// send records to endpoint and reset p.records to empty
func (p *Publisher) send() error {
	p.rw.Lock()
	defer p.rw.Unlock()

	body := bytes.Join(p.requests, []byte("\n"))

	resp, err := p.post(body)
	if err != nil {
		return fmt.Errorf("unable to post to endpoint: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		return fmt.Errorf("unable to post to endpoint: %s (%d)", string(b), resp.StatusCode)
	}

	atomic.AddUint64(&p.nrSubmits, 1)

	log.Info().Uint64("submits", p.nrSubmits).Uint64("processed", p.nrRequests).Msg("bulk publishing progress")

	p.requests = nil

	return nil
}

func (p *Publisher) post(body []byte) (*http.Response, error) {
	return p.client.Post(p.endpoint(), "text/plain", body)
}

func (p *Publisher) incrementRevision(orgID, datasetID string) error {
	req := Request{
		OrgID:     orgID,
		DatasetID: datasetID,
		Action:    "increment_revision",
	}

	err := p.Append(&req)
	if err != nil {
		return err
	}

	return p.send()
}

func (p *Publisher) dropOrphans(orgID, datasetID string) error {
	req := Request{
		OrgID:     orgID,
		DatasetID: datasetID,
		Action:    "drop_orphans",
	}

	err := p.Append(&req)
	if err != nil {
		return err
	}

	return p.send()
}
