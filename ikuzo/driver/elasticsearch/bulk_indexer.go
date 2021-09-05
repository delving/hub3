package elasticsearch

import (
	"context"
	"expvar"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

func (c *Client) NewBulkIndexer(orgs []domain.OrganizationConfig, workers int) (esutil.BulkIndexer, error) {
	// create default mappings
	indexNames, err := c.CreateDefaultMappings(orgs, true, false)
	if err != nil {
		c.log.Error().Err(err).Msg("unable to create mappings")
		return nil, err
	}

	c.log.Info().Strs("created indices", indexNames).Msg("created or updated mappings")

	flushBytes := 5 * 1024 * 1024 // 5 MB
	numWorkers := workers

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        c.index,         // The Elasticsearch client
		NumWorkers:    numWorkers,      // The number of worker goroutines
		FlushBytes:    flushBytes,      // The flush threshold in bytes
		FlushInterval: 5 * time.Second, // The periodic flush interval
		OnError: func(ctx context.Context, err error) {
			c.log.Error().Err(err).Msg("flush: bulk indexing error")
		},
	})

	if !c.cfg.DisableMetrics {
		expvar.Publish("go-elasticsearch-bulk", expvar.Func(func() interface{} { m := bi.Stats(); return m }))
	}

	return bi, err
}
