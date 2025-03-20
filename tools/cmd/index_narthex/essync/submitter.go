package essync

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/tools/cmd/index_narthex/stats"
)

type BulkProcessor struct {
	config  Config
	stats   *stats.Stats
	logger  *slog.Logger
	workers []*worker
	wg      sync.WaitGroup
	metrics struct {
		recordsSeen      *stats.Metric
		recordsSubmitted *stats.Metric
		bulkRequests     *stats.Metric
		errors           *stats.Metric
	}
}

type Config struct {
	NumWorkers    int
	MaxBatchSize  int
	UpdateTimeout time.Duration
	ESClient      *elasticsearch.Client
	Reporter      stats.Reporter
}

type BulkRequest struct {
	Meta   map[string]interface{}
	Source map[string]interface{}
}

func NewBulkProcessor(config Config) *BulkProcessor {
	s := stats.NewStats()
	bp := &BulkProcessor{
		config:  config,
		stats:   s,
		logger:  slog.Default(),
		workers: make([]*worker, config.NumWorkers),
	}

	// Register metrics
	bp.metrics.recordsSeen = s.RegisterMetric("records_seen")
	bp.metrics.recordsSubmitted = s.RegisterMetric("records_submitted")
	bp.metrics.bulkRequests = s.RegisterMetric("bulk_requests")
	bp.metrics.errors = s.RegisterMetric("errors")

	for i := 0; i < config.NumWorkers; i++ {
		bp.workers[i] = &worker{
			id:         i,
			buffer:     make([]BulkRequest, 0, config.MaxBatchSize),
			processor:  bp,
			bufferChan: make(chan BulkRequest, config.MaxBatchSize),
			flushChan:  make(chan struct{}),
		}
	}

	return bp
}

func (bp *BulkProcessor) Process(ctx context.Context, reader io.Reader) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := bp.config.Reporter.Start(ctx, "bulk processing"); err != nil {
		return fmt.Errorf("failed to start reporter: %w", err)
	}

	// Start workers
	for _, w := range bp.workers {
		bp.wg.Add(1)
		go w.run(ctx)
	}

	// Start stats updater
	go bp.updateStats(ctx)

	scanner := bufio.NewScanner(reader)

	// Increase the buffer size
	const maxCapacity = 1024 * 1024 * 10 // 10MB or adjust as needed
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	currentWorker := 0

	for scanner.Scan() {
		meta := make(map[string]interface{})
		if err := json.Unmarshal(scanner.Bytes(), &meta); err != nil {
			bp.logger.Error("failed to unmarshal meta", "error", err)
			bp.metrics.errors.Value.Add(1)
			continue
		}

		if !scanner.Scan() {
			bp.logger.Error("missing source document")
			bp.metrics.errors.Value.Add(1)
			break
		}

		source := make(map[string]interface{})
		if err := json.Unmarshal(scanner.Bytes(), &source); err != nil {
			bp.logger.Error("failed to unmarshal source", "error", err)
			bp.metrics.errors.Value.Add(1)
			continue
		}

		bp.metrics.recordsSeen.Value.Add(1)

		w := bp.workers[currentWorker]
		select {
		case w.bufferChan <- BulkRequest{Meta: meta, Source: source}:
		case <-ctx.Done():
			return ctx.Err()
		}
		currentWorker = (currentWorker + 1) % bp.config.NumWorkers
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// Signal workers to flush and stop
	for _, w := range bp.workers {
		close(w.bufferChan)
	}

	bp.wg.Wait()
	bp.config.Reporter.Complete(bp.stats.GetValues())
	return nil
}

func (bp *BulkProcessor) updateStats(ctx context.Context) {
	ticker := time.NewTicker(bp.config.UpdateTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bp.config.Reporter.Update(bp.stats.GetValues())
		case <-ctx.Done():
			return
		}
	}
}

type worker struct {
	id         int
	buffer     []BulkRequest
	processor  *BulkProcessor
	bufferChan chan BulkRequest
	flushChan  chan struct{}
}

func (w *worker) run(ctx context.Context) {
	defer w.processor.wg.Done()

	for {
		select {
		case req, ok := <-w.bufferChan:
			if !ok {
				// Channel closed, flush remaining and exit
				if len(w.buffer) > 0 {
					w.flush()
				}
				return
			}

			w.buffer = append(w.buffer, req)
			if len(w.buffer) >= w.processor.config.MaxBatchSize {
				w.flush()
			}

		case <-ctx.Done():
			if len(w.buffer) > 0 {
				w.flush()
			}
			return
		}
	}
}

func (w *worker) flush() {
	if len(w.buffer) == 0 {
		return
	}

	// Build bulk request body
	var bulkBody bytes.Buffer
	for _, req := range w.buffer {
		metaJSON, err := json.Marshal(req.Meta)
		if err != nil {
			w.processor.logger.Error("failed to marshal meta", "error", err)
			w.processor.metrics.errors.Value.Add(1)
			continue
		}
		bulkBody.Write(metaJSON)
		bulkBody.WriteByte('\n')

		if len(req.Source) == 0 {
			continue
		}

		sourceJSON, err := json.Marshal(req.Source)
		if err != nil {
			w.processor.logger.Error("failed to marshal source", "error", err)
			w.processor.metrics.errors.Value.Add(1)
			continue
		}
		bulkBody.Write(sourceJSON)
		bulkBody.WriteByte('\n')
	}

	// Submit bulk request
	res, err := w.processor.config.ESClient.BulkWithResponse(context.Background(), bytes.NewReader(bulkBody.Bytes()))
	if err != nil {
		w.processor.logger.Error("bulk request failed", "error", err)
		w.processor.metrics.errors.Value.Add(1)
		return
	}

	// Parse response
	var bulkResponse struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index struct {
				Status int    `json:"status"`
				Error  string `json:"error,omitempty"`
			} `json:"index"`
		} `json:"items"`
	}

	if err := json.NewDecoder(res).Decode(&bulkResponse); err != nil {
		var body []byte
		body, readErr := io.ReadAll(res)
		if readErr != nil {
			w.processor.logger.Error("failed to read bulk response body", "error", readErr)
		}
		w.processor.logger.Error("failed to parse bulk response", "error", err, "response", string(body))
		w.processor.metrics.errors.Value.Add(1)
		return
	}

	// Count successful operations
	successCount := 0
	for _, item := range bulkResponse.Items {
		if item.Index.Status >= 200 && item.Index.Status < 300 {
			successCount++
		} else {
			w.processor.metrics.errors.Value.Add(1)
			w.processor.logger.Error("document indexing failed",
				"status", item.Index.Status,
				"error", item.Index.Error)
		}
	}

	w.processor.metrics.recordsSubmitted.Value.Add(int64(successCount))
	w.processor.metrics.bulkRequests.Value.Add(1)

	// Clear buffer
	w.buffer = w.buffer[:0]
}
