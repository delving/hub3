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

// The Indexer contains all services elements for indexing RDF data in ElasticSearch

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/delving/rapid-saas/config"
	//elastic "github.com/olivere/elastic"
	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	service   *elastic.BulkProcessorService
	processor *elastic.BulkProcessor
	once      sync.Once
)

// CreateBulkProcessor creates an Elastic BulkProcessorService
func CreateBulkProcessor(ctx context.Context) *elastic.BulkProcessor {
	p, err := service.Do(ctx)
	if err != nil {
		log.Printf("Unable to connect start BulkProcessor. %s", err)
	}
	return p
}

// CreateBulkProcessorService creates a service instance
func CreateBulkProcessorService() *elastic.BulkProcessorService {
	return ESClient().BulkProcessor().
		Name("RAPID-backgroundworker").
		Workers(4).
		BulkActions(2000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). // commit every 30s
		//After(elastic.BulkAfterFunc{afterFn}). // after Execution callback
		After(afterFn). // after Execution callback
		//Before(beforeFn).
		Stats(true) // enable statistics

}

func beforeFn(executionID int64, requests []elastic.BulkableRequest) {
	//log.Println("starting bulk.")
}

func afterFn(executionID int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	log.Println("After processor")
	if config.Config.ElasticSearch.IndexV1 && response.Errors {
		log.Println("Errors in bulk request")
		for _, item := range response.Failed() {
			log.Printf("errored item: %#v errors: %#v", item, item.Error)
		}
	}
}

// FlushIndexProcesser flushes all workers and creates a new consistent index snapshot
func FlushIndexProcesser() error {
	return IndexingProcessor().Flush()
}

// IndexingProcessor returns a pointer to the running BulkProcessor
func IndexingProcessor() *elastic.BulkProcessor {
	if !config.Config.ElasticSearch.Enabled {
		log.Fatal("When elasticsearch is not enabled IndexingProcessor should never be called.")
	}
	once.Do(func() {
		// Setup a bulk processor service
		service = CreateBulkProcessorService()

		// Setup a bulk processor
		log.Println("Creating BulkProcessorService")
		processor = CreateBulkProcessor(ctx)
	})
	return processor
}

// BulkIndexStatistics returns access to statistics in an indexing snapshot
func BulkIndexStatistics(p *elastic.BulkProcessor) elastic.BulkProcessorStats {
	stats := p.Stats()
	fmt.Printf("Number of times flush has been invoked: %d\n", stats.Flushed)
	fmt.Printf("Number of times workers committed reqs: %d\n", stats.Committed)
	fmt.Printf("Number of requests indexed            : %d\n", stats.Indexed)
	fmt.Printf("Number of requests reported as created: %d\n", stats.Created)
	fmt.Printf("Number of requests reported as updated: %d\n", stats.Updated)
	fmt.Printf("Number of requests reported as success: %d\n", stats.Succeeded)
	fmt.Printf("Number of requests reported as failed : %d\n", stats.Failed)

	for i, w := range stats.Workers {
		fmt.Printf("Worker %d: Number of requests queued: %d\n", i, w.Queued)
		fmt.Printf("           Last response time       : %v\n", w.LastDuration)
	}
	return stats
}
