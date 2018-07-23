package ead

import "github.com/olivere/elastic"

type NodeBuilder struct {
	fragments chan<- *elastic.BulkIndexRequest
	records   chan<- *elastic.BulkIndexRequest
	spec      string
	orgID     string
	//r = elastic.NewBulkIndexRequest().
	//Index(c.Config.ElasticSearch.IndexName).
	//Type(fragments.DocType).
	//RetryOnConflict(3).
	//Id(action.HubID).
	//Doc(fb.Doc())
}
