package fragments

import (
	"encoding/json"
	"strings"
	"testing"

	elastic "github.com/olivere/elastic/v7"
)

// A record can hold several entries with the same value (a photographer named
// in two roles); a facet count is the number of records, not of entries.
func TestFacetCountsRecords(t *testing.T) {
	fb, err := NewFacetURIBuilder("", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, byName := range []bool{false, true} {
		agg, err := CreateAggregationBySearchLabel("resources.entries", &FacetField{Field: "schema_creator", Size: 10, ByName: byName}, false, fb)
		if err != nil {
			t.Fatal(err)
		}
		src, _ := agg.Source()
		js, _ := json.Marshal(src)
		if !strings.Contains(string(js), `"reverse_nested":{}`) {
			t.Errorf("byName=%v: aggregation counts no records: %s", byName, js)
		}
	}

	res := &elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 3000}}}
	if err := json.Unmarshal([]byte(`{"schema_creator": {"doc_count": 3000, "filter": {"doc_count": 9000,
		"inner": {"doc_count": 6000,
			"value": {"buckets": [{"key": "fotografen", "doc_count": 5956, "records": {"doc_count": 2961}}]},
			"sortedValues": {"buckets": [{"key": {"sortkey": null, "value": "verteller"}, "doc_count": 1182, "records": {"doc_count": 1110}}]}
		}}}}`), &res.Aggregations); err != nil {
		t.Fatal(err)
	}
	qfs, err := DecodeFacets(res, fb)
	if err != nil || len(qfs) != 1 {
		t.Fatalf("DecodeFacets: %v, %d facets", err, len(qfs))
	}
	got := map[string]int64{}
	for _, l := range qfs[0].Links {
		got[l.Value] = l.Count
	}
	if got["fotografen"] != 2961 || got["verteller"] != 1110 {
		t.Errorf("counts %v, want fotografen 2961 and verteller 1110 records", got)
	}
}
