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
	"fmt"
	"sync"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/domain"
	es "github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/service/x/esproxy"
	"github.com/delving/hub3/ikuzo/service/x/index"
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
	client *es.Client
	// BulkIndexer
	bi *esutil.BulkIndexer
	// IndexService
	is *index.Service
	// base of the index aliases
	IndexName string
	// if non-empty digital objects will be indexed in a dedicated v2 index
	DigitalObjectSuffix string
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
	FastHTTP bool `json:"fastHTTP"`
	// OrphanWait is the duration in seconds that the orphanDelete will wait for the cluster to be in sync
	OrphanWait int
	// once makes sure that createmapping is only run once
	once sync.Once
	// UserName is the BasicAuth username
	UserName string `json:"userName"`
	// Password is the BasicAuth password
	Password string `json:"password"`
}

func (e *ElasticSearch) AddOptions(cfg *Config) error {
	if !e.Enabled || len(e.Urls) == 0 {
		return nil
	}

	client, err := e.NewCustomClient(&cfg.logger)
	if err != nil {
		return fmt.Errorf("unable to create elasticsearch.Client: %w", err)
	}

	if _, pingErr := client.Ping(); pingErr != nil {
		return fmt.Errorf("unable to connect to elasticsearch; %w", pingErr)
	}

	if e.Proxy {
		proxySvc, proxyErr := esproxy.NewService(
			esproxy.SetElasticClient(client),
			esproxy.SetEnableIntrospect(cfg.Logging.DevMode),
		)
		if proxyErr != nil {
			return fmt.Errorf("unable to create ES proxy: %w", proxyErr)
		}
		cfg.options = append(
			cfg.options,
			ikuzo.RegisterService(proxySvc),
		)

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
		bulk.SetBlobConfig(cfg.Bulk.Minio),
		bulk.SetLogRequests(cfg.Bulk.StoreRequests),
	)
	if bulkErr != nil {
		return fmt.Errorf("unable to create bulk service; %w", bulkErr)
	}

	cfg.options = append(
		cfg.options,
		ikuzo.RegisterService(bulkSvc),
	)

	if e.UseRemoteIndexer {
		orgSvc, err := cfg.getOrganisationService("")
		if err != nil {
			return err
		}

		cfgs, err := orgSvc.Configs(context.TODO())
		if err != nil {
			return err
		}

		_, err = e.CreateDefaultMappings(client, cfgs, true, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *ElasticSearch) NewBulkIndexer(orgs *organization.Service) (*esutil.BulkIndexer, error) {
	if e.bi != nil {
		return e.bi, nil
	}

	esClient, err := e.NewCustomClient(e.logger)
	if err != nil {
		return nil, err
	}

	cfgs, err := orgs.Configs(context.TODO())
	if err != nil {
		return nil, err
	}

	bi, err := esClient.NewBulkIndexer(cfgs, e.Workers)
	if err != nil {
		return nil, err
	}

	e.bi = &bi

	return e.bi, nil
}

func (e *ElasticSearch) CreateDefaultMappings(esClient *es.Client, orgs []domain.OrganizationConfig, withAlias, withReset bool) ([]string, error) {
	e.once.Do(func() {
		if _, err := esClient.CreateDefaultMappings(orgs, withAlias, withReset); err != nil {
			log.Error().Err(err).Msg("unable to create default mappings for elasticsearch")
		}
	})

	return []string{}, nil
}

func (e *ElasticSearch) NewCustomClient(l *logger.CustomLogger) (*es.Client, error) {
	if e.client != nil {
		return e.client, nil
	}

	cfg := es.DefaultConfig()
	cfg.Urls = e.Urls
	cfg.UserName = e.UserName
	cfg.Password = e.Password
	cfg.MaxRetries = e.MaxRetries
	cfg.Timeout = e.ClientTimeOut
	logConv := l.With().Logger()
	cfg.Logger = &logConv

	esClient, err := es.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	e.client = esClient

	return e.client, nil
}

func (e *ElasticSearch) IndexService(cfg *Config, ncfg *index.NatsConfig) (*index.Service, error) {
	if e.is != nil {
		return e.is, nil
	}

	orgs, err := cfg.getOrganisationService("")
	if err != nil {
		return nil, err
	}

	options := []index.Option{
		index.SetOrganisationService(orgs),
	}

	if !e.UseRemoteIndexer || ncfg == nil {
		cfg.logger.Info().Msg("setting up bulk indexer")

		bi, bulkErr := e.NewBulkIndexer(orgs)
		if bulkErr != nil {
			return nil, bulkErr
		}

		options = append(
			options,
			index.SetBulkIndexer(*bi, ncfg == nil),
			index.SetOrphanWait(e.OrphanWait),
			index.SetDisableMetrics(!e.Metrics),
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

	return e.is, nil
}
