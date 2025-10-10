package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/tools/cmd/index_narthex/essync"
	"github.com/delving/hub3/tools/cmd/index_narthex/stats"
	"github.com/spf13/cobra"
)

// V2Client handles communication with external v2 API
type V2Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// V2Response matches the v2 API ScrollResultV4 structure
type V2Response struct {
	Pager *struct {
		NextScrollID string `json:"nextScrollID"`
		Total        int64  `json:"total"`
	} `json:"pager"`
	Items []*fragments.FragmentGraph `json:"items"`
}

// NewV2Client creates a new v2 API client
func NewV2Client(baseURL string) *V2Client {
	return &V2Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: slog.Default(),
	}
}

// FetchPage fetches a single page from the v2 API.
// If scrollID is empty, builds initial query from params.
// Otherwise uses scrollID for pagination.
func (c *V2Client) FetchPage(ctx context.Context, scrollID string, params url.Values) (*V2Response, error) {
	var requestURL string

	if scrollID == "" {
		// Initial request with query parameters
		requestURL = fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	} else {
		// Pagination request with scrollID
		requestURL = fmt.Sprintf("%s?scrollID=%s", c.baseURL, url.QueryEscape(scrollID))
	}

	c.logger.Debug("fetching v2 page", "url", requestURL)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var v2Resp V2Response
	if err := json.NewDecoder(resp.Body).Decode(&v2Resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &v2Resp, nil
}

// streamToIndexer converts FragmentGraph items to NDJSON bulk format.
// Format: {"index":{"_index":"...","_id":"..."}} \n {...source...} \n
func streamToIndexer(items []*fragments.FragmentGraph, writer io.Writer, indexName string, targetOrgID string) error {
	for _, item := range items {
		if item == nil || item.Meta == nil {
			continue
		}

		// Override orgID if target is different from source
		if targetOrgID != "" {
			item.Meta.OrgID = targetOrgID
		}

		// Write bulk metadata line
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    item.Meta.HubID,
			},
		}

		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		if _, err := writer.Write(metaJSON); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}

		// Write document source line
		sourceJSON, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal source: %w", err)
		}

		if _, err := writer.Write(sourceJSON); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

// runSync is the main entry point for the sync command
func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create v2 client
	client := NewV2Client(sFlags.sourceURL)

	// Build initial query parameters
	params := url.Values{}
	params.Set("orgID", sFlags.sourceOrgID)
	params.Set("rows", fmt.Sprintf("%d", sFlags.batchSize))
	params.Set("itemFormat", "fragmentGraph") // Critical: get full documents

	if sFlags.query != "" {
		params.Set("q", sFlags.query)
	}

	slog.Info("starting sync from external v2 API",
		"source", sFlags.sourceURL,
		"orgID", sFlags.sourceOrgID,
		"query", sFlags.query,
		"batchSize", sFlags.batchSize)

	if sFlags.dryRun {
		slog.Info("DRY RUN MODE: will fetch but not index data")
		return runSyncDryRun(ctx, client, params)
	}

	// Create ES client for indexing
	esClient, err := elasticsearch.NewClient(&elasticsearch.Config{
		Urls: []string{sFlags.esURL},
	})
	if err != nil {
		return fmt.Errorf("failed to create ES client: %w", err)
	}

	// Create pipe for streaming data to bulk processor
	pipeReader, pipeWriter := io.Pipe()

	// Setup stats and reporters
	reporters := stats.NewMultiReporter(
		stats.NewSlogReporter(slog.Default(), sFlags.statsInterval),
		stats.NewProgressReporter(sFlags.statsInterval),
	)

	// Start bulk processor in goroutine
	processorDone := make(chan error, 1)
	go func() {
		processor := essync.NewBulkProcessor(essync.Config{
			NumWorkers:    sFlags.workers,
			MaxBatchSize:  sFlags.batchSize,
			UpdateTimeout: sFlags.statsInterval,
			ESClient:      esClient,
			Reporter:      reporters,
		})

		processorDone <- processor.Process(ctx, pipeReader)
	}()

	// Fetch and stream data
	targetOrg := sFlags.targetOrgID
	if targetOrg == "" {
		targetOrg = sFlags.sourceOrgID
	}

	var scrollID string
	var totalFetched int64
	var totalExpected int64

	for {
		// Fetch page
		resp, err := client.FetchPage(ctx, scrollID, params)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return fmt.Errorf("failed to fetch page: %w", err)
		}

		// Update totals on first page
		if totalExpected == 0 && resp.Pager != nil {
			totalExpected = resp.Pager.Total
			slog.Info("sync started", "totalRecords", totalExpected)
		}

		// Stream items to indexer
		if len(resp.Items) > 0 {
			if err := streamToIndexer(resp.Items, pipeWriter, sFlags.indexName, targetOrg); err != nil {
				pipeWriter.CloseWithError(err)
				return fmt.Errorf("failed to stream items: %w", err)
			}

			totalFetched += int64(len(resp.Items))
			slog.Debug("fetched page", "items", len(resp.Items), "total", totalFetched)
		}

		// Check if more pages
		if resp.Pager == nil || resp.Pager.NextScrollID == "" {
			slog.Info("fetch complete", "totalFetched", totalFetched)
			break
		}

		scrollID = resp.Pager.NextScrollID
	}

	// Close pipe and wait for processor
	pipeWriter.Close()
	if err := <-processorDone; err != nil {
		return fmt.Errorf("bulk processing failed: %w", err)
	}

	slog.Info("sync completed successfully",
		"fetched", totalFetched,
		"expected", totalExpected)

	return nil
}

// runSyncDryRun fetches data without indexing (for testing)
func runSyncDryRun(ctx context.Context, client *V2Client, params url.Values) error {
	var scrollID string
	var totalFetched int64
	var pageCount int

	for {
		resp, err := client.FetchPage(ctx, scrollID, params)
		if err != nil {
			return fmt.Errorf("failed to fetch page: %w", err)
		}

		pageCount++
		totalFetched += int64(len(resp.Items))

		hasMore := resp.Pager != nil && resp.Pager.NextScrollID != ""

		slog.Info("dry-run fetched page",
			"page", pageCount,
			"items", len(resp.Items),
			"total", totalFetched,
			"hasMore", hasMore)

		// Sample first item from first page
		if pageCount == 1 && len(resp.Items) > 0 {
			slog.Info("sample record",
				"hubID", resp.Items[0].Meta.HubID,
				"orgID", resp.Items[0].Meta.OrgID,
				"spec", resp.Items[0].Meta.Spec)
		}

		if !hasMore {
			break
		}

		scrollID = resp.Pager.NextScrollID
	}

	slog.Info("dry-run completed",
		"totalPages", pageCount,
		"totalFetched", totalFetched)

	return nil
}
