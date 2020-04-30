package internal

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/x/index"
	eshub "github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch/mapping"
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
	// IndexService
	is *index.Service
	// base of the index aliases
	IndexName string
	// number of shards. default 1
	Shards int
	// number of replicas. default 0
	Replicas int
	// logger
	logger *logger.CustomLogger
	// UseRemoteIndexer is true when a separate process reads of the queue
	UseRemoteIndexer bool
}

func (e *ElasticSearch) AddOptions(cfg *Config) error {
	if !e.Enabled || len(e.Urls) == 0 {
		return nil
	}

	client, err := e.NewClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("unable to create elasticsearch.Client: %w", err)
	}

	if e.Proxy {
		esProxy, proxyErr := eshub.NewProxy(client)
		if proxyErr != nil {
			return fmt.Errorf("unable to create ES proxy: %w", proxyErr)
		}

		cfg.options = append(cfg.options, ikuzo.SetElasticSearchProxy(esProxy))
	}

	_, err = e.CreateDefaultMappings(client, true)
	if err != nil {
		return err
	}

	return nil
}

func (e *ElasticSearch) CreateDefaultMappings(es *elasticsearch.Client, withAlias bool) ([]string, error) {
	mappings := map[string]func(shards, replicas int) string{
		fmt.Sprintf("%sv1", e.IndexName): mapping.V1ESMapping,
		fmt.Sprintf("%sv2", e.IndexName): mapping.V2ESMapping,
	}

	indexNames := []string{}

	for indexName, m := range mappings {
		createName, err := eshub.IndexCreate(
			es,
			indexName,
			m(e.Shards, e.Replicas),
			withAlias,
		)

		if err != nil && !errors.Is(err, eshub.ErrIndexAlreadyCreated) {
			return []string{}, err
		}

		indexNames = append(indexNames, createName)
	}

	return indexNames, nil
}

func (e *ElasticSearch) NewClient(l *logger.CustomLogger) (*elasticsearch.Client, error) {
	if e.client != nil {
		return e.client, nil
	}

	e.logger = l

	retryBackoff := backoff.NewExponentialBackOff()

	client, err := elasticsearch.NewClient(
		elasticsearch.Config{
			// Connect to ElasticSearch URLS
			//
			Addresses: e.Urls,

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

func (e *ElasticSearch) NewBulkIndexer(es *elasticsearch.Client) (esutil.BulkIndexer, error) {
	if e.bi != nil {
		return e.bi, nil
	}

	if es == nil {
		return nil, fmt.Errorf("cannot start BulkIndexer without valid es client")
	}

	// create default mappings
	_, err := e.CreateDefaultMappings(es, true)
	if err != nil {
		return nil, err
	}

	flushBytes := 5 * 1024 * 1024 // 5 MB
	numWorkers := e.Workers

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        es,              // The Elasticsearch client
		NumWorkers:    numWorkers,      // The number of worker goroutines
		FlushBytes:    flushBytes,      // The flush threshold in bytes
		FlushInterval: 5 * time.Second, // The periodic flush interval
		OnError: func(ctx context.Context, err error) {
			e.logger.Error().Err(err).Msg("flush: bulk indexing error")
		},
	})

	if e.Metrics {
		expvar.Publish("go-elasticsearch-bulk", expvar.Func(func() interface{} { m := bi.Stats(); return m }))
	}

	e.bi = bi

	return e.bi, err
}

func (e *ElasticSearch) IndexService(l *logger.CustomLogger, ncfg *index.NatsConfig) (*index.Service, error) {
	if e.is != nil {
		return e.is, nil
	}

	options := []index.Option{}

	if !e.UseRemoteIndexer {
		l.Info().Msg("setting up bulk indexer")

		es, err := e.NewClient(l)
		if err != nil {
			return nil, err
		}

		bi, err := e.NewBulkIndexer(es)
		if err != nil {
			return nil, err
		}

		options = append(options, index.SetBulkIndexer(bi))
	}

	if ncfg != nil {
		options = append(options, index.SetNatsConfiguration(ncfg))
	}

	is, err := index.NewService(options...)
	if err != nil {
		return nil, err
	}

	if !e.UseRemoteIndexer {
		err := is.Start(context.Background(), 1)
		if err != nil {
			return nil, err
		}
	}

	if e.Metrics {
		expvar.Publish("hub3-index-service", expvar.Func(func() interface{} { m := is.Metrics(); return m }))
	}

	return is, nil
}
