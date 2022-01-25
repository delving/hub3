package internal

import (
	"expvar"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticConfig struct {
	Urls       []string
	Client     *http.Client
	UserName   string
	Password   string
	Logger     *logger.CustomLogger
	Metrics    bool
	MaxRetries int
	FastHTTP   bool
	Timeout    int
}

// NewElasticClient returns a github.com/elastic/go-elasticsearch client.
func NewElasticClient(cfg *ElasticConfig) (*elasticsearch.Client, error) {
	retryBackoff := backoff.NewExponentialBackOff()

	innerCfg := elasticsearch.Config{
		// Connect to ElasticSearch URLS
		//
		Addresses: cfg.Urls,

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
		EnableMetrics: cfg.Metrics,

		// Retry up to MaxRetries attempts
		//
		MaxRetries: cfg.MaxRetries,
	}

	if cfg.Logger != nil {
		// Custom rs/zerolog structured logger
		innerCfg.Logger = cfg.Logger
	}

	if cfg.UserName != "" && cfg.Password != "" {
		innerCfg.Username = cfg.UserName
		innerCfg.Password = cfg.Password
	}

	if cfg.FastHTTP {
		// Custom transport based on fasthttp
		innerCfg.Transport = &Transport{}
	}

	client, err := elasticsearch.NewClient(innerCfg)

	// Publish client metrics to expvar
	if cfg.Metrics {
		expvar.Publish("go-elasticsearch", expvar.Func(func() interface{} { m, _ := client.Metrics(); return m }))
	}

	return client, err
}
