package elasticsearch

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal"
	"github.com/elastic/go-elasticsearch/v8"
	externalAPI "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

type Response = externalAPI.Response

// Client is a client to interact with the ElasticSearch cluster.
// Must be used via NewClient to have proper initialisation.
type Client struct {
	cfg            *Config
	search         *elastic.Client
	index          *elasticsearch.Client
	disableMetrics bool
	log            zerolog.Logger
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
		EnableTrace:      false,
		EnableInfo:       false,
	}

	client := Client{
		cfg:    cfg,
		search: internal.NewOlivereClient(&searchCfg),
		log:    cfg.Logger.With().Str("svc", "elasticsearch").Logger(),
	}

	indexCfg := internal.ElasticConfig{
		Urls:       cfg.Urls,
		UserName:   cfg.UserName,
		Password:   cfg.Password,
		Metrics:    !cfg.DisableMetrics,
		MaxRetries: cfg.MaxRetries,
		Timeout:    cfg.Timeout,
		FastHTTP:   false,
	}

	esclient, err := internal.NewElasticClient(&indexCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create elastic search client; %w", err)
	}

	client.index = esclient

	return &client, nil
}

func (c *Client) Ping() (*externalAPI.Response, error) {
	return c.index.Info()
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
