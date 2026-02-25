package elasticsearch8

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// roundTripFunc adapts a function to http.RoundTripper for mocking.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newMockClient creates a Client backed by a mock HTTP transport.
func newMockClient(fn roundTripFunc) *Client {
	es, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: fn,
	})

	return NewClientFromExisting(es, zerolog.Nop())
}

// testContext returns a context with an Organization set.
func testContext() context.Context {
	ctx := context.Background()

	return domain.SetOrganizationInContext(ctx, domain.Organization{
		ID: domain.OrganizationID("testorg"),
	})
}

// mockResponse creates a simple HTTP response with the given status and body.
func mockResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestStore_implementsSearchStore(t *testing.T) {
	// Compile-time interface check.
	var _ semantic.SearchStore = (*Store)(nil)
}

func TestStore_Search(t *testing.T) {
	cannedResponse := `{
		"took": 5,
		"timed_out": false,
		"hits": {
			"total": {"value": 2, "relation": "eq"},
			"hits": [
				{
					"_id": "doc1",
					"_score": 1.5,
					"_source": {"@id": "http://example.org/1", "dc_title": "First"}
				},
				{
					"_id": "doc2",
					"_score": 1.2,
					"_source": {"@id": "http://example.org/2", "dc_title": "Second"}
				}
			]
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	result, err := store.Search(ctx, &semantic.QueryOptions{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}

	if len(result.ResultIDs) != 2 {
		t.Errorf("expected 2 result IDs, got %d", len(result.ResultIDs))
	}

	if result.ResultIDs[0] != "doc1" {
		t.Errorf("expected first result ID %q, got %q", "doc1", result.ResultIDs[0])
	}

	if result.Results[0]["dc_title"] != "First" {
		t.Errorf("expected first title %q, got %v", "First", result.Results[0]["dc_title"])
	}
}

func TestStore_Search_WithFacets(t *testing.T) {
	cannedResponse := `{
		"took": 3,
		"hits": {
			"total": {"value": 10, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"dc_type": {
				"buckets": [
					{"key": "painting", "doc_count": 5},
					{"key": "sculpture", "doc_count": 3}
				],
				"sum_other_doc_count": 2
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	opts := &semantic.QueryOptions{
		Facets: []semantic.FacetRequest{
			{Field: "dc_type", Limit: 10},
		},
		Peek: true,
	}

	result, err := store.Search(ctx, opts, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 10 {
		t.Errorf("expected total 10, got %d", result.Total)
	}

	if len(result.Facets) != 1 {
		t.Fatalf("expected 1 facet, got %d", len(result.Facets))
	}

	facet := result.Facets[0]

	if facet.Field != "dc_type" {
		t.Errorf("expected facet field %q, got %q", "dc_type", facet.Field)
	}

	if len(facet.Values) != 2 {
		t.Errorf("expected 2 facet values, got %d", len(facet.Values))
	}

	if facet.SumOther != 2 {
		t.Errorf("expected sum_other 2, got %d", facet.SumOther)
	}
}

func TestStore_GetByID(t *testing.T) {
	cannedResponse := `{
		"_index": "testorgv2",
		"_id": "doc1",
		"found": true,
		"_source": {
			"@id": "http://example.org/1",
			"dc_title": "Test Document",
			"dc_creator": "Test Author"
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	doc, err := store.GetByID(ctx, "doc1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc["@id"] != "http://example.org/1" {
		t.Errorf("expected @id %q, got %v", "http://example.org/1", doc["@id"])
	}

	if doc["dc_title"] != "Test Document" {
		t.Errorf("expected dc_title %q, got %v", "Test Document", doc["dc_title"])
	}

	if doc["dc_creator"] != "Test Author" {
		t.Errorf("expected dc_creator %q, got %v", "Test Author", doc["dc_creator"])
	}
}

func TestStore_GetByID_NotFound(t *testing.T) {
	cannedResponse := `{
		"_index": "testorgv2",
		"_id": "missing",
		"found": false
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(404, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	_, err := store.GetByID(ctx, "missing", nil)
	if err == nil {
		t.Fatal("expected error for missing document, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}
}

func TestStore_Health(t *testing.T) {
	cannedResponse := `{
		"cluster_name": "test-cluster",
		"status": "green",
		"number_of_nodes": 1
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())

	err := store.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStore_Health_Unhealthy(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(503, `{"error": "cluster unavailable"}`), nil
	})

	store := NewStore(client, zerolog.Nop())

	err := store.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for unhealthy cluster, got nil")
	}
}

func TestStore_SearchContext(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{}`), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := context.Background()

	searchCtx := &semantic.SearchContext{
		ID:           "ctx-1",
		Token:        "token-abc",
		ResultIDs:    []string{"doc1", "doc2", "doc3"},
		TotalResults: 3,
	}

	// Save
	if err := store.SaveSearchContext(ctx, searchCtx); err != nil {
		t.Fatalf("SaveSearchContext failed: %v", err)
	}

	// Get
	got, err := store.GetSearchContext(ctx, "token-abc")
	if err != nil {
		t.Fatalf("GetSearchContext failed: %v", err)
	}

	if got.ID != "ctx-1" {
		t.Errorf("expected context ID %q, got %q", "ctx-1", got.ID)
	}

	if got.Token != "token-abc" {
		t.Errorf("expected token %q, got %q", "token-abc", got.Token)
	}

	if len(got.ResultIDs) != 3 {
		t.Errorf("expected 3 result IDs, got %d", len(got.ResultIDs))
	}

	if got.TotalResults != 3 {
		t.Errorf("expected total results 3, got %d", got.TotalResults)
	}

	// Delete
	if err := store.DeleteSearchContext(ctx, "token-abc"); err != nil {
		t.Fatalf("DeleteSearchContext failed: %v", err)
	}

	// Get after delete should fail
	_, err = store.GetSearchContext(ctx, "token-abc")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != semantic.ErrContextNotFound {
		t.Errorf("expected ErrContextNotFound, got %v", err)
	}
}

func TestStore_SearchContext_NotFound(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{}`), nil
	})

	store := NewStore(client, zerolog.Nop())

	_, err := store.GetSearchContext(context.Background(), "nonexistent")
	if err != semantic.ErrContextNotFound {
		t.Errorf("expected ErrContextNotFound, got %v", err)
	}
}

func TestStore_Search_NoOrganization(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{}`), nil
	})

	store := NewStore(client, zerolog.Nop())

	// Use a context without organization
	_, err := store.Search(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for missing organization, got nil")
	}

	if !strings.Contains(err.Error(), "no organization") {
		t.Errorf("expected 'no organization' in error, got %q", err.Error())
	}
}

func TestStore_Aggregate(t *testing.T) {
	cannedResponse := `{
		"took": 2,
		"hits": {
			"total": {"value": 50, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"dc_type": {
				"buckets": [
					{"key": "painting", "doc_count": 20},
					{"key": "sculpture", "doc_count": 15},
					{"key": "drawing", "doc_count": 10}
				],
				"sum_other_doc_count": 5
			},
			"dc_creator": {
				"buckets": [
					{"key": "Rembrandt", "doc_count": 8},
					{"key": "Vermeer", "doc_count": 5}
				],
				"sum_other_doc_count": 37
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	facets := []semantic.FacetRequest{
		{Field: "dc_type", Limit: 10},
		{Field: "dc_creator", Limit: 5},
	}

	results, err := store.Aggregate(ctx, facets, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 facet results, got %d", len(results))
	}

	// Check first facet
	if results[0].Field != "dc_type" {
		t.Errorf("expected first facet field %q, got %q", "dc_type", results[0].Field)
	}

	if len(results[0].Values) != 3 {
		t.Errorf("expected 3 values for dc_type, got %d", len(results[0].Values))
	}

	if results[0].SumOther != 5 {
		t.Errorf("expected sum_other 5, got %d", results[0].SumOther)
	}

	// Check second facet
	if results[1].Field != "dc_creator" {
		t.Errorf("expected second facet field %q, got %q", "dc_creator", results[1].Field)
	}

	if len(results[1].Values) != 2 {
		t.Errorf("expected 2 values for dc_creator, got %d", len(results[1].Values))
	}
}

func TestStore_implementsSimilarStore(t *testing.T) {
	// Compile-time interface check.
	var _ semantic.SimilarStore = (*Store)(nil)
}

func TestStore_FindSimilar(t *testing.T) {
	cannedResponse := `{
		"took": 3,
		"timed_out": false,
		"hits": {
			"total": {"value": 3, "relation": "eq"},
			"hits": [
				{
					"_id": "similar-1",
					"_score": 2.5,
					"_source": {"@id": "http://example.org/s1", "dc_title": "Similar One"}
				},
				{
					"_id": "similar-2",
					"_score": 2.1,
					"_source": {"@id": "http://example.org/s2", "dc_title": "Similar Two"}
				},
				{
					"_id": "similar-3",
					"_score": 1.8,
					"_source": {"@id": "http://example.org/s3", "dc_title": "Similar Three"}
				}
			]
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	opts := &semantic.SimilarOptions{
		Count:       5,
		Fields:      []string{"dc:creator", "dc:title"},
		MinTermFreq: 2,
		MinDocFreq:  5,
	}

	result, err := store.FindSimilar(ctx, "doc-123", opts, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}

	if len(result.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(result.Results))
	}

	if len(result.ResultIDs) != 3 {
		t.Errorf("expected 3 result IDs, got %d", len(result.ResultIDs))
	}

	if result.ResultIDs[0] != "similar-1" {
		t.Errorf("expected first result ID %q, got %q", "similar-1", result.ResultIDs[0])
	}

	if result.ResultIDs[1] != "similar-2" {
		t.Errorf("expected second result ID %q, got %q", "similar-2", result.ResultIDs[1])
	}

	if result.ResultIDs[2] != "similar-3" {
		t.Errorf("expected third result ID %q, got %q", "similar-3", result.ResultIDs[2])
	}
}

func TestStore_FindSimilar_Defaults(t *testing.T) {
	cannedResponse := `{
		"took": 2,
		"hits": {
			"total": {"value": 1, "relation": "eq"},
			"hits": [
				{
					"_id": "default-1",
					"_score": 1.0,
					"_source": {"@id": "http://example.org/d1", "dc_title": "Default"}
				}
			]
		}
	}`

	var capturedBody []byte

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Body != nil {
			capturedBody, _ = io.ReadAll(req.Body)
		}

		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	// Pass nil opts to exercise defaults.
	result, err := store.FindSimilar(ctx, "doc-456", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}

	// Verify defaults were applied in the query body.
	var query map[string]any
	if err := json.Unmarshal(capturedBody, &query); err != nil {
		t.Fatalf("failed to unmarshal captured body: %v", err)
	}

	size, ok := query["size"].(float64)
	if !ok || int(size) != 5 {
		t.Errorf("expected default size 5, got %v", query["size"])
	}

	mlt, ok := query["query"].(map[string]any)["more_like_this"].(map[string]any)
	if !ok {
		t.Fatal("expected more_like_this query in body")
	}

	fields, ok := mlt["fields"].([]any)
	if !ok || len(fields) != 1 || fields[0] != "full_text" {
		t.Errorf("expected default fields [full_text], got %v", mlt["fields"])
	}

	minTermFreq, ok := mlt["min_term_freq"].(float64)
	if !ok || int(minTermFreq) != 2 {
		t.Errorf("expected default min_term_freq 2, got %v", mlt["min_term_freq"])
	}

	minDocFreq, ok := mlt["min_doc_freq"].(float64)
	if !ok || int(minDocFreq) != 5 {
		t.Errorf("expected default min_doc_freq 5, got %v", mlt["min_doc_freq"])
	}
}

func TestStore_resolveIndices(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{}`), nil
	})
	store := NewStore(client, zerolog.Nop())

	t.Run("single org returns one index", func(t *testing.T) {
		indices := store.resolveIndices("testorg", nil)
		if len(indices) != 1 {
			t.Fatalf("expected 1 index, got %d", len(indices))
		}
		if indices[0] != "testorgv2" {
			t.Errorf("expected 'testorgv2', got %q", indices[0])
		}
	})

	t.Run("with context orgs returns multiple indices", func(t *testing.T) {
		indices := store.resolveIndices("testorg", []string{"other-org", "third-org"})
		if len(indices) != 3 {
			t.Fatalf("expected 3 indices, got %d", len(indices))
		}
		if indices[0] != "testorgv2" {
			t.Errorf("expected 'testorgv2', got %q", indices[0])
		}
		if indices[1] != "other-orgv2" {
			t.Errorf("expected 'other-orgv2', got %q", indices[1])
		}
		if indices[2] != "third-orgv2" {
			t.Errorf("expected 'third-orgv2', got %q", indices[2])
		}
	})

	t.Run("duplicate org is deduplicated", func(t *testing.T) {
		indices := store.resolveIndices("testorg", []string{"testorg"})
		if len(indices) != 1 {
			t.Fatalf("expected 1 index (duplicate removed), got %d: %v", len(indices), indices)
		}
	})
}

func TestStore_FindSimilar_CapsAt20(t *testing.T) {
	cannedResponse := `{
		"took": 1,
		"hits": {
			"total": {"value": 0, "relation": "eq"},
			"hits": []
		}
	}`

	var capturedBody []byte

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Body != nil {
			capturedBody, _ = io.ReadAll(req.Body)
		}

		return mockResponse(200, cannedResponse), nil
	})

	store := NewStore(client, zerolog.Nop())
	ctx := testContext()

	opts := &semantic.SimilarOptions{
		Count:       100, // way over the max
		MinTermFreq: 2,
		MinDocFreq:  5,
	}

	_, err := store.FindSimilar(ctx, "doc-789", opts, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var query map[string]any
	if err := json.Unmarshal(capturedBody, &query); err != nil {
		t.Fatalf("failed to unmarshal captured body: %v", err)
	}

	size, ok := query["size"].(float64)
	if !ok || int(size) != 20 {
		t.Errorf("expected capped size 20, got %v", query["size"])
	}
}
