package elasticsearch8

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestIntrospectionStore_implementsInterface(t *testing.T) {
	// Compile-time interface check.
	var _ semantic.IntrospectionStore = (*IntrospectionStore)(nil)
}

func TestIntrospectionStore_IntrospectClasses(t *testing.T) {
	cannedResponse := `{
		"took": 3,
		"hits": {
			"total": {"value": 100, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"classes": {
				"doc_count": 100,
				"class_types": {
					"buckets": [
						{"key": "http://www.europeana.eu/schemas/edm/ProvidedCHO", "doc_count": 50},
						{"key": "http://www.w3.org/2004/02/skos/core#Concept", "doc_count": 30},
						{"key": "http://xmlns.com/foaf/0.1/Agent", "doc_count": 20}
					]
				}
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())
	ctx := testContext()

	classes, err := store.IntrospectClasses(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(classes) != 3 {
		t.Fatalf("expected 3 classes, got %d", len(classes))
	}

	// Check first class.
	if classes[0].URI != "http://www.europeana.eu/schemas/edm/ProvidedCHO" {
		t.Errorf("expected URI %q, got %q",
			"http://www.europeana.eu/schemas/edm/ProvidedCHO", classes[0].URI)
	}

	if classes[0].Label != "edm:ProvidedCHO" {
		t.Errorf("expected label %q, got %q", "edm:ProvidedCHO", classes[0].Label)
	}

	if classes[0].Count != 50 {
		t.Errorf("expected count 50, got %d", classes[0].Count)
	}

	// Check second class (hash-delimited namespace).
	if classes[1].Label != "skos:Concept" {
		t.Errorf("expected label %q, got %q", "skos:Concept", classes[1].Label)
	}

	if classes[1].Count != 30 {
		t.Errorf("expected count 30, got %d", classes[1].Count)
	}

	// Check third class.
	if classes[2].Label != "foaf:Agent" {
		t.Errorf("expected label %q, got %q", "foaf:Agent", classes[2].Label)
	}
}

func TestIntrospectionStore_IntrospectClasses_WithQuery(t *testing.T) {
	cannedResponse := `{
		"took": 2,
		"hits": {
			"total": {"value": 10, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"classes": {
				"doc_count": 10,
				"class_types": {
					"buckets": [
						{"key": "http://www.europeana.eu/schemas/edm/ProvidedCHO", "doc_count": 10}
					]
				}
			}
		}
	}`

	var capturedBody string

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Body != nil {
			data, _ := io.ReadAll(req.Body)
			capturedBody = string(data)
		}

		return mockResponse(200, cannedResponse), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())
	ctx := testContext()

	opts := &semantic.QueryOptions{
		Query: &semantic.TextQuery{Value: "painting"},
	}

	classes, err := store.IntrospectClasses(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(classes))
	}

	if classes[0].Count != 10 {
		t.Errorf("expected count 10, got %d", classes[0].Count)
	}

	// Verify query_string was used instead of match_all.
	if !strings.Contains(capturedBody, "query_string") {
		t.Error("expected query_string in request body")
	}

	if !strings.Contains(capturedBody, "painting") {
		t.Error("expected 'painting' in request body")
	}
}

func TestIntrospectionStore_IntrospectProperties(t *testing.T) {
	// Response without class filter (direct entries under properties).
	cannedResponse := `{
		"took": 2,
		"hits": {
			"total": {"value": 50, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"properties": {
				"doc_count": 50,
				"entries": {
					"doc_count": 200,
					"predicates": {
						"buckets": [
							{"key": "http://purl.org/dc/elements/1.1/title", "doc_count": 50},
							{"key": "http://purl.org/dc/elements/1.1/creator", "doc_count": 40},
							{"key": "http://purl.org/dc/terms/created", "doc_count": 30}
						]
					}
				}
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())
	ctx := testContext()

	properties, err := store.IntrospectProperties(ctx, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(properties))
	}

	if properties[0].Predicate != "http://purl.org/dc/elements/1.1/title" {
		t.Errorf("expected predicate %q, got %q",
			"http://purl.org/dc/elements/1.1/title", properties[0].Predicate)
	}

	if properties[0].Label != "dc:title" {
		t.Errorf("expected label %q, got %q", "dc:title", properties[0].Label)
	}

	if properties[0].Count != 50 {
		t.Errorf("expected count 50, got %d", properties[0].Count)
	}

	if properties[2].Label != "dcterms:created" {
		t.Errorf("expected label %q, got %q", "dcterms:created", properties[2].Label)
	}
}

func TestIntrospectionStore_IntrospectProperties_WithClassFilter(t *testing.T) {
	// Response with class filter (entries nested under class_filter).
	cannedResponse := `{
		"took": 2,
		"hits": {
			"total": {"value": 50, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"properties": {
				"doc_count": 50,
				"class_filter": {
					"doc_count": 25,
					"entries": {
						"doc_count": 100,
						"predicates": {
							"buckets": [
								{"key": "http://purl.org/dc/elements/1.1/title", "doc_count": 25},
								{"key": "http://purl.org/dc/elements/1.1/creator", "doc_count": 20}
							]
						}
					}
				}
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())
	ctx := testContext()

	properties, err := store.IntrospectProperties(ctx, "http://www.europeana.eu/schemas/edm/ProvidedCHO", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(properties))
	}

	if properties[0].Count != 25 {
		t.Errorf("expected count 25, got %d", properties[0].Count)
	}
}

func TestIntrospectionStore_IntrospectField(t *testing.T) {
	cannedResponse := `{
		"took": 1,
		"hits": {
			"total": {"value": 100, "relation": "eq"},
			"hits": []
		},
		"aggregations": {
			"field_values": {
				"doc_count": 100,
				"entries": {
					"doc_count": 500,
					"filtered": {
						"doc_count": 80,
						"values": {
							"buckets": [
								{"key": "painting", "doc_count": 30},
								{"key": "sculpture", "doc_count": 25},
								{"key": "drawing", "doc_count": 15}
							]
						}
					}
				}
			}
		}
	}`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, cannedResponse), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())
	ctx := testContext()

	info, err := store.IntrospectField(ctx, "http://purl.org/dc/elements/1.1/type", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info == nil {
		t.Fatal("expected non-nil PropertyInfo")
	}

	if info.Predicate != "http://purl.org/dc/elements/1.1/type" {
		t.Errorf("expected predicate %q, got %q",
			"http://purl.org/dc/elements/1.1/type", info.Predicate)
	}

	if info.Label != "dc:type" {
		t.Errorf("expected label %q, got %q", "dc:type", info.Label)
	}

	if info.Count != 80 {
		t.Errorf("expected count 80, got %d", info.Count)
	}
}

func TestIntrospectionStore_NoOrganization(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{}`), nil
	})

	store := NewIntrospectionStore(client, zerolog.Nop())

	// Use a context without organization.
	_, err := store.IntrospectClasses(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for missing organization, got nil")
	}

	if !strings.Contains(err.Error(), "no organization") {
		t.Errorf("expected 'no organization' in error, got %q", err.Error())
	}
}
