package fragments

import (
	"context"
	fmt "fmt"
	"log"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/index"
	elastic "gopkg.in/olivere/elastic.v5"
)

// TreeStats holds all the information for a navigation tree for a Dataset
type TreeStats struct {
	Spec     string
	Leafs    int64
	Depth    []StatCounter
	Children []StatCounter
	Type     []StatCounter
}

// StatCounter holds value counters for statistics overviews
type StatCounter struct {
	Value    string `json:"value"`
	DocCount int    `json:"docCount"`
}

// createStatCounters creates counters from an ElasticSearch aggregation
func createStatCounters(aggs elastic.Aggregations, name string) ([]StatCounter, error) {
	counters := []StatCounter{}
	aggCount, found := aggs.Terms(name)
	if !found {
		log.Printf("Expected to find %s aggregations but got: %v", name, aggs)
		return counters, fmt.Errorf("expected %s aggregrations", name)
	}
	for _, keyCount := range aggCount.Buckets {
		var key string
		switch keyCount.Key.(type) {
		case float64:
			key = fmt.Sprintf("%d", int(keyCount.Key.(float64)))
		case string:
			key = keyCount.Key.(string)
		}
		counters = append(counters, StatCounter{
			Value:    key,
			DocCount: int(keyCount.DocCount),
		})
	}
	return counters, nil
}

// CreateTreeStats creates a statistics overview
func CreateTreeStats(ctx context.Context, spec string) (*TreeStats, error) {
	tree := &TreeStats{
		Spec: spec,
	}

	// Counters
	//depth := []StatCounter{}

	// Aggregations
	depthAgg := elastic.NewTermsAggregation().Field("tree.depth").Size(30).OrderByCountDesc()
	childAgg := elastic.NewTermsAggregation().Field("tree.children").Size(100).OrderByCountDesc()
	typeAgg := elastic.NewTermsAggregation().Field("tree.type").Size(100).OrderByCountDesc()

	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewMatchPhraseQuery(c.Config.ElasticSearch.SpecKey, spec),
		elastic.NewTermQuery("meta.docType", FragmentGraphDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID),
	)
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.IndexName).
		Query(q).
		Size(0).
		Aggregation("depth", depthAgg).
		Aggregation("children", childAgg).
		Aggregation("type", typeAgg).
		Do(ctx)
	if err != nil {
		log.Printf("Unable to get TreeStat for dataset %s; %s", spec, err)
		return nil, err
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return nil, err
	}
	tree.Leafs = res.Hits.TotalHits

	aggs := res.Aggregations
	buckets := []string{"depth", "children", "type"}

	for _, a := range buckets {
		counter, err := createStatCounters(aggs, a)
		if err != nil {
			return tree, err
		}

		switch a {
		case "depth":
			tree.Depth = counter
		case "children":
			tree.Children = counter
		case "type":
			tree.Type = counter
		}
	}

	return tree, nil
}
