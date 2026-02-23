package elasticsearch8

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestResultParser_ParseSearchResponse(t *testing.T) {
	rp := &ResultParser{}

	t.Run("basic search response", func(t *testing.T) {
		esResponse := `{
			"took": 5,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{"_id": "doc1", "_score": 1.5, "_source": {"@id": "doc1", "title": "Test 1"}},
					{"_id": "doc2", "_score": 1.0, "_source": {"title": "Test 2"}}
				]
			}
		}`

		result, err := rp.ParseSearchResponse([]byte(esResponse))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Total != 2 {
			t.Errorf("expected Total=2, got %d", result.Total)
		}

		if len(result.Results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result.Results))
		}

		if len(result.ResultIDs) != 2 {
			t.Fatalf("expected 2 result IDs, got %d", len(result.ResultIDs))
		}

		// doc1 has @id in source, should keep it.
		if id, ok := result.Results[0]["@id"]; !ok || id != "doc1" {
			t.Errorf("expected result[0][@id]='doc1', got %v", id)
		}

		// doc2 missing @id in source, should get it from _id.
		if id, ok := result.Results[1]["@id"]; !ok || id != "doc2" {
			t.Errorf("expected result[1][@id]='doc2', got %v", id)
		}

		// Check metadata.
		took, ok := result.Metadata["elasticsearch_took_ms"]
		if !ok {
			t.Fatal("expected metadata to contain elasticsearch_took_ms")
		}

		if took != 5 {
			t.Errorf("expected elasticsearch_took_ms=5, got %v", took)
		}

		// Check ResultIDs.
		if result.ResultIDs[0] != "doc1" || result.ResultIDs[1] != "doc2" {
			t.Errorf("unexpected ResultIDs: %v", result.ResultIDs)
		}
	})

	t.Run("empty hits", func(t *testing.T) {
		esResponse := `{
			"took": 1,
			"hits": {
				"total": {"value": 0, "relation": "eq"},
				"hits": []
			}
		}`

		result, err := rp.ParseSearchResponse([]byte(esResponse))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Total != 0 {
			t.Errorf("expected Total=0, got %d", result.Total)
		}

		if len(result.Results) != 0 {
			t.Errorf("expected 0 results, got %d", len(result.Results))
		}

		if len(result.ResultIDs) != 0 {
			t.Errorf("expected 0 result IDs, got %d", len(result.ResultIDs))
		}
	})

	t.Run("malformed source is skipped", func(t *testing.T) {
		esResponse := `{
			"took": 2,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{"_id": "good", "_score": 1.0, "_source": {"title": "Good"}},
					{"_id": "bad", "_score": 0.5, "_source": "not-a-json-object"}
				]
			}
		}`

		result, err := rp.ParseSearchResponse([]byte(esResponse))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Only the valid doc should be in results; the malformed one is skipped.
		if len(result.Results) != 1 {
			t.Errorf("expected 1 result (malformed skipped), got %d", len(result.Results))
		}

		if len(result.ResultIDs) != 1 {
			t.Errorf("expected 1 result ID, got %d", len(result.ResultIDs))
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := rp.ParseSearchResponse([]byte(`{invalid`))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestResultParser_ParseFacets(t *testing.T) {
	rp := &ResultParser{}

	t.Run("flat terms aggregation", func(t *testing.T) {
		aggsJSON := map[string]json.RawMessage{
			"dc_creator": json.RawMessage(`{
				"buckets": [
					{"key": "Picasso", "doc_count": 10},
					{"key": "Monet", "doc_count": 5}
				],
				"sum_other_doc_count": 3
			}`),
		}

		facets := []semantic.FacetRequest{
			{Field: "dc_creator"},
		}

		results := rp.ParseFacets(aggsJSON, facets)

		if len(results) != 1 {
			t.Fatalf("expected 1 facet result, got %d", len(results))
		}

		fr := results[0]
		if fr.Field != "dc_creator" {
			t.Errorf("expected field='dc_creator', got %s", fr.Field)
		}

		if len(fr.Values) != 2 {
			t.Fatalf("expected 2 values, got %d", len(fr.Values))
		}

		if fr.Values[0].Value != "Picasso" || fr.Values[0].Count != 10 {
			t.Errorf("unexpected first value: %+v", fr.Values[0])
		}

		if fr.Values[1].Value != "Monet" || fr.Values[1].Count != 5 {
			t.Errorf("unexpected second value: %+v", fr.Values[1])
		}

		if fr.Values[0].Label != "Picasso" {
			t.Errorf("expected Label='Picasso', got %s", fr.Values[0].Label)
		}

		if fr.SumOther != 3 {
			t.Errorf("expected SumOther=3, got %d", fr.SumOther)
		}
	})

	t.Run("nested aggregation", func(t *testing.T) {
		aggsJSON := map[string]json.RawMessage{
			"dc_creator": json.RawMessage(`{
				"doc_count": 100,
				"filtered": {
					"doc_count": 50,
					"values": {
						"buckets": [
							{"key": "Picasso", "doc_count": 10}
						],
						"sum_other_doc_count": 0
					}
				}
			}`),
		}

		facets := []semantic.FacetRequest{
			{Field: "dc_creator"},
		}

		results := rp.ParseFacets(aggsJSON, facets)

		if len(results) != 1 {
			t.Fatalf("expected 1 facet result, got %d", len(results))
		}

		fr := results[0]
		if len(fr.Values) != 1 {
			t.Fatalf("expected 1 value, got %d", len(fr.Values))
		}

		if fr.Values[0].Value != "Picasso" || fr.Values[0].Count != 10 {
			t.Errorf("unexpected value: %+v", fr.Values[0])
		}

		if fr.SumOther != 0 {
			t.Errorf("expected SumOther=0, got %d", fr.SumOther)
		}
	})

	t.Run("missing aggregation for facet", func(t *testing.T) {
		aggsJSON := map[string]json.RawMessage{
			"dc_subject": json.RawMessage(`{"buckets": [{"key": "art", "doc_count": 1}], "sum_other_doc_count": 0}`),
		}

		facets := []semantic.FacetRequest{
			{Field: "dc_creator"}, // Not in aggs.
		}

		results := rp.ParseFacets(aggsJSON, facets)
		if len(results) != 0 {
			t.Errorf("expected 0 results for missing field, got %d", len(results))
		}
	})

	t.Run("empty aggregations", func(t *testing.T) {
		results := rp.ParseFacets(nil, []semantic.FacetRequest{{Field: "dc_creator"}})
		if len(results) != 0 {
			t.Errorf("expected 0 results for nil aggs, got %d", len(results))
		}
	})

	t.Run("empty facet requests", func(t *testing.T) {
		aggsJSON := map[string]json.RawMessage{
			"dc_creator": json.RawMessage(`{"buckets": [], "sum_other_doc_count": 0}`),
		}

		results := rp.ParseFacets(aggsJSON, nil)
		if len(results) != 0 {
			t.Errorf("expected 0 results for nil facets, got %d", len(results))
		}
	})
}

func TestResultParser_ParseGetResponse(t *testing.T) {
	rp := &ResultParser{}

	t.Run("source with @id", func(t *testing.T) {
		getResp := `{
			"_id": "doc1",
			"found": true,
			"_source": {"@id": "http://example.com/doc1", "title": "Test"}
		}`

		doc, err := rp.ParseGetResponse([]byte(getResp))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc["@id"] != "http://example.com/doc1" {
			t.Errorf("expected @id='http://example.com/doc1', got %v", doc["@id"])
		}

		if doc["title"] != "Test" {
			t.Errorf("expected title='Test', got %v", doc["title"])
		}
	})

	t.Run("source missing @id", func(t *testing.T) {
		getResp := `{
			"_id": "doc2",
			"found": true,
			"_source": {"title": "No ID"}
		}`

		doc, err := rp.ParseGetResponse([]byte(getResp))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc["@id"] != "doc2" {
			t.Errorf("expected @id='doc2', got %v", doc["@id"])
		}
	})

	t.Run("not found", func(t *testing.T) {
		getResp := `{
			"_id": "missing",
			"found": false
		}`

		_, err := rp.ParseGetResponse([]byte(getResp))
		if err == nil {
			t.Fatal("expected error for not-found document")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := rp.ParseGetResponse([]byte(`{}`))
		if err == nil {
			t.Fatal("expected error for empty response (found=false)")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := rp.ParseGetResponse([]byte(`{invalid`))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}
