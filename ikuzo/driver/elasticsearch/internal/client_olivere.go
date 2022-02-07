package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/delving/hub3/ikuzo/logger"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

type OlivereConfig struct {
	Urls             []string
	Logger           *zerolog.Logger
	TimeoutInSeconds int // only respected when Client == nil
	HTTPRetries      int // only respected when Client == nil
	Client           *http.Client
	UserName         string
	Password         string
	EnableTrace      bool
	EnableInfo       bool
}

func (cfg *OlivereConfig) HasAuthentication() bool {
	return len(cfg.UserName) > 0 && len(cfg.Password) > 0
}

// ListIndexes returns a list of all the ElasticSearch Indices.
// func ListIndexes() ([]string, error) {
// return ESClient().IndexNames()
// }

func NewOlivereClient(cfg *OlivereConfig) *elastic.Client {
	if cfg.Client == nil {
		// TODO(kiivihal): add support for leveled logger
		cfg.Client = NewClient(cfg.HTTPRetries, cfg.TimeoutInSeconds).StandardClient()
	}

	errLog := logger.NewWrapError(cfg.Logger)

	options := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Urls...),                      // set elastic urs from config
		elastic.SetSniff(false),                          // disable sniffing
		elastic.SetHealthcheckInterval(10 * time.Second), // do healthcheck every 10 seconds
		elastic.SetErrorLog(errLog),                      // error log
		elastic.SetHttpClient(cfg.Client),
	}

	if cfg.HasAuthentication() {
		options = append(options, elastic.SetBasicAuth(cfg.UserName, cfg.Password))
	}

	if cfg.EnableTrace {
		traceLog := logger.NewWrapTrace(cfg.Logger)
		options = append(options, elastic.SetTraceLog(traceLog))
	}

	if cfg.EnableInfo {
		infoLog := logger.NewWrapInfo(cfg.Logger)
		options = append(options, elastic.SetInfoLog(infoLog)) // info log
	}

	c, err := elastic.NewClient(options...)
	if err != nil {
		fmt.Printf("Unable to connect to ElasticSearch. %s\n", err)
	}

	return c
}

// ESClient creates or returns an ElasticSearch Client.
// This function should always be used to perform any ElasticSearch action.
// func ESClient() *elastic.Client {
// if client == nil {
// if config.Config.ElasticSearch.Enabled {
// // setting up execution context
// ctx = context.Background()

// // setup ElasticSearch client
// client = createESClient()
// } else {
// config.Config.Logger.Fatal().
// Str("component", "elasticsearch").
// Msg("FATAL: trying to call elasticsearch when not enabled.")
// }
// }

// return client
// }
