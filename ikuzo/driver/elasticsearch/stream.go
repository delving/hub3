package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/olivere/elastic/v7"
)

var ErrEndOfScroll = errors.New("reached end of scroll")

type StreamConfig struct {
	Query      elastic.Query
	IndexNames []string
	Fsc        *elastic.FetchSourceContext
	pitID      string
}

// Stream can be used to stream process results of a query from ElasticSearch
func (c *Client) Stream(
	ctx context.Context, cfg *StreamConfig,
	fn func(hit *elastic.SearchHit) error,
) (seen int, err error) {
	slog.Info("starting streaming", "cfg", cfg, "urls", c.cfg.Urls)

	// Check ES version to determine whether to use PIT or scroll
	pitClient, err := NewPITClient(c.cfg.Urls, c.cfg.UserName, c.cfg.Password)
	if err != nil {
		return 0, fmt.Errorf("unable to create PIT client: %w", err)
	}

	supportsPIT, err := pitClient.SupportsPIT(ctx)
	if err != nil {
		slog.Warn("unable to check ES version, falling back to scroll API", "error", err)
		supportsPIT = false
	}

	if supportsPIT {
		return c.streamWithPIT(ctx, cfg, fn, pitClient)
	}

	slog.Info("ES version < 7.10, using scroll API instead of PIT")
	return c.streamWithScroll(ctx, cfg, fn)
}

// streamWithPIT uses Point in Time API (ES 7.10+)
func (c *Client) streamWithPIT(
	ctx context.Context, cfg *StreamConfig,
	fn func(hit *elastic.SearchHit) error,
	pitClient *PITClient,
) (seen int, err error) {
	indexName := strings.Join(cfg.IndexNames, ",")
	pitID, err := pitClient.CreatePIT(ctx, indexName, "1m")
	if err != nil {
		return 0, fmt.Errorf("unable to open point in time: %w", err)
	}

	defer func() {
		if closeErr := pitClient.ClosePIT(context.Background(), pitID); closeErr != nil {
			c.log.Error().Err(closeErr).Msg("unable to close point in time")
		}
		slog.Debug("closed point in time", "id", pitID)
	}()

	cfg.pitID = pitID

	var (
		esPage      int
		totalSeen   int
		searchAfter []interface{}
		hits        []*elastic.SearchHit
	)

	for {
		hits, searchAfter, err = c.next(cfg, searchAfter)
		if err != nil {
			break
		}

		for _, hit := range hits {
			if fnErr := fn(hit); fnErr != nil {
				return 0, fmt.Errorf("error processing hit: %w", fnErr)
			}
		}
		esPage++
	}

	if err != nil {
		if !errors.Is(err, ErrEndOfScroll) {
			return 0, fmt.Errorf("unable to process page %d: %w", esPage, err)
		}

		if totalSeen == 0 {
			return totalSeen, nil
		}
	}

	return totalSeen, nil
}

// streamWithScroll uses the scroll API (for ES < 7.10)
func (c *Client) streamWithScroll(
	ctx context.Context, cfg *StreamConfig,
	fn func(hit *elastic.SearchHit) error,
) (seen int, err error) {
	pageSize := 1000
	scrollKeepAlive := "5m"

	scroll := c.search.Scroll(cfg.IndexNames...).
		Size(pageSize).
		Query(cfg.Query).
		Scroll(scrollKeepAlive).
		Sort("_doc", true)

	if cfg.Fsc != nil {
		scroll = scroll.FetchSourceContext(cfg.Fsc)
	}

	var (
		esPage    int
		totalSeen int
	)

	for {
		results, err := scroll.Do(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return totalSeen, fmt.Errorf("scroll error on page %d: %w", esPage, err)
		}

		if results.Hits == nil || len(results.Hits.Hits) == 0 {
			break
		}

		for _, hit := range results.Hits.Hits {
			if fnErr := fn(hit); fnErr != nil {
				return totalSeen, fmt.Errorf("error processing hit: %w", fnErr)
			}
			totalSeen++
		}
		esPage++
	}

	return totalSeen, nil
}

func (c *Client) next(cfg *StreamConfig, searchAfter []interface{}) ([]*elastic.SearchHit, []interface{}, error) {
	pageSize := 1000

	search := c.search.Search().
		Size(pageSize).
		Query(cfg.Query).
		PointInTime(elastic.NewPointInTimeWithKeepAlive(cfg.pitID, "10m")).
		Sort("meta.hubID", true)

	if cfg.Fsc != nil {
		search = search.FetchSourceContext(cfg.Fsc)
	}

	if len(searchAfter) > 0 {
		search = search.SearchAfter(searchAfter...)
	}

	resp, err := search.Do(context.Background())
	if err != nil {
		return nil, nil, err
	}

	if resp.Error != nil {
		return nil, nil, fmt.Errorf("%s", resp.Error.Reason)
	}

	if len(resp.Hits.Hits) == 0 {
		return nil, nil, ErrEndOfScroll
	}

	newSearchAfter := resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort

	return resp.Hits.Hits, newSearchAfter, nil
}
