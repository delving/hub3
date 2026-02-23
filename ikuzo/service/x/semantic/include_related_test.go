package semantic

import (
	"context"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

type mockSimilarStore struct {
	result   *semantic.SearchResult
	err      error
	lastID   string
	lastOpts *semantic.SimilarOptions
}

func (m *mockSimilarStore) FindSimilar(_ context.Context, id string, opts *semantic.SimilarOptions, _ *semantic.ResourceConfig) (*semantic.SearchResult, error) {
	m.lastID = id
	m.lastOpts = opts
	return m.result, m.err
}

func TestRelatedItemsProvider_Name(t *testing.T) {
	p := NewRelatedItemsProvider(nil)
	if p.Name() != "relatedItems" {
		t.Errorf("Name() = %q, want %q", p.Name(), "relatedItems")
	}
}

func TestRelatedItemsProvider_NoStore(t *testing.T) {
	p := NewRelatedItemsProvider(nil)
	_, err := p.Provide(context.Background(), &semantic.IncludeRequest{
		DocumentID: "test",
		Document:   map[string]any{},
	})
	if err == nil {
		t.Error("expected error when no store configured")
	}
}

func TestRelatedItemsProvider_EmptyResults(t *testing.T) {
	store := &mockSimilarStore{
		result: &semantic.SearchResult{
			Total:   0,
			Results: []map[string]any{},
		},
	}
	p := NewRelatedItemsProvider(store)

	result, err := p.Provide(context.Background(), &semantic.IncludeRequest{
		DocumentID: "test",
		Document:   map[string]any{},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty results")
	}
}

func TestRelatedItemsProvider_WithResults(t *testing.T) {
	store := &mockSimilarStore{
		result: &semantic.SearchResult{
			Total: 2,
			Results: []map[string]any{
				{"@id": "doc-1", "dc:title": "Related Doc 1"},
				{"@id": "doc-2", "dc:title": "Related Doc 2"},
			},
		},
	}
	p := NewRelatedItemsProvider(store)

	result, err := p.Provide(context.Background(), &semantic.IncludeRequest{
		DocumentID: "test",
		Document:   map[string]any{},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	collection, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result is %T, want map[string]any", result)
	}

	if collection["@type"] != "hydra:Collection" {
		t.Errorf("@type = %v, want hydra:Collection", collection["@type"])
	}

	members, ok := collection["member"].([]any)
	if !ok {
		t.Fatalf("member is %T", collection["member"])
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

func TestRelatedItemsProvider_ParsesParams(t *testing.T) {
	store := &mockSimilarStore{
		result: &semantic.SearchResult{
			Total:   1,
			Results: []map[string]any{{"@id": "doc-1"}},
		},
	}
	p := NewRelatedItemsProvider(store)

	_, err := p.Provide(context.Background(), &semantic.IncludeRequest{
		DocumentID: "test-doc",
		Document:   map[string]any{},
		Params: map[string]string{
			"count":  "10",
			"fields": "dc:title,dc:creator",
		},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if store.lastOpts.Count != 10 {
		t.Errorf("Count = %d, want 10", store.lastOpts.Count)
	}
	if len(store.lastOpts.Fields) != 2 {
		t.Errorf("Fields = %v, want 2 fields", store.lastOpts.Fields)
	}
}

func TestRelatedItemsProvider_CapsCountAt20(t *testing.T) {
	store := &mockSimilarStore{
		result: &semantic.SearchResult{
			Total:   1,
			Results: []map[string]any{{"@id": "doc-1"}},
		},
	}
	p := NewRelatedItemsProvider(store)

	_, err := p.Provide(context.Background(), &semantic.IncludeRequest{
		DocumentID: "test",
		Document:   map[string]any{},
		Params:     map[string]string{"count": "100"},
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if store.lastOpts.Count != 20 {
		t.Errorf("Count = %d, want 20 (capped)", store.lastOpts.Count)
	}
}
