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
	"syscall"
	"time"

	"bitbucket.org/delving/rapid/config"
	elastic "gopkg.in/olivere/elastic.v5"
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
			ensureESIndex("")
		} else {
			stdlog.Fatal("FATAL: trying to call elasticsearch when not enabled.")
		}
	}
	return client
}

func ensureESIndex(index string) {
	if index == "" {
		index = config.Config.ElasticSearch.IndexName
	}
	exists, err := ESClient().IndexExists(index).Do(ctx)
	if err != nil {
		// Handle error
		stdlog.Fatal(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			// Handle error
			stdlog.Fatal(err)
		}
		if !createIndex.Acknowledged {
			stdlog.Println(createIndex.Acknowledged)
			// Not acknowledged
		}
	}
}

// ListIndexes returns a list of all the ElasticSearch Indices.
func ListIndexes() ([]string, error) {
	return ESClient().IndexNames()
}

func createESClient() *elastic.Client {
	if client == nil {
		c, err := elastic.NewClient(
			elastic.SetURL(config.Config.ElasticSearch.Urls...), // set elastic urs from config
			elastic.SetSniff(false),                             // disable sniffing
			elastic.SetHealthcheckInterval(10*time.Second),      // do healthcheck every 10 seconds
			elastic.SetRetrier(NewCustomRetrier()),              // set custom retrier that tries 5 times. Default is 0
			// todo replace with logrus logger later
			elastic.SetErrorLog(stdlog.New(os.Stderr, "ELASTIC ", stdlog.LstdFlags)), // error log
			elastic.SetInfoLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags)),          // info log
			//elastic.SetTraceLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags)),         // trace log
		)
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

// Create a new index.
// TODO: add other mappings to the default mapping
var mapping = `{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"rdfrecord":{
			"properties":{
				"spec":{
					"type":"string"
				},
				"graph":{
					"index": "no",
					"type": "string",
					"doc_values": false
				}
			}
		}
	}
}`
