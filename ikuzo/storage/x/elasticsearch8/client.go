// Package elasticsearch8 provides search stores using the official
// go-elasticsearch/v8 client directly, without the olivere wrapper.
package elasticsearch8

import (
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"
)

// Config holds the configuration for creating a new Client.
type Config struct {
	Addresses   []string
	Username    string
	Password    string
	IndexPrefix string
	MaxRetries  int
	Logger      zerolog.Logger
}

// Client wraps an elasticsearch.Client with convenience methods
// for index naming and logging.
type Client struct {
	es          *elasticsearch.Client
	indexPrefix string
	log         zerolog.Logger
}

// NewClient creates a new Client from the given Config.
func NewClient(cfg Config) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses:  cfg.Addresses,
		Username:   cfg.Username,
		Password:   cfg.Password,
		MaxRetries: cfg.MaxRetries,
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create elasticsearch8 client: %w", err)
	}

	return &Client{
		es:          es,
		indexPrefix: cfg.IndexPrefix,
		log:         cfg.Logger.With().Str("svc", "elasticsearch8").Logger(),
	}, nil
}

// NewClientFromExisting wraps an existing elasticsearch.Client.
func NewClientFromExisting(es *elasticsearch.Client, logger zerolog.Logger) *Client {
	return &Client{
		es:  es,
		log: logger.With().Str("svc", "elasticsearch8").Logger(),
	}
}

// IndexName returns the Elasticsearch index name for the given orgID.
// The convention is strings.ToLower(orgID) + "v2". When an IndexPrefix
// is configured, it is prepended with an underscore separator.
func (c *Client) IndexName(orgID string) string {
	name := strings.ToLower(orgID) + "v2"
	if c.indexPrefix != "" {
		return c.indexPrefix + "_" + name
	}

	return name
}

// ES returns the underlying elasticsearch.Client for direct API access.
func (c *Client) ES() *elasticsearch.Client {
	return c.es
}
