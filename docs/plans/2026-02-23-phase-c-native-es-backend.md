# Phase C: Native ES Backend Implementation Plan

> **Current status, 2026-06-04:** Future development. See
> [ADR 0001](../adr/0001-semantic-v1-wraps-v2.md). The current Semantic V1
> delivery scope is a JSON-LD/Hydra wrapper around the existing V2 search
> implementation, not a customer-visible native Elasticsearch backend.

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a new `ikuzo/storage/x/elasticsearch8/` package that implements `SearchStore`, `SimilarStore`, and `IntrospectionStore` using `go-elasticsearch/v8` directly — no olivere dependency. Queries the same v2 index structure.

**Architecture:** Direct JSON query building against ES via `esapi`. A `QueryBuilder` constructs ES query JSON from `semantic.QueryOptions`, an `Executor` sends requests via `go-elasticsearch/v8`, and a `ResultParser` transforms raw ES JSON responses into domain types. The existing v2 nested document model (resources.entries with searchLabel) is preserved.

**Tech Stack:** `github.com/elastic/go-elasticsearch/v8` (already in go.mod v8.16.0), `encoding/json`, `rs/zerolog`

**Constraint:** Side-by-side with existing code. No modifications to `ikuzo/storage/x/elasticsearch/`, `ikuzo/storage/x/v2adapter/`, or `hub3/fragments/`. Config wiring adds a new option to select this backend.

---

## Task 1: Package scaffold and ES client wrapper

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/client.go`
- Test: `ikuzo/storage/x/elasticsearch8/client_test.go`

**Step 1: Write the failing test**

```go
package elasticsearch8

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with defaults", func(t *testing.T) {
		c, err := NewClient(Config{
			Addresses: []string{"http://localhost:9200"},
		})
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if c == nil {
			t.Fatal("NewClient() returned nil")
		}
		if c.indexPrefix != "" {
			t.Errorf("indexPrefix = %q, want empty", c.indexPrefix)
		}
	})

	t.Run("creates client with index prefix", func(t *testing.T) {
		c, err := NewClient(Config{
			Addresses:   []string{"http://localhost:9200"},
			IndexPrefix: "test",
		})
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if c.indexPrefix != "test" {
			t.Errorf("indexPrefix = %q, want %q", c.indexPrefix, "test")
		}
	})

	t.Run("resolves index name", func(t *testing.T) {
		c, err := NewClient(Config{
			Addresses: []string{"http://localhost:9200"},
		})
		if err != nil {
			t.Fatal(err)
		}
		got := c.IndexName("museum")
		want := "museumv2"
		if got != want {
			t.Errorf("IndexName() = %q, want %q", got, want)
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/storage/x/elasticsearch8/ -run TestNewClient -v`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
package elasticsearch8

import (
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"
)

// Config holds configuration for the ES8 client.
type Config struct {
	Addresses   []string
	Username    string
	Password    string
	IndexPrefix string // Optional prefix before orgID
	MaxRetries  int
	Logger      zerolog.Logger
}

// Client wraps go-elasticsearch/v8 for the semantic store.
type Client struct {
	es          *elasticsearch.Client
	indexPrefix string
	log         zerolog.Logger
}

// NewClient creates a new ES8 client.
func NewClient(cfg Config) (*Client, error) {
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:  cfg.Addresses,
		Username:   cfg.Username,
		Password:   cfg.Password,
		MaxRetries: maxRetries,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}

	return &Client{
		es:          es,
		indexPrefix: cfg.IndexPrefix,
		log:         cfg.Logger.With().Str("component", "es8_semantic").Logger(),
	}, nil
}

// NewClientFromExisting wraps an existing go-elasticsearch client.
// This allows reusing the client already created by ikuzo/driver/elasticsearch.
func NewClientFromExisting(es *elasticsearch.Client, logger zerolog.Logger) *Client {
	return &Client{
		es:  es,
		log: logger.With().Str("component", "es8_semantic").Logger(),
	}
}

// IndexName returns the ES index name for an organization.
// Convention: lowercase(orgID) + "v2"
func (c *Client) IndexName(orgID string) string {
	base := strings.ToLower(orgID) + "v2"
	if c.indexPrefix != "" {
		return c.indexPrefix + "_" + base
	}
	return base
}

// ES returns the underlying go-elasticsearch client for direct access.
func (c *Client) ES() *elasticsearch.Client {
	return c.es
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/storage/x/elasticsearch8/ -run TestNewClient -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/storage/x/elasticsearch8/client.go ikuzo/storage/x/elasticsearch8/client_test.go
git commit -m "feat(es8): add elasticsearch8 package scaffold with client wrapper"
```

---

## Task 2: JSON query builder — bool queries and text search

Build ES query JSON from `semantic.QueryOptions`. No olivere — pure `map[string]any` → JSON.

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/query_builder.go`
- Test: `ikuzo/storage/x/elasticsearch8/query_builder_test.go`

**Step 1: Write the failing test**

```go
package elasticsearch8

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestQueryBuilder_BuildQuery(t *testing.T) {
	t.Run("empty options returns match_all", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		body, err := qb.BuildQuery(&semantic.QueryOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		query := m["query"].(map[string]any)
		if _, ok := query["match_all"]; !ok {
			t.Errorf("expected match_all, got %v", query)
		}
	})

	t.Run("text query produces query_string", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Query: &semantic.TextQuery{Value: "Rembrandt"},
		}
		body, err := qb.BuildQuery(opts)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		query := m["query"].(map[string]any)
		boolQ := query["bool"].(map[string]any)
		must := boolQ["must"].([]any)
		if len(must) < 1 {
			t.Fatal("expected at least 1 must clause")
		}
	})

	t.Run("text query with fields produces multi_match", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Query: &semantic.TextQuery{
				Value:  "art",
				Fields: []string{"dc_creator", "dc_title"},
			},
		}
		body, err := qb.BuildQuery(opts)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		query := m["query"].(map[string]any)
		boolQ := query["bool"].(map[string]any)
		must := boolQ["must"].([]any)
		found := false
		for _, clause := range must {
			cm := clause.(map[string]any)
			if _, ok := cm["multi_match"]; ok {
				found = true
			}
		}
		if !found {
			t.Error("expected multi_match clause")
		}
	})

	t.Run("fuzzy text query adds fuzziness", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Query: &semantic.TextQuery{Value: "Rembrant", Fuzzy: true},
		}
		body, err := qb.BuildQuery(opts)
		if err != nil {
			t.Fatal(err)
		}
		s := string(body)
		if !containsSubstring(s, "fuzziness") {
			t.Error("expected fuzziness in query")
		}
	})

	t.Run("includes orgID filter", func(t *testing.T) {
		qb := NewQueryBuilder("museum")
		body, err := qb.BuildQuery(&semantic.QueryOptions{
			Query: &semantic.TextQuery{Value: "test"},
		})
		if err != nil {
			t.Fatal(err)
		}
		s := string(body)
		if !containsSubstring(s, "meta.orgID") || !containsSubstring(s, "museum") {
			t.Errorf("expected orgID filter, got %s", s)
		}
	})

	t.Run("includes docType filter", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		body, err := qb.BuildQuery(&semantic.QueryOptions{
			Query: &semantic.TextQuery{Value: "test"},
		})
		if err != nil {
			t.Fatal(err)
		}
		s := string(body)
		if !containsSubstring(s, "meta.docType") || !containsSubstring(s, "fragmentGraph") {
			t.Errorf("expected docType filter, got %s", s)
		}
	})

	t.Run("includes pagination", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Pagination: &semantic.Pagination{Page: 2, Size: 25},
		}
		body, err := qb.BuildQuery(opts)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		from := m["from"].(float64)
		size := m["size"].(float64)
		if from != 25 {
			t.Errorf("from = %v, want 25", from)
		}
		if size != 25 {
			t.Errorf("size = %v, want 25", size)
		}
	})

	t.Run("includes sorting", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Sort: []semantic.SortField{
				{Field: "dc_date", Direction: semantic.SortDesc},
			},
		}
		body, err := qb.BuildQuery(opts)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		if _, ok := m["sort"]; !ok {
			t.Error("expected sort in query")
		}
	})
}

func containsSubstring(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && json.Valid([]byte(s)) && contains(s, sub)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/storage/x/elasticsearch8/ -run TestQueryBuilder -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package elasticsearch8

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// QueryBuilder constructs Elasticsearch query JSON from semantic.QueryOptions.
type QueryBuilder struct {
	orgID string
}

// NewQueryBuilder creates a new query builder for the given organization.
func NewQueryBuilder(orgID string) *QueryBuilder {
	return &QueryBuilder{orgID: orgID}
}

// BuildQuery constructs the full ES search request body as JSON.
func (qb *QueryBuilder) BuildQuery(opts *semantic.QueryOptions) ([]byte, error) {
	body := map[string]any{}

	// Build the query section
	query := qb.buildQueryClause(opts)
	body["query"] = query

	// Pagination
	if opts.Pagination != nil {
		from := (opts.Pagination.Page - 1) * opts.Pagination.Size
		body["from"] = from
		body["size"] = opts.Pagination.Size
	}

	// Sorting
	if len(opts.Sort) > 0 {
		body["sort"] = qb.buildSort(opts.Sort)
	}

	// Track total hits
	body["track_total_hits"] = true

	return json.Marshal(body)
}

// buildQueryClause builds the "query" portion of the ES request.
func (qb *QueryBuilder) buildQueryClause(opts *semantic.QueryOptions) map[string]any {
	must := []any{}
	filter := []any{}

	// Always filter by docType and orgID
	filter = append(filter,
		map[string]any{"term": map[string]any{"meta.docType": "fragmentGraph"}},
		map[string]any{"term": map[string]any{"meta.orgID": qb.orgID}},
	)

	// Text query
	if opts.Query != nil && opts.Query.Value != "" {
		must = append(must, qb.buildTextQuery(opts.Query))
	}

	// Filters
	for _, f := range opts.Filters {
		fq := qb.buildFilter(f)
		if fq != nil {
			filter = append(filter, fq)
		}
	}

	// If only filters (no must clauses), use match_all as base
	if len(must) == 0 && len(filter) == 0 {
		return map[string]any{"match_all": map[string]any{}}
	}

	boolQ := map[string]any{}
	if len(must) > 0 {
		boolQ["must"] = must
	}
	if len(filter) > 0 {
		boolQ["filter"] = filter
	}

	return map[string]any{"bool": boolQ}
}

// buildTextQuery builds a text search clause.
func (qb *QueryBuilder) buildTextQuery(q *semantic.TextQuery) map[string]any {
	if len(q.Fields) > 0 {
		// Multi-match across specified fields
		mm := map[string]any{
			"query":  q.Value,
			"fields": q.Fields,
		}
		if q.Fuzzy {
			mm["fuzziness"] = "AUTO"
		}
		if q.Operator != "" {
			mm["operator"] = strings.ToLower(q.Operator)
		}
		return map[string]any{"multi_match": mm}
	}

	// Default: query_string across all fields (uses full_text)
	qs := map[string]any{
		"query":            q.Value,
		"default_field":    "full_text",
		"default_operator": "and",
	}
	if q.Fuzzy {
		qs["fuzziness"] = "AUTO"
		qs["fuzzy_prefix_length"] = 2
	}
	return map[string]any{"query_string": qs}
}

// buildSort builds the "sort" array for ES.
func (qb *QueryBuilder) buildSort(sorts []semantic.SortField) []any {
	result := make([]any, 0, len(sorts))
	for _, s := range sorts {
		order := "asc"
		if s.Direction == semantic.SortDesc {
			order = "desc"
		}
		result = append(result, map[string]any{
			s.Field: map[string]any{"order": order},
		})
	}
	return result
}

// buildFilter converts a semantic.Filter to an ES query clause.
func (qb *QueryBuilder) buildFilter(f semantic.Filter) map[string]any {
	switch ft := f.(type) {
	case *semantic.PropertyFilter:
		return qb.buildPropertyFilter(ft)
	case *semantic.RangeFilter:
		return qb.buildRangeFilter(ft)
	case *semantic.ExistsFilter:
		return map[string]any{
			"exists": map[string]any{"field": translateField(ft.FieldName)},
		}
	case *semantic.GeoBBoxFilter:
		return qb.buildGeoBBoxFilter(ft)
	case *semantic.GeoDistanceFilter:
		return qb.buildGeoDistanceFilter(ft)
	case *semantic.GeoPolygonFilter:
		return qb.buildGeoPolygonFilter(ft)
	default:
		return nil
	}
}

// buildPropertyFilter converts a PropertyFilter to ES query.
func (qb *QueryBuilder) buildPropertyFilter(f *semantic.PropertyFilter) map[string]any {
	field := translateField(f.FieldName)

	switch f.OperatorType {
	case semantic.OpEqual:
		return map[string]any{"term": map[string]any{field + ".keyword": f.Value}}
	case semantic.OpNotEqual:
		return map[string]any{"bool": map[string]any{
			"must_not": []any{
				map[string]any{"term": map[string]any{field + ".keyword": f.Value}},
			},
		}}
	case semantic.OpIn:
		return map[string]any{"terms": map[string]any{field + ".keyword": f.Value}}
	case semantic.OpNotIn:
		return map[string]any{"bool": map[string]any{
			"must_not": []any{
				map[string]any{"terms": map[string]any{field + ".keyword": f.Value}},
			},
		}}
	case semantic.OpContains:
		return map[string]any{"match": map[string]any{field: f.Value}}
	case semantic.OpStartsWith:
		return map[string]any{"prefix": map[string]any{field + ".keyword": f.Value}}
	case semantic.OpGreaterThan:
		return map[string]any{"range": map[string]any{field: map[string]any{"gt": f.Value}}}
	case semantic.OpGreaterEqual:
		return map[string]any{"range": map[string]any{field: map[string]any{"gte": f.Value}}}
	case semantic.OpLessThan:
		return map[string]any{"range": map[string]any{field: map[string]any{"lt": f.Value}}}
	case semantic.OpLessEqual:
		return map[string]any{"range": map[string]any{field: map[string]any{"lte": f.Value}}}
	default:
		return nil
	}
}

// buildRangeFilter converts a RangeFilter to ES range query.
func (qb *QueryBuilder) buildRangeFilter(f *semantic.RangeFilter) map[string]any {
	field := translateField(f.FieldName)
	rangeQ := map[string]any{}

	if f.Min != nil {
		switch f.OperatorType {
		case semantic.OpGreaterThan:
			rangeQ["gt"] = f.Min
		default:
			rangeQ["gte"] = f.Min
		}
	}

	if f.Max != nil {
		switch f.OperatorType {
		case semantic.OpLessThan:
			rangeQ["lt"] = f.Max
		default:
			rangeQ["lte"] = f.Max
		}
	}

	return map[string]any{"range": map[string]any{field: rangeQ}}
}

// buildGeoBBoxFilter converts GeoBBoxFilter to ES geo_bounding_box query.
func (qb *QueryBuilder) buildGeoBBoxFilter(f *semantic.GeoBBoxFilter) map[string]any {
	field := translateField(f.FieldName)
	return map[string]any{
		"geo_bounding_box": map[string]any{
			field: map[string]any{
				"top_left": map[string]any{
					"lat": f.Bounds.North,
					"lon": f.Bounds.West,
				},
				"bottom_right": map[string]any{
					"lat": f.Bounds.South,
					"lon": f.Bounds.East,
				},
			},
		},
	}
}

// buildGeoDistanceFilter converts GeoDistanceFilter to ES geo_distance query.
func (qb *QueryBuilder) buildGeoDistanceFilter(f *semantic.GeoDistanceFilter) map[string]any {
	field := translateField(f.FieldName)
	return map[string]any{
		"geo_distance": map[string]any{
			"distance": f.Distance,
			field: map[string]any{
				"lat": f.Point.Lat,
				"lon": f.Point.Lon,
			},
		},
	}
}

// buildGeoPolygonFilter converts GeoPolygonFilter to ES geo_polygon query.
func (qb *QueryBuilder) buildGeoPolygonFilter(f *semantic.GeoPolygonFilter) map[string]any {
	if f.Polygon == nil || len(f.Polygon.Coordinates) == 0 {
		return nil
	}

	ring := f.Polygon.Coordinates[0]
	points := make([]map[string]any, 0, len(ring))
	for _, coord := range ring {
		if len(coord) >= 2 {
			points = append(points, map[string]any{
				"lat": coord[1], // GeoJSON: [lon, lat]
				"lon": coord[0],
			})
		}
	}

	field := translateField(f.FieldName)
	return map[string]any{
		"geo_polygon": map[string]any{
			field: map[string]any{
				"points": points,
			},
		},
	}
}

// translateField converts API field names to ES field names.
// Example: dc:creator → dc_creator
func translateField(field string) string {
	return strings.ReplaceAll(field, ":", "_")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/storage/x/elasticsearch8/ -run TestQueryBuilder -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/storage/x/elasticsearch8/query_builder.go ikuzo/storage/x/elasticsearch8/query_builder_test.go
git commit -m "feat(es8): add query builder with bool queries, text search, filters, geo, sorting"
```

---

## Task 3: Aggregation builder — facets

Build ES aggregation JSON for facet requests.

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/aggregation_builder.go`
- Test: `ikuzo/storage/x/elasticsearch8/aggregation_builder_test.go`

**Step 1: Write the failing test**

```go
package elasticsearch8

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestAggregationBuilder_BuildAggregations(t *testing.T) {
	t.Run("single facet produces terms aggregation", func(t *testing.T) {
		ab := &AggregationBuilder{}
		facets := []semantic.FacetRequest{
			{Field: "dc_creator", Limit: 20},
		}
		aggs := ab.BuildAggregations(facets)
		data, _ := json.Marshal(aggs)
		s := string(data)
		if !contains(s, "dc_creator") {
			t.Errorf("expected dc_creator in aggs, got %s", s)
		}
		if !contains(s, "terms") {
			t.Errorf("expected terms agg, got %s", s)
		}
	})

	t.Run("multiple facets produce multiple aggregations", func(t *testing.T) {
		ab := &AggregationBuilder{}
		facets := []semantic.FacetRequest{
			{Field: "dc_creator", Limit: 20},
			{Field: "dc_type", Limit: 10},
		}
		aggs := ab.BuildAggregations(facets)
		if len(aggs) != 2 {
			t.Errorf("expected 2 aggs, got %d", len(aggs))
		}
	})

	t.Run("uses nested aggregation for entries fields", func(t *testing.T) {
		ab := &AggregationBuilder{UseNested: true}
		facets := []semantic.FacetRequest{
			{Field: "dc_creator", Limit: 20},
		}
		aggs := ab.BuildAggregations(facets)
		data, _ := json.Marshal(aggs)
		s := string(data)
		if !contains(s, "nested") {
			t.Errorf("expected nested agg, got %s", s)
		}
	})

	t.Run("respects facet sort", func(t *testing.T) {
		ab := &AggregationBuilder{}
		facets := []semantic.FacetRequest{
			{Field: "dc_creator", Limit: 20, Sort: "index"},
		}
		aggs := ab.BuildAggregations(facets)
		data, _ := json.Marshal(aggs)
		s := string(data)
		if !contains(s, "_key") {
			t.Errorf("expected _key sort order, got %s", s)
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/storage/x/elasticsearch8/ -run TestAggregationBuilder -v`

**Step 3: Write implementation**

The `AggregationBuilder` should support both flat field aggregations (for the `fields.*` pattern) and nested aggregations (for the `resources.entries` pattern used in v2 index). The `UseNested` flag determines which pattern to use.

For nested aggregations, the structure mirrors v2:
```json
{
  "dc_creator": {
    "nested": {"path": "resources.entries"},
    "aggs": {
      "filtered": {
        "filter": {"term": {"resources.entries.searchLabel": "dc_creator"}},
        "aggs": {
          "values": {
            "terms": {"field": "resources.entries.@value.keyword", "size": 20}
          }
        }
      }
    }
  }
}
```

For flat fields (fields.* pattern):
```json
{
  "dc_creator": {
    "terms": {"field": "fields.dc_creator.keyword", "size": 20}
  }
}
```

**Step 4: Run test**
**Step 5: Commit**

```bash
git commit -m "feat(es8): add aggregation builder for facets with nested and flat modes"
```

---

## Task 4: Collapse and debug support in query builder

Extend `QueryBuilder.BuildQuery` to support collapse, peek, debug, and facet bool type.

**Files:**
- Modify: `ikuzo/storage/x/elasticsearch8/query_builder.go`
- Modify: `ikuzo/storage/x/elasticsearch8/query_builder_test.go`

**Step 1: Write the failing test**

```go
func TestQueryBuilder_Collapse(t *testing.T) {
	t.Run("collapse produces collapse clause", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Collapse: &semantic.CollapseOptions{Field: "meta.spec", Size: 3},
		}
		body, _ := qb.BuildQuery(opts)
		var m map[string]any
		json.Unmarshal(body, &m)
		collapse, ok := m["collapse"]
		if !ok {
			t.Fatal("expected collapse in query")
		}
		cm := collapse.(map[string]any)
		if cm["field"] != "meta.spec" {
			t.Errorf("collapse field = %v, want meta.spec", cm["field"])
		}
	})

	t.Run("collapse with inner_hits", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Collapse: &semantic.CollapseOptions{Field: "meta.spec", Size: 5},
		}
		body, _ := qb.BuildQuery(opts)
		s := string(body)
		if !contains(s, "inner_hits") {
			t.Error("expected inner_hits in collapse")
		}
	})
}

func TestQueryBuilder_Peek(t *testing.T) {
	t.Run("peek sets size to 0", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Peek: true,
			Facets: []semantic.FacetRequest{{Field: "dc_creator"}},
		}
		body, _ := qb.BuildQuery(opts)
		var m map[string]any
		json.Unmarshal(body, &m)
		size := m["size"].(float64)
		if size != 0 {
			t.Errorf("size = %v, want 0", size)
		}
	})
}

func TestQueryBuilder_Debug(t *testing.T) {
	t.Run("debug adds explain", func(t *testing.T) {
		qb := NewQueryBuilder("testorg")
		opts := &semantic.QueryOptions{
			Debug: "query",
			Query: &semantic.TextQuery{Value: "test"},
		}
		body, _ := qb.BuildQuery(opts)
		var m map[string]any
		json.Unmarshal(body, &m)
		if m["explain"] != true {
			t.Error("expected explain=true for debug mode")
		}
	})
}
```

**Step 2: Run test to verify it fails**
**Step 3: Extend BuildQuery to handle Collapse, Peek, Debug**

In `BuildQuery`:
- If `opts.Collapse != nil`, add `"collapse"` with `"field"` and `"inner_hits"`
- If `opts.Peek`, force `"size": 0`
- If `opts.Debug != ""`, add `"explain": true`

**Step 4: Run test**
**Step 5: Commit**

```bash
git commit -m "feat(es8): add collapse, peek, and debug support to query builder"
```

---

## Task 5: Result parser — parse raw ES JSON responses

Parse ES search response JSON into `semantic.SearchResult`.

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/result_parser.go`
- Test: `ikuzo/storage/x/elasticsearch8/result_parser_test.go`

**Step 1: Write the failing test**

Test with a realistic ES response JSON fixture:
- Parse hits into `[]map[string]any`
- Parse total hits count
- Parse aggregation buckets into `[]FacetResult`
- Handle missing _source gracefully
- Extract `_id` when `@id` not in source

**Step 2: Run test to verify it fails**

**Step 3: Write implementation**

Define ES response structs:
```go
type esSearchResponse struct {
    Took     int              `json:"took"`
    Hits     esHits           `json:"hits"`
    Aggs     map[string]json.RawMessage `json:"aggregations"`
}

type esHits struct {
    Total    esTotal    `json:"total"`
    Hits     []esHit    `json:"hits"`
}

type esTotal struct {
    Value    int64  `json:"value"`
    Relation string `json:"relation"`
}

type esHit struct {
    ID        string          `json:"_id"`
    Score     *float64        `json:"_score"`
    Source    json.RawMessage `json:"_source"`
    InnerHits map[string]esInnerHits `json:"inner_hits"`
}

type esInnerHits struct {
    Hits esHits `json:"hits"`
}
```

The `ResultParser` reads the raw JSON, unmarshals into these structs, then maps to `semantic.SearchResult`.

**Step 4: Run test**
**Step 5: Commit**

```bash
git commit -m "feat(es8): add result parser for ES search responses"
```

---

## Task 6: SearchStore implementation — Search, GetByID, Health

Wire the query builder, executor, and result parser into a `Store` that implements `semantic.SearchStore`.

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/store.go`
- Test: `ikuzo/storage/x/elasticsearch8/store_test.go`

**Step 1: Write the failing test**

Unit test with a mock HTTP transport that returns canned ES responses. This avoids needing a live ES instance.

```go
func TestStore_implementsSearchStore(t *testing.T) {
	var _ semantic.SearchStore = (*Store)(nil)
}

func TestStore_implementsSimilarStore(t *testing.T) {
	var _ semantic.SimilarStore = (*Store)(nil)
}
```

Plus functional tests using a `roundTripFunc` to mock HTTP responses for:
- `Search` — returns items and facets
- `GetByID` — returns single document
- `Health` — cluster health check
- `Search` with empty results — returns zero results

**Step 2: Run test to verify it fails**

**Step 3: Write implementation**

```go
type Store struct {
    client    *Client
    aggMode   AggregationMode // Nested vs Flat
    log       zerolog.Logger
    contextMu sync.RWMutex
    contexts  map[string]*semantic.SearchContext
}
```

Key methods:
- `Search`: extract orgID from context → build query JSON → POST to `/{index}/_search` → parse response
- `GetByID`: `GET /{index}/_doc/{id}` → parse _source
- `Aggregate`: build query with `"size": 0` + aggregations → parse agg response
- `Health`: `GET /_cluster/health`
- `SaveSearchContext` / `GetSearchContext` / `DeleteSearchContext`: in-memory map (same as v2adapter)

**Step 4: Run test**
**Step 5: Commit**

```bash
git commit -m "feat(es8): implement SearchStore with Search, GetByID, Aggregate, Health"
```

---

## Task 7: SimilarStore implementation — FindSimilar (MLT)

**Files:**
- Modify: `ikuzo/storage/x/elasticsearch8/store.go`
- Modify: `ikuzo/storage/x/elasticsearch8/store_test.go`

**Step 1: Write the failing test**

Test `FindSimilar` with mock HTTP transport returning MLT results.

**Step 2: Run test**

**Step 3: Write implementation**

Build MLT query JSON:
```json
{
  "query": {
    "more_like_this": {
      "fields": ["fields.dc_creator", "fields.dc_title"],
      "like": [{"_index": "museumv2", "_id": "doc-123"}],
      "min_term_freq": 2,
      "min_doc_freq": 5,
      "max_query_terms": 15
    }
  },
  "size": 5
}
```

**Step 4: Run test**
**Step 5: Commit**

```bash
git commit -m "feat(es8): implement SimilarStore with MLT query support"
```

---

## Task 8: IntrospectionStore implementation

**Files:**
- Create: `ikuzo/storage/x/elasticsearch8/introspection.go`
- Test: `ikuzo/storage/x/elasticsearch8/introspection_test.go`

**Step 1: Write the failing test**

Test all 4 methods with mock responses.

**Step 2: Write implementation**

Replicate the same aggregation patterns from `v2adapter/introspect.go` but as raw JSON:

- `IntrospectClasses`: nested terms agg on `resources.types`
- `IntrospectProperties`: nested agg with class filter on `resources.entries.predicate`
- `IntrospectField`: terms agg on specific field values
- `IntrospectPaths`: composite agg on predicate paths

**Step 3: Run test**
**Step 4: Commit**

```bash
git commit -m "feat(es8): implement IntrospectionStore with class and property discovery"
```

---

## Task 9: Config wiring — add es8 backend option

Wire the new store into the config so it can be selected.

**Files:**
- Modify: `ikuzo/ikuzoctl/cmd/config/semantic.go`

**Step 1: Write the failing test**

Test that the config creates the new store when `UseES8Backend` is true.

**Step 2: Write implementation**

Add a new config option:
```go
type Semantic struct {
    Enabled       bool   `json:"enabled"`
    BaseURL       string `json:"baseURL"`
    UseV2Adapter  bool   `json:"useV2Adapter"`
    UseES8Backend bool   `json:"useES8Backend"` // NEW: use native go-elasticsearch/v8 backend
}
```

In `AddOptions`, add a third branch:
```go
if s.UseES8Backend {
    es8Client := elasticsearch8.NewClientFromExisting(esClient.ES(), logger)
    store = elasticsearch8.NewStore(es8Client, logger)
    // Also create introspection store
    introspect = elasticsearch8.NewIntrospectionStore(es8Client, logger)
}
```

**Step 3: Run test**
**Step 4: Commit**

```bash
git commit -m "feat(config): add UseES8Backend option for native elasticsearch8 backend"
```

---

## Task 10: Build verification

**Step 1: Run full build**

```bash
go build ./...
```

**Step 2: Run all tests in new package**

```bash
go test ./ikuzo/storage/x/elasticsearch8/... -v -count=1
```

**Step 3: Run all semantic tests**

```bash
go test ./ikuzo/domain/semantic/... -count=1
go test ./ikuzo/service/x/semantic/... -count=1
```

**Step 4: Run staticcheck**

```bash
make staticcheck
```

**Step 5: Verify existing tests still pass**

```bash
go test ./ikuzo/storage/x/v2adapter/ -run "TestQueryTranslator|TestTranslateMLT" -count=1
go test ./ikuzo/storage/x/elasticsearch/ -count=1
```

Expected: All pass, no regressions.
