package elasticsearch

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/olivere/elastic/v7"
)

// Client is a client to interact with the ElasticSearch cluster.
// Must be used via NewClient to have proper initialisation.
type Client struct {
	cfg    *Config
	search *elastic.Client
	index  *elasticsearch.Client
}

func NewClient(cfg *Config) (*Client, error) {
	if validErr := cfg.Valid(); validErr != nil {
		return nil, validErr
	}

	searchCfg := internal.OlivereConfig{
		Urls:             []string{},
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
	}

	indexCfg := internal.ElasticConfig{
		Urls:       []string{},
		UserName:   cfg.UserName,
		Password:   cfg.Password,
		Metrics:    true,
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

func (c *Client) Ping() (*esapi.Response, error) {
	return c.index.Info()
}
