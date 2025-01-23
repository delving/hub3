package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	// Create a Point In Time
	pitReq := c.search.OpenPointInTime(cfg.IndexNames...).
		KeepAlive("1m").
		Pretty(true)

	if err := pitReq.Validate(); err != nil {
		slog.Error("unable to validate point in time request", "error", err)
		return 0, fmt.Errorf("unable to validate point in time request: %w", err)
	}

	openResp, err := pitReq.Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to open point in time: %w", err)
	}

	defer func() {
		resp, err := c.search.ClosePointInTime(openResp.Id).Pretty(true).Do(context.Background())
		if err != nil {
			c.log.Error().Err(err).Msg("unable to close point in time")
		}
		slog.Debug("closing point in time", "id", openResp.Id, "succeeded", resp.Succeeded, "freeing resources", resp.NumFreed)
	}()

	cfg.pitID = openResp.Id

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
