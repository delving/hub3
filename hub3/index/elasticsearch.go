// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/logger"
	elastic "github.com/olivere/elastic/v7"
)

const (
	fragmentIndexFmt = "%s_frag"
)

// CustomRetrier for configuring the retrier for the ElasticSearch client.
type CustomRetrier struct {
	backoff elastic.Backoff
}

var (
	client *elastic.Client
	ctx    context.Context
)

// ESClient creates or returns an ElasticSearch Client.
// This function should always be used to perform any ElasticSearch action.
func ESClientDeprecated() *elastic.Client {
	if client == nil {
		if config.Config.ElasticSearch.Enabled {
			// setting up execution context
			ctx = context.Background()

			// setup ElasticSearch client
			client = createESClient()
		} else {
			config.Config.Logger.Fatal().
				Str("component", "elasticsearch").
				Msg("FATAL: trying to call elasticsearch when not enabled.")
		}
	}

	return client
}

// ListIndexes returns a list of all the ElasticSearch Indices.
// func ListIndexes() ([]string, error) {
// return ESClient().IndexNames()
// }

func createESClient() *elastic.Client {
	timeout := time.Duration(config.Config.ElasticSearch.RequestTimeout) * time.Second
	httpclient := &http.Client{
		Timeout: timeout,
	}

	errLog := logger.NewWrapError(config.Config.Logger)

	options := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Config.ElasticSearch.Urls...), // set elastic urs from config
		elastic.SetSniff(false),                             // disable sniffing
		elastic.SetHealthcheckInterval(10 * time.Second),    // do healthcheck every 10 seconds
		elastic.SetRetrier(NewCustomRetrier()),              // set custom retrier that tries 5 times. Default is 0
		elastic.SetErrorLog(errLog),                         // error log
		elastic.SetHttpClient(httpclient),
	}

	if config.Config.ElasticSearch.HasAuthentication() {
		es := config.Config.ElasticSearch
		options = append(options, elastic.SetBasicAuth(es.UserName, es.Password))
	}

	if config.Config.ElasticSearch.EnableTrace {
		traceLog := logger.NewWrapTrace(config.Config.Logger)
		options = append(options, elastic.SetTraceLog(traceLog))
	}

	if config.Config.ElasticSearch.EnableInfo {
		infoLog := logger.NewWrapInfo(config.Config.Logger)
		options = append(options, elastic.SetInfoLog(infoLog)) // info log
	}

	if client == nil {
		c, err := elastic.NewClient(options...)
		if err != nil {
			fmt.Printf("Unable to connect to ElasticSearch. %s\n", err)
		}

		client = c
	}

	return client
}

// NewCustomRetrier creates custom retrier for elasticsearch
func NewCustomRetrier() *CustomRetrier {
	return &CustomRetrier{
		backoff: elastic.NewExponentialBackoff(10*time.Millisecond, 8*time.Second),
	}
}

// Retry defines how the retrier should deal with retrying the elasticsearch connection.
func (r *CustomRetrier) Retry(
	ctx context.Context,
	retry int,
	req *http.Request,
	resp *http.Response,
	err error) (time.Duration, bool, error) {
	// Fail hard on a specific error
	if err == syscall.ECONNREFUSED {
		return 0, false, errors.New("elasticsearch or network down")
	}

	// Stop after 5 retries
	if retry >= 5 {
		return 0, false, nil
	}

	// Let the backoff strategy decide how long to wait and whether to stop
	wait, stop := r.backoff.Next(retry)

	return wait, stop, nil
}
