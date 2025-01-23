package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
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
)

type processFlags struct {
	path        string
	targetIndex string
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
}

type submitFlags struct {
	changesFile   string
	esURL         string
	workers       int
	batchSize     int
	statsInterval time.Duration
}

var (
	pFlags   processFlags
	dFlags   downloadFlags
	detFlags detectFlags
	subFlags submitFlags
)

func init() {
	// process command flags
	processCmd.Flags().StringVar(&pFlags.path, "path", "", "directory that contains the narthex rdf-xml files")
	processCmd.Flags().StringVar(&pFlags.targetIndex, "index", "", "elasticsearch index to write to")
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
	detectCmd.Flags().StringVarP(&detFlags.inputFile, "input", "i", "", "Input file path (required)")
	detectCmd.Flags().StringVarP(&detFlags.outputFile, "output", "o", "", "Output file path (required)")
	detectCmd.Flags().StringVarP(&detFlags.esURL, "url", "u", "http://localhost:9200", "Elasticsearch URL")
	detectCmd.Flags().StringVarP(&detFlags.indexName, "index", "n", "", "Index name (required)")
	detectCmd.Flags().StringVarP(&detFlags.queryString, "query", "q", "", "Query JSON string")
	detectCmd.Flags().StringVarP(&detFlags.fnameFilter, "fnameFilter", "", "", "Filter for lastest containing substring in input file path is a directory")
	detectCmd.Flags().DurationVar(&detFlags.statsInterval, "stats-interval", 1*time.Second, "Statistics update interval")
	detectCmd.MarkFlagRequired("input")
	detectCmd.MarkFlagRequired("output")
	detectCmd.MarkFlagRequired("index")

	// submit command flags
	submitCmd.Flags().StringVarP(&subFlags.changesFile, "changes", "c", "", "Changes file path (required)")
	submitCmd.Flags().StringVarP(&subFlags.esURL, "url", "u", "http://localhost:9200", "Elasticsearch URL")
	submitCmd.Flags().IntVarP(&subFlags.workers, "workers", "w", 4, "Number of worker goroutines")
	submitCmd.Flags().IntVar(&subFlags.batchSize, "batch-size", 1000, "Maximum batch size for bulk requests")
	submitCmd.Flags().DurationVar(&subFlags.statsInterval, "stats-interval", 1*time.Second, "Statistics update interval")
	submitCmd.MarkFlagRequired("changes")

	// Add all commands to root
	rootCmd.AddCommand(processCmd, downloadCmd, detectCmd, submitCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	cfg := elasticsearch.Config{
		Urls: []string{detFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	// Setup reporters
	logger := slog.Default()
	reporter := stats.NewMultiReporter(
		stats.NewSlogReporter(logger, detFlags.statsInterval),
		// stats.NewConsoleReporter(5*time.Second),
	)

	detector, err := essync.NewChangeDetector(detFlags.outputFile, reporter)
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

	if err := detector.DetectChanges(detFlags.inputFile, opts); err != nil {
		return fmt.Errorf("error detecting changes: %w", err)
	}

	return nil
}

func runProcess(cmd *cobra.Command, args []string) error {
	rdfXML, err := essync.FindLatestPath(pFlags.path, "__processed.rdf")
	if err != nil {
		return fmt.Errorf("unable to find processed file; %w", err)
	}

	slog.Info("path to the processed rdf file", "resolvedPath", rdfXML, "sourcePath", pFlags.path)

	if rdfXML == "" {
		return fmt.Errorf("no '_processed.rdf' file found in %s", pFlags.path)
	}

	r, err := openXMLReader(rdfXML)
	if err != nil {
		return fmt.Errorf("unable to open processed mapping file: %w", err)
	}
	defer r.Close()
	cfg := Config{IndexName: pFlags.targetIndex}
	cfg.OutputPath, cfg.DatePrefix = processPath(rdfXML)
	if cfg.DatePrefix == "" {
		return fmt.Errorf("unable to extract date prefix from path (expected prefix before '__' in the filename)")
	}
	stats, err := ParseNarthex(r, cfg)
	if err != nil {
		return fmt.Errorf("unable to parse narthex: %w", err)
	}
	slog.Info("finished processing narthex file",
		"records", stats.Records,
		"lines", stats.Lines,
		"converted", stats.Converted)
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
	cfg := elasticsearch.Config{
		Urls: []string{subFlags.esURL},
	}
	client, err := elasticsearch.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	var reader io.ReadCloser
	file, err := os.Open(subFlags.changesFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if strings.HasSuffix(subFlags.changesFile, ".zst") {
		decoder, err := zstd.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		defer decoder.Close()
		reader = io.NopCloser(decoder)
	} else {
		reader = file
	}

	// Create multiple reporters
	reporters := stats.NewMultiReporter(
		stats.NewSlogReporter(slog.Default(), subFlags.statsInterval),
		// essync.NewConsoleReporter(subFlags.statsInterval),
	)

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

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
