package fragments

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"log"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
	elastic "github.com/olivere/elastic"
)

// TreeStats holds all the information for a navigation tree for a Dataset
type TreeStats struct {
	Spec     string
	Leafs    int64
	Depth    []StatCounter
	Children []StatCounter
	Type     []StatCounter
	PhysDesc []StatCounter
}

// TreeDescription describes the meta-information for an Archival Finding Aid tree
type TreeDescription struct {
	Name        string   `json:"name"`
	InventoryID string   `json:"inventoryID"`
	Abstract    []string `json:"abstract"`
	Owner       string   `json:"owner"`
	Period      []string `json:"period"`
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

func TreeNode(ctx context.Context, hubID string) (*Tree, error) {
	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewTermQuery("tree.hubID", hubID),
	)
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.IndexName).
		Query(q).
		Size(10).
		Do(ctx)
	if err != nil {
		log.Printf("Unable to get hubID %s; %s", hubID, err)
		return nil, err
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return nil, err
	}
	if res.TotalHits() == int64(0) {
		log.Printf("Unable to get hubID %s; %s", hubID, err)
		return nil, fmt.Errorf("hudId %s not found", hubID)

	}
	fg, err := decodeFragmentGraph(res.Hits.Hits[0].Source)
	if err != nil {
		return nil, err
	}

	return fg.Tree, nil
}

func decodeFragmentGraph(hit *json.RawMessage) (*FragmentGraph, error) {
	r := new(FragmentGraph)
	if err := json.Unmarshal(*hit, r); err != nil {
		return nil, err
	}
	return r, nil
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
	childAgg := elastic.NewTermsAggregation().Field("tree.childCount").Size(100).OrderByCountDesc()
	typeAgg := elastic.NewTermsAggregation().Field("tree.type").Size(100).OrderByCountDesc()

	fub, err := NewFacetURIBuilder("", []*QueryFilter{})
	if err != nil {
		return nil, err
	}

	//resourceFields := []string{"ead-rdf_physdesc"}
	physDescField, err := NewFacetField("ead-rdf_physdesc")
	if err != nil {
		return nil, err
	}
	physDescAgg, err := CreateAggregationBySearchLabel(
		"resources.entries",
		physDescField,
		false,
		fub,
	)
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
		Aggregation("physdesc", physDescAgg).
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

	physdescCounter := []StatCounter{}
	ct, ok := aggs.Nested("physdesc")
	if ok {
		facet, ok := ct.Filter("filter")
		if ok {
			inner, ok := facet.Filter("inner")
			if ok {
				value, ok := inner.Terms("value")
				if ok {
					for _, keyCount := range value.Buckets {
						physdescCounter = append(physdescCounter, StatCounter{
							Value:    fmt.Sprintf("%s", keyCount.Key),
							DocCount: int(keyCount.DocCount),
						})
					}
				}
			}
		}
	}
	tree.PhysDesc = physdescCounter

	return tree, nil
}
