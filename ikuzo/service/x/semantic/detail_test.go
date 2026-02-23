package semantic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/go-chi/chi/v5"
)

func TestHandleResourceDetail(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid resource ID",
			path:           "/api/semantic/v1/resource/test-123",
			expectedStatus: 200,
		},
		{
			name:           "valid resource with context token",
			path:           "/api/semantic/v1/resource/test-123?context=ctx_12345",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			svc.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestBuildNavigationContext(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	searchCtx := &semantic.SearchContext{
		Token:        "test-token",
		ResultIDs:    []string{"id1", "id2", "id3", "id4", "id5"},
		TotalResults: 100,
	}

	tests := []struct {
		name        string
		currentID   string
		wantNil     bool
		wantPos     int
		wantHasNext bool
		wantHasPrev bool
	}{
		{
			name:        "first item",
			currentID:   "id1",
			wantNil:     false,
			wantPos:     1,
			wantHasNext: true,
			wantHasPrev: false,
		},
		{
			name:        "middle item",
			currentID:   "id3",
			wantNil:     false,
			wantPos:     3,
			wantHasNext: true,
			wantHasPrev: true,
		},
		{
			name:        "last item",
			currentID:   "id5",
			wantNil:     false,
			wantPos:     5,
			wantHasNext: false,
			wantHasPrev: true,
		},
		{
			name:      "not in results",
			currentID: "unknown",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/resource/"+tt.currentID, nil)
			navCtx := svc.buildNavigationContext(req, tt.currentID, searchCtx)

			if tt.wantNil {
				if navCtx != nil {
					t.Error("Expected nil navigation context")
				}
				return
			}

			if navCtx == nil {
				t.Fatal("Expected non-nil navigation context")
			}

			if navCtx.Position != tt.wantPos {
				t.Errorf("Position = %d, want %d", navCtx.Position, tt.wantPos)
			}

			if navCtx.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", navCtx.HasNext, tt.wantHasNext)
			}

			if navCtx.HasPrevious != tt.wantHasPrev {
				t.Errorf("HasPrevious = %v, want %v", navCtx.HasPrevious, tt.wantHasPrev)
			}

			if navCtx.TotalResults != searchCtx.TotalResults {
				t.Errorf("TotalResults = %d, want %d", navCtx.TotalResults, searchCtx.TotalResults)
			}

			// Check that URLs are generated
			if navCtx.First == "" {
				t.Error("Expected First URL to be set")
			}

			if tt.wantHasNext && navCtx.Next == "" {
				t.Error("Expected Next URL to be set")
			}

			if tt.wantHasPrev && navCtx.Previous == "" {
				t.Error("Expected Previous URL to be set")
			}

			if navCtx.Last == "" {
				t.Error("Expected Last URL to be set")
			}
		})
	}
}

func TestCreateSearchContext(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	opts := &semantic.QueryOptions{
		Query: &semantic.TextQuery{
			Value: "test query",
		},
		Pagination: &semantic.Pagination{
			Page: 1,
			Size: 20,
		},
	}

	result := &semantic.SearchResult{
		Total:     100,
		ResultIDs: []string{"id1", "id2", "id3"},
	}

	searchCtx := svc.createSearchContext(opts, result)

	if searchCtx == nil {
		t.Fatal("Expected non-nil search context")
	}

	if searchCtx.Token == "" {
		t.Error("Expected token to be generated")
	}

	if searchCtx.Query != opts {
		t.Error("Expected query to be preserved")
	}

	if len(searchCtx.ResultIDs) != len(result.ResultIDs) {
		t.Errorf("ResultIDs length = %d, want %d", len(searchCtx.ResultIDs), len(result.ResultIDs))
	}

	if searchCtx.TotalResults != result.Total {
		t.Errorf("TotalResults = %d, want %d", searchCtx.TotalResults, result.Total)
	}

	if searchCtx.ExpiresAt == "" {
		t.Error("Expected expiration time to be set")
	}
}

func TestBuildSearchURL(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	searchCtx := &semantic.SearchContext{
		Query: &semantic.QueryOptions{
			Query: &semantic.TextQuery{
				Value: "rembrandt",
			},
			Pagination: &semantic.Pagination{
				Page: 2,
				Size: 50,
			},
			Filters: []semantic.Filter{
				&semantic.PropertyFilter{
					FieldName:    "creator",
					OperatorType: semantic.OpEqual,
					Value:        "Rembrandt",
				},
			},
		},
	}

	req := httptest.NewRequest("GET", "/resource/123", nil)
	url := svc.buildSearchURL(req, searchCtx)

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// Check that URL contains expected query parameters
	expectedParams := []string{"query=rembrandt", "page=2", "size=50", "filter"}
	for _, param := range expectedParams {
		if !contains(url, param) {
			t.Errorf("Expected URL to contain '%s', got: %s", param, url)
		}
	}
}

func TestGenerateToken(t *testing.T) {
	token1 := generateToken()
	token2 := generateToken()

	if token1 == "" {
		t.Error("Expected non-empty token")
	}

	if token1 == token2 {
		t.Error("Expected different tokens for consecutive calls")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandleResourceDetail_WithIncludeProvider(t *testing.T) {
	// Create a mock include provider
	mockProvider := &mockDetailIncludeProvider{
		name: "testSection",
		result: map[string]any{
			"@type": "hub3:TestSection",
			"value": "test-data",
		},
	}

	store := semantic.NewMockStore()
	svc, err := NewService(
		WithStore(store),
		WithIncludeProvider(mockProvider),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	router := chi.NewRouter()
	svc.Routes("", router)

	req := httptest.NewRequest("GET", "/api/semantic/v1/resource/test-id?include=testSection", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var doc map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&doc); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	section, ok := doc["hub3:testSection"]
	if !ok {
		t.Fatal("expected hub3:testSection in response")
	}

	sectionMap, ok := section.(map[string]any)
	if !ok {
		t.Fatalf("section is %T, want map[string]any", section)
	}
	if sectionMap["@type"] != "hub3:TestSection" {
		t.Errorf("section @type = %v, want hub3:TestSection", sectionMap["@type"])
	}
}

func TestHandleResourceDetail_WithIncludeParams(t *testing.T) {
	mockProvider := &mockDetailIncludeProvider{
		name:          "relatedItems",
		captureParams: true,
	}

	store := semantic.NewMockStore()
	svc, err := NewService(
		WithStore(store),
		WithIncludeProvider(mockProvider),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	router := chi.NewRouter()
	svc.Routes("", router)

	req := httptest.NewRequest("GET", "/api/semantic/v1/resource/test-id?include=relatedItems&relatedItems.count=10&relatedItems.fields=dc:title", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Verify params were passed
	if mockProvider.lastParams["count"] != "10" {
		t.Errorf("count param = %q, want %q", mockProvider.lastParams["count"], "10")
	}
	if mockProvider.lastParams["fields"] != "dc:title" {
		t.Errorf("fields param = %q, want %q", mockProvider.lastParams["fields"], "dc:title")
	}
}

func TestHandleResourceDetail_NoInclude(t *testing.T) {
	store := semantic.NewMockStore()
	svc, err := NewService(WithStore(store))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	router := chi.NewRouter()
	svc.Routes("", router)

	req := httptest.NewRequest("GET", "/api/semantic/v1/resource/test-id", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// mockDetailIncludeProvider is a test double for IncludeProvider
type mockDetailIncludeProvider struct {
	name          string
	result        any
	err           error
	captureParams bool
	lastParams    map[string]string
}

func (m *mockDetailIncludeProvider) Name() string { return m.name }
func (m *mockDetailIncludeProvider) Provide(_ context.Context, req *semantic.IncludeRequest) (any, error) {
	if m.captureParams {
		m.lastParams = req.Params
		return map[string]any{"captured": true}, nil
	}
	return m.result, m.err
}
