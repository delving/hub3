# V2-to-Semantic API Migration & New ES Backend

## Context

The semantic API v1 (`/api/semantic/v1`) is the sole public search interface going forward. The v2 search API (`/api/search/v2`) is feature-frozen for frontend migration only. The current v2adapter bridge translates semantic API calls to v2 backend calls using the deprecated `olivere/elastic` library. This document defines the migration strategy: close feature gaps in the semantic API, enrich the detail endpoint, build a clean native ES backend, and deprecate v2.

## V2 vs Semantic API: Feature Comparison Matrix

### Core Search & Filtering

| Capability | V2 | Semantic v1 | Status |
|---|---|---|---|
| Free-text query | `q=`, `query=` | `query=` (GET), `query.value` (POST) | Covered |
| Query refinement | `rq=` (AND appended) | — | Drop (expressible as filters) |
| Search fields | `searchFields=` | `query.fields` (POST) | Covered |
| Property filter (eq, neq, in, nin) | `qf=field:value` | `filter[field][op]=value` | Covered |
| Exists filter | `qf.exist=field` | `filter[field][exists]` | Covered |
| Range filter (gt, gte, lt, lte) | Range syntax in qf | `filter[field][gt]=` etc. | Covered |
| Date range | `qf.dateRange=field:[a TO b]` | RangeFilter with dates | Covered |
| Contains / starts-with | Wildcard in qf | `filter[field][contains]=` | Covered |
| Hidden filters | `hqf=field:value` | — | **Gap** |
| ID filter (case-insensitive) | `qf.id=field:value` | — | **Gap** (minor) |

### Faceting

| Capability | V2 | Semantic v1 | Status |
|---|---|---|---|
| Basic facet request | `facet.field=` | `facet[field]=limit` | Covered |
| Facet limit | `facet.size=`, `facet.limit=` | Limit in facet value | Covered |
| Facet sort | `facet.sort=count\|alpha` | POST body only | Partial |
| Facet AND/OR logic | `facetBoolType=and\|or` | — | **Gap** |
| Facet OR between fields | `facet.orBetween=true` | — | **Gap** |
| Facet merge | `facet.mergeFilter=` | — | Drop (rarely used) |
| Facet expand/drill-down | `facet.expand=field` | — | **Gap** |
| Facet cursor | `facet.cursor=encoded` | — | **Gap** |
| Facet value filter | `facet.filter=pattern` | POST body only | Partial |
| Full facet (limit=2000) | `facet.full=field` | — | Drop (expressible as limit) |

### Pagination & Sorting

| Capability | V2 | Semantic v1 | Status |
|---|---|---|---|
| Page-based | `page=`, `rows=` | `page=`, `size=` | Covered |
| Offset-based | `start=` | `pagination.offset` (POST) | Covered |
| Cursor pagination | `searchAfter=` | `cursor=` | Covered |
| Session restore | `scrollID=hex` (full state) | — | **Gap** |
| Sort | `sortBy=`, `sortAsc=` | `sort=-field` | Covered |
| Multi-sort | Single only | Multiple `sort` params | v1 is richer |

### Response Format & Control

| Capability | V2 | Semantic v1 | Status |
|---|---|---|---|
| Item format | 8 formats (summary, fragmentGraph, etc.) | JSON-LD only | By design |
| Response format | JSON, protobuf, ldjson, bulkaction | JSON-LD only | By design |
| Language preference | `lang=en,nl` | `languages=en,nl` | Covered |
| Block enable/disable | `enable=ITEMS` / `disable=LAYOUT` | `detailLevel=` | Different approach |
| Field selection | — | `fields=field1,field2` | v1 is richer |
| Expand relations | — | `expand=relatedAgents` | v1 is richer |
| Cross-index search | `contextIndex=` | — | **Gap** |
| JSONP callback | `callback=fn` | — | Drop (legacy) |

### Specialized Features

| Capability | V2 | Semantic v1 | Decision |
|---|---|---|---|
| Collapse/grouping | `collapseOn`, `collapseSize`, `collapseSort` | — | **Migrate** |
| Peek mode | `peek=field` | — | **Migrate** |
| More-Like-This | `mlt=true`, `mlt.count=5`, `mlt.filterkey=` | — | **Migrate + improve** |
| Geo bbox | `min_x`, `min_y`, `max_x`, `max_y` | `filter[geo][bbox]=` (stubbed) | **Complete** |
| Geo distance | `pt=`, `d=` | `GeoDistanceFilter` (stubbed) | **Complete** |
| Geo clustering | `geoType=CLUSTER` | — | Defer |
| Tree/EAD | 15+ params | — | **Drop** (different domain) |
| Debug/echo | `echo=searchService\|searchResponse` | — | **Migrate** |
| Cache control | `noCache`, `cacheRefresh` | — | **Migrate** |
| Introspection | — | `/introspect/*` | v1 is richer |
| Query context | — | `/contexts/query/*` | v1 is richer |
| API documentation | — | `/docs`, `/` entry point | v1 is richer |
| Field suggestions | — | Levenshtein typo detection | v1 is richer |

### Summary

- **Covered**: 20+ capabilities fully translated
- **Gaps to close**: 12 capabilities to migrate
- **Drop**: 4 capabilities (query refinement, facet merge/full, tree, JSONP)
- **v1 richer**: 7 capabilities that v2 lacks

---

## Architecture

```
                    +-------------------+
   Clients --------+  Semantic API v1   +-------- New clients
                    |  (sole public     |
                    |   interface)       |
                    +--------+----------+
                             |
              +--------------+--------------+
              |              |              |
    +---------v--------+  +--v-------+  +--v-----------+
    | Native ES        |  |V2Adapter |  | Include      |
    | Backend (new)    |  |(bridge,  |  | Providers    |
    | go-elasticsearch |  | frozen)  |  | (MLT, KP...) |
    +------------------+  +----------+  +--------------+
                             |
                    +--------v----------+
                    | V2 API (frozen)   +-------- Legacy frontend
                    | (no new features) |         (during migration)
                    +-------------------+
```

The semantic API v1 is the sole public interface. The v2adapter bridge remains operational during frontend migration but receives no new features. A new native ES backend replaces both the v2adapter and the deprecated `olivere/elastic` dependency. Include providers handle detail enrichment (related items, knowledge panels).

---

## Detail Endpoint Enrichment

The detail endpoint (`GET /resource/{id}`) becomes a composable view. The base document is always returned; additional sections are opt-in via the `include` query parameter.

### Request format

```
GET /api/semantic/v1/resource/{id}?include=relatedItems&include=knowledgeContext&context=ctx_abc
```

### `include` sections

| Section | Description | Source |
|---|---|---|
| `relatedItems` | Similar/related items as Hydra Collection | ES More-Like-This query |
| `knowledgeContext` | Knowledge panels for linked entities | Entity extraction from document graph |
| `navigation` | Prev/next in search results | Already exists via `?context=` |

Future sections: `provenance`, `versions`, `relatedCollections`.

### `relatedItems`

Uses ES More-Like-This to find similar documents based on the current document's content.

**Optional parameters:**
- `relatedItems.count=5` — number of items (default: 5, max: 20)
- `relatedItems.fields=dc:title,dc:creator` — similarity fields (default: derived from ResourceConfig)

**Response structure:**

```json
{
  "@id": "abc",
  "@type": "edm:ProvidedCHO",
  "dc:title": "The Night Watch",

  "hub3:relatedItems": {
    "@type": "hydra:Collection",
    "totalItems": 12,
    "member": [
      {
        "@id": "def",
        "dc:title": "The Anatomy Lesson",
        "hub3:similarityScore": 0.87
      }
    ],
    "hub3:similarityFields": ["dc:title", "dc:creator", "dc:subject"]
  }
}
```

MLT improvements over v2:
- Configurable per-request similarity fields (v2 uses global `mlt.filterkey`)
- Similarity score exposed per result
- Standard Hydra Collection format (not a separate response structure)
- Integrated into the document view (not a separate request)

### `knowledgeContext`

Extracts and groups linked entities from the document's own graph, organized by EDM type. Initially requires no additional ES queries — built from the document's inlined resources.

**Response structure:**

```json
{
  "hub3:knowledgeContext": {
    "@type": "hub3:KnowledgePanel",
    "hub3:agents": [
      {
        "@id": "http://example.org/agent/rembrandt",
        "@type": "edm:Agent",
        "skos:prefLabel": "Rembrandt van Rijn",
        "hub3:role": "dc:creator"
      }
    ],
    "hub3:places": [
      {
        "@id": "http://example.org/place/amsterdam",
        "@type": "edm:Place",
        "skos:prefLabel": "Amsterdam"
      }
    ],
    "hub3:concepts": [
      {
        "@id": "http://example.org/concept/oil-painting",
        "@type": "skos:Concept",
        "skos:prefLabel": "Oil painting"
      }
    ],
    "hub3:timespans": [
      {
        "@id": "http://example.org/timespan/17th-century",
        "@type": "edm:TimeSpan",
        "skos:prefLabel": "17th century"
      }
    ]
  }
}
```

Future enrichment: external knowledge bases, Wikidata links, collection statistics per entity.

### SearchStore interface extension

```go
type SearchStore interface {
    // ... existing methods ...

    // FindSimilar returns documents similar to the given document.
    FindSimilar(ctx context.Context, id string, opts *SimilarOptions, config *ResourceConfig) (*SearchResult, error)
}

type SimilarOptions struct {
    // Count is the maximum number of similar items to return.
    Count int
    // Fields are the fields to use for similarity computation.
    // If empty, defaults are derived from ResourceConfig.
    Fields []string
    // MinTermFreq is the minimum term frequency for MLT.
    MinTermFreq int
    // MinDocFreq is the minimum document frequency for MLT.
    MinDocFreq int
}
```

Knowledge context extraction does not need a store method — it is built from the document's own graph data at the service layer.

---

## Phase A: Feature Gap Closure

Close the remaining search feature gaps so the semantic API can fully replace v2.

### A1: Collapse/Result Grouping

Add `CollapseOptions` to `QueryOptions`:

```go
type CollapseOptions struct {
    Field string // Field to collapse on (deduplicate by)
    Size  int    // Inner hits per group (default: 1)
    Sort  []SortField // Sort for inner hits
}
```

GET: `?collapse=field&collapse.size=3&collapse.sort=-dc:date`
POST: `"collapse": {"field": "edm:dataProvider", "size": 3}`

### A2: Facet Boolean Logic

Add `FacetBoolType` to `QueryOptions`:

```go
type FacetBoolType string

const (
    FacetBoolOr  FacetBoolType = "or"  // default: selected values broaden results
    FacetBoolAnd FacetBoolType = "and" // selected values narrow results
)
```

GET: `?facetBool=and`
POST: `"facetBoolType": "and"`

### A3: Facet Expansion & Cursor

Add cursor-based pagination for facet values:

GET: `?facet[dc:creator]=50&facet[dc:creator].cursor=abc&facet[dc:creator].sort=count`
POST: `"facets": [{"field": "dc:creator", "limit": 50, "cursor": "abc"}]`

Response includes `hub3:nextCursor` in the facet result for continuation.

### A4: Hidden Filters

Add `Hidden` flag to filters:

```go
type PropertyFilter struct {
    // ... existing fields ...
    Hidden bool // If true, not shown in activeFilters response
}
```

GET: `?hfilter[orgID][eq]=museum-x` (h-prefix = hidden)
POST: `"filters": [{"field": "orgID", "operator": "eq", "value": "museum-x", "hidden": true}]`

### A5: Peek Mode

Facet-only query that returns zero items:

GET: `?peek=dc:creator,dc:type` (shorthand: request only these facets, size=0)
POST: `"peek": true` or `"pagination": {"size": 0}` with facets

### A6: Debug Mode

GET: `?debug=query` (returns the ES query DSL in response metadata)
POST: `"debug": "query"`

Response includes `hub3:debug` section with the generated ES query, timing breakdown, and shard info.

---

## Phase B: Detail Enrichment

### B1: Include Parameter & Provider Architecture

Add `include` query parameter to `handleResourceDetail`. Each include section is handled by a provider:

```go
type IncludeProvider interface {
    Name() string
    Provide(ctx context.Context, doc map[string]any, config *ResourceConfig) (any, error)
}
```

Providers are registered on the service and invoked when their name appears in `?include=`.

### B2: Related Items (MLT)

Implement `FindSimilar` in the native ES backend using ES More-Like-This query. Register as `relatedItems` include provider.

### B3: Knowledge Context

Implement entity extraction from the document's graph. Group linked resources by EDM type (`edm:Agent`, `edm:Place`, `skos:Concept`, `edm:TimeSpan`). Register as `knowledgeContext` include provider.

---

## Phase C: New Native ES Backend

Replace the v2adapter bridge and `olivere/elastic` with a clean implementation using the official `go-elasticsearch/v8` client.

### C1: ES Client Layer

New package: `ikuzo/storage/x/elasticsearch8/`

- Direct JSON query building (no ORM, no reflection)
- Uses `esapi` for typed requests
- Connection pooling, retries, circuit breaking
- ES 7.x and 8.x compatibility

### C2: Native SearchStore

Implement `semantic.SearchStore` interface directly against ES:

- `Search`: Build ES query JSON from `QueryOptions`, execute, transform results
- `GetByID`: Direct `_doc` get
- `Aggregate`: Build aggregation JSON from `FacetRequest`
- `FindSimilar`: ES More-Like-This query
- `SaveSearchContext` / `GetSearchContext` / `DeleteSearchContext`: ES or Redis

### C3: Native IntrospectionStore

Implement `semantic.IntrospectionStore` directly:

- `IntrospectClasses`: Nested terms aggregation on `resources.types`
- `IntrospectProperties`: Terms aggregation on `resources.entries.searchLabel` with class filter
- `IntrospectField`: Terms aggregation on specific field values
- `IntrospectPaths`: Composite aggregation on predicate paths

### C4: Complete Geo Support

Implement geo filter translation in the native backend:

- Bounding box: ES `geo_bounding_box` query
- Distance: ES `geo_distance` query
- Polygon: ES `geo_polygon` query

### C5: Collapse, Facet Logic, Debug

Implement the Phase A features natively in the ES backend:

- Collapse: ES `collapse` field with `inner_hits`
- Facet AND/OR: Post-filter vs query-filter strategy
- Debug: Return the generated ES query JSON in response metadata

---

## Phase D: V2 Deprecation

### D1: Feature Freeze

Announce v2 API feature freeze. Document the semantic v1 equivalent for every v2 parameter.

### D2: Frontend Migration Guide

Provide a parameter mapping table:

| V2 Parameter | Semantic v1 Equivalent |
|---|---|
| `q=rembrandt` | `query=rembrandt` |
| `qf=dc_creator:Rembrandt` | `filter[dc_creator][eq]=Rembrandt` |
| `facet.field=dc_type` | `facet[dc_type]=20` |
| `rows=20&page=2` | `size=20&page=2` |
| `sortBy=dc_date&sortAsc=false` | `sort=-dc_date` |
| `mlt=true&mlt.count=5` | `include=relatedItems&relatedItems.count=5` |
| `collapseOn=edm_dataProvider` | `collapse=edm_dataProvider` |
| `facetBoolType=and` | `facetBool=and` |

### D3: Cross-Index Search

Add `contextIndex` support to semantic API for multi-tenant cross-search.

### D4: Remove V2 Adapter

Once frontend has fully migrated, remove:
- `ikuzo/storage/x/v2adapter/` package
- `olivere/elastic` dependency
- V2 route registrations

---

## Implementation Priority

```
Phase A (Feature Gaps)        Phase B (Detail Enrichment)
  A1: Collapse                  B1: Include providers
  A2: Facet bool logic          B2: Related items (MLT)
  A3: Facet expansion           B3: Knowledge context
  A4: Hidden filters
  A5: Peek mode                 Phase D (Deprecation)
  A6: Debug mode                  D1: Feature freeze
                                  D2: Migration guide
Phase C (Native ES Backend)     D3: Cross-index
  C1: ES client layer            D4: Remove v2adapter
  C2: Native SearchStore
  C3: Native IntrospectionStore
  C4: Geo support
  C5: Collapse/facet/debug

Dependency graph:
  A* -> C5 (features defined first, then implemented natively)
  B1 -> B2, B3 (provider architecture first)
  C1 -> C2, C3, C4, C5 (client layer first)
  C* -> D4 (native backend replaces adapter)
```

Phase A and B can proceed in parallel since they touch different parts of the codebase (query domain vs detail handler). Phase C depends on Phase A for feature definitions. Phase D is sequential after C.

---

## Decisions

1. **Semantic API v1 is the sole public interface** — all features expressed through it
2. **V2 API is feature-frozen** — kept alive only for frontend migration
3. **olivere/elastic must go** — replaced by `go-elasticsearch/v8`
4. **Tree/EAD is out of scope** — separate domain, not migrated
5. **MLT lives on the detail endpoint** — `?include=relatedItems`, not a separate endpoint
6. **Knowledge panels are graph-derived** — no additional queries initially, enrichable later
7. **Include mechanism is extensible** — new sections added without breaking existing clients
