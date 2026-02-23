package semantic

import (
	"context"
	"testing"
)

// mockIncludeProvider is a test double.
type mockIncludeProvider struct {
	name   string
	result any
	err    error
}

func (m *mockIncludeProvider) Name() string { return m.name }
func (m *mockIncludeProvider) Provide(_ context.Context, _ *IncludeRequest) (any, error) {
	return m.result, m.err
}

func TestIncludeProvider_Interface(t *testing.T) {
	// Verify the interface can be implemented
	var provider IncludeProvider = &mockIncludeProvider{
		name:   "relatedItems",
		result: map[string]any{"@type": "hydra:Collection", "totalItems": 5},
	}

	if provider.Name() != "relatedItems" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "relatedItems")
	}

	result, err := provider.Provide(context.Background(), &IncludeRequest{
		DocumentID: "test-id",
		Document:   map[string]any{"@id": "test-id"},
	})
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestIncludeRequest_Params(t *testing.T) {
	req := &IncludeRequest{
		DocumentID: "abc",
		Document:   map[string]any{"@id": "abc", "@type": "edm:ProvidedCHO"},
		Params:     map[string]string{"count": "5", "fields": "dc:title,dc:creator"},
	}

	if req.DocumentID != "abc" {
		t.Errorf("DocumentID = %q, want %q", req.DocumentID, "abc")
	}
	if req.Params["count"] != "5" {
		t.Errorf("Params[count] = %q, want %q", req.Params["count"], "5")
	}
}

func TestDefaultSimilarOptions(t *testing.T) {
	opts := DefaultSimilarOptions()
	if opts.Count != 5 {
		t.Errorf("Count = %d, want 5", opts.Count)
	}
	if opts.MinTermFreq != 2 {
		t.Errorf("MinTermFreq = %d, want 2", opts.MinTermFreq)
	}
	if opts.MinDocFreq != 5 {
		t.Errorf("MinDocFreq = %d, want 5", opts.MinDocFreq)
	}
}

func TestSimilarOptions_DefaultFields(t *testing.T) {
	opts := DefaultSimilarOptions()
	if len(opts.Fields) != 0 {
		t.Errorf("Fields should be empty by default, got %v", opts.Fields)
	}
}
