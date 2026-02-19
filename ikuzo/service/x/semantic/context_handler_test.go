package semantic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestQueryContextCRUD(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Create context
	body := `{"query": {"text": "amsterdam"}, "totalResults": 100}`
	req := httptest.NewRequest(http.MethodPost, "/api/semantic/v1/contexts/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201. Body: %s", w.Code, w.Body.String())
	}

	var createResp map[string]any
	json.NewDecoder(w.Body).Decode(&createResp)

	ctxID, ok := createResp["id"].(string)
	if !ok || ctxID == "" {
		t.Fatal("expected context ID in create response")
	}

	// Retrieve context
	req = httptest.NewRequest(http.MethodGet, "/api/semantic/v1/contexts/query/"+ctxID, nil)
	w = httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get: status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	// Delete context
	req = httptest.NewRequest(http.MethodDelete, "/api/semantic/v1/contexts/query/"+ctxID, nil)
	w = httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want 204", w.Code)
	}

	// Verify deleted — should be 404
	req = httptest.NewRequest(http.MethodGet, "/api/semantic/v1/contexts/query/"+ctxID, nil)
	w = httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("after delete: status = %d, want 404", w.Code)
	}
}

func TestQueryContextGet_NotFound(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/contexts/query/nonexistent", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
