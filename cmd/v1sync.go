// Copyright © 2018 Delving B.V. <info@delving.eu>
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

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/delving/hub3/hub3/mapping"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	pb "gopkg.in/cheggaaa/pb.v1"
	elastic "gopkg.in/olivere/elastic.v5"
)

// v1syncCmd represents the v1sync command
var (
	v1syncCmd = &cobra.Command{
		Use:   "v1sync",
		Short: "sync between two v5 elasticsearch index clusters",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := synchronise()
			if err != nil {
				stdlog.Println(synchronise())
			}
		},
	}

	sourceURL      string
	sourceIndex    string
	sourceUserName string
	sourcePassword string
	backup         bool

	targetUserName string
	targetPassword string
	targetURL      string
	targetOrgID    string
	trace          bool
)

// CustomRetrier for configuring the retrier for the ElasticSearch client.
type CustomRetrier struct {
	backoff elastic.Backoff
}

func init() {

	v1syncCmd.Flags().StringVarP(&sourceURL, "sourceURL", "s", "", "source URL for ElasticSearch")
	v1syncCmd.Flags().StringVarP(&sourceIndex, "sourceIndex", "i", "", "source indexname")
	v1syncCmd.Flags().StringVarP(&sourceUserName, "sourceUser", "", "", "source username for ElasticSearch")
	v1syncCmd.Flags().StringVarP(&sourcePassword, "sourcePassword", "", "", "source password ElasticSearch")
	v1syncCmd.Flags().BoolVarP(&backup, "backup", "", false, "backup records to disk")

	v1syncCmd.Flags().StringVarP(&targetURL, "targetURL", "t", "", "target URL for ElasticSearch")
	v1syncCmd.Flags().StringVarP(&targetUserName, "user", "u", "", "target username for ElasticSearch")
	v1syncCmd.Flags().StringVarP(&targetPassword, "password", "p", "", "target password ElasticSearch")
	v1syncCmd.Flags().StringVarP(&targetOrgID, "orgID", "o", "", "target orgID and indexname")
	v1syncCmd.Flags().BoolVarP(&trace, "trace", "v", false, "show trace information")

	// set required
	v1syncCmd.MarkFlagRequired("sourceURL")
	v1syncCmd.MarkFlagRequired("sourceIndex")
	//v1syncCmd.MarkFlagRequired("orgID")
	//v1syncCmd.MarkFlagRequired("targetURL")

	RootCmd.AddCommand(v1syncCmd)

}

func getESClient(url, user, password string) (*elastic.Client, error) {
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(url),                                                      // set elastic urs from config
		elastic.SetSniff(false),                                                  // disable sniffing
		elastic.SetHealthcheckInterval(10 * time.Second),                         // do healthcheck every 10 seconds
		elastic.SetRetrier(NewCustomRetrier()),                                   // set custom retrier that tries 5 times. Default is 0
		elastic.SetErrorLog(stdlog.New(os.Stderr, "ELASTIC ", stdlog.LstdFlags)), // error log
	}

	if user != "" && password != "" {
		options = append(options, elastic.SetBasicAuth(user, password))
	}
	if trace {
		options = append(options, elastic.SetInfoLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags))) // info log
		options = append(options, elastic.SetTraceLog(stdlog.New(os.Stdout, "", stdlog.LstdFlags)))
	}

	return elastic.NewClient(options...)
}

func createBulkProcessor(ctx context.Context) (*elastic.BulkProcessor, error) {
	if backup {
		return nil, nil
	}
	client, err := getESClient(targetURL, targetUserName, targetPassword)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create target ES client")
	}

	stdlog.Printf("created client")

	indices, err := client.IndexNames()
	if err != nil {
		return nil, err
	}
	stdlog.Printf("target indices: %#v", indices)
	stdlog.Printf("list indices")

	err = ensureESIndex(client, targetOrgID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create index")
	}
	indices, _ = client.IndexNames()
	stdlog.Printf("target indices: %#v", indices)

	if backup {
		return nil, nil
	}
	return client.BulkProcessor().
		Name("ES bulk worker").
		Workers(4).
		BulkActions(2000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). // commit every 30s
		Do(ctx)
}

func synchronise() error {
	bulkCtx := context.Background()

	p, err := createBulkProcessor(bulkCtx)
	if err != nil {
		return errors.Wrap(err, "unable to create target bulk processor ")
	}

	ctx := context.Background()
	sourceClient, err := getESClient(sourceURL, sourceUserName, sourcePassword)
	if err != nil {
		return errors.Wrap(err, "unable to create source elasticsearch client")
	}
	// Count total and setup progress
	total, err := sourceClient.Count(sourceIndex).Type("").Do(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to get count from elasticsearch")
	}
	stdlog.Printf("%s has %d records", sourceIndex, total)

	bar := pb.StartNew(int(total))

	// 1st goroutine sends individual hits to channel.
	hits := make(chan *elastic.SearchHit)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		defer close(hits)
		// Initialize scroller. Just don't call Do yet.
		scroll := sourceClient.Scroll(sourceIndex).Size(100)
		for {
			results, err := scroll.Do(ctx)
			if err == io.EOF {
				return nil // all results retrieved
			}
			if err != nil {
				return err // something went wrong
			}

			// Send the hits to the hits channel
			for _, hit := range results.Hits.Hits {
				select {
				case hits <- hit:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	})

	// 2nd goroutine receives hits and deserializes them.
	//
	for i := 0; i < 10; i++ {
		g.Go(func() error {
			for hit := range hits {
				//stdlog.Printf("hit: \n %#v", hit)
				switch backup {
				case true:
					storeHit(hit)
				case false:
					r := elastic.NewBulkIndexRequest().
						Index(targetOrgID).
						Type(hit.Type).
						Id(hit.Id).
						Doc(hit.Source)
					p.Add(r)
				}

				bar.Increment()

				// Terminate early?
				select {
				default:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	// Check whether any goroutines failed.
	if err := g.Wait(); err != nil {
		return err
	}

	if !backup {
		err = p.Flush()
		if err != nil {
			return errors.Wrap(err, "unable to flush records to index")
		}
		err = p.Close()
		if err != nil {
			return errors.Wrap(err, "unable to close bulk processor")
		}

	}

	// Done.
	bar.Finish()
	return nil
}

func storeHit(hit *elastic.SearchHit) error {
	outputDir := filepath.Join("/tmp", hit.Index)
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}
	b, err := json.Marshal(hit)
	if err != nil {
		return err
	}
	fname := filepath.Join(outputDir, fmt.Sprintf("%s.json", hit.Id))
	return ioutil.WriteFile(fname, b, 0644)
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

func ensureESIndex(client *elastic.Client, index string) error {
	ctx := context.Background()
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		// Handle error
		return err
	}

	if !exists {
		// Create a new index.
		esMapping := mapping.V1ESMapping
		createIndex, err := client.CreateIndex(index).BodyJson(esMapping).Do(ctx)
		if err != nil {
			// Handle error
			return err
		}
		if !createIndex.Acknowledged {
			stdlog.Println(createIndex.Acknowledged)
			// Not acknowledged
		}

		return nil
	}
	return nil
}
