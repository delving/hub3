package semantic

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	mockStore := domainSemantic.NewMockStore()

	svc, err := NewService(WithStore(mockStore))
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	return svc
}

func TestBuildCollection_IncludesTiming(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=amsterdam", nil)
	result := &domainSemantic.SearchResult{
		Total:   100,
		Results: []map[string]any{},
		Took:    42 * time.Millisecond,
	}
	opts := &domainSemantic.QueryOptions{
		Query:      &domainSemantic.TextQuery{Value: "amsterdam"},
		Pagination: domainSemantic.DefaultPagination(),
	}
	config := domainSemantic.NewResourceConfig("test:Resource", "Test")

	collection := svc.buildCollection(r, result, opts, config)

	// Verify timing is present
	if collection.Timing == nil {
		t.Fatal("expected hub3:timing in collection")
	}

	if collection.Timing.Took != 42 {
		t.Errorf("Timing.Took = %d, want 42", collection.Timing.Took)
	}

	if collection.Timing.Unit != "ms" {
		t.Errorf("Timing.Unit = %s, want ms", collection.Timing.Unit)
	}
}

func TestBuildCollection_TimingSubMillisecond(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test", nil)
	result := &domainSemantic.SearchResult{
		Total:   1,
		Results: []map[string]any{},
		Took:    500 * time.Microsecond, // sub-millisecond
	}
	opts := &domainSemantic.QueryOptions{
		Query:      &domainSemantic.TextQuery{Value: "test"},
		Pagination: domainSemantic.DefaultPagination(),
	}
	config := domainSemantic.NewResourceConfig("test:Resource", "Test")

	collection := svc.buildCollection(r, result, opts, config)

	if collection.Timing == nil {
		t.Fatal("expected hub3:timing in collection")
	}

	// Sub-millisecond should round to 0 or 1, not panic
	if collection.Timing.Took < 0 {
		t.Errorf("Timing.Took = %d, should not be negative", collection.Timing.Took)
	}
}

func TestBuildPartialCollectionView_NoLastLink(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test&page=2", nil)
	result := &domainSemantic.SearchResult{
		Total: 100,
	}
	opts := &domainSemantic.QueryOptions{
		Pagination: &domainSemantic.Pagination{Page: 2, Size: 20},
	}

	view := svc.buildPartialCollectionView(r, result, opts)

	if view.Last != "" {
		t.Errorf("expected no Last link, got %s", view.Last)
	}

	if view.Next == "" {
		t.Error("expected Next link to be set")
	}

	if view.Previous == "" {
		t.Error("expected Previous link to be set")
	}

	if view.First == "" {
		t.Error("expected First link to be set")
	}
}

func TestBuildPartialCollectionView_FirstPage(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test", nil)
	result := &domainSemantic.SearchResult{
		Total: 100,
	}
	opts := &domainSemantic.QueryOptions{
		Pagination: &domainSemantic.Pagination{Page: 1, Size: 20},
	}

	view := svc.buildPartialCollectionView(r, result, opts)

	if view.Last != "" {
		t.Errorf("expected no Last link, got %s", view.Last)
	}

	if view.Next == "" {
		t.Error("expected Next link on first page with more results")
	}

	if view.Previous != "" {
		t.Errorf("expected no Previous link on first page, got %s", view.Previous)
	}
}

func TestBuildPartialCollectionView_LastPage(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test&page=5", nil)
	result := &domainSemantic.SearchResult{
		Total: 100, // 5 pages of 20
	}
	opts := &domainSemantic.QueryOptions{
		Pagination: &domainSemantic.Pagination{Page: 5, Size: 20},
	}

	view := svc.buildPartialCollectionView(r, result, opts)

	if view.Next != "" {
		t.Errorf("expected no Next link on last page, got %s", view.Next)
	}

	if view.Previous == "" {
		t.Error("expected Previous link on last page")
	}
}

func TestFacetValueIncludesFilterString(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test", nil)
	facetResults := []domainSemantic.FacetResult{
		{
			Field:       "dc_creator",
			FacetType:   "enum",
			TotalValues: 1,
			Values: []domainSemantic.FacetValueResult{
				{Value: "Rembrandt van Rijn", Count: 234},
			},
		},
	}
	opts := &domainSemantic.QueryOptions{
		Query: &domainSemantic.TextQuery{Value: "test"},
	}

	facets := svc.buildFacets(r, facetResults, opts)

	if len(facets) != 1 {
		t.Fatalf("expected 1 facet, got %d", len(facets))
	}

	if len(facets[0].Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(facets[0].Values))
	}

	val := facets[0].Values[0]
	if val.Filter == "" {
		t.Error("expected filter string on facet value")
	}

	expected := "filter[dc_creator][eq]=Rembrandt+van+Rijn"
	if val.Filter != expected {
		t.Errorf("Filter = %q, want %q", val.Filter, expected)
	}
}

func TestFacetValueFilterStringSpecialChars(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test", nil)
	facetResults := []domainSemantic.FacetResult{
		{
			Field:       "dc_subject",
			FacetType:   "enum",
			TotalValues: 1,
			Values: []domainSemantic.FacetValueResult{
				{Value: "art & culture", Count: 50},
			},
		},
	}
	opts := &domainSemantic.QueryOptions{
		Query: &domainSemantic.TextQuery{Value: "test"},
	}

	facets := svc.buildFacets(r, facetResults, opts)

	val := facets[0].Values[0]
	expected := "filter[dc_subject][eq]=art+%26+culture"
	if val.Filter != expected {
		t.Errorf("Filter = %q, want %q", val.Filter, expected)
	}
}

func TestFacetValueRetainsApplyAndRemoveURLs(t *testing.T) {
	svc := newTestService(t)

	r := httptest.NewRequest(http.MethodGet, "/api/semantic/v1/search?query=test", nil)
	facetResults := []domainSemantic.FacetResult{
		{
			Field:       "dc_creator",
			FacetType:   "enum",
			TotalValues: 1,
			Values: []domainSemantic.FacetValueResult{
				{Value: "Rembrandt", Count: 100},
			},
		},
	}
	opts := &domainSemantic.QueryOptions{
		Query: &domainSemantic.TextQuery{Value: "test"},
	}

	facets := svc.buildFacets(r, facetResults, opts)

	val := facets[0].Values[0]

	// Backwards compat: ApplyURL and RemoveURL should still be set
	if val.ApplyURL == "" {
		t.Error("expected ApplyURL to still be set for backwards compatibility")
	}

	if val.RemoveURL == "" {
		t.Error("expected RemoveURL to still be set for backwards compatibility")
	}

	// New Filter field should also be set
	if val.Filter == "" {
		t.Error("expected Filter string to be set")
	}
}

func TestQueryContextRefJSONFields(t *testing.T) {
	ref := &domainSemantic.QueryContextRef{
		ID:      "/contexts/query/ctx_a7f3",
		Expires: "2026-02-18T15:30:00Z",
	}

	if ref.ID != "/contexts/query/ctx_a7f3" {
		t.Errorf("ID = %s, want /contexts/query/ctx_a7f3", ref.ID)
	}

	if ref.Expires != "2026-02-18T15:30:00Z" {
		t.Errorf("Expires = %s, want 2026-02-18T15:30:00Z", ref.Expires)
	}
}

func TestTimingJSONFields(t *testing.T) {
	timing := &domainSemantic.Timing{
		Took: 42,
		Unit: "ms",
	}

	if timing.Took != 42 {
		t.Errorf("Took = %d, want 42", timing.Took)
	}

	if timing.Unit != "ms" {
		t.Errorf("Unit = %s, want ms", timing.Unit)
	}
}
