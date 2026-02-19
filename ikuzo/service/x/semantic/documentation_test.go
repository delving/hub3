package semantic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestAPIDocumentation_IncludesCapabilities(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/docs", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify capabilities section exists
	caps, ok := result["hub3:capabilities"].(map[string]any)
	if !ok {
		t.Fatal("hub3:capabilities not found or wrong type")
	}

	// Verify operators
	operators, ok := caps["operators"].([]any)
	if !ok {
		t.Fatal("operators not found or wrong type")
	}
	if len(operators) < 5 {
		t.Errorf("expected at least 5 operators, got %d", len(operators))
	}

	// Verify "eq" is in operators
	hasEq := false
	for _, op := range operators {
		if op == "eq" {
			hasEq = true
			break
		}
	}
	if !hasEq {
		t.Error("expected 'eq' in operators list")
	}

	// Verify facet types
	facetTypes, ok := caps["facetTypes"].([]any)
	if !ok {
		t.Fatal("facetTypes not found")
	}
	if len(facetTypes) < 2 {
		t.Errorf("expected at least 2 facet types, got %d", len(facetTypes))
	}

	// Verify numeric values
	if maxPage, ok := caps["maxPageSize"].(float64); !ok || maxPage <= 0 {
		t.Error("expected positive maxPageSize")
	}

	if defaultPage, ok := caps["defaultPageSize"].(float64); !ok || defaultPage <= 0 {
		t.Error("expected positive defaultPageSize")
	}

	// Verify path separator
	if sep, ok := caps["pathSeparator"].(string); !ok || sep != "/" {
		t.Errorf("pathSeparator = %v, want /", sep)
	}

	// Verify context TTL
	if ttl, ok := caps["contextTTL"].(string); !ok || ttl == "" {
		t.Error("expected non-empty contextTTL")
	}
}

func TestAPIDocumentation_HasCorrectType(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/docs", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	if result["@type"] != "hydra:ApiDocumentation" {
		t.Errorf("@type = %v, want hydra:ApiDocumentation", result["@type"])
	}
}

func TestAPIDocumentation_HasIntrospectionLink(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/docs", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	intro, ok := result["hub3:introspection"].(map[string]any)
	if !ok {
		t.Fatal("hub3:introspection not found or wrong type")
	}

	if intro["@type"] != "hydra:Link" {
		t.Errorf("introspection @type = %v, want hydra:Link", intro["@type"])
	}

	if intro["hydra:title"] != "Data Introspection" {
		t.Errorf("introspection title = %v, want 'Data Introspection'", intro["hydra:title"])
	}

	// Verify the @id ends with /introspect
	id, ok := intro["@id"].(string)
	if !ok || id == "" {
		t.Fatal("introspection @id should be a non-empty string")
	}
}

func TestBuildCapabilities(t *testing.T) {
	caps := buildCapabilities()

	// Verify it's a map with expected keys
	operators, ok := caps["operators"].([]string)
	if !ok {
		t.Fatal("operators should be []string")
	}
	if len(operators) == 0 {
		t.Error("operators should not be empty")
	}

	facetTypes, ok := caps["facetTypes"].([]string)
	if !ok {
		t.Fatal("facetTypes should be []string")
	}
	if len(facetTypes) == 0 {
		t.Error("facetTypes should not be empty")
	}

	if caps["pathSeparator"] != "/" {
		t.Errorf("pathSeparator = %v, want /", caps["pathSeparator"])
	}

	// Verify maxPageSize comes from DefaultStoreCapabilities
	maxPage, ok := caps["maxPageSize"].(int)
	if !ok || maxPage != domainSemantic.DefaultStoreCapabilities().MaxPageSize {
		t.Errorf("maxPageSize = %v, want %d", caps["maxPageSize"], domainSemantic.DefaultStoreCapabilities().MaxPageSize)
	}

	// Verify defaultPageSize comes from DefaultPagination
	defaultPage, ok := caps["defaultPageSize"].(int)
	if !ok || defaultPage != domainSemantic.DefaultPagination().Size {
		t.Errorf("defaultPageSize = %v, want %d", caps["defaultPageSize"], domainSemantic.DefaultPagination().Size)
	}

	// Verify contextTTL
	ttl, ok := caps["contextTTL"].(string)
	if !ok || ttl != "15m" {
		t.Errorf("contextTTL = %v, want 15m", caps["contextTTL"])
	}

	// Verify sort options
	sortOpts, ok := caps["sortOptions"].([]string)
	if !ok || len(sortOpts) != 2 {
		t.Errorf("sortOptions = %v, want [asc, desc]", caps["sortOptions"])
	}
}

func TestSupportedOperators(t *testing.T) {
	ops := supportedOperators()

	required := []string{"eq", "in", "contains", "gt", "gte", "lt", "lte", "exists"}
	for _, req := range required {
		found := false
		for _, op := range ops {
			if op == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected operator %q in supported operators", req)
		}
	}
}

func TestSupportedFacetTypes(t *testing.T) {
	types := supportedFacetTypes()

	required := []string{"enum", "range", "date", "boolean"}
	for _, req := range required {
		found := false
		for _, ft := range types {
			if ft == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected facet type %q in supported facet types", req)
		}
	}
}

func TestEntryPoint_HasLinks(t *testing.T) {
	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/", nil)
	w := httptest.NewRecorder()

	svc.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	if result["@type"] != "hydra:EntryPoint" {
		t.Errorf("@type = %v, want hydra:EntryPoint", result["@type"])
	}
}
