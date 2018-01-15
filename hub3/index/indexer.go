package index

// The Indexer contains all services elements for indexing RDF data in ElasticSearch

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	. "bitbucket.org/delving/rapid/config"
	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	service   *elastic.BulkProcessorService
	processor *elastic.BulkProcessor
	once      sync.Once
)

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
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). // commit every 30s
		After(afterFn).                  // after Excecution callback
		//Before(beforeFn).
		Stats(true) // enable statistics

}

func beforeFn(executionId int64, requests []elastic.BulkableRequest) {
	//log.Println("starting bulk.")
}

func afterFn(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	log.Println("After processor")
	if response.Errors {
		log.Println("Errors in bulk request")
		log.Println(response.Failed())
	}
}

// IndexingProcessor returns a pointer to the running BulkProcessor
func IndexingProcessor() *elastic.BulkProcessor {
	if !Config.ElasticSearch.Enabled {
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

// IndexStatistics returns access to statistics in an indexing snapshot
func IndexStatistics(p *elastic.BulkProcessor) elastic.BulkProcessorStats {
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
		fmt.Printf("           Last response time       : %v\n", i, w.LastDuration)
	}
	return stats
}
