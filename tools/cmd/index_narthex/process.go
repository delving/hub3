package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
	"github.com/delving/hub3/ikuzo/rdf/index"
	"github.com/delving/hub3/tools/cmd/index_narthex/essync"
	"github.com/klauspost/compress/zstd"
	"github.com/tidwall/sjson"
	"golang.org/x/sync/errgroup"
)

var ErrUnsupportedExtension = fmt.Errorf("unsupported file extension: file must be .rdf or .rdf.zst")

func processPath(path string) (string, string) {
	baseDir := filepath.Dir(path)
	filename := filepath.Base(path)

	prefix, _, found := strings.Cut(filename, "__")
	if !found {
		return baseDir, ""
	}

	return baseDir, prefix
}

// openXMLReader opens either a plain XML file or a zst-compressed XML file
// and returns an io.ReadCloser. It validates the file extension before opening.
func openXMLReader(filename string) (io.ReadCloser, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	baseFile := strings.TrimSuffix(filename, ext)

	switch {
	case ext == ".zst" && (strings.HasSuffix(baseFile, ".xml") || strings.HasSuffix(baseFile, ".rdf")):
		return openZstReader(filename)
	case ext == ".xml":
		return openPlainReader(filename)
	default:
		return nil, fmt.Errorf("%w: got %q", ErrUnsupportedExtension, filepath.Ext(filename))
	}
}

// openPlainReader opens a regular XML file
func openPlainReader(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open XML file: %w", err)
	}
	return file, nil
}

// openZstReader opens a zst-compressed file
func openZstReader(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open zst file: %w", err)
	}

	decoder, err := zstd.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create zst decoder: %w", err)
	}

	return &zstReadCloser{
		Reader:  decoder,
		decoder: decoder,
		file:    file,
	}, nil
}

// zstReadCloser implements io.ReadCloser for zst compressed files
type zstReadCloser struct {
	io.Reader
	decoder *zstd.Decoder
	file    *os.File
}

// Close closes both the decoder and the underlying file
func (z *zstReadCloser) Close() error {
	z.decoder.Close()
	return z.file.Close()
}

type NarthexStats struct {
	Lines     int64
	Records   int64
	Converted int64
	OrgID     string
	DatasetID string
	Errors    []error
	mu        sync.Mutex // Mutex for protecting Errors slice
	writer    *FileWriter
	cfg       Config
}

type recordJob struct {
	hubID  string
	record []byte
}

type Config struct {
	IndexName  string
	DatePrefix string
	OutputPath string
}

var (
	linePrefix = []byte("<!--<")
	lineSuffix = []byte(">-->")
)

func extractID(b []byte) (hubID, hash string, err error) {
	parts := bytes.SplitN(b, []byte("__"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid line separator: %q", string(b))
	}
	hubID = string(bytes.TrimPrefix(parts[0], linePrefix))
	hash = string(bytes.TrimSuffix(parts[1], lineSuffix))
	return hubID, hash, nil
}

func ParseNarthex(r io.Reader, cfg Config) (*NarthexStats, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 5*1024*1024)

	// Create job channel and errgroup
	jobs := make(chan recordJob, 150)
	g, ctx := errgroup.WithContext(context.Background())

	fw, err := NewFileWriter(cfg.OutputPath, cfg.DatePrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to create file writer: %w", err)
	}
	stats := &NarthexStats{
		writer: fw,
		cfg:    cfg,
	}

	// Start 4 worker goroutines
	for i := 0; i < 4; i++ {
		g.Go(func() error {
			return processRecords(ctx, stats, jobs)
		})
	}

	// Start scanner goroutine
	g.Go(func() error {
		defer close(jobs)
		return scanRecords(scanner, stats, jobs)
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return stats, err
	}

	if err := stats.writer.Close(); err != nil {
		return nil, fmt.Errorf("unable to close stats writer; %w", err)
	}

	return stats, nil
}

func scanRecords(scanner *bufio.Scanner, stats *NarthexStats, jobs chan<- recordJob) error {
	var record bytes.Buffer

	for scanner.Scan() {
		b := scanner.Bytes()
		if bytes.HasPrefix(b, linePrefix) {
			atomic.AddInt64(&stats.Records, 1)
			if atomic.LoadInt64(&stats.Records)%5000 == 0 {
				slog.Info("progress update",
					"records", atomic.LoadInt64(&stats.Records),
					"lines", atomic.LoadInt64(&stats.Lines),
					"converted", atomic.LoadInt64(&stats.Converted),
					"errors", len(stats.Errors))
			}

			hubID, _, err := extractID(b)
			if err != nil {
				return err
			}

			// Send the record to workers
			recordCopy := make([]byte, record.Len())
			copy(recordCopy, record.Bytes())

			select {
			case jobs <- recordJob{hubID: hubID, record: recordCopy}:
			default:
				// If channel is full, process inline
				if err := processRecord(stats, hubID, record.Bytes()); err != nil {
					stats.mu.Lock()
					slog.Error("unable to process record", "hubID", hubID, "error", err)
					stats.Errors = append(stats.Errors, err)
					stats.mu.Unlock()
				}
			}

			record.Reset()
			continue
		}

		if _, err := record.Write(b); err != nil {
			return err
		}
		atomic.AddInt64(&stats.Lines, 1)
	}

	return scanner.Err()
}

func processRecords(ctx context.Context, stats *NarthexStats, jobs <-chan recordJob) error {
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return nil
			}
			if err := processRecord(stats, job.hubID, job.record); err != nil {
				stats.mu.Lock()
				slog.Error("unable to process record", "hubID", job.hubID, "error", err)
				stats.Errors = append(stats.Errors, err)
				stats.mu.Unlock()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func fixRDFHeader(data []byte) []byte {
	pattern := `([<\/])\w+:(RDF)`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAll(data, []byte(`${1}rdf:$2`))
}

func processRecord(stats *NarthexStats, hubID string, recordData []byte) error {
	recordData = fixRDFHeader(recordData)
	g, err := rdfxml.Parse(bytes.NewReader(recordData), nil, hubID)
	if err != nil {
		slog.Error("unable to parse record",
			"hubID", hubID,
			"error", err,
			"data", string(recordData))
		return err
	}
	g.GraphName = hubID
	atomic.AddInt64(&stats.Converted, 1)

	if err := stats.writer.WriteNTriples(g); err != nil {
		return fmt.Errorf("write ntriples: %w", err)
	}

	subjectURI := g.SubjectSafe()

	header, err := createHeader(hubID, subjectURI.RawValue())
	if err != nil {
		return fmt.Errorf("unable to create header: %w", err)
	}

	fg, err := index.NewGraph(header)
	if err != nil {
		return fmt.Errorf("unable to create graph: %w", err)
	}

	if err := fg.AddGraph(g); err != nil {
		return fmt.Errorf("unable to add graph: %w", err)
	}

	mesg, err := fg.IndexMessage()
	if err != nil {
		return fmt.Errorf("unable to create index message: %w", err)
	}
	mesg.IndexName = stats.cfg.IndexName

	rec, err := toBulkRecord(mesg)
	if err != nil {
		return fmt.Errorf("unable to create bulk record: %w", err)
	}

	if err := stats.writer.WriteBulk(rec); err != nil {
		return fmt.Errorf("write bulk: %w", err)
	}

	return nil
}

type SourceStats struct {
	Records uint64
}

func createHeader(hubID, subject string) (index.Header, error) {
	s := strings.TrimSuffix(strings.TrimPrefix(hubID, "urn:"), "/graph")
	id, err := domain.NewHubID(s)
	if err != nil {
		return index.Header{}, fmt.Errorf("unable to create hubID: %w", err)
	}

	header := index.Header{
		OrgID:    id.OrgID().String(),
		Spec:     id.DatasetID.ID,
		EntryURI: subject,
		Tags:     []string{"narthex"},
		HubID:    id.String(),
	}

	return header, nil
}

func toBulkRecord(msg *domainpb.IndexMessage) ([]byte, error) {
	var buf bytes.Buffer

	// Determine operation type
	operation := "index"
	if msg.Deleted {
		operation = "delete"
	}

	// Create action line
	action := map[string]interface{}{
		operation: map[string]interface{}{
			"_index": msg.IndexName,
			"_id":    msg.RecordID,
		},
	}

	if err := json.NewEncoder(&buf).Encode(action); err != nil {
		return nil, fmt.Errorf("encoding action: %w", err)
	}

	// For delete operations, we don't need a source document
	if msg.Deleted {
		return buf.Bytes(), nil
	}

	source, err := setChecksum(msg.Source)
	if err != nil {
		return nil, fmt.Errorf("setting checksum: %w", err)
	}

	_, err = buf.Write(source)
	if err != nil {
		return nil, fmt.Errorf("writing source: %w", err)
	}
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

func setChecksum(data []byte) ([]byte, error) {
	source, err := sjson.DeleteBytes(data, "_checksum")
	if err != nil {
		return []byte{}, fmt.Errorf("unable to remove existing checksum: %w", err)
	}

	source, err = sjson.DeleteBytes(source, "meta.modified")
	if err != nil {
		return []byte{}, fmt.Errorf("unable to remove modified date: %w", err)
	}

	checksum := essync.Checksum(source)
	data, err = sjson.SetBytes(data, "_checksum", checksum)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to set checksum: %w", err)
	}

	return data, nil
}

func calculateChecksum(data map[string]interface{}) string {
	// Remove existing checksum if present
	delete(data, "_checksum")
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}
