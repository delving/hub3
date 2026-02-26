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

func TestResultParser_toSemanticDoc(t *testing.T) {
	rp := &ResultParser{}

	t.Run("plain document passes through unchanged", func(t *testing.T) {
		source := json.RawMessage(`{"@id": "doc1", "dc_title": "Test", "@type": ["edm:ProvidedCHO"]}`)

		doc, err := rp.toSemanticDoc(source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if doc["@id"] != "doc1" {
			t.Errorf("expected @id='doc1', got %v", doc["@id"])
		}

		if doc["dc_title"] != "Test" {
			t.Errorf("expected dc_title='Test', got %v", doc["dc_title"])
		}
	})

	t.Run("FragmentGraph is converted to semantic view", func(t *testing.T) {
		// Minimal FragmentGraph with resources that will produce a semantic view.
		source := json.RawMessage(`{
			"meta": {
				"orgID": "test-org",
				"spec": "test-spec",
				"hubID": "test-org_test-spec_1",
				"docType": "fragmentGraph",
				"entryURI": "http://example.org/resource/1",
				"namedGraphURI": "http://example.org/resource/1/graph",
				"modified": 1000
			},
			"resources": [
				{
					"id": "http://example.org/resource/1",
					"types": ["http://www.europeana.eu/schemas/edm/ProvidedCHO"],
					"entries": [
						{
							"searchLabel": "dc_title",
							"predicate": "http://purl.org/dc/elements/1.1/title",
							"value": "Night Watch",
							"entryType": "literal",
							"order": 0
						}
					]
				}
			]
		}`)

		doc, err := rp.toSemanticDoc(source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The result should be a semantic view, not the raw FragmentGraph.
		// It should have @id and the dc_title field.
		if _, hasResources := doc["resources"]; hasResources {
			t.Error("expected semantic view, but got raw FragmentGraph (has 'resources' key)")
		}

		if _, hasMeta := doc["meta"]; hasMeta {
			t.Error("expected semantic view, but got raw FragmentGraph (has 'meta' key)")
		}

		// Should have @id from the resource.
		if doc["@id"] == nil {
			t.Error("expected @id in semantic view")
		}
	})

	t.Run("FragmentGraph with cycle detection", func(t *testing.T) {
		// Two resources that reference each other.
		source := json.RawMessage(`{
			"meta": {
				"orgID": "test-org",
				"spec": "test-spec",
				"hubID": "test-org_test-spec_cycle",
				"docType": "fragmentGraph",
				"entryURI": "http://example.org/resource/A",
				"namedGraphURI": "http://example.org/resource/A/graph",
				"modified": 1000
			},
			"resources": [
				{
					"id": "http://example.org/resource/A",
					"types": ["http://www.europeana.eu/schemas/edm/ProvidedCHO"],
					"entries": [
						{
							"searchLabel": "dc_title",
							"predicate": "http://purl.org/dc/elements/1.1/title",
							"value": "Resource A",
							"entryType": "literal",
							"order": 0
						},
						{
							"searchLabel": "dcterms_isPartOf",
							"predicate": "http://purl.org/dc/terms/isPartOf",
							"@id": "http://example.org/resource/B",
							"entryType": "resourceType",
							"order": 1
						}
					]
				},
				{
					"id": "http://example.org/resource/B",
					"types": ["ore:Aggregation"],
					"entries": [
						{
							"searchLabel": "dc_title",
							"predicate": "http://purl.org/dc/elements/1.1/title",
							"value": "Collection B",
							"entryType": "literal",
							"order": 0
						},
						{
							"searchLabel": "dcterms_hasPart",
							"predicate": "http://purl.org/dc/terms/hasPart",
							"@id": "http://example.org/resource/A",
							"entryType": "resourceType",
							"order": 1
						}
					]
				}
			]
		}`)

		doc, err := rp.toSemanticDoc(source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should not have raw FragmentGraph fields.
		if _, hasResources := doc["resources"]; hasResources {
			t.Error("expected semantic view, got raw FragmentGraph")
		}

		// The response should be finite (cycle detection prevents infinite nesting).
		// Marshal and check it's valid JSON (no stack overflow from infinite recursion).
		data, err := json.Marshal(doc)
		if err != nil {
			t.Fatalf("failed to marshal result (possible infinite nesting): %v", err)
		}

		if len(data) == 0 {
			t.Error("expected non-empty JSON output")
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
