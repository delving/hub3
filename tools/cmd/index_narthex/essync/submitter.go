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

// lineTrackingReader wraps a reader and tracks bytes read since last newline
type lineTrackingReader struct {
	reader          io.Reader
	currentLineSize int64
	linePreview     []byte
	previewSize     int
	mu              sync.Mutex
}

const maxPreviewSize = 1024 // Store first 1KB of each line for identification

func (r *lineTrackingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < n; i++ {
		if p[i] == '\n' {
			r.currentLineSize = 0
			r.linePreview = r.linePreview[:0]
			r.previewSize = 0
		} else {
			r.currentLineSize++
			if r.previewSize < maxPreviewSize {
				r.linePreview = append(r.linePreview, p[i])
				r.previewSize++
			}
		}
	}
	return n, err
}

func (r *lineTrackingReader) CurrentLineSize() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentLineSize
}

func (r *lineTrackingReader) LinePreview() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.linePreview) == 0 {
		return ""
	}
	preview := string(r.linePreview)
	if r.currentLineSize > int64(len(r.linePreview)) {
		preview += "..."
	}
	return preview
}

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
	MaxRecordSize int // max record size in MB (default: 10)
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

	// Use configurable max record size (default 10MB)
	maxRecordSize := bp.config.MaxRecordSize
	if maxRecordSize <= 0 {
		maxRecordSize = 10
	}
	maxCapacity := maxRecordSize * 1024 * 1024

	// Wrap reader to track current line size for better error reporting
	trackingReader := &lineTrackingReader{reader: reader}
	scanner := bufio.NewScanner(trackingReader)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	currentWorker := 0
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		meta := make(map[string]interface{})
		if err := json.Unmarshal(scanner.Bytes(), &meta); err != nil {
			bp.logger.Error("failed to unmarshal meta", "error", err, "line", lineNumber)
			bp.metrics.errors.Value.Add(1)
			continue
		}

		if !scanner.Scan() {
			bp.logger.Error("missing source document", "metaLine", lineNumber, "expectedSourceLine", lineNumber+1)
			bp.metrics.errors.Value.Add(1)
			break
		}
		lineNumber++
		sourceSize := len(scanner.Bytes())

		source := make(map[string]interface{})
		if err := json.Unmarshal(scanner.Bytes(), &source); err != nil {
			bp.logger.Error("failed to unmarshal source", "error", err, "line", lineNumber, "sizeBytes", sourceSize)
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
		recordNumber := (lineNumber / 2) + 1
		if err.Error() == "bufio.Scanner: token too long" {
			actualSize := trackingReader.CurrentLineSize()
			actualSizeMB := float64(actualSize) / (1024 * 1024)
			suggestedSize := int(actualSizeMB) + 5 // Add 5MB buffer
			if suggestedSize < maxRecordSize*2 {
				suggestedSize = maxRecordSize * 2
			}
			bp.logger.Error("record too large for buffer",
				"line", lineNumber+1,
				"recordNumber", recordNumber,
				"actualSizeBytes", actualSize,
				"actualSizeMB", fmt.Sprintf("%.2f", actualSizeMB),
				"maxRecordSizeMB", maxRecordSize,
				"maxRecordSizeBytes", maxCapacity,
				"suggestion", fmt.Sprintf("try --max-record-size %d", suggestedSize),
				"linePreview", trackingReader.LinePreview())
		}
		return fmt.Errorf("scanner error at line %d (record ~%d): %w", lineNumber+1, recordNumber, err)
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
				Status int         `json:"status"`
				Error  interface{} `json:"error,omitempty"` // Using interface{} to handle both string and object error formats
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
			
			// Handle error that can be either a string or an object
			var errorMsg string
			switch err := item.Index.Error.(type) {
			case string:
				errorMsg = err
			case map[string]interface{}:
				// If error is an object, try to extract useful information
				if reason, ok := err["reason"].(string); ok {
					errorMsg = reason
				} else {
					// Fall back to JSON representation
					errorBytes, _ := json.Marshal(err) 
					errorMsg = string(errorBytes)
					// Print the full error object for debugging
					w.processor.logger.Debug("full error object", "error_details", err)
				}
			default:
				// Fall back to %v format for other cases
				errorMsg = fmt.Sprintf("%v", err)
			}
			
			w.processor.logger.Error("document indexing failed",
				"status", item.Index.Status,
				"error", errorMsg)
		}
	}

	w.processor.metrics.recordsSubmitted.Value.Add(int64(successCount))
	w.processor.metrics.bulkRequests.Value.Add(1)

	// Clear buffer
	w.buffer = w.buffer[:0]
}
