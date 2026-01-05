package essync

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/tools/cmd/index_narthex/stats"
	"github.com/klauspost/compress/zstd"
	"github.com/olivere/elastic/v7"
)

type ChangeDetector struct {
	existingDocs map[string]string // map[id]checksum
	encoder      *zstd.Encoder
	outputFile   *os.File
	reporter     stats.Reporter
	scanStats    *stats.Stats // stats.Stats for document scanning
	changeStats  *stats.Stats // stats.Stats for change detection
	metrics      struct {
		// Scan metrics
		scanned *stats.Metric

		// Change metrics
		unchanged     *stats.Metric
		newRecords    *stats.Metric
		updates       *stats.Metric
		deletesFound  *stats.Metric
		sourceDeletes *stats.Metric
	}
}

func NewChangeDetector(outputPath string, reporter stats.Reporter) (*ChangeDetector, error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create output file: %w", err)
	}

	enc, err := zstd.NewWriter(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}

	cd := &ChangeDetector{
		existingDocs: make(map[string]string),
		encoder:      enc,
		outputFile:   f,
		reporter:     reporter,
		scanStats:    stats.NewStats(),
		changeStats:  stats.NewStats(),
	}

	// Register scan metrics
	cd.metrics.scanned = cd.scanStats.RegisterMetric("records_scanned")

	// Register change metrics
	cd.metrics.unchanged = cd.changeStats.RegisterMetric("unchanged")
	cd.metrics.newRecords = cd.changeStats.RegisterMetric("new_records")
	cd.metrics.updates = cd.changeStats.RegisterMetric("updates")
	cd.metrics.deletesFound = cd.changeStats.RegisterMetric("deletes_found")
	cd.metrics.sourceDeletes = cd.changeStats.RegisterMetric("source_deletes")

	return cd, nil
}

func (cd *ChangeDetector) Close() error {
	if err := cd.encoder.Close(); err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}
	return cd.outputFile.Close()
}

func (cd *ChangeDetector) LoadExistingDocs(client *elasticsearch.Client, opts LoadOptions) error {
	if len(opts.Query) == 0 {
		opts.Query = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	q, err := convertMapToElasticQuery(opts.Query)
	if err != nil {
		return fmt.Errorf("unable to convert query: %w", err)
	}

	cfg := &elasticsearch.StreamConfig{
		Query:      q,
		IndexNames: []string{opts.Index},
		Fsc:        elastic.NewFetchSourceContext(true).Include("_checksum"),
	}

	ctx := context.Background()
	if err := cd.reporter.Start(ctx, "scanning existing documents"); err != nil {
		return fmt.Errorf("start reporter: %w", err)
	}

	// Start scan stats.Stats reporting
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cd.reporter.Update(cd.scanStats.GetValues())
			case <-ctx.Done():
				return
			}
		}
	}()

	_, err = client.Stream(ctx, cfg, func(hit *elastic.SearchHit) error {
		var source struct {
			Checksum string `json:"_checksum"`
		}
		if err := json.Unmarshal(hit.Source, &source); err != nil {
			return fmt.Errorf("unmarshal source: %w", err)
		}
		cd.existingDocs[hit.Id] = source.Checksum
		cd.metrics.scanned.Value.Add(1)
		return nil
	})
	if err != nil {
		return fmt.Errorf("stream documents: %w", err)
	}

	cd.reporter.Complete(cd.scanStats.GetValues())
	return nil
}

func (cd *ChangeDetector) DetectChanges(inputPath string, opts LoadOptions) error {
	ctx := context.Background()
	if err := cd.reporter.Start(ctx, "detecting changes"); err != nil {
		return fmt.Errorf("start reporter: %w", err)
	}

	indexFname, err := FindLatestPath(inputPath, opts.FnameFilter)
	if err != nil {
		return fmt.Errorf("unable to find input file; %w", err)
	}

	slog.Info("using index file", "file", indexFname)

	file, err := os.Open(indexFname)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(indexFname, ".zst") {
		decoder, err := zstd.NewReader(file)
		if err != nil {
			return fmt.Errorf("create zstd decoder: %w", err)
		}
		defer decoder.Close()
		reader = decoder
	}

	// Start change stats.Stats reporting
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cd.reporter.Update(cd.changeStats.GetValues())
			case <-ctx.Done():
				return
			}
		}
	}()

	// Use configurable max record size (default 10MB)
	maxRecordSize := opts.MaxRecordSize
	if maxRecordSize <= 0 {
		maxRecordSize = 10
	}
	maxCapacity := maxRecordSize * 1024 * 1024

	// Wrap reader to track current line size for better error reporting
	trackingReader := &lineTrackingReader{reader: reader}
	scanner := bufio.NewScanner(trackingReader)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()

		// Skip empty lines
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var meta map[string]interface{}
		if err := json.Unmarshal(line, &meta); err != nil {
			continue
		}

		if err := cd.processLine(meta, scanner, &lineNumber); err != nil {
			return fmt.Errorf("process line %d: %w", lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		recordNumber := (lineNumber / 2) + 1
		if err.Error() == "bufio.Scanner: token too long" {
			actualSize := trackingReader.CurrentLineSize()
			actualSizeMB := float64(actualSize) / (1024 * 1024)
			suggestedSize := int(actualSizeMB) + 5
			if suggestedSize < maxRecordSize*2 {
				suggestedSize = maxRecordSize * 2
			}
			slog.Error("record too large for buffer",
				"line", lineNumber+1,
				"recordNumber", recordNumber,
				"actualSizeBytes", actualSize,
				"actualSizeMB", fmt.Sprintf("%.2f", actualSizeMB),
				"maxRecordSizeMB", maxRecordSize,
				"suggestion", fmt.Sprintf("try --max-record-size %d", suggestedSize),
				"linePreview", trackingReader.LinePreview())
		}
		return fmt.Errorf("scanner error at line %d (record ~%d): %w", lineNumber+1, recordNumber, err)
	}

	// Process deletes for documents that exist in ES but not in source
	for id := range cd.existingDocs {
		cd.metrics.deletesFound.Value.Add(1)
		deleteAction := struct {
			Delete struct {
				Index string `json:"_index"`
				ID    string `json:"_id"`
			} `json:"delete"`
		}{
			Delete: struct {
				Index string `json:"_index"`
				ID    string `json:"_id"`
			}{
				Index: opts.Index,
				ID:    id,
			},
		}

		if err := json.NewEncoder(cd.encoder).Encode(deleteAction); err != nil {
			return fmt.Errorf("encode delete action: %w", err)
		}
	}

	cd.reporter.Complete(cd.changeStats.GetValues())
	return nil
}

func (cd *ChangeDetector) processLine(meta map[string]interface{}, scanner *bufio.Scanner, lineNumber *int) error {
	var operation string
	var indexAction map[string]interface{}

	if action, exists := meta["index"]; exists {
		operation = "index"
		indexAction = action.(map[string]interface{})
	} else if action, exists := meta["delete"]; exists {
		operation = "delete"
		indexAction = action.(map[string]interface{})
	} else {
		return nil
	}

	id := indexAction["_id"].(string)
	index := indexAction["_index"].(string)

	if operation == "delete" {
		cd.metrics.sourceDeletes.Value.Add(1)
		deleteAction := map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": index,
				"_id":    id,
			},
		}
		return json.NewEncoder(cd.encoder).Encode(deleteAction)
	}

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read source document for id=%s: %w", id, err)
		}
		return fmt.Errorf("unexpected EOF: missing source document for id=%s", id)
	}
	*lineNumber++

	var sourceDoc map[string]interface{}
	if err := json.Unmarshal(scanner.Bytes(), &sourceDoc); err != nil {
		return fmt.Errorf("unmarshal source: %w", err)
	}

	// Check if source already has a checksum
	var checksum string
	if existingChecksum, ok := sourceDoc["_checksum"].(string); ok {
		checksum = existingChecksum
	} else {
		// Generate and add checksum if not present
		checksum = calculateChecksum(sourceDoc)
		sourceDoc["_checksum"] = checksum
	}

	existingChecksum, exists := cd.existingDocs[id]
	// Delete from existingDocs to track what's missing in source
	delete(cd.existingDocs, id)

	if !exists {
		cd.metrics.newRecords.Value.Add(1)
	} else if existingChecksum != checksum {
		cd.metrics.updates.Value.Add(1)
	} else {
		cd.metrics.unchanged.Value.Add(1)
		return nil
	}

	// Write index action + source in bulk format
	indexAction = map[string]interface{}{
		"index": map[string]interface{}{
			"_index": index,
			"_id":    id,
		},
	}
	if err := json.NewEncoder(cd.encoder).Encode(indexAction); err != nil {
		return fmt.Errorf("encode index action: %w", err)
	}
	return json.NewEncoder(cd.encoder).Encode(sourceDoc)
}

func convertMapToElasticQuery(queryMap map[string]interface{}) (elastic.Query, error) {
	queryJSON, err := json.Marshal(queryMap)
	if err != nil {
		return nil, fmt.Errorf("marshal query map: %w", err)
	}
	return elastic.RawStringQuery(string(queryJSON)), nil
}
