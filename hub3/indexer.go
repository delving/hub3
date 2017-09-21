package hub3

// The Indexer contains all services elements for indexing RDF data in ElasticSearch

import (
	"context"
	"fmt"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// RdfObject holds all the fields for the RDF triples that will be indexed by ElasticSearch
type RdfObject struct {
	Subject             string    `json:"subject"`                       // URI of the subject
	SubjectClass        []string  `json:"class"`                         // RDF types of the subject
	Predicate           string    `json:"predicate"`                     // URI of the predicate
	SearchLabel         string    `json:"searchLabel"`                   // Label of predicate. Used for user facing searching
	Object              string    `json:"object"`                        // URI or Literal value of the object the object
	ObjectLang          string    `json:"language,omitempty"`            // The RDF language of the object
	ObjectContentType   string    `json:"objectContentType,omitempty"`   // The XSD:type of the Object Literal value
	IsResource          bool      `json:"isResource"`                    // Boolean to determine if the RdfObject is a Literal or URI reference
	Value               string    `json:"value"`                         // RDF label of the resource. This can be a URI, a label or an inlined label from the URI.
	LatLong             float64   `json:"geoHash,omitempty"`             // A field that can be used for searching
	Polygon             []float64 `json:"polygon,omitempty"`             // A field that contains geo polygons
	Facet               string    `json:"facet,omitempty"`               // Raw non-analysed field of value that can be used for facetting and aggregations. Will not contain a URI.
	Level               int       `json:"level"`                         // The level of the triple compared to the root subject. This is used for relevance ranking in the query.
	ReferrerSubject     string    `json:"refererSubject,omitempty"`      // if on level 2 or 3 list the URI of the referring subject.
	ReferrerPredicate   string    `json:"referrerPredicate,omitempty"`   // the predicate of the referring label
	ReferrerSearchLabel string    `json:"referrerSearchLabel,omitempty"` // the searchLabel of the predicate. This can be used for searching
	NamedGraph          string    `json:"namedGraph,omitempty"`          // the NamedGraph that this object belongs to
	ResourceSortOrder   int       `json:"sortOrder,omitempty"`           // the order in which the resource (for the subject) is sorted when inlined from the referer

}

// RdfRecord holds all the fields for the result of a SPARQL query for a subject.
// The SPARQL query contains data in three levels. The each triple gets assigned a level for additional weighting at search time.
type RdfRecord struct {
	HubID       string      `json:"hubId"`
	SourceURI   string      `json:"sourceUri"`
	DataSetName string      `json:"spec"`
	DataSetURI  string      `json:"datasetUri"`
	Tag         []string    `json:"tag"`
	RecordType  string      `json:"recordType"`
	Modified    time.Time   `json:"modified"`
	Created     time.Time   `json:"created"`
	Triples     []RdfObject `json:"triples"`
}

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
