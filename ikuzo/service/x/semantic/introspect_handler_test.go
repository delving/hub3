package semantic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestIntrospectClassesHandler(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Classes: []domainSemantic.ClassInfo{
			{URI: "http://www.europeana.eu/schemas/edm/ProvidedCHO", Label: "edm:ProvidedCHO", Count: 100},
			{URI: "http://www.europeana.eu/schemas/edm/Agent", Label: "edm:Agent", Count: 50},
		},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/classes", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	classes, ok := result["hub3:classes"].([]any)
	if !ok {
		t.Fatalf("hub3:classes not found or wrong type in response")
	}

	if len(classes) != 2 {
		t.Errorf("expected 2 classes, got %d", len(classes))
	}
}

func TestIntrospectClassesHandler_NoStore(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(
		WithStore(mockStore),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/classes", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestIntrospectPropertiesHandler(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Properties: map[string][]domainSemantic.PropertyInfo{
			"edm:ProvidedCHO": {
				{Field: "dc_creator", Predicate: "http://purl.org/dc/elements/1.1/creator", Label: "dc:creator", Count: 42000},
			},
		},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/classes/edm:ProvidedCHO/properties", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	props, ok := result["hub3:properties"].([]any)
	if !ok {
		t.Fatalf("hub3:properties not found or wrong type")
	}

	if len(props) != 1 {
		t.Errorf("expected 1 property, got %d", len(props))
	}
}

func TestIntrospectFieldHandler(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Properties: map[string][]domainSemantic.PropertyInfo{
			"edm:ProvidedCHO": {
				{Field: "dc_creator", Predicate: "http://purl.org/dc/elements/1.1/creator", Label: "dc:creator", Count: 42000},
			},
		},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/fields/dc_creator", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := result["hub3:property"]; !ok {
		t.Error("hub3:property not found in response")
	}
}

func TestIntrospectFieldHandler_NotFound(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Properties: map[string][]domainSemantic.PropertyInfo{},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/fields/nonexistent", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestIntrospectPathsHandler(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Paths: []domainSemantic.PathInfo{
			{Path: "dc_creator/foaf_name", FromClass: "edm:ProvidedCHO", ToClass: "edm:Agent", Count: 100},
		},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/paths", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := result["hub3:paths"]; !ok {
		t.Error("hub3:paths not found in response")
	}
}

func TestIntrospectRootRedirectsToClasses(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()
	mockIntrospect := &domainSemantic.MockIntrospectionStore{
		Classes: []domainSemantic.ClassInfo{
			{URI: "http://example.org/Thing", Label: "Thing", Count: 10},
		},
	}

	svc, err := NewService(
		WithStore(mockStore),
		WithIntrospectionStore(mockIntrospect),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/introspect/", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := result["hub3:classes"]; !ok {
		t.Error("hub3:classes not found - root should return classes")
	}
}
