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
	"sort"
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

// findProcessableDirectories discovers all subdirectories containing processed RDF files
func findProcessableDirectories(parentPath string) ([]string, error) {
	return findProcessableDirectoriesWithPattern(parentPath, "")
}

// findProcessableDirectoriesWithPattern discovers subdirectories with a custom or default pattern
func findProcessableDirectoriesWithPattern(parentPath, customPattern string) ([]string, error) {
	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read parent directory %s: %w", parentPath, err)
	}

	var directories []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(parentPath, entry.Name())

		// Check if this directory contains processed RDF files using all patterns
		_, err := findProcessedFileWithPattern(dirPath, customPattern)
		if err != nil {
			slog.Debug("no processed files found in directory", "dir", dirPath, "error", err)
			continue
		}

		// If we found processed files, add this directory to the list
		directories = append(directories, dirPath)
	}

	// Sort directories for consistent processing order
	sort.Strings(directories)

	if len(directories) == 0 {
		return nil, fmt.Errorf("no subdirectories with processed RDF files found in %s", parentPath)
	}

	return directories, nil
}

// findProcessedFile tries multiple patterns to find a processed RDF file in a directory
func findProcessedFile(dirPath string) (string, error) {
	return findProcessedFileWithPattern(dirPath, "")
}

// findProcessedFileWithPattern tries to find a processed RDF file using custom or default patterns
func findProcessedFileWithPattern(dirPath, customPattern string) (string, error) {
	var patterns []string

	if customPattern != "" {
		// Use only the custom pattern if provided
		patterns = []string{customPattern}
	} else {
		// Use default patterns - ordered from most specific to most general
		patterns = []string{
			"__processed_",    // Matches __processed_<model>.rdf.zst (your pattern)
			"__processed.rdf", // Original exact pattern
			"_processed_",     // Alternative with single underscore
			"processed_",      // Without leading underscores
			".rdf",            // Any .rdf file
			"_processed.rdf",
			"processed.rdf",
			"__processed.xml",
			"_processed.xml",
			"processed.xml",
		}
	}

	var allErrors []string
	for _, pattern := range patterns {
		rdfPath, err := essync.FindLatestPath(dirPath, pattern)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("pattern '%s': %v", pattern, err))
			continue
		}
		if rdfPath != "" {
			slog.Debug("found processed file", "dir", dirPath, "file", rdfPath, "pattern", pattern)
			return rdfPath, nil
		}
		allErrors = append(allErrors, fmt.Sprintf("pattern '%s': no matching files", pattern))
	}

	patternDesc := "default patterns"
	if customPattern != "" {
		patternDesc = fmt.Sprintf("custom pattern '%s'", customPattern)
	}

	return "", fmt.Errorf("no processed files found in %s using %s. Details: %s", dirPath, patternDesc, strings.Join(allErrors, "; "))
}

// processSingleDirectory processes a single directory containing RDF files
// This is kept for backward compatibility
func processSingleDirectory(dirPath, targetIndex, customPattern string) (*NarthexStats, error) {
	return processSingleDirectoryWithTagMap(dirPath, targetIndex, customPattern, nil, 0) // 0 uses default 5MB
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
	IndexName     string
	DatePrefix    string
	OutputPath    string
	TagMap        *index.TagMap
	MaxRecordSize int // max record size in MB (default: 5)
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

	// Use configurable max record size (default 5MB)
	maxRecordSize := cfg.MaxRecordSize
	if maxRecordSize <= 0 {
		maxRecordSize = 5
	}
	scanner.Buffer(buf, maxRecordSize*1024*1024)

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
	for range 4 {
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

	// Print newline after progress updates to move to next line
	fmt.Println()

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
				fmt.Printf("\rProgress: records=%d lines=%d converted=%d errors=%d",
					atomic.LoadInt64(&stats.Records),
					atomic.LoadInt64(&stats.Lines),
					atomic.LoadInt64(&stats.Converted),
					len(stats.Errors))
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

	// Add debugging info about the RDF graph
	slog.Debug("RDF graph info before index conversion",
		"hubID", hubID,
		"subjectURI", subjectURI.RawValue(),
		"triples", len(g.Triples()),
		"resources", len(g.Resources()),
		"graphName", g.GraphName)

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

	// Debug the graph resources before IndexMessage
	entryURIExists := resourceExists(fg, fg.Header.EntryURI)
	slog.Debug("Index Graph info before IndexMessage",
		"resources_count", len(fg.Resources),
		"entryURI", fg.Header.EntryURI,
		"entryURI_in_resources", entryURIExists)

	// If EntryURI doesn't exist in resources, try to find alternative subject
	if !entryURIExists && len(fg.Resources) > 0 {
		slog.Warn("EntryURI not found in graph resources, attempting to fix",
			"entryURI", fg.Header.EntryURI,
			"resources_count", len(fg.Resources))

		// Use the subject URI from the RDF graph as EntryURI if available
		if subjectURI.RawValue() != "" && resourceExists(fg, subjectURI.RawValue()) {
			slog.Info("Using subjectURI as EntryURI",
				"original", fg.Header.EntryURI,
				"new", subjectURI.RawValue())
			fg.Header.EntryURI = subjectURI.RawValue()
		} else if len(fg.Resources) > 0 {
			// Use the first resource as EntryURI if no better option
			newEntryURI := fg.Resources[0].ID
			slog.Info("Using first resource as EntryURI",
				"original", fg.Header.EntryURI,
				"new", newEntryURI)
			fg.Header.EntryURI = newEntryURI
		}
	}

	// Use the TagMap if provided
	mesg, err := fg.IndexMessage(stats.cfg.TagMap)
	if err != nil {
		slog.Error("Error creating index message", "error", err, "entryURI", fg.Header.EntryURI)
		return fmt.Errorf("unable to create index message: %w", err)
	}

	// Debug the graph resources after IndexMessage
	// Check if we can detect context references in resources
	hasContextRefs := false
	contextCount := 0
	hasExternalContextRefs := false
	externalContextCount := 0

	// Check if context is set in the graph
	slog.Debug("Graph context status", "contextIsSet", fg.ContextIsSet())

	// Print details of the first few resources
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("Detailed resource information after IndexMessage:")

		// Examine the message source
		sourceBytes := mesg.Source
		sampleSize := 256 // Show first 256 bytes as sample
		if len(sourceBytes) < sampleSize {
			sampleSize = len(sourceBytes)
		}
		slog.Debug("Source bytes sample", "sample", string(sourceBytes[:sampleSize]))

		// Try to understand the actual structure
		var data interface{}
		if err := json.Unmarshal(sourceBytes, &data); err != nil {
			slog.Error("Failed to unmarshal Source as JSON (any type)", "error", err)
		} else {
			// Determine the actual type of the data
			switch v := data.(type) {
			case map[string]interface{}:
				// It's an object as expected
				keys := make([]string, 0, len(v))
				for k := range v {
					keys = append(keys, k)
				}
				slog.Debug("Source is a JSON object", "keys", keys)

				// Check for resources field
				if resources, ok := v["resources"]; ok {
					resourcesType := fmt.Sprintf("%T", resources)
					slog.Debug("Resources field found", "type", resourcesType)

					// If resources is an array, examine the first item
					if resourcesArr, ok := resources.([]interface{}); ok && len(resourcesArr) > 0 {
						slog.Debug("Resources is an array", "count", len(resourcesArr))

						if firstResource, ok := resourcesArr[0].(map[string]interface{}); ok {
							// Get resource keys
							resourceKeys := make([]string, 0, len(firstResource))
							for k := range firstResource {
								resourceKeys = append(resourceKeys, k)
							}
							slog.Debug("First resource", "keys", resourceKeys)

							// Check for context and external context
							examineContextField(firstResource, "context")
							examineContextField(firstResource, "graphExternalContext")
						} else {
							slog.Debug("First resource is not an object",
								"type", fmt.Sprintf("%T", resourcesArr[0]))
						}
					}
				}
			case []interface{}:
				// It's an array
				slog.Debug("Source is a JSON array", "length", len(v))
				if len(v) > 0 {
					slog.Debug("First array element type", "type", fmt.Sprintf("%T", v[0]))
				}
			case string:
				// It's a string, which might be the issue
				slog.Debug("Source is a JSON string", "length", len(v))

				// Try to parse the string as JSON
				var nestedData interface{}
				if err := json.Unmarshal([]byte(v), &nestedData); err != nil {
					slog.Debug("String is not valid JSON", "error", err)
				} else {
					slog.Debug("String contains valid JSON", "type", fmt.Sprintf("%T", nestedData))
				}
			default:
				slog.Debug("Source is an unexpected JSON type", "type", fmt.Sprintf("%T", data))
			}
		}
	}

	// Limit to first 3 resources to avoid flooding logs
	resourcesLimit := 3
	for i, r := range fg.Resources {
		if i >= resourcesLimit {
			break
		}

		// Check if context references have correct levels
		contextLevels := []int32{}
		for _, ctx := range r.Context {
			contextLevels = append(contextLevels, ctx.Level)
		}

		// Check if external context references have correct levels
		externalLevels := []int32{}
		for _, ctx := range r.GraphExternalContext {
			externalLevels = append(externalLevels, ctx.Level)
		}

		slog.Debug(fmt.Sprintf("Resource #%d", i),
			"id", r.ID,
			"types", r.Types,
			"entries_count", len(r.Entries),
			"context_count", len(r.Context),
			"context_levels", contextLevels,
			"external_context_count", len(r.GraphExternalContext),
			"external_context_levels", externalLevels)
	}

	for _, r := range fg.Resources {
		if len(r.Context) > 0 {
			hasContextRefs = true
			contextCount += len(r.Context)
		}

		if len(r.GraphExternalContext) > 0 {
			hasExternalContextRefs = true
			externalContextCount += len(r.GraphExternalContext)
		}
	}

	// Look for resource type entries and external contexts
	resourceEntryCount := 0
	resourcesWithContext := 0
	resourcesWithExternalContext := 0

	for _, r := range fg.Resources {
		for _, e := range r.Entries {
			if e.EntryType == index.ResourceType || e.EntryType == index.Bnode {
				resourceEntryCount++
			}
		}

		if len(r.Context) > 0 {
			resourcesWithContext++
		}

		if len(r.GraphExternalContext) > 0 {
			resourcesWithExternalContext++
		}
	}

	slog.Debug("Index Graph after IndexMessage",
		"context_refs_found", hasContextRefs,
		"context_count", contextCount,
		"resources_with_context", resourcesWithContext,
		"external_context_refs_found", hasExternalContextRefs,
		"external_context_count", externalContextCount,
		"resources_with_external_context", resourcesWithExternalContext,
		"resource_entry_count", resourceEntryCount,
		"resource_count", len(fg.Resources))
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

// Helper function to check if a resource with a given ID exists in the graph
func resourceExists(g *index.Graph, id string) bool {
	for _, r := range g.Resources {
		if r.ID == id {
			return true
		}
	}
	return false
}

// Helper function to get keys from a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to examine context fields in resources
func examineContextField(resource map[string]interface{}, fieldName string) {
	if field, exists := resource[fieldName]; exists {
		fieldType := fmt.Sprintf("%T", field)
		slog.Debug(fmt.Sprintf("Resource %s field found", fieldName), "type", fieldType)

		if contextArr, ok := field.([]interface{}); ok {
			slog.Debug(fmt.Sprintf("Resource %s is an array", fieldName), "count", len(contextArr))

			// Examine first context item if available
			if len(contextArr) > 0 {
				contextItemType := fmt.Sprintf("%T", contextArr[0])
				slog.Debug(fmt.Sprintf("First %s item type", fieldName), "type", contextItemType)

				// If it's an object, show its keys
				if contextItem, ok := contextArr[0].(map[string]interface{}); ok {
					contextKeys := make([]string, 0, len(contextItem))
					for k := range contextItem {
						contextKeys = append(contextKeys, k)
					}
					slog.Debug(fmt.Sprintf("First %s item keys", fieldName), "keys", contextKeys)
				}
			}
		} else {
			slog.Debug(fmt.Sprintf("Resource %s is not an array", fieldName))
		}
	} else {
		slog.Debug(fmt.Sprintf("Resource %s field not found", fieldName))
	}
}

