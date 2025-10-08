package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal"
	"github.com/elastic/go-elasticsearch/v8"
	externalAPI "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/olivere/elastic/v7"
	"github.com/opensearch-project/opensearch-go"
	"github.com/rs/zerolog"
)

type Response = externalAPI.Response

// Client is a client to interact with the ElasticSearch cluster.
// Must be used via NewClient to have proper initialisation.
type Client struct {
	cfg            *Config
	search         *elastic.Client
	es             *elasticsearch.Client
	opensearch     *opensearch.Client
	disableMetrics bool
	log            zerolog.Logger
	pit            *PITClient
}

func NewClient(cfg *Config) (*Client, error) {
	if validErr := cfg.Valid(); validErr != nil {
		return nil, validErr
	}

	defaults := DefaultConfig()
	if cfg.Urls == nil {
		cfg.Urls = defaults.Urls
	}

	if cfg.Logger == nil {
		cfg.Logger = defaults.Logger
	}

	searchCfg := internal.OlivereConfig{
		Urls:             cfg.Urls,
		Logger:           cfg.Logger,
		TimeoutInSeconds: cfg.Timeout,
		HTTPRetries:      cfg.MaxRetries,
		UserName:         cfg.UserName,
		Password:         cfg.Password,
		EnableTrace:      true,
		EnableInfo:       true,
	}

	client := Client{
		cfg:    cfg,
		search: internal.NewOlivereClient(&searchCfg),
		log:    cfg.Logger.With().Str("svc", "elasticsearch").Logger(),
	}

	var err error

	client.pit, err = NewPITClient(cfg.Urls, cfg.UserName, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to create pit client; %w", err)
	}

	indexCfg := internal.ElasticConfig{
		Urls:       cfg.Urls,
		UserName:   cfg.UserName,
		Password:   cfg.Password,
		Metrics:    !cfg.DisableMetrics,
		MaxRetries: cfg.MaxRetries,
		Timeout:    cfg.Timeout,
	}

	esclient, err := internal.NewElasticClient(&indexCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create elastic search client; %w", err)
	}

	client.es = esclient

	return &client, nil
}

func (c *Client) Ping() (*externalAPI.Response, error) {
	return c.es.Ping()
}

// Bulk sends a bulk request to Elasticsearch.
// `bulkData` should be a JSON-formatted byte slice containing bulk operations.
func (c *Client) Bulk(ctx context.Context, bulkData io.Reader) error {
	// Use the Bulk API
	res, err := c.es.Bulk(bulkData, c.es.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	// Check for request errors
	if res.IsError() {
		return fmt.Errorf("bulk request failed: %s", res.String())
	}

	// Success
	// fmt.Println("Bulk request completed successfully.")
	return nil
}

// Bulk sends a bulk request to Elasticsearch.
// `bulkData` should be a JSON-formatted byte slice containing bulk operations.
func (c *Client) BulkWithResponse(ctx context.Context, bulkData io.Reader) (io.Reader, error) {
	// Use the Bulk API
	res, err := c.es.Bulk(bulkData, c.es.Bulk.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	// Check for request errors
	if res.IsError() {
		return nil, fmt.Errorf("bulk request failed: %s", res.String())
	}

	// Success
	var buf bytes.Buffer
	_, err = io.Copy(&buf, res.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Bulk request completed successfully.")
	return &buf, nil
}

// CreateDefaultMappings creates index mappings for all supplied organizations
func (c *Client) CreateDefaultMappings(orgs []domain.OrganizationConfig, withAlias, withReset bool) (indices []string, err error) {
	indexNames := []string{}

	for _, cfg := range orgs {
		indices, err := c.createDefaultMappings(cfg, withAlias, withReset)
		if err != nil {
			return []string{}, fmt.Errorf("error with default mapping %s; %w", cfg.OrgID(), err)
		}

		indexNames = append(indexNames, indices...)
	}

	return indexNames, nil
}

// SearchClient returns the underlying olivere elastic.Client for advanced search operations.
func (c *Client) SearchClient() *elastic.Client {
	return c.search
}
