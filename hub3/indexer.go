package hub3

// The Indexer contains all services elements for indexing RDF data in ElasticSearch

import (
	"context"
	"fmt"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	service   *elastic.BulkProcessorService
	processor *elastic.BulkProcessor
)

func init() {
	// setup ElasticSearch client
	client = createESClient()

	// Setup a bulk processor service
	service = createBulkProcessorService()

	// Setup a bulk processor
	processor = createBulkProcesor()
}

func createBulkProcesor() *elastic.BulkProcessor {
	p, err := service.Do(context.Background())
	if err != nil {
		// todo: change with proper logging later
		fmt.Printf("Unable to connect start BulkProcessor. %s", err)
	}
	return p
}

func createBulkProcessorService() *elastic.BulkProcessorService {
	return client.BulkProcessor().
		Name("RAPID-backgroundworker-1").
		Workers(2).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). // commit every 30s
		Stats(true)                      // enable statistics
}

// IndexingProcessor returns a pointer to the running BulkProcessor
func IndexingProcessor() *elastic.BulkProcessor {
	return processor
}

// IndexStatistics returns access to statistics in an indexing snapshot
func IndexStatistics() elastic.BulkProcessorStats {
	return processor.Stats()
}
