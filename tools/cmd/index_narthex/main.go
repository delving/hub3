package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/ikuzo/rdf/index"
	"github.com/delving/hub3/tools/cmd/index_narthex/essync"
	"github.com/delving/hub3/tools/cmd/index_narthex/stats"
	"github.com/klauspost/compress/zstd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hub3-cli",
	Short: "CLI tool for Hub3 data processing",
}

// Existing command variables
var (
	processCmd = &cobra.Command{
		Use:   "process",
		Short: "Process RDF/XML files",
		RunE:  runProcess,
	}
	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download and process SIP files",
		RunE:  runDownload,
	}
	// essync commands
	detectCmd = &cobra.Command{
		Use:   "detect",
		Short: "Detect changes between source and Elasticsearch",
		RunE:  runDetect,
	}
	submitCmd = &cobra.Command{
		Use:   "submit",
		Short: "Submit detected changes to Elasticsearch",
		RunE:  runSubmit,
	}
	analyzeCmd = &cobra.Command{
		Use:   "analyze",
		Short: "Analyze downloaded datasets to detect RDF-XML data",
		RunE:  runAnalyze,
	}
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync data from external v2 API to local Elasticsearch",
		RunE:  runSync,
	}
)

type processFlags struct {
	path        string
	targetIndex string
	all         bool
	pattern     string
	tagMapFile  string
	debug       bool
}

type downloadFlags struct {
	url      string
	username string
	password string
	output   string
	orgID    string
}

type detectFlags struct {
	inputFile     string
	outputFile    string
	esURL         string
	indexName     string
	queryString   string
	fnameFilter   string
	statsInterval time.Duration
	all           bool
}

type submitFlags struct {
	changesFile   string
	esURL         string
	workers       int
	batchSize     int
	statsInterval time.Duration
	all           bool
}

type analyzeFlags struct {
	path       string
	all        bool
	jsonOutput string
	rdfOnly    bool
}

type syncFlags struct {
	sourceURL     string
	sourceOrgID   string
	targetOrgID   string
	esURL         string
	indexName     string
	query         string
	batchSize     int
	workers       int
	statsInterval time.Duration
	dryRun        bool
}


var (
	pFlags   processFlags
	dFlags   downloadFlags
	detFlags detectFlags
	subFlags submitFlags
	aFlags   analyzeFlags
	sFlags   syncFlags
)

func init() {
	// process command flags
	processCmd.Flags().StringVar(&pFlags.path, "path", "", "directory that contains the narthex rdf-xml files (or parent directory when using --all)")
	processCmd.Flags().StringVar(&pFlags.targetIndex, "index", "", "elasticsearch index to write to")
	processCmd.Flags().BoolVar(&pFlags.all, "all", false, "process all subdirectories in the specified path")
	processCmd.Flags().StringVar(&pFlags.pattern, "pattern", "", "custom pattern to search for (overrides default patterns)")
	processCmd.Flags().StringVar(&pFlags.tagMapFile, "tagmap", "", "path to TOML file containing RDF tag mappings")
	processCmd.Flags().BoolVar(&pFlags.debug, "debug", false, "enable debug logging")
	processCmd.MarkFlagRequired("path")
	processCmd.MarkFlagRequired("index")

	// download command flags
	downloadCmd.Flags().StringVar(&dFlags.url, "url", "", "URL of the XML file")
	downloadCmd.Flags().StringVar(&dFlags.username, "username", "", "Basic auth username")
	downloadCmd.Flags().StringVar(&dFlags.password, "password", "", "Basic auth password")
	downloadCmd.Flags().StringVar(&dFlags.output, "output", "", "Output directory for processed files")
	downloadCmd.Flags().StringVar(&dFlags.orgID, "orgID", "", "The default orgID to replace empty mapping orgID entries")
	downloadCmd.MarkFlagRequired("url")
	downloadCmd.MarkFlagRequired("output")

	// detect command flags
	detectCmd.Flags().StringVarP(&detFlags.inputFile, "input", "i", "", "Input file path (or parent directory when using --all) (required)")
	detectCmd.Flags().StringVarP(&detFlags.outputFile, "output", "o", "", "Output file path (base name when using --all) (required)")
	detectCmd.Flags().StringVarP(&detFlags.esURL, "url", "u", "http://localhost:9200", "Elasticsearch URL")
	detectCmd.Flags().StringVarP(&detFlags.indexName, "index", "n", "", "Index name (required)")
	detectCmd.Flags().StringVarP(&detFlags.queryString, "query", "q", "", "Query JSON string")
	detectCmd.Flags().StringVarP(&detFlags.fnameFilter, "fnameFilter", "", "", "Filter for lastest containing substring in input file path is a directory")
	detectCmd.Flags().DurationVar(&detFlags.statsInterval, "stats-interval", 1*time.Second, "Statistics update interval")
	detectCmd.Flags().BoolVar(&detFlags.all, "all", false, "detect changes for all subdirectories in the specified input path")
	detectCmd.MarkFlagRequired("input")
	detectCmd.MarkFlagRequired("output")
	detectCmd.MarkFlagRequired("index")

	// submit command flags
	submitCmd.Flags().StringVarP(&subFlags.changesFile, "changes", "c", "", "Changes file path (or parent directory when using --all) (required)")
	submitCmd.Flags().StringVarP(&subFlags.esURL, "url", "u", "http://localhost:9200", "Elasticsearch URL")
	submitCmd.Flags().IntVarP(&subFlags.workers, "workers", "w", 4, "Number of worker goroutines")
	submitCmd.Flags().IntVar(&subFlags.batchSize, "batch-size", 1000, "Maximum batch size for bulk requests")
	submitCmd.Flags().DurationVar(&subFlags.statsInterval, "stats-interval", 1*time.Second, "Statistics update interval")
	submitCmd.Flags().BoolVar(&subFlags.all, "all", false, "submit all changes files found in subdirectories of the specified path")
	submitCmd.MarkFlagRequired("changes")

	// analyze command flags
	analyzeCmd.Flags().StringVar(&aFlags.path, "path", "", "directory that contains downloaded datasets (or parent directory when using --all)")
	analyzeCmd.Flags().BoolVar(&aFlags.all, "all", false, "analyze all subdirectories in the specified path")
	analyzeCmd.Flags().StringVar(&aFlags.jsonOutput, "json", "", "output JSON file with first 100 lines of each dataset for inspection")
	analyzeCmd.Flags().BoolVar(&aFlags.rdfOnly, "rdf-only", false, "when used with --json, only include RDF datasets in the output (useful for false positive inspection)")
	analyzeCmd.MarkFlagRequired("path")

	// sync command flags
	syncCmd.Flags().StringVar(&sFlags.sourceURL, "url", "", "Source v2 API URL (e.g., https://data.hub3.org/api/search/v2) (required)")
	syncCmd.Flags().StringVar(&sFlags.sourceOrgID, "source-org", "", "Source organization ID (required)")
	syncCmd.Flags().StringVar(&sFlags.targetOrgID, "target-org", "", "Target organization ID (defaults to source-org)")
	syncCmd.Flags().StringVar(&sFlags.esURL, "es-url", "http://localhost:9200", "Target Elasticsearch URL")
	syncCmd.Flags().StringVar(&sFlags.indexName, "index", "", "Target index name (required)")
	syncCmd.Flags().StringVar(&sFlags.query, "query", "*", "V2 query string (e.g., 'meta.spec:dataset-name')")
	syncCmd.Flags().IntVar(&sFlags.batchSize, "batch-size", 100, "Records per page")
	syncCmd.Flags().IntVar(&sFlags.workers, "workers", 4, "Bulk indexing workers")
	syncCmd.Flags().DurationVar(&sFlags.statsInterval, "stats-interval", 1*time.Second, "Statistics update interval")
	syncCmd.Flags().BoolVar(&sFlags.dryRun, "dry-run", false, "Fetch data but don't index (for testing)")
	syncCmd.MarkFlagRequired("url")
	syncCmd.MarkFlagRequired("source-org")
	syncCmd.MarkFlagRequired("index")

	// Add all commands to root
	rootCmd.AddCommand(processCmd, downloadCmd, detectCmd, submitCmd, analyzeCmd, syncCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	if detFlags.all {
		return runDetectAll()
	}
	return runDetectSingle()
}

func runDetectSingle() error {
	return detectChangesForPath(detFlags.inputFile, detFlags.outputFile)
}

func runDetectAll() error {
	directories, err := findProcessableDirectories(detFlags.inputFile)
	if err != nil {
		return err
	}

	slog.Info("found directories to detect changes", "count", len(directories), "parentPath", detFlags.inputFile)

	// Create a shared Elasticsearch client to avoid expvar collision
	cfg := elasticsearch.Config{
		Urls: []string{detFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	successCount := 0
	for i, dirPath := range directories {
		dirName := filepath.Base(dirPath)
		outputFile := generateOutputFileName(detFlags.outputFile, dirName)
		
		slog.Info("detecting changes for directory", "progress", fmt.Sprintf("%d/%d", i+1, len(directories)), "dir", dirPath, "output", outputFile)
		
		if err := detectChangesForPathWithClient(client, dirPath, outputFile); err != nil {
			slog.Error("failed to detect changes for directory", "dir", dirPath, "error", err)
			continue
		}
		
		successCount++
		slog.Info("completed change detection", "dir", dirPath, "output", outputFile)
	}

	slog.Info("finished detecting changes for all directories",
		"successfulDirs", successCount,
		"totalDirs", len(directories))

	return nil
}

func detectChangesForPath(inputPath, outputPath string) error {
	cfg := elasticsearch.Config{
		Urls: []string{detFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	return detectChangesForPathWithClient(client, inputPath, outputPath)
}

func detectChangesForPathWithClient(client *elasticsearch.Client, inputPath, outputPath string) error {
	// Setup reporters
	logger := slog.Default()
	reporter := stats.NewMultiReporter(
		stats.NewSlogReporter(logger, detFlags.statsInterval),
		stats.NewProgressReporter(detFlags.statsInterval),
	)

	detector, err := essync.NewChangeDetector(outputPath, reporter)
	if err != nil {
		return fmt.Errorf("error creating detector: %w", err)
	}
	defer detector.Close()

	opts := essync.DefaultLoadOptions(detFlags.indexName)
	if detFlags.queryString != "" {
		var query map[string]interface{}
		if err := json.Unmarshal([]byte(detFlags.queryString), &query); err != nil {
			return fmt.Errorf("invalid query JSON: %w", err)
		}
		opts.Query = query
	}
	opts.FnameFilter = detFlags.fnameFilter

	if err := detector.LoadExistingDocs(client, opts); err != nil {
		return fmt.Errorf("error loading existing docs: %w", err)
	}

	if err := detector.DetectChanges(inputPath, opts); err != nil {
		return fmt.Errorf("error detecting changes: %w", err)
	}

	return nil
}

func generateOutputFileName(baseName, dirName string) string {
	ext := filepath.Ext(baseName)
	
	// Generate timestamp in format YYYY-MM-DDTHHMISS
	timestamp := time.Now().Format("2006-01-02T150405")
	
	// Use consistent naming: timestamp__index.jsonl.zst
	if strings.HasSuffix(baseName, ".zst") {
		return fmt.Sprintf("%s__%s__index.jsonl.zst", timestamp, dirName)
	} else if strings.HasSuffix(baseName, ".jsonl") {
		return fmt.Sprintf("%s__%s__index.jsonl", timestamp, dirName)
	} else {
		// Fallback to old naming if extension is unexpected
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		return fmt.Sprintf("%s_%s%s", nameWithoutExt, dirName, ext)
	}
}

func runProcess(cmd *cobra.Command, args []string) error {
	if pFlags.all {
		return runProcessAll()
	}
	return runProcessSingle()
}

func runProcessSingle() error {
	// Set debug logging if requested
	if pFlags.debug {
		logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(logHandler)
		slog.SetDefault(logger)
		slog.Debug("Debug logging enabled")
	}

	// Load tag map if provided
	var tagMap *index.TagMap
	var err error
	
	tagMapFile := pFlags.tagMapFile
	if tagMapFile == "" {
		// Try default locations
		defaultLocations := []string{
			"hub3.toml",
			"./config/hub3.toml",
		}
		
		for _, loc := range defaultLocations {
			if _, err := os.Stat(loc); err == nil {
				tagMapFile = loc
				slog.Info("using default tag map file", "path", tagMapFile)
				break
			}
		}
	}
	
	if tagMapFile != "" {
		tagMap, err = LoadRDFTagMapFromTOML(tagMapFile)
		if err != nil {
			return fmt.Errorf("failed to load tag map: %w", err)
		}
		slog.Info("loaded tag map", "file", tagMapFile, "entries", tagMap.Len())
	}
	
	stats, err := processSingleDirectoryWithTagMap(pFlags.path, pFlags.targetIndex, pFlags.pattern, tagMap)
	if err != nil {
		return err
	}

	slog.Info("finished processing narthex file",
		"records", stats.Records,
		"lines", stats.Lines,
		"converted", stats.Converted)
	return nil
}

func runProcessAll() error {
	// Set debug logging if requested
	if pFlags.debug {
		logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(logHandler)
		slog.SetDefault(logger)
		slog.Debug("Debug logging enabled")
	}

	directories, err := findProcessableDirectoriesWithPattern(pFlags.path, pFlags.pattern)
	if err != nil {
		return err
	}

	slog.Info("found directories to process", "count", len(directories), "parentPath", pFlags.path)

	// Load tag map if provided
	var tagMap *index.TagMap
	
	tagMapFile := pFlags.tagMapFile
	if tagMapFile == "" {
		// Try default locations
		defaultLocations := []string{
			"hub3.toml",
			"./config/hub3.toml",
		}
		
		for _, loc := range defaultLocations {
			if _, err := os.Stat(loc); err == nil {
				tagMapFile = loc
				slog.Info("using default tag map file", "path", tagMapFile)
				break
			}
		}
	}
	
	if tagMapFile != "" {
		tagMap, err = LoadRDFTagMapFromTOML(tagMapFile)
		if err != nil {
			return fmt.Errorf("failed to load tag map: %w", err)
		}
		slog.Info("loaded tag map", "file", tagMapFile, "entries", tagMap.Len())
	}

	var totalStats NarthexStats
	var processedCount int

	for i, dirPath := range directories {
		slog.Info("processing directory", "progress", fmt.Sprintf("%d/%d", i+1, len(directories)), "dir", dirPath)
		
		stats, err := processSingleDirectoryWithTagMap(dirPath, pFlags.targetIndex, pFlags.pattern, tagMap)
		if err != nil {
			slog.Error("failed to process directory", "dir", dirPath, "error", err)
			continue
		}

		// Aggregate statistics
		totalStats.Records += stats.Records
		totalStats.Lines += stats.Lines
		totalStats.Converted += stats.Converted
		totalStats.Errors = append(totalStats.Errors, stats.Errors...)
		processedCount++

		slog.Info("completed directory", 
			"dir", dirPath,
			"records", stats.Records,
			"lines", stats.Lines,
			"converted", stats.Converted,
			"errors", len(stats.Errors))
	}

	slog.Info("finished processing all directories",
		"processedDirs", processedCount,
		"totalDirs", len(directories),
		"totalRecords", totalStats.Records,
		"totalLines", totalStats.Lines,
		"totalConverted", totalStats.Converted,
		"totalErrors", len(totalStats.Errors))

	return nil
}

func runDownload(cmd *cobra.Command, args []string) error {
	config := HarvestConfig{
		URL:      dFlags.url,
		Username: dFlags.username,
		Password: dFlags.password,
		Output:   dFlags.output,
		OrgID:    dFlags.orgID,
	}
	return ProcessAllFiles(config)
}

func runSubmit(cmd *cobra.Command, args []string) error {
	if subFlags.all {
		return runSubmitAll()
	}
	return runSubmitSingle()
}

func runSubmitSingle() error {
	return submitChangesFile(subFlags.changesFile)
}

func runSubmitAll() error {
	changesFiles, err := findChangesFiles(subFlags.changesFile)
	if err != nil {
		return err
	}

	slog.Info("found changes files to submit", "count", len(changesFiles), "parentPath", subFlags.changesFile)

	// Create a shared Elasticsearch client to avoid expvar collision
	cfg := elasticsearch.Config{
		Urls: []string{subFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	successCount := 0
	for i, changesFile := range changesFiles {
		// Print progress on a single line
		fmt.Printf("\rSubmitting changes files: %d/%d | Current: %s", i+1, len(changesFiles), filepath.Base(changesFile))
		
		if err := submitChangesFileWithClientAndReporter(client, changesFile, false); err != nil {
			fmt.Printf("\n") // Move to new line for error
			slog.Error("failed to submit changes file", "file", changesFile, "error", err)
			continue
		}
		
		successCount++
	}
	
	// Move to new line and print final result
	fmt.Printf("\n")

	slog.Info("finished submitting all changes files",
		"successfulFiles", successCount,
		"totalFiles", len(changesFiles))

	return nil
}

func submitChangesFile(changesFilePath string) error {
	cfg := elasticsearch.Config{
		Urls: []string{subFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	return submitChangesFileWithClient(client, changesFilePath)
}

func submitChangesFileWithClient(client *elasticsearch.Client, changesFilePath string) error {
	return submitChangesFileWithClientAndReporter(client, changesFilePath, true)
}

func submitChangesFileWithClientAndReporter(client *elasticsearch.Client, changesFilePath string, useSlogReporter bool) error {
	var reader io.ReadCloser
	file, err := os.Open(changesFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if strings.HasSuffix(changesFilePath, ".zst") {
		decoder, err := zstd.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		defer decoder.Close()
		reader = io.NopCloser(decoder)
	} else {
		reader = file
	}

	// Create reporters based on context
	var reporters stats.Reporter
	if useSlogReporter {
		reporters = stats.NewMultiReporter(
			stats.NewSlogReporter(slog.Default(), subFlags.statsInterval),
			stats.NewProgressReporter(subFlags.statsInterval),
		)
	} else {
		// When processing multiple files, only use ProgressReporter to avoid interfering with batch progress
		reporters = stats.NewProgressReporter(subFlags.statsInterval)
	}

	processor := essync.NewBulkProcessor(essync.Config{
		NumWorkers:    subFlags.workers,
		MaxBatchSize:  subFlags.batchSize,
		UpdateTimeout: subFlags.statsInterval,
		ESClient:      client,
		Reporter:      reporters,
	})

	ctx := context.Background()
	if err := processor.Process(ctx, reader); err != nil {
		return fmt.Errorf("bulk processing failed: %w", err)
	}

	return nil
}

// isIndexOrChangesFile checks if a filename matches index or changes file patterns
func isIndexOrChangesFile(fileName string) bool {
	// Check for timestamp__index pattern (e.g., 2025-05-21T232728__index.jsonl.zst)
	if strings.Contains(fileName, "__index") && (strings.HasSuffix(fileName, ".jsonl") || strings.HasSuffix(fileName, ".jsonl.zst") || strings.HasSuffix(fileName, ".zst")) {
		return true
	}
	// Check for legacy changes pattern
	if strings.Contains(fileName, "changes") && (strings.HasSuffix(fileName, ".jsonl") || strings.HasSuffix(fileName, ".jsonl.zst") || strings.HasSuffix(fileName, ".zst")) {
		return true
	}
	return false
}

// findLatestIndexFile finds the most recent index file in a directory based on timestamp
func findLatestIndexFile(dirPath string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		slog.Debug("error reading directory for latest index file", "dir", dirPath, "error", err)
		return ""
	}

	var latestFile string
	var latestTime string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		fileName := entry.Name()
		if !isIndexOrChangesFile(fileName) {
			continue
		}

		// Extract timestamp from filename (format: YYYY-MM-DDTHHMISS__)
		if strings.Contains(fileName, "__index") {
			parts := strings.Split(fileName, "__")
			if len(parts) >= 2 {
				timestamp := parts[0]
				// Compare timestamps lexicographically (works for ISO format)
				if timestamp > latestTime {
					latestTime = timestamp
					latestFile = filepath.Join(dirPath, fileName)
				}
			}
		} else {
			// For legacy changes files, just take the first one found
			if latestFile == "" {
				latestFile = filepath.Join(dirPath, fileName)
			}
		}
	}

	return latestFile
}

func findChangesFiles(parentPath string) ([]string, error) {
	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read parent directory %s: %w", parentPath, err)
	}

	var changesFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Look for index/changes files in subdirectories and find the latest one
			subDirPath := filepath.Join(parentPath, entry.Name())
			latestFile := findLatestIndexFile(subDirPath)
			if latestFile != "" {
				changesFiles = append(changesFiles, latestFile)
			}
		} else {
			// Look for index/changes files directly in the parent directory
			fileName := entry.Name()
			if isIndexOrChangesFile(fileName) {
				changesFiles = append(changesFiles, filepath.Join(parentPath, fileName))
			}
		}
	}

	if len(changesFiles) == 0 {
		return nil, fmt.Errorf("no index or changes files found in %s (looking for patterns like YYYY-MM-DDTHHMISS__index.jsonl.zst or *changes*.jsonl/.zst)", parentPath)
	}

	return changesFiles, nil
}

// DatasetAnalysis holds the analysis results for a single dataset
type DatasetAnalysis struct {
	Spec       string   `json:"spec"`
	IsRDF      bool     `json:"is_rdf"`
	Aggregator string   `json:"aggregator,omitempty"`
	Lines      []string `json:"lines"`
	Error      string   `json:"error,omitempty"`
	SourceFile string   `json:"source_file"`
}

// AnalysisResults holds all dataset analysis results
type AnalysisResults struct {
	TotalDatasets       int                       `json:"total_datasets"`
	RDFDatasets         int                       `json:"rdf_datasets"`
	NonRDFDatasets      int                      `json:"non_rdf_datasets"`
	RDFSpecs            []string                 `json:"rdf_specs"`
	Aggregators         map[string]int           `json:"aggregators"`
	RDFAggregators      map[string]int           `json:"rdf_aggregators"`
	AggregatorDatasets  map[string][]string      `json:"aggregator_datasets"`
	RDFAggregatorDatasets map[string][]string    `json:"rdf_aggregator_datasets"`
	DatasetAnalyses     []DatasetAnalysis        `json:"dataset_analyses"`
}

// findDatasetDirectories discovers all subdirectories that could contain dataset files
func findDatasetDirectories(parentPath string) ([]string, error) {
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
		directories = append(directories, dirPath)
	}

	if len(directories) == 0 {
		return nil, fmt.Errorf("no subdirectories found in %s", parentPath)
	}

	// Sort directories for consistent processing order
	sort.Strings(directories)

	return directories, nil
}

// analyzeDatasetDirectory analyzes a single dataset directory for RDF-XML content
// Returns: DatasetAnalysis struct with all details
func analyzeDatasetDirectory(dirPath string) DatasetAnalysis {
	spec := filepath.Base(dirPath)
	analysis := DatasetAnalysis{
		Spec: spec,
	}

	sourceFile, err := findSourceXMLFile(dirPath)
	if err != nil {
		analysis.Error = fmt.Sprintf("no source.xml.gz file found in %s: %v", dirPath, err)
		return analysis
	}

	analysis.SourceFile = sourceFile

	isRDF, aggregator, lines, err := detectRDFXML(sourceFile)
	if err != nil {
		analysis.Error = fmt.Sprintf("failed to detect RDF-XML in %s: %v", sourceFile, err)
		return analysis
	}

	analysis.IsRDF = isRDF
	analysis.Aggregator = aggregator
	analysis.Lines = lines
	return analysis
}

// findSourceXMLFile finds source.xml.gz file in the given directory
func findSourceXMLFile(dirPath string) (string, error) {
	sourceFile := filepath.Join(dirPath, "source.xml.gz")
	if _, err := os.Stat(sourceFile); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source.xml.gz not found")
		}
		return "", fmt.Errorf("error checking source.xml.gz: %w", err)
	}
	return sourceFile, nil
}

// detectAggregator identifies the data aggregator based on identifier patterns
func detectAggregator(lines []string) string {
	// Aggregator patterns to look for in identifier elements
	aggregatorPatterns := map[string]*regexp.Regexp{
		"dcn":         regexp.MustCompile(`<identifier>dcn_`),
		"za":          regexp.MustCompile(`<identifier>za_`),
		"brabantcloud": regexp.MustCompile(`<identifier>brabantcloud_`),
	}
	
	for _, line := range lines {
		for aggregator, pattern := range aggregatorPatterns {
			if pattern.MatchString(line) {
				return aggregator
			}
		}
	}
	
	return "unknown"
}

// detectRDFXML reads the first 100 lines of a gzipped XML file and detects RDF-XML patterns
// Returns: isRDF, aggregator, lines read, error
func detectRDFXML(filePath string) (bool, string, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, "", nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return false, "", nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	scanner := bufio.NewScanner(gzipReader)
	lineCount := 0
	maxLines := 100
	var lines []string
	isRDF := false

	// RDF-XML detection patterns - now requires stricter criteria
	// Basic RDF indicators (namespace declarations)
	namespacePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)xmlns:rdf\s*=`),                  // xmlns:rdf=
		regexp.MustCompile(`(?i)http://www\.w3\.org/1999/02/22-rdf-syntax-ns#`), // RDF namespace URI
	}
	
	// Strict RDF patterns (actual RDF usage - required for qualification)
	strictRDFPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<\s*rdf:RDF`),                    // <rdf:RDF root element
		regexp.MustCompile(`(?i)rdf:about\s*=`),                  // rdf:about attribute
		regexp.MustCompile(`(?i)rdf:resource\s*=`),               // rdf:resource attribute
	}
	
	// Additional semantic patterns (also qualify as RDF)
	semanticPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)xmlns:rdfs\s*=`),                 // xmlns:rdfs=
		regexp.MustCompile(`(?i)xmlns:owl\s*=`),                  // xmlns:owl=
		regexp.MustCompile(`(?i)<\s*rdf:Description`),            // <rdf:Description (moved here)
	}

	var hasNamespace bool
	var hasStrictRDF bool
	var hasSemantic bool

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		lineCount++
		lines = append(lines, line)

		// Check namespace patterns
		if !hasNamespace {
			for _, pattern := range namespacePatterns {
				if pattern.MatchString(line) {
					hasNamespace = true
					slog.Debug("RDF namespace detected", "file", filePath, "line", lineCount, "pattern", pattern.String())
					break
				}
			}
		}

		// Check strict RDF patterns
		if !hasStrictRDF {
			for _, pattern := range strictRDFPatterns {
				if pattern.MatchString(line) {
					hasStrictRDF = true
					slog.Debug("Strict RDF pattern detected", "file", filePath, "line", lineCount, "pattern", pattern.String())
					break
				}
			}
		}

		// Check semantic patterns 
		if !hasSemantic {
			for _, pattern := range semanticPatterns {
				if pattern.MatchString(line) {
					hasSemantic = true
					slog.Debug("Semantic pattern detected", "file", filePath, "line", lineCount, "pattern", pattern.String())
					break
				}
			}
		}

		// Set RDF flag but continue reading to get all 100 lines
		if hasNamespace && hasStrictRDF {
			isRDF = true
		}
	}

	// Final determination: need namespace + strict RDF usage (not just semantic namespaces)
	isRDF = hasNamespace && hasStrictRDF

	if err := scanner.Err(); err != nil {
		return false, "", lines, fmt.Errorf("error reading file: %w", err)
	}

	// Detect aggregator from the lines we've read
	aggregator := detectAggregator(lines)

	slog.Debug("RDF detection completed", 
		"file", filePath, 
		"linesChecked", lineCount, 
		"isRDF", isRDF,
		"aggregator", aggregator,
		"hasNamespace", hasNamespace,
		"hasStrictRDF", hasStrictRDF, 
		"hasSemantic", hasSemantic)
	return isRDF, aggregator, lines, nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	if aFlags.all {
		return runAnalyzeAll()
	}
	return runAnalyzeSingle()
}

func runAnalyzeAll() error {
	directories, err := findDatasetDirectories(aFlags.path)
	if err != nil {
		return err
	}

	slog.Info("found directories to analyze", "count", len(directories), "parentPath", aFlags.path)

	results := AnalysisResults{
		DatasetAnalyses:       make([]DatasetAnalysis, 0, len(directories)),
		Aggregators:           make(map[string]int),
		RDFAggregators:        make(map[string]int),
		AggregatorDatasets:    make(map[string][]string),
		RDFAggregatorDatasets: make(map[string][]string),
	}
	successCount := 0

	for i, dirPath := range directories {
		dirName := filepath.Base(dirPath)
		
		slog.Info("analyzing dataset directory", "progress", fmt.Sprintf("%d/%d", i+1, len(directories)), "dir", dirPath, "spec", dirName)
		
		analysis := analyzeDatasetDirectory(dirPath)
		results.DatasetAnalyses = append(results.DatasetAnalyses, analysis)
		
		if analysis.Error == "" {
			results.TotalDatasets++
			if analysis.IsRDF {
				results.RDFDatasets++
				results.RDFSpecs = append(results.RDFSpecs, dirName)
				// Track RDF-specific aggregator counts and datasets
				if analysis.Aggregator != "" {
					results.RDFAggregators[analysis.Aggregator]++
					results.RDFAggregatorDatasets[analysis.Aggregator] = append(results.RDFAggregatorDatasets[analysis.Aggregator], dirName)
				}
			}
			// Track overall aggregator counts and datasets
			if analysis.Aggregator != "" {
				results.Aggregators[analysis.Aggregator]++
				results.AggregatorDatasets[analysis.Aggregator] = append(results.AggregatorDatasets[analysis.Aggregator], dirName)
			}
			successCount++
			slog.Info("completed analysis", "dir", dirPath, "spec", dirName, "isRDF", analysis.IsRDF, "aggregator", analysis.Aggregator)
		} else {
			slog.Error("failed to analyze directory", "dir", dirPath, "error", analysis.Error)
		}
	}

	results.NonRDFDatasets = results.TotalDatasets - results.RDFDatasets

	// Write JSON output if requested
	if aFlags.jsonOutput != "" {
		if err := writeJSONOutput(aFlags.jsonOutput, results, aFlags.rdfOnly); err != nil {
			slog.Error("failed to write JSON output", "file", aFlags.jsonOutput, "error", err)
		} else {
			datasetsInOutput := len(results.DatasetAnalyses)
			if aFlags.rdfOnly {
				datasetsInOutput = results.RDFDatasets
			}
			slog.Info("wrote JSON output", "file", aFlags.jsonOutput, "datasets", datasetsInOutput, "rdfOnly", aFlags.rdfOnly)
		}
	}

	// Print summary
	fmt.Printf("\nAnalysis Results:\n")
	fmt.Printf("================\n")
	fmt.Printf("Total datasets: %d\n", results.TotalDatasets)
	fmt.Printf("RDF-XML datasets: %d\n", results.RDFDatasets)
	fmt.Printf("Non-RDF datasets: %d\n", results.NonRDFDatasets)
	
	fmt.Printf("\nData aggregators (all datasets):\n")
	for aggregator, count := range results.Aggregators {
		fmt.Printf("- %s: %d datasets\n", aggregator, count)
	}
	
	fmt.Printf("\nRDF aggregators (RDF datasets only):\n")
	for aggregator, count := range results.RDFAggregators {
		fmt.Printf("- %s: %d RDF datasets\n", aggregator, count)
		if datasets, exists := results.RDFAggregatorDatasets[aggregator]; exists {
			for _, dataset := range datasets {
				fmt.Printf("  * %s\n", dataset)
			}
		}
	}
	
	fmt.Printf("\nRDF-XML dataset specs:\n")
	for _, spec := range results.RDFSpecs {
		fmt.Printf("- %s\n", spec)
	}

	if aFlags.jsonOutput != "" {
		fmt.Printf("\nDetailed analysis written to: %s\n", aFlags.jsonOutput)
	}

	slog.Info("finished analyzing all directories",
		"successfulDirs", successCount,
		"totalDirs", len(directories),
		"rdfDatasets", results.RDFDatasets,
		"nonRdfDatasets", results.NonRDFDatasets)

	return nil
}

func runAnalyzeSingle() error {
	dirName := filepath.Base(aFlags.path)
	
	slog.Info("analyzing single dataset directory", "dir", aFlags.path, "spec", dirName)
	
	analysis := analyzeDatasetDirectory(aFlags.path)
	
	if analysis.Error != "" {
		return fmt.Errorf("analysis failed: %s", analysis.Error)
	}

	// Write JSON output if requested
	if aFlags.jsonOutput != "" {
		results := AnalysisResults{
			TotalDatasets:         1,
			DatasetAnalyses:       []DatasetAnalysis{analysis},
			Aggregators:           make(map[string]int),
			RDFAggregators:        make(map[string]int),
			AggregatorDatasets:    make(map[string][]string),
			RDFAggregatorDatasets: make(map[string][]string),
		}
		if analysis.IsRDF {
			results.RDFDatasets = 1
			results.RDFSpecs = []string{dirName}
			// Track RDF aggregator
			if analysis.Aggregator != "" {
				results.RDFAggregators[analysis.Aggregator] = 1
				results.RDFAggregatorDatasets[analysis.Aggregator] = []string{dirName}
			}
		} else {
			results.NonRDFDatasets = 1
		}
		// Track overall aggregator
		if analysis.Aggregator != "" {
			results.Aggregators[analysis.Aggregator] = 1
			results.AggregatorDatasets[analysis.Aggregator] = []string{dirName}
		}
		
		if err := writeJSONOutput(aFlags.jsonOutput, results, aFlags.rdfOnly); err != nil {
			slog.Error("failed to write JSON output", "file", aFlags.jsonOutput, "error", err)
		} else {
			datasetsInOutput := 1
			if aFlags.rdfOnly && !analysis.IsRDF {
				datasetsInOutput = 0
			}
			slog.Info("wrote JSON output", "file", aFlags.jsonOutput, "datasets", datasetsInOutput, "rdfOnly", aFlags.rdfOnly)
		}
	}

	// Print result
	fmt.Printf("\nAnalysis Result:\n")
	fmt.Printf("===============\n")
	fmt.Printf("Dataset spec: %s\n", dirName)
	fmt.Printf("Is RDF-XML: %t\n", analysis.IsRDF)
	fmt.Printf("Aggregator: %s\n", analysis.Aggregator)
	fmt.Printf("Lines analyzed: %d\n", len(analysis.Lines))

	if aFlags.jsonOutput != "" {
		fmt.Printf("Detailed analysis written to: %s\n", aFlags.jsonOutput)
	}

	slog.Info("completed analysis", "dir", aFlags.path, "spec", dirName, "isRDF", analysis.IsRDF)

	return nil
}

// writeJSONOutput writes the analysis results to a JSON file
// If rdfOnly is true, only includes RDF datasets in the output
func writeJSONOutput(filename string, results AnalysisResults, rdfOnly bool) error {
	outputResults := results
	
	if rdfOnly {
		// Filter to only include RDF datasets
		var rdfAnalyses []DatasetAnalysis
		for _, analysis := range results.DatasetAnalyses {
			if analysis.IsRDF {
				rdfAnalyses = append(rdfAnalyses, analysis)
			}
		}
		
		outputResults = AnalysisResults{
			TotalDatasets:         results.TotalDatasets,         // Keep original total for context
			RDFDatasets:           results.RDFDatasets,           // Keep RDF count
			NonRDFDatasets:        results.NonRDFDatasets,       // Keep non-RDF count for context
			RDFSpecs:              results.RDFSpecs,              // Keep RDF specs list
			Aggregators:           results.Aggregators,           // Keep all aggregator data for context
			RDFAggregators:        results.RDFAggregators,        // Keep RDF aggregator data
			AggregatorDatasets:    results.AggregatorDatasets,    // Keep dataset lists per aggregator
			RDFAggregatorDatasets: results.RDFAggregatorDatasets, // Keep RDF dataset lists per aggregator
			DatasetAnalyses:       rdfAnalyses,                  // Only RDF analyses
		}
		
		slog.Info("filtered JSON output to RDF datasets only", 
			"totalAnalyses", len(results.DatasetAnalyses),
			"rdfAnalyses", len(rdfAnalyses))
	}

	// Create a buffer and encoder with HTML escaping disabled
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(outputResults); err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
