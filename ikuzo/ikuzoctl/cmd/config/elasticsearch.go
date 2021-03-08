// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/service/x/index"
	eshub "github.com/delving/hub3/ikuzo/storage/x/elasticsearch"
	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch/mapping"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/rs/zerolog/log"
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
	// IndexTypes options are v1, v2, fragment
	IndexTypes []string
	// use FastHTTP transport for communication with the ElasticSearch cluster
	FastHTTP bool
	// OrphanWait is the duration in seconds that the orphanDelete will wait for the cluster to be in sync
	OrphanWait int
	// once makes sure that createmapping is only run once
	once sync.Once
	// UserName is the BasicAuth username
	UserName string `json:"userName"`
	// Password is the BasicAuth password
	Password string `json:"password"`
}

func (e *ElasticSearch) normalizedIndexName() string {
	return strings.ToLower(e.IndexName)
}

func (e *ElasticSearch) AddOptions(cfg *Config) error {
	if !e.Enabled || len(e.Urls) == 0 {
		return nil
	}

	client, err := e.NewClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("unable to create elasticsearch.Client: %w", err)
	}

	if _, infoErr := client.Info(); infoErr != nil {
		return fmt.Errorf("unable to connect to elasticsearch; %w", infoErr)
	}

	if e.Proxy {
		esProxy, proxyErr := eshub.NewProxy(client)
		if proxyErr != nil {
			return fmt.Errorf("unable to create ES proxy: %w", proxyErr)
		}

		cfg.options = append(cfg.options, ikuzo.SetElasticSearchProxy(esProxy))
	}

	// when not in datanode mode no service should be started
	if !cfg.IsDataNode() {
		return nil
	}

	// enable bulk indexer
	is, isErr := cfg.GetIndexService()
	if isErr != nil {
		return fmt.Errorf("unable to create index service; %w", isErr)
	}

	postHooks, phErr := cfg.getPostHookServices()
	if phErr != nil {
		return fmt.Errorf("unable to create posthook service; %w", phErr)
	}

	bulkSvc, bulkErr := bulk.NewService(
		bulk.SetIndexService(is),
		bulk.SetIndexTypes(e.IndexTypes...),
		bulk.SetPostHookService(postHooks...),
	)
	if bulkErr != nil {
		return fmt.Errorf("unable to create bulk service; %w", isErr)
	}

	cfg.options = append(
		cfg.options,
		ikuzo.SetBulkService(bulkSvc),
		ikuzo.SetShutdownHook("elasticsearch", is),
	)

	_, err = e.CreateDefaultMappings(client, true, false)
	if err != nil {
		return err
	}

	return nil
}

func (e *ElasticSearch) ResetAll(w http.ResponseWriter, r *http.Request) {
	// reset elasticsearch
	_, err := e.createDefaultMappings(e.client, true, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// reset Key Value Store
	models.ResetStorm()

	// reset EAD cache
	models.ResetEADCache()
}

func (e *ElasticSearch) CreateDefaultMappings(es *elasticsearch.Client, withAlias, withReset bool) ([]string, error) {
	e.once.Do(func() {
		if _, err := e.createDefaultMappings(es, withAlias, withReset); err != nil {
			log.Error().Err(err).Msg("unable to create default mappings for elasticsearch")
		}
	})

	return []string{}, nil
}

func (e *ElasticSearch) createDefaultMappings(es *elasticsearch.Client, withAlias, withReset bool) ([]string, error) {
	mappings := map[string]func(shards, replicas int) string{}

	for _, indexType := range e.IndexTypes {
		switch indexType {
		case "v1":
			mappings[fmt.Sprintf("%sv1", e.normalizedIndexName())] = mapping.V1ESMapping
		case "v2":
			mappings[fmt.Sprintf("%sv2", e.normalizedIndexName())] = mapping.V2ESMapping
		case "fragments":
			mappings[fmt.Sprintf("%sv2_frag", e.normalizedIndexName())] = mapping.FragmentESMapping
		default:
			log.Warn().Msgf("ignoring unknown indexType %s during mapping creation", indexType)
		}
	}

	indexNames := []string{}

	for indexName, m := range mappings {
		if withReset {
			storedIndexName, aliasErr := eshub.AliasGet(es, indexName)
			if aliasErr != nil && !errors.Is(aliasErr, eshub.ErrAliasNotFound) {
				return []string{}, aliasErr
			}

			if storedIndexName != "" {
				if err := eshub.AliasDelete(es, storedIndexName, indexName); err != nil {
					log.Error().Err(err).Str("alias", indexName).
						Str("index", storedIndexName).Msg("unable to delete alias")

					return []string{}, err
				}

				resp, deleteErr := es.Indices.Delete([]string{storedIndexName})
				if deleteErr != nil && resp.IsError() {
					log.Error().Err(deleteErr).Str("alias", indexName).
						Str("index", storedIndexName).Msg("unable to delete index")
					return []string{}, deleteErr
				}
			}
		}

		createName, err := eshub.IndexCreate(
			es,
			indexName,
			m(e.Shards, e.Replicas),
			withAlias,
		)

		if err != nil && !errors.Is(err, eshub.ErrIndexAlreadyCreated) {
			return []string{}, err
		}

		if errors.Is(err, eshub.ErrIndexAlreadyCreated) {
			if strings.HasSuffix(indexName, "v2") {
				valid, err := eshub.IsMappingValid(es, createName)
				if err != nil {
					return []string{}, err
				}

				if !valid {
					if err := eshub.MappingUpdate(es, createName, mapping.V2MappingUpdate()); err != nil {
						log.Error().Err(err).Msg("unable to apply v2 mapping update")
						return []string{}, err
					}

					log.Warn().Str("index", createName).Msg("applying elasticsearch mapping update")
				}
			}
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

	cfg := elasticsearch.Config{
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

		// Custom rs/zerolog structured logger
		//
		Logger: l,
	}

	if e.hasAuthentication() {
		cfg.Username = e.UserName
		cfg.Password = e.Password
	}

	if e.FastHTTP {
		// Custom transport based on fasthttp
		cfg.Transport = &eshub.Transport{}
	}

	client, err := elasticsearch.NewClient(cfg)

	// Publish client metrics to expvar
	if e.Metrics {
		expvar.Publish("go-elasticsearch", expvar.Func(func() interface{} { m, _ := client.Metrics(); return m }))
	}

	e.client = client

	return e.client, err
}

// hasAuthentication returns if ElasticSearch has authentication enabled.
func (e *ElasticSearch) hasAuthentication() bool {
	return len(e.UserName) > 0 && len(e.Password) > 0
}

func (e *ElasticSearch) NewBulkIndexer(es *elasticsearch.Client) (esutil.BulkIndexer, error) {
	if e.bi != nil {
		return e.bi, nil
	}

	if es == nil {
		return nil, fmt.Errorf("cannot start BulkIndexer without valid es client")
	}

	// create default mappings
	_, err := e.CreateDefaultMappings(es, true, false)
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

func (e *ElasticSearch) IndexService(cfg *Config, ncfg *index.NatsConfig) (*index.Service, error) {
	if e.is != nil {
		return e.is, nil
	}

	var err error

	options := []index.Option{}

	if !e.UseRemoteIndexer || ncfg == nil {
		cfg.logger.Info().Msg("setting up bulk indexer")

		es, clientErr := e.NewClient(&cfg.logger)
		if clientErr != nil {
			return nil, clientErr
		}

		bi, bulkErr := e.NewBulkIndexer(es)
		if bulkErr != nil {
			return nil, bulkErr
		}

		options = append(
			options,
			index.SetBulkIndexer(bi, ncfg == nil),
			index.SetOrphanWait(e.OrphanWait),
		)
	}

	if ncfg != nil {
		options = append(options, index.SetNatsConfiguration(ncfg))
	}

	postHooks, phErr := cfg.getPostHookServices()
	if phErr != nil {
		return nil, fmt.Errorf("unable to create posthook service; %w", phErr)
	}

	if len(postHooks) != 0 {
		options = append(options, index.SetPostHookService(postHooks...))
	}

	e.is, err = index.NewService(options...)
	if err != nil {
		return nil, err
	}

	if !e.UseRemoteIndexer && ncfg != nil {
		err := e.is.Start(context.Background(), 1)
		if err != nil {
			return nil, err
		}
	}

	if e.Metrics {
		expvar.Publish("hub3-index-service", expvar.Func(func() interface{} { m := e.is.Metrics(); return m }))
	}

	return e.is, nil
}
