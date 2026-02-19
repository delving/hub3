# Semantic API v1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the semantic API v1 as designed in `docs/plans/2026-02-18-semantic-api-design.md` -- a self-describing, schema-agnostic query platform with introspection, query contexts, and path-based querying.

**Architecture:** Two-sided design (structure vs content). The structure side (search, introspect, query context, docs) is backend-agnostic via the `SearchStore` interface. Content (JSON-LD documents) passes through opaque. The introspection layer bridges the two by examining indexed data. Initial backend is the existing V2SearchAdapter.

**Tech Stack:** Go 1.22+, Elasticsearch 7.x (via olivere/elastic), chi router, zerolog, protobuf (for existing API types). Tests use stdlib `testing` with table-driven patterns.

---

## Phase 1: Introspection Foundation

The introspection layer is the core innovation. Build it first because it validates the design and enables everything else.

### Task 1: Introspection Domain Types

**Files:**
- Create: `ikuzo/domain/semantic/introspect.go`
- Test: `ikuzo/domain/semantic/introspect_test.go`

**Step 1: Write the failing test**

```go
package semantic

import "testing"

func TestClassInfo_Basic(t *testing.T) {
	ci := ClassInfo{
		URI:   "http://www.europeana.eu/schemas/edm/ProvidedCHO",
		Label: "edm:ProvidedCHO",
		Count: 45000,
	}

	if ci.Label != "edm:ProvidedCHO" {
		t.Errorf("Label = %v, want edm:ProvidedCHO", ci.Label)
	}
}

func TestPropertyInfo_HasResolvedLabels(t *testing.T) {
	pi := PropertyInfo{
		Field:             "dc_creator",
		Predicate:         "http://purl.org/dc/elements/1.1/creator",
		Label:             "dc:creator",
		ValueTypes:        []string{"Literal", "Resource"},
		Count:             42000,
		Languages:         []string{"nl", "en"},
		HasResolvedLabels: true,
		Paths:             []string{"dc_creator", "dc_creator/foaf_name"},
	}

	if !pi.HasResolvedLabels {
		t.Error("Expected HasResolvedLabels to be true")
	}

	if len(pi.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(pi.Paths))
	}
}

func TestIntrospectionResult_WithScope(t *testing.T) {
	ir := IntrospectionResult{
		Scope: IntrospectionScope{
			Type:           "query",
			ContextID:      "ctx_a7f3",
			TotalDocuments: 12847,
		},
	}

	if ir.Scope.Type != "query" {
		t.Errorf("Scope.Type = %v, want query", ir.Scope.Type)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run TestClassInfo -v`
Expected: FAIL with compilation error (types not defined)

**Step 3: Write minimal implementation**

```go
package semantic

// IntrospectionResult contains the result of an introspection query.
type IntrospectionResult struct {
	Scope      IntrospectionScope `json:"hub3:scope"`
	Classes    []ClassInfo        `json:"hub3:classes,omitempty"`
	Schemas    []SchemaInfo       `json:"hub3:schemas,omitempty"`
	Properties []PropertyInfo     `json:"hub3:properties,omitempty"`
}

// IntrospectionScope describes what data the introspection covers.
type IntrospectionScope struct {
	Type           string `json:"type"`                     // "index" or "query"
	ContextID      string `json:"context,omitempty"`        // query context ID when type="query"
	TotalDocuments int64  `json:"totalDocuments"`
}

// ClassInfo describes an RDF class found in the data.
type ClassInfo struct {
	URI            string `json:"uri"`
	Label          string `json:"label"`
	Count          int64  `json:"count"`
	PropertiesLink string `json:"properties,omitempty"` // URL to property introspection
}

// PropertyInfo describes a property found on a class.
type PropertyInfo struct {
	Field             string      `json:"field"`
	Predicate         string      `json:"predicate"`
	Label             string      `json:"label"`
	ValueTypes        []string    `json:"valueTypes"`
	Count             int64       `json:"count"`
	Languages         []string    `json:"languages,omitempty"`
	HasResolvedLabels bool        `json:"hasResolvedLabels,omitempty"`
	DataType          string      `json:"dataType,omitempty"`
	Range             *FieldRange `json:"range,omitempty"`
	Paths             []string    `json:"paths,omitempty"`
	Schema            *SchemaRef  `json:"schema,omitempty"`
}

// FieldRange describes the min/max range for numeric or date fields.
type FieldRange struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

// SchemaInfo describes a record-definition schema found in the data.
type SchemaInfo struct {
	RecDefID      string   `json:"recDefID"`
	DocumentCount int64    `json:"documentCount"`
	Specs         []string `json:"spec"`
}

// SchemaRef links a property to its record-definition documentation.
type SchemaRef struct {
	RecDefID      string `json:"recDefID"`
	Documentation string `json:"documentation,omitempty"`
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run "TestClassInfo|TestPropertyInfo|TestIntrospectionResult" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/domain/semantic/introspect.go ikuzo/domain/semantic/introspect_test.go
git commit -m "feat(semantic): add introspection domain types"
```

---

### Task 2: IntrospectionStore Interface

**Files:**
- Modify: `ikuzo/domain/semantic/store.go`
- Test: `ikuzo/domain/semantic/introspect_test.go` (append)

**Step 1: Write the failing test**

Append to `introspect_test.go`:

```go
func TestMockIntrospectionStore(t *testing.T) {
	store := &MockIntrospectionStore{
		Classes: []ClassInfo{
			{URI: "http://www.europeana.eu/schemas/edm/ProvidedCHO", Label: "edm:ProvidedCHO", Count: 100},
		},
	}

	ctx := context.Background()
	result, err := store.IntrospectClasses(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 class, got %d", len(result))
	}

	if result[0].Count != 100 {
		t.Errorf("Count = %d, want 100", result[0].Count)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run TestMockIntrospectionStore -v`
Expected: FAIL (MockIntrospectionStore not defined)

**Step 3: Write minimal implementation**

Add to `store.go`:

```go
// IntrospectionStore provides introspection capabilities for indexed data.
// It can optionally scope results to a query context.
type IntrospectionStore interface {
	// IntrospectClasses returns all RDF classes found in the data.
	// If opts is non-nil, results are scoped to that query's result set.
	IntrospectClasses(ctx context.Context, opts *QueryOptions) ([]ClassInfo, error)

	// IntrospectProperties returns properties for a given class.
	IntrospectProperties(ctx context.Context, classURI string, opts *QueryOptions) ([]PropertyInfo, error)

	// IntrospectField returns value distribution for a specific field.
	IntrospectField(ctx context.Context, field string, opts *QueryOptions) (*PropertyInfo, error)

	// IntrospectPaths returns predicate paths between classes.
	IntrospectPaths(ctx context.Context, opts *QueryOptions) ([]PathInfo, error)
}
```

Add `PathInfo` to `introspect.go`:

```go
// PathInfo describes a predicate path between classes.
type PathInfo struct {
	Path      string `json:"path"`      // e.g., "dc_creator/foaf_name"
	FromClass string `json:"fromClass"` // e.g., "edm:ProvidedCHO"
	ToClass   string `json:"toClass"`   // e.g., "edm:Agent"
	Count     int64  `json:"count"`
}
```

Add `MockIntrospectionStore` to `store.go`:

```go
// MockIntrospectionStore is a test implementation of IntrospectionStore.
type MockIntrospectionStore struct {
	Classes    []ClassInfo
	Properties map[string][]PropertyInfo
	Paths      []PathInfo
}

func (m *MockIntrospectionStore) IntrospectClasses(_ context.Context, _ *QueryOptions) ([]ClassInfo, error) {
	return m.Classes, nil
}

func (m *MockIntrospectionStore) IntrospectProperties(_ context.Context, classURI string, _ *QueryOptions) ([]PropertyInfo, error) {
	if m.Properties == nil {
		return nil, nil
	}
	return m.Properties[classURI], nil
}

func (m *MockIntrospectionStore) IntrospectField(_ context.Context, field string, _ *QueryOptions) (*PropertyInfo, error) {
	for _, props := range m.Properties {
		for _, p := range props {
			if p.Field == field {
				return &p, nil
			}
		}
	}
	return nil, nil
}

func (m *MockIntrospectionStore) IntrospectPaths(_ context.Context, _ *QueryOptions) ([]PathInfo, error) {
	return m.Paths, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run TestMockIntrospectionStore -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/domain/semantic/store.go ikuzo/domain/semantic/introspect.go ikuzo/domain/semantic/introspect_test.go
git commit -m "feat(semantic): add IntrospectionStore interface and mock"
```

---

### Task 3: Introspection HTTP Routes

**Files:**
- Modify: `ikuzo/service/x/semantic/service.go` (add routes)
- Create: `ikuzo/service/x/semantic/introspect_handler.go`
- Test: `ikuzo/service/x/semantic/introspect_handler_test.go`

**Step 1: Write the failing test**

```go
package semantic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestIntrospectClassesHandler(t *testing.T) {
	mockStore := &domainSemantic.MockStore{}
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run TestIntrospectClassesHandler -v`
Expected: FAIL (WithIntrospectionStore not defined)

**Step 3: Write minimal implementation**

Add `WithIntrospectionStore` option to `options.go`:

```go
// WithIntrospectionStore sets the introspection store for the service.
func WithIntrospectionStore(store semantic.IntrospectionStore) Option {
	return func(s *Service) error {
		s.introspect = store
		return nil
	}
}
```

Add `introspect` field to `Service` struct in `service.go`:

```go
type Service struct {
	store      semantic.SearchStore
	introspect semantic.IntrospectionStore
	registry   *semantic.ConfigRegistry
	log        zerolog.Logger
	baseURL    string
}
```

Register routes in `service.go` `Routes()` method:

```go
// Introspection routes
router.Route("/introspect", func(r chi.Router) {
	r.Get("/", s.handleIntrospect)
	r.Get("/classes", s.handleIntrospectClasses)
	r.Get("/classes/{class}/properties", s.handleIntrospectProperties)
	r.Get("/fields/{field}", s.handleIntrospectField)
	r.Get("/paths", s.handleIntrospectPaths)
})
```

Create `introspect_handler.go`:

```go
package semantic

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) handleIntrospectClasses(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	classes, err := s.introspect.IntrospectClasses(r.Context(), nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect classes")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":        "hub3:IntrospectionResult",
		"hub3:scope":   map[string]any{"type": "index"},
		"hub3:classes": classes,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Service) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	s.handleIntrospectClasses(w, r)
}

func (s *Service) handleIntrospectProperties(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	classURI := chi.URLParam(r, "class")

	props, err := s.introspect.IntrospectProperties(r.Context(), classURI, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect properties")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":           "hub3:PropertyIntrospection",
		"hub3:class":      classURI,
		"hub3:properties": props,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Service) handleIntrospectField(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	field := chi.URLParam(r, "field")

	prop, err := s.introspect.IntrospectField(r.Context(), field, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect field")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	if prop == nil {
		http.Error(w, "field not found", http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"@type":         "hub3:PropertyIntrospection",
		"hub3:property": prop,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Service) handleIntrospectPaths(w http.ResponseWriter, r *http.Request) {
	if s.introspect == nil {
		http.Error(w, "introspection not available", http.StatusServiceUnavailable)
		return
	}

	paths, err := s.introspect.IntrospectPaths(r.Context(), nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to introspect paths")
		http.Error(w, "introspection failed", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"@type":      "hub3:IntrospectionResult",
		"hub3:paths": paths,
	}

	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(resp)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run TestIntrospectClassesHandler -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/service/x/semantic/introspect_handler.go ikuzo/service/x/semantic/introspect_handler_test.go ikuzo/service/x/semantic/options.go ikuzo/service/x/semantic/service.go
git commit -m "feat(semantic): add introspection HTTP routes and handlers"
```

---

### Task 4: Elasticsearch Introspection Backend

**Files:**
- Create: `ikuzo/storage/x/v2adapter/introspect.go`
- Test: `ikuzo/storage/x/v2adapter/introspect_test.go`

This task implements `IntrospectionStore` using Elasticsearch aggregations. It queries the actual indexed data to discover classes, properties, and paths.

**Step 1: Write the failing test**

```go
package v2adapter

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestBuildClassAggregation(t *testing.T) {
	q := buildClassAggregation()

	// Verify it produces a valid ES aggregation source
	src, err := q.Source()
	if err != nil {
		t.Fatalf("failed to build aggregation source: %v", err)
	}

	agg, ok := src.(map[string]any)
	if !ok {
		t.Fatal("expected map source from aggregation")
	}

	if _, ok := agg["nested"]; !ok {
		t.Error("expected nested aggregation for resources")
	}
}

func TestBuildPropertyAggregation(t *testing.T) {
	q := buildPropertyAggregation("edm:ProvidedCHO")

	src, err := q.Source()
	if err != nil {
		t.Fatalf("failed to build aggregation source: %v", err)
	}

	if src == nil {
		t.Error("expected non-nil aggregation source")
	}
}

func TestClassInfoFromBucket(t *testing.T) {
	// Test the conversion from ES aggregation bucket to ClassInfo
	ci := classInfoFromBucket("http://www.europeana.eu/schemas/edm/ProvidedCHO", 45000)

	if ci.Label != "edm:ProvidedCHO" {
		t.Errorf("Label = %v, want edm:ProvidedCHO", ci.Label)
	}

	if ci.Count != 45000 {
		t.Errorf("Count = %d, want 45000", ci.Count)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/storage/x/v2adapter/ -run "TestBuildClass|TestBuildProperty|TestClassInfo" -v`
Expected: FAIL (functions not defined)

**Step 3: Write minimal implementation**

Create `introspect.go` with ES aggregation builders and a `V2IntrospectionAdapter` that implements `IntrospectionStore`. The class discovery uses a nested aggregation on `resources.types`:

```go
package v2adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	elastic "github.com/olivere/elastic/v7"
)

// V2IntrospectionAdapter implements semantic.IntrospectionStore using ES aggregations.
type V2IntrospectionAdapter struct {
	client *elastic.Client
	index  string
}

// NewV2IntrospectionAdapter creates a new introspection adapter.
func NewV2IntrospectionAdapter(client *elastic.Client, index string) *V2IntrospectionAdapter {
	return &V2IntrospectionAdapter{client: client, index: index}
}

func buildClassAggregation() *elastic.NestedAggregation {
	return elastic.NewNestedAggregation().Path("resources").
		SubAggregation("class_types",
			elastic.NewTermsAggregation().Field("resources.types").Size(100),
		)
}

func buildPropertyAggregation(classFilter string) *elastic.NestedAggregation {
	agg := elastic.NewNestedAggregation().Path("resources")

	entriesAgg := elastic.NewNestedAggregation().Path("resources.entries").
		SubAggregation("labels",
			elastic.NewTermsAggregation().Field("resources.entries.searchLabel").Size(200),
		)

	if classFilter != "" {
		filterAgg := elastic.NewFilterAggregation().
			Filter(elastic.NewTermQuery("resources.types", classFilter)).
			SubAggregation("entries", entriesAgg)
		agg.SubAggregation("class_filter", filterAgg)
	} else {
		agg.SubAggregation("entries", entriesAgg)
	}

	return agg
}

func classInfoFromBucket(uri string, count int64) semantic.ClassInfo {
	label := uri
	// Extract compact form from URI (e.g., "edm:ProvidedCHO" from full URI)
	if idx := strings.LastIndex(uri, "/"); idx != -1 {
		localName := uri[idx+1:]
		// Try to find prefix (simplified; real impl uses NamespaceManager)
		for prefix, ns := range commonNamespaces {
			if strings.HasPrefix(uri, ns) {
				label = prefix + ":" + localName
				break
			}
		}
	}

	return semantic.ClassInfo{
		URI:   uri,
		Label: label,
		Count: count,
	}
}

var commonNamespaces = map[string]string{
	"edm":     "http://www.europeana.eu/schemas/edm/",
	"dc":      "http://purl.org/dc/elements/1.1/",
	"dcterms": "http://purl.org/dc/terms/",
	"skos":    "http://www.w3.org/2004/02/skos/core#",
	"foaf":    "http://xmlns.com/foaf/0.1/",
	"rdf":     "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":    "http://www.w3.org/2000/01/rdf-schema#",
	"ore":     "http://www.openarchives.org/ore/terms/",
}

func (a *V2IntrospectionAdapter) IntrospectClasses(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.ClassInfo, error) {
	search := a.client.Search(a.index).Size(0)

	// Apply query scope if provided
	if opts != nil && opts.Query != nil && opts.Query.Value != "" {
		search = search.Query(elastic.NewQueryStringQuery(opts.Query.Value))
	}

	search = search.Aggregation("classes", buildClassAggregation())

	result, err := search.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("introspect classes: %w", err)
	}

	nested, found := result.Aggregations.Nested("classes")
	if !found {
		return nil, nil
	}

	terms, found := nested.Terms("class_types")
	if !found {
		return nil, nil
	}

	classes := make([]semantic.ClassInfo, 0, len(terms.Buckets))
	for _, bucket := range terms.Buckets {
		uri, ok := bucket.Key.(string)
		if !ok {
			continue
		}
		classes = append(classes, classInfoFromBucket(uri, bucket.DocCount))
	}

	return classes, nil
}

func (a *V2IntrospectionAdapter) IntrospectProperties(ctx context.Context, classURI string, opts *semantic.QueryOptions) ([]semantic.PropertyInfo, error) {
	search := a.client.Search(a.index).Size(0)
	search = search.Aggregation("properties", buildPropertyAggregation(classURI))

	result, err := search.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("introspect properties: %w", err)
	}

	// Extract property info from nested aggregation result
	// (Implementation details depend on exact ES response shape)
	_ = result
	return nil, nil // TODO: implement result parsing
}

func (a *V2IntrospectionAdapter) IntrospectField(ctx context.Context, field string, opts *semantic.QueryOptions) (*semantic.PropertyInfo, error) {
	return nil, nil // TODO: implement
}

func (a *V2IntrospectionAdapter) IntrospectPaths(ctx context.Context, opts *semantic.QueryOptions) ([]semantic.PathInfo, error) {
	return nil, nil // TODO: implement
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/storage/x/v2adapter/ -run "TestBuildClass|TestBuildProperty|TestClassInfo" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/storage/x/v2adapter/introspect.go ikuzo/storage/x/v2adapter/introspect_test.go
git commit -m "feat(semantic): add ES introspection backend with class/property discovery"
```

---

## Phase 2: Query Context

Replace base64 scroll tokens with named, short-lived query context resources.

### Task 5: Query Context Store Interface Enhancement

**Files:**
- Modify: `ikuzo/domain/semantic/pagination.go`
- Test: `ikuzo/domain/semantic/pagination_test.go`

Enhance `SearchContext` to include a short, human-readable ID and explicit TTL management. The existing `SearchContext` struct is close but needs the short ID generation and the query preservation for introspect-by-query.

**Step 1: Write the failing test**

```go
func TestNewQueryContext(t *testing.T) {
	opts := &QueryOptions{
		Query: &TextQuery{Value: "amsterdam"},
	}

	ctx := NewQueryContext(opts, 12847)

	if ctx.ID == "" {
		t.Error("expected non-empty context ID")
	}

	if len(ctx.ID) > 12 {
		t.Errorf("context ID should be short, got %d chars: %s", len(ctx.ID), ctx.ID)
	}

	if ctx.TotalResults != 12847 {
		t.Errorf("TotalResults = %d, want 12847", ctx.TotalResults)
	}

	if ctx.Query == nil {
		t.Error("expected query to be preserved in context")
	}

	if ctx.IsExpired() {
		t.Error("freshly created context should not be expired")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run TestNewQueryContext -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Add to `pagination.go`:

```go
import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const defaultContextTTL = 15 * time.Minute

// NewQueryContext creates a new query context with a short, human-readable ID.
func NewQueryContext(opts *QueryOptions, totalResults int64) *SearchContext {
	id := generateShortID()
	return &SearchContext{
		ID:           id,
		Token:        id,
		Query:        opts,
		TotalResults: totalResults,
		ExpiresAt:    time.Now().Add(defaultContextTTL).Format(time.RFC3339),
	}
}

// IsExpired checks if the context has expired.
func (sc *SearchContext) IsExpired() bool {
	if sc.ExpiresAt == "" {
		return false
	}
	expires, err := time.Parse(time.RFC3339, sc.ExpiresAt)
	if err != nil {
		return true
	}
	return time.Now().After(expires)
}

// ExtendTTL extends the context's expiration time.
func (sc *SearchContext) ExtendTTL() {
	sc.ExpiresAt = time.Now().Add(defaultContextTTL).Format(time.RFC3339)
}

func generateShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "ctx_" + hex.EncodeToString(b)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run TestNewQueryContext -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/domain/semantic/pagination.go ikuzo/domain/semantic/pagination_test.go
git commit -m "feat(semantic): add NewQueryContext with short IDs and TTL management"
```

---

### Task 6: Query Context HTTP Routes

**Files:**
- Modify: `ikuzo/service/x/semantic/service.go` (add routes)
- Create: `ikuzo/service/x/semantic/context_handler.go`
- Test: `ikuzo/service/x/semantic/context_handler_test.go`

Add CRUD routes for query contexts: POST to create, GET to retrieve, DELETE to release.

**Step 1: Write the failing test**

```go
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
	mockStore := &domainSemantic.MockStore{}

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
}
```

**Step 2-5: Implement, test, commit**

Register context routes in `service.go`, implement handlers in `context_handler.go` that use the existing `SearchContextStore` methods on `SearchStore`.

```bash
git commit -m "feat(semantic): add query context CRUD routes"
```

---

## Phase 3: Introspect-by-Query

### Task 7: Wire Query Context to Introspection

**Files:**
- Modify: `ikuzo/service/x/semantic/introspect_handler.go`
- Test: `ikuzo/service/x/semantic/introspect_handler_test.go` (append)

Add `?context={ctx}` parameter support to all introspect handlers. When present, look up the stored query and pass it to `IntrospectionStore` methods.

```bash
git commit -m "feat(semantic): support introspect-by-query via context parameter"
```

---

## Phase 4: Response Envelope Alignment

### Task 8: Hydra Collection Response Builder

**Files:**
- Modify: `ikuzo/service/x/semantic/response.go`
- Test: `ikuzo/service/x/semantic/response_test.go`

Align the search response with the design doc contracts: context-based pagination (no `last` link), `hub3:queryContext` in every response, `hub3:activeFilters` with `remove` URLs, `hub3:facets` with `filter` strings.

```bash
git commit -m "feat(semantic): align search response with v1 design contracts"
```

---

### Task 9: API Capabilities Endpoint

**Files:**
- Modify: `ikuzo/service/x/semantic/documentation.go`
- Test: `ikuzo/service/x/semantic/documentation_test.go`

Replace the current docs endpoint with a code-generated capabilities response matching the design doc's `hub3:capabilities` contract.

```bash
git commit -m "feat(semantic): add code-generated API capabilities endpoint"
```

---

## Phase 5: Label Resolution (Index-Time)

### Task 10: Add Resolution Fields to Entry Mapping

**Files:**
- Modify: `ikuzo/driver/elasticsearch/internal/mapping/v2.go`
- Modify: `ikuzo/rdf/index/entry.go`
- Test: `ikuzo/rdf/index/entry_test.go`

Add `resolvedFrom` (keyword) and `resolvedLevel` (integer) to the ES mapping and the `Entry` struct.

```bash
git commit -m "feat(index): add resolvedFrom and resolvedLevel to Entry and ES mapping"
```

---

### Task 11: Label Resolution in Indexing Pipeline

**Files:**
- Create: `ikuzo/rdf/index/label_resolver.go`
- Test: `ikuzo/rdf/index/label_resolver_test.go`

Implement the label resolution logic: when building an `index.Resource` and encountering a Resource-type entry, check if the linked resource is in the same graph. If so, resolve the best label using `ResourceLabelPredicates` priority order, and populate `@value`, `resolvedFrom`, `resolvedLevel`.

```bash
git commit -m "feat(index): add index-time label resolution with provenance tracking"
```

---

## Phase 6: Error Suggestions

### Task 12: Field Name Suggestion on Unknown Fields

**Files:**
- Create: `ikuzo/domain/semantic/suggest.go`
- Test: `ikuzo/domain/semantic/suggest_test.go`

Implement Levenshtein distance or similar for suggesting correct field names when a filter references an unknown field. Uses the introspection store to get valid field names.

```bash
git commit -m "feat(semantic): add field name suggestions for typo detection"
```

---

## Phase 7: Integration and Wiring

### Task 13: Wire Introspection into Service Configuration

**Files:**
- Modify: `ikuzo/ikuzoctl/cmd/config/semantic.go`

Wire the `V2IntrospectionAdapter` into the semantic service configuration, alongside the existing `V2SearchAdapter`.

```bash
git commit -m "feat(config): wire introspection adapter into semantic service"
```

---

### Task 14: End-to-End Integration Test

**Files:**
- Modify: `ikuzo/service/x/semantic/integration_test.go`

Add integration tests that exercise the full flow: search -> get context -> introspect with context -> detail with navigation.

```bash
git commit -m "test(semantic): add end-to-end integration tests for v1 API flow"
```

---

## Task Dependency Graph

```
Phase 1: Introspection
  Task 1 (domain types) → Task 2 (store interface) → Task 3 (HTTP routes) → Task 4 (ES backend)

Phase 2: Query Context
  Task 5 (context enhancement) → Task 6 (HTTP routes)

Phase 3: Introspect-by-Query
  Task 4 + Task 6 → Task 7 (wire context to introspect)

Phase 4: Response Alignment
  Task 6 → Task 8 (Hydra response)
  Task 9 (capabilities) — independent

Phase 5: Label Resolution
  Task 10 (mapping) → Task 11 (resolver)

Phase 6: Error Suggestions
  Task 2 → Task 12 (suggestions)

Phase 7: Integration
  All above → Task 13 (wiring) → Task 14 (e2e tests)
```

## Notes for Implementer

- **Test commands**: `go test ./ikuzo/domain/semantic/... -v` and `go test ./ikuzo/service/x/semantic/... -v`
- **Build check**: `go build ./...` after every task to catch compilation issues
- **Existing tests**: Run `make test` periodically to ensure no regressions
- **Code style**: Follow `CLAUDE.md` — gofmt, 140 char max, organized imports (stdlib, external, internal)
- **Path queries**: The `ikuzo/rdf/index/path_query.go` parser already exists and is comprehensive. Task 7+ should reuse it, not reimplement.
- **ES mapping changes** (Task 10): Adding new keyword/integer fields to the mapping is backward-compatible — existing indices will auto-create the fields on next document write.
