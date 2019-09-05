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
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/mapping"
	elastic "github.com/olivere/elastic"
)

const (
	fragmentIndexFmt = "%s_frag"
)

// CustomRetrier for configuring the retrier for the ElasticSearch client.
type CustomRetrier struct {
	backoff elastic.Backoff
}

func init() {
	stdlog.SetFlags(stdlog.LstdFlags | stdlog.Lshortfile)
}

var (
	client *elastic.Client
	ctx    context.Context
)

// ESClient creates or returns an ElasticSearch Client.
// This function should always be used to perform any ElasticSearch action.
func ESClient() *elastic.Client {
	if client == nil {
		if config.Config.ElasticSearch.Enabled {
			// setting up execution context
			ctx = context.Background()

			// setup ElasticSearch client
			client = createESClient()
			//defer client.Stop()
			ensureESIndex(config.Config.ElasticSearch.IndexName, false)
			ensureESIndex(fmt.Sprintf(fragmentIndexFmt, config.Config.ElasticSearch.IndexName), false)
		} else {
			stdlog.Fatal("FATAL: trying to call elasticsearch when not enabled.")
		}
	}
	return client
}

func IndexReset(index string) error {
	if index == "" {
		index = config.Config.ElasticSearch.IndexName
	}
	ensureESIndex(index, true)
	ensureESIndex(fmt.Sprintf(fragmentIndexFmt, index), true)
	return nil
}

func ensureESIndex(index string, reset bool) {
	if index == "" {
		index = config.Config.ElasticSearch.IndexName
	}
	exists, err := ESClient().IndexExists(index).Do(ctx)
	if err != nil {
		// Handle error
		stdlog.Fatalf("unable to find index for %s: %#v", index, err)
	}
	if exists && reset {
		deleteIndex, err := ESClient().DeleteIndex(index).Do(ctx)
		if err != nil {
			stdlog.Fatal(err)
		}
		if !deleteIndex.Acknowledged {
			stdlog.Printf("Unable to delete index %s", index)
		}
		exists = false
	}

	if !exists {
		// Create a new index.
		indexMapping := mapping.ESMapping
		if config.Config.ElasticSearch.IndexV1 {
			indexMapping = mapping.V1ESMapping
		}
		if strings.HasSuffix(index, "_frag") {
			indexMapping = mapping.ESFragmentMapping
		}
		createIndex, err := client.
			CreateIndex(index).
			BodyJson(
				fmt.Sprintf(
					indexMapping,
					config.Config.ElasticSearch.Shards,
					config.Config.ElasticSearch.Replicas,
				),
			).
			Do(ctx)
		if err != nil {
			// Handle error
			stdlog.Fatal(err)
		}
		if !createIndex.Acknowledged {
			stdlog.Println(createIndex.Acknowledged)
			// Not acknowledged
		}

	}

	// add mapping updates
	updateIndex, err := elastic.NewIndicesPutMappingService(client).
		Index(index).
		Type("doc").
		BodyString(mapping.ESMappingUpdate).
		Do(ctx)
	if err != nil {
		stdlog.Fatalf("unable to patch ES mapping: %#v", err)
		return
	}
	if !updateIndex.Acknowledged {
		stdlog.Println(updateIndex.Acknowledged)
		// Not acknowledged
	}
	return
}

// ListIndexes returns a list of all the ElasticSearch Indices.
func ListIndexes() ([]string, error) {
	return ESClient().IndexNames()
}

func createESClient() *elastic.Client {
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Config.ElasticSearch.Urls...), // set elastic urs from config
		elastic.SetSniff(false),                             // disable sniffing
		elastic.SetHealthcheckInterval(10 * time.Second),    // do healthcheck every 10 seconds
		elastic.SetRetrier(NewCustomRetrier()),              // set custom retrier that tries 5 times. Default is 0
		// todo replace with logrus logger later
		elastic.SetErrorLog(stdlog.New(os.Stderr, "ELASTIC ", stdlog.LstdFlags)), // error log
	}

	if config.Config.ElasticSearch.HasAuthentication() {
		es := config.Config.ElasticSearch
		options = append(options, elastic.SetBasicAuth(es.UserName, es.Password))
	}
	if config.Config.ElasticSearch.EnableTrace {
		options = append(options, elastic.SetTraceLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags)))
	}
	if config.Config.ElasticSearch.EnableInfo {
		options = append(options, elastic.SetInfoLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags))) // info log
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
		return 0, false, errors.New("Elasticsearch or network down")
	}

	// Stop after 5 retries
	if retry >= 5 {
		return 0, false, nil
	}

	// Let the backoff strategy decide how long to wait and whether to stop
	wait, stop := r.backoff.Next(retry)
	return wait, stop, nil
}
