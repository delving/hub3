package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/delving/hub3/hub3/fragments"
	elastic "github.com/olivere/elastic/v7"
)

func expandRequest(sort string) *fragments.SearchRequest {
	return &fragments.SearchRequest{OrgID: "nde", OrgIDKey: "meta.orgID", FacetExpand: "schema_creator", FacetSort: sort, FacetLimit: 10}
}

// A record can name one value in several entries (a photographer in two
// roles); the expanded facet, like the others, counts records.
func TestExpandFacetCountsRecords(t *testing.T) {
	for _, sort := range []string{"", "count"} {
		cfg, err := newNestedQueryAgg(expandRequest(sort), nil)
		if err != nil {
			t.Fatal(err)
		}
		src, _ := cfg.aggs.Source()
		js, _ := json.Marshal(src)
		if !strings.Contains(string(js), `"records":{"reverse_nested":{}}`) {
			t.Errorf("sort %q: aggregation counts no records: %s", sort, js)
		}
		if sort == "count" && !strings.Contains(string(js), `"order":[{"records":"desc"}]`) {
			t.Errorf("sort by count does not order by records: %s", js)
		}
	}

	for _, tc := range []struct{ sort, aggs string }{
		{"", `{"comp_agg":{"doc_count":3000,"nested_agg":{"doc_count":9000,"filtered_labels":{"doc_count":6000,
			"unique_searchLabels":{"value":1},
			"values_agg":{"buckets":[{"key":{"sortkey":null,"value":"fotografen"},"doc_count":5956,"records":{"doc_count":2961}}]}}}}}`},
		{"count", `{"comp_agg":{"doc_count":3000,"nested_agg":{"doc_count":9000,"filtered_labels":{"doc_count":6000,
			"unique_searchLabels":{"value":1},
			"values_agg":{"buckets":[{"key":"fotografen","doc_count":5956,"records":{"doc_count":2961}}]}}}}}`},
	} {
		sr := expandRequest(tc.sort)
		cfg, err := newNestedQueryAgg(sr, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp := &elastic.SearchResult{}
		if err := json.Unmarshal([]byte(tc.aggs), &resp.Aggregations); err != nil {
			t.Fatal(err)
		}
		facet := decodeExpandedFacet(resp, sr, cfg)
		if len(facet.Links) != 1 || facet.Links[0].Count != 2961 || facet.Total != 1 {
			t.Errorf("sort %q: %+v, want fotografen with 2961 records", tc.sort, facet)
		}
	}
}

// An extra filter, such as an integration layer's own search, narrows the
// records whose values are counted.
func TestExpandFacetExtraFilter(t *testing.T) {
	extra := elastic.NewTermQuery("meta.spec", "uu")
	cfg, err := newNestedQueryAgg(expandRequest(""), extra)
	if err != nil {
		t.Fatal(err)
	}
	for name, q := range map[string]interface{ Source() (interface{}, error) }{"query": cfg.q, "aggregation": cfg.aggs} {
		src, _ := q.Source()
		js, _ := json.Marshal(src)
		if !strings.Contains(string(js), `{"term":{"meta.spec":"uu"}}`) {
			t.Errorf("%s lacks the extra filter: %s", name, js)
		}
	}
}
