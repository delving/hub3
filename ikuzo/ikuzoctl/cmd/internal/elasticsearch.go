package internal

import (
	"expvar"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	eshub "github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type ElasticSearch struct {
	// enable elasticsearch client
	Enabled bool
	// urls to connect to elasticsearch cluster
	Urls []string
	// enable elasticsearch caching proxy
	Proxy bool
	// number of elasticsearch workers. default: 1
	Workers int
	// maxRetries number of client retries. default: 5
	MaxRetries int
	// clientTimeOut seconds for the client to time out. default 10
	ClientTimeOut int
	// gather elasticsearch metrics
	Metrics bool
	// elasticsearch client
	client *elasticsearch.Client
	// BulkIndexer
	bi esutil.BulkIndexer
}

func (e *ElasticSearch) AddOptions(cfg *Config) error {
	if !e.Enabled || len(e.Urls) == 0 {
		return nil
	}

	client, err := e.newClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("unable to create elasticsearch.Client: %w", err)
	}

	if e.Proxy {
		esProxy, err := eshub.NewProxy(client)
		if err != nil {
			return fmt.Errorf("unable to create ES proxy: %w", err)
		}

		cfg.options = append(cfg.options, ikuzo.SetElasticSearchProxy(esProxy))
	}

	return nil
}

func (e *ElasticSearch) newClient(l *logger.CustomLogger) (*elasticsearch.Client, error) {
	if e.client != nil {
		return e.client, nil
	}

	retryBackoff := backoff.NewExponentialBackOff()

	client, err := elasticsearch.NewClient(
		elasticsearch.Config{
			// Retry on 429 TooManyRequests statuses
			//
			RetryOnStatus: []int{502, 503, 504, 429},

			// Configure the backoff function
			//
			RetryBackoff: func(i int) time.Duration {
				if i == 1 {
					retryBackoff.Reset()
				}
				return retryBackoff.NextBackOff()
			},

			// Enable client metrics
			//
			EnableMetrics: e.Metrics,

			// Retry up to MaxRetries attempts
			//
			MaxRetries: e.MaxRetries,

			// Custom transport based on fasthttp
			//
			Transport: &eshub.Transport{},

			// Custom rs/zerolog structured logger
			//
			Logger: l,
		},
	)

	// Publish client metrics to expvar
	if e.Metrics {
		expvar.Publish("go-elasticsearch", expvar.Func(func() interface{} { m, _ := client.Metrics(); return m }))
	}

	e.client = client

	return e.client, err
}

func (e *ElasticSearch) newBulkIndexer(es *elasticsearch.Client) (esutil.BulkIndexer, error) {
	if e.bi != nil {
		return e.bi, nil
	}

	var err error

	// TODO(kiivihal): check and create index mapping

	flushBytes := 5e+6 // 5 MB
	numWorkers := e.Workers

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        es,               // The Elasticsearch client
		NumWorkers:    numWorkers,       // The number of worker goroutines
		FlushBytes:    int(flushBytes),  // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})

	if e.Metrics {
		expvar.Publish("go-elasticsearch-bulk", expvar.Func(func() interface{} { m := bi.Stats(); return m }))
	}

	e.bi = bi

	return e.bi, err
}
