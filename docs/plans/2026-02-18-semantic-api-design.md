# Semantic API v1: Design Document

**Date:** 2026-02-18
**Status:** Approved
**Scope:** Hub3 backend API design; informs DIW frontend integration

---

## 1. Vision and Goals

The Semantic API is a self-describing, schema-agnostic data platform for querying RDF records. It is designed as a 10+ year foundation with Go v1-level stability guarantees.

### Core Principles

1. **Self-documenting** -- the API explains its capabilities and the data it serves, generated from code and index state, never maintained separately
2. **Self-discovering** -- clients explore what classes, properties, paths, and values exist in the data without prior knowledge
3. **Request-driven configuration** -- clients send query + config together via POST; the server adapts to what is requested rather than exposing a fixed format
4. **Backend agnostic** -- Elasticsearch today, SPARQL or graph stores tomorrow; the API surface stays stable
5. **Easy, predictable, reusable** -- simple things are simple (Level 1 queries); complex things are possible (path queries); everything is discoverable

### Non-Goals

- Replacing the v2 search API (which continues to serve existing consumers)
- Implementing a full SPARQL endpoint
- Real-time data ingestion (the bulk API handles that)

---

## 2. Architecture: Two Sides

The API separates **structure** (what the API does) from **content** (what the data contains).

### Side A: Structure

- Operations: search, detail, aggregate, introspect
- Query language: filters, facets, sort, pagination, paths
- Response shapes: Hydra Collection, detail view, introspection result
- Self-documentation: capabilities, operators, endpoints
- Query context management: stored queries for pagination and follow-up

Structure never hardcodes knowledge about content. No vocabulary-specific code in the API layer.

### Side B: Content

- JSON-LD documents produced by the semantic format (SemanticView)
- Opaque to the API structure -- the API doesn't interpret document contents
- Discoverable via the introspection layer
- Optionally annotated with record-definition schemas (recDefID)

### How They Connect

The **introspection layer** bridges the two sides. It examines the actual indexed data and reports what is there: classes, properties, value distributions, paths. When records were ingested with a record-definition (e.g., `edm-v2.1`), introspection enriches its responses with that schema knowledge. Without a recdef, introspection still works -- just without the documentation overlay.

### Virtual Datasets (Future)

The architecture naturally supports virtual datasets: a stored query config with a hidden base query that wraps all user queries transparently. The API serves slices of ingested data as if they were standalone collections. This is not in scope for v1 but the query context design supports it without changes.

---

## 3. API Surface

Base path: `/api/semantic/v1`

### 3.1 Search

| Method | Path | Description |
|--------|------|-------------|
| GET | `/search` | Simple search with query parameters |
| POST | `/search` | Search with inline config |
| GET | `/search/{id}` | Detail view with optional navigation context |

**GET parameters:**

| Parameter | Example | Description |
|-----------|---------|-------------|
| `q` | `q=amsterdam` | Full-text query |
| `filter[field][op]` | `filter[dc_creator][eq]=Rembrandt` | Filter by field |
| `facet[field]` | `facet[dc_creator]=20` | Request facet with size |
| `sort` | `sort=-dc_date` | Sort (prefix `-` for descending) |
| `page` | `page=3` | Page number (1-based) |
| `size` | `size=20` | Results per page (max 100) |
| `context` | `context=ctx_a7f3` | Query context for pagination |

**POST body:**

```json
{
  "query": { "text": "amsterdam", "operator": "AND" },
  "filters": [
    { "field": "dc_creator", "operator": "eq", "value": "Rembrandt" }
  ],
  "config": {
    "facets": [
      { "field": "dc_creator", "size": 20 },
      { "field": "edm_type", "size": 10 }
    ],
    "fields": ["dc_title", "dc_creator", "edm_object"],
    "sort": [{ "field": "dc_date", "order": "desc" }],
    "detailLevel": "standard"
  }
}
```

The `config` block replaces server-side ConfigRegistry. The server validates it against what actually exists in the data.

### 3.2 Introspect

All introspect endpoints accept an optional `?context={ctx}` parameter to scope introspection to a specific result set.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/introspect` | Overview: classes, schemas, dataset stats |
| GET | `/introspect/classes` | All classes with document counts |
| GET | `/introspect/classes/{class}/properties` | Properties for a class |
| GET | `/introspect/fields/{field}` | Value distribution for a field |
| GET | `/introspect/paths` | Predicate paths between classes |

### 3.3 Query Context

| Method | Path | Description |
|--------|------|-------------|
| POST | `/contexts/query` | Create context from search state |
| GET | `/contexts/query/{id}` | Retrieve context state |
| DELETE | `/contexts/query/{id}` | Release context |

Query contexts are short-lived resources (default 15 minute TTL, extended on use). They capture query state, result metadata, and enable:
- Consistent pagination across pages
- Detail navigation (previous/next within results)
- Introspection of a specific result set
- Future: export, virtual datasets

Context IDs are short, human-readable, URL-safe strings. No base64 tokens.

### 3.4 Documentation

| Method | Path | Description |
|--------|------|-------------|
| GET | `/docs` | API capabilities (generated from code) |

### 3.5 JSON-LD Context

| Method | Path | Description |
|--------|------|-------------|
| GET | `/contexts/jsonld/{name}` | JSON-LD @context documents |

---

## 4. Path-Based Querying

RDF data is a graph. The API supports progressive levels of path complexity for querying, filtering, faceting, and sorting.

### Level 1: Simple Field (80% of use cases)

```
filter[dc_creator][eq]=Rembrandt
facet[dc_creator]=20
```

Matches the searchLabel on any resource in the document. No path knowledge required.

### Level 2: Class-Scoped Field

```
filter[edm:ProvidedCHO/dc_creator][eq]=Rembrandt
facet[edm:ProvidedCHO/dc_subject]=20
```

Restricts to a specific resource type. Uses the context path stored on each resource.

### Level 3: Multi-Hop Path

```
filter[dc_creator/foaf_name][eq]=Rembrandt
facet[dc_subject/skos_prefLabel]=20
sort=dc_date/edm_begin
```

Traverses the resource graph: follow `dc_creator` to the linked resource, get its `foaf_name`.

### Level 4: Typed Multi-Hop (Maximum Precision)

```
filter[edm:ProvidedCHO/dc_creator/edm:Agent/foaf_name][eq]=Rembrandt
```

Class and predicate alternating. Rarely needed but available for precision.

### How Paths Map to Elasticsearch

The FragmentGraph stores context paths on each resource. Path queries translate to nested queries matching context + entries:

- Level 1: nested on `resources.entries.searchLabel`
- Level 2: nested on `resources.types` + `resources.entries.searchLabel`
- Level 3: nested matching context path + entry searchLabel
- Level 4: nested matching context path with type constraints at each hop

The `/introspect/paths` endpoint reports what paths actually exist in the data so clients do not have to guess.

### POST Config Paths

In POST body, paths are plain strings with slash separators:

```json
{
  "config": {
    "facets": [
      { "field": "dc_creator", "size": 20 },
      { "field": "dc_creator/foaf_name", "size": 20 }
    ]
  }
}
```

---

## 5. Label Resolution for Mixed Literal/Resource Fields

### The Problem

When faceting on `dc_creator`, results may contain:
- Literal values: `"Rembrandt van Rijn"` (direct string)
- Resource links: `http://data.rkd.nl/artists/66219` (URI pointing to an Agent)

Users expect readable labels in both cases.

### Solution: Index-Time Label Merging

When the indexing pipeline encounters a Resource-type entry whose linked resource is inlined in the same graph, it resolves the best label and merges it into the entry.

**Before (current):**

```json
{
  "entrytype": "Resource",
  "@id": "http://data.rkd.nl/artists/66219",
  "searchLabel": "dc_creator"
}
```

**After (with label resolution):**

```json
{
  "entrytype": "Resource",
  "@id": "http://data.rkd.nl/artists/66219",
  "@value": "Rembrandt van Rijn",
  "searchLabel": "dc_creator",
  "resolvedFrom": "foaf_name",
  "resolvedLevel": 1
}
```

### New Entry Fields for Provenance

| Field | Type | Description |
|-------|------|-------------|
| `resolvedFrom` | keyword | SearchLabel of the property the label was resolved from. Absent on direct literals. |
| `resolvedLevel` | integer | How many hops away the label was found (1 = direct child of linked resource) |

### Label Resolution Priority

At index time, when resolving a resource link to a label, predicates are checked in order:

1. `skos:prefLabel`
2. `rdfs:label`
3. `foaf:name`
4. `dc:title`
5. `schema:name`

First match wins. Multiple language variants are stored as separate entries with different `@language` values.

### Benefits

- Faceting on `dc_creator` returns readable values regardless of literal vs resource
- Clients distinguish direct literals from resolved labels via `resolvedFrom` presence
- The `@id` is preserved for linking to source resources
- No query-time cost -- resolution happens during indexing
- The introspection endpoint reports `hasResolvedLabels: true` on fields where this was applied

### Not Solved (Future Design Work)

- **External resource resolution** -- linked resources NOT inlined in the graph require a lookup service at index time or a separate enrichment step
- **Label preference per query** -- letting clients choose which label predicate to display. The path query system (Level 3) serves as an escape hatch: `facet[dc_creator/foaf_name]=20`
- **Staleness** -- resolved values are stale until re-index if the source label changes. Acceptable for most use cases.

### Impact on Existing Code

- Enhancement to the indexing pipeline (`ResourceMap.SetResources` or `index.Entry` creation)
- ES mapping already has `@id` and `@value` on entries, so Resource-type entries can carry `@value` without mapping changes
- `resolvedFrom` and `resolvedLevel` fields need to be added to entry mapping

---

## 6. Response Contracts

### 6.1 Base Envelope

Every response includes:

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "...",
  "hub3:timing": { "took": 42, "unit": "ms" }
}
```

The `@context` is always a URL reference. Content-specific namespaces (dc, edm, etc.) are included in the context document.

### 6.2 Search Response (hydra:Collection)

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "hydra:Collection",
  "hydra:totalItems": 12847,

  "hydra:member": [
    { "@id": "...", "@type": ["edm:ProvidedCHO"], "dc:title": "..." }
  ],

  "hydra:view": {
    "@type": "hydra:PartialCollectionView",
    "hydra:first": "/search?context=ctx_a7f3&page=1",
    "hydra:previous": "/search?context=ctx_a7f3&page=2",
    "hydra:next": "/search?context=ctx_a7f3&page=4"
  },

  "hub3:facets": [
    {
      "field": "dc_creator",
      "type": "enum",
      "totalValues": 3421,
      "values": [
        {
          "value": "Rembrandt van Rijn",
          "count": 234,
          "filter": "filter[dc_creator][eq]=Rembrandt+van+Rijn"
        }
      ]
    }
  ],

  "hub3:activeFilters": [
    {
      "field": "edm_type",
      "operator": "eq",
      "value": "IMAGE",
      "remove": "/search?context=ctx_a7f3&q=amsterdam"
    }
  ],

  "hub3:queryContext": {
    "@id": "/contexts/query/ctx_a7f3",
    "expires": "2026-02-18T15:30:00Z"
  },

  "hub3:timing": { "took": 42, "unit": "ms" }
}
```

**Pagination:** context-based by default. Links always reference the query context. Three links maximum: `first`, `previous` (absent on page 1), `next` (absent on last page). No `last` link -- computing exact last page is expensive and rarely useful. The client knows it reached the end when `next` is absent.

**Facet values** include ready-to-use `filter` strings. Active filters include `remove` URLs. Clients do not need to construct filter URLs.

**Content** (`hydra:member`) contains JSON-LD documents as-is. The API does not reshape them.

### 6.3 Detail Response (hub3:DetailView)

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "hub3:DetailView",

  "hub3:item": {
    "@id": "https://example.org/record/123",
    "@type": ["edm:ProvidedCHO"],
    "dc:title": { "nl": "Nachtwacht", "en": "Night Watch" },
    "dc:creator": [
      {
        "@id": "http://data.rkd.nl/artists/66219",
        "@type": ["edm:Agent"],
        "foaf:name": "Rembrandt van Rijn"
      }
    ],
    "@graph": []
  },

  "hub3:navigation": {
    "context": "ctx_a7f3",
    "position": 47,
    "totalResults": 12847,
    "previous": "/search/record-122?context=ctx_a7f3",
    "next": "/search/record-124?context=ctx_a7f3",
    "backToSearch": "/search?context=ctx_a7f3&page=3"
  },

  "hub3:meta": {
    "spec": "rijksmuseum",
    "orgID": "rijks",
    "hubID": "rijks_rijksmuseum_SK-C-5",
    "recDefID": "edm-v2.1",
    "modified": "2026-01-15T10:30:00Z"
  },

  "hub3:timing": { "took": 12, "unit": "ms" }
}
```

Navigation uses the query context. Without a context parameter on the request, `hub3:navigation` is absent.

### 6.4 Introspection Response (hub3:IntrospectionResult)

**Overview (`/introspect`):**

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "hub3:IntrospectionResult",

  "hub3:scope": {
    "type": "index",
    "totalDocuments": 450000
  },

  "hub3:classes": [
    {
      "uri": "http://www.europeana.eu/schemas/edm/ProvidedCHO",
      "label": "edm:ProvidedCHO",
      "count": 450000,
      "properties": "/introspect/classes/edm:ProvidedCHO/properties"
    }
  ],

  "hub3:schemas": [
    {
      "recDefID": "edm-v2.1",
      "documentCount": 420000,
      "spec": ["rijksmuseum", "boijmans"]
    }
  ]
}
```

When scoped by query context: `hub3:scope.type` is `"query"` and includes the context ID.

**Property introspection (`/introspect/classes/{class}/properties`):**

```json
{
  "@type": "hub3:PropertyIntrospection",
  "hub3:class": "edm:ProvidedCHO",

  "hub3:properties": [
    {
      "field": "dc_creator",
      "predicate": "http://purl.org/dc/elements/1.1/creator",
      "label": "dc:creator",
      "valueTypes": ["Literal", "Resource"],
      "count": 42000,
      "languages": ["nl", "en", "de"],
      "hasResolvedLabels": true,
      "paths": ["dc_creator", "dc_creator/foaf_name", "dc_creator/skos_prefLabel"],
      "schema": {
        "recDefID": "edm-v2.1",
        "documentation": "The person or organization that created the resource"
      }
    },
    {
      "field": "dc_date",
      "predicate": "http://purl.org/dc/elements/1.1/date",
      "label": "dc:date",
      "valueTypes": ["Literal"],
      "count": 38000,
      "dataType": "date",
      "range": { "min": "1400-01-01", "max": "2024-12-31" },
      "paths": ["dc_date"]
    }
  ]
}
```

### 6.5 Error Response (hydra:Error)

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "hydra:Error",
  "hydra:title": "Invalid filter field",
  "hydra:description": "Field 'dc_creatorrr' does not exist. Did you mean 'dc_creator'?",
  "hub3:code": "UNKNOWN_FIELD",
  "hub3:details": {
    "field": "dc_creatorrr",
    "suggestions": ["dc_creator"]
  }
}
```

Errors include field suggestions when applicable. `hub3:code` is machine-readable, `hydra:description` is human-readable.

### 6.6 API Documentation Response (hydra:ApiDocumentation)

```json
{
  "@context": "{baseURL}/contexts/jsonld/api",
  "@type": "hydra:ApiDocumentation",
  "hydra:title": "Semantic Search API",

  "hub3:capabilities": {
    "operators": ["eq", "in", "contains", "gt", "gte", "lt", "lte", "between", "exists"],
    "facetTypes": ["enum", "range", "date", "boolean"],
    "sortOptions": ["asc", "desc"],
    "maxPageSize": 100,
    "defaultPageSize": 20,
    "pathSeparator": "/",
    "contextTTL": "15m"
  }
}
```

Generated from code. Never maintained separately.

### Response Type Summary

| Response | `@type` | Content Location |
|----------|---------|-----------------|
| Search results | `hydra:Collection` | `hydra:member` (opaque JSON-LD) |
| Detail view | `hub3:DetailView` | `hub3:item` (opaque JSON-LD) |
| Introspection | `hub3:IntrospectionResult` | `hub3:classes`, `hub3:properties` |
| Error | `hydra:Error` | `hub3:details` |
| API docs | `hydra:ApiDocumentation` | `hub3:capabilities` |

Content (JSON-LD documents) only appears inside `hydra:member` and `hub3:item`. Everything else is structure. The two sides never mix.

---

## 7. Multi-Tenancy

### Organization Isolation

All operations are scoped by `orgID`, extracted from the request context (set by authentication middleware). Queries always filter by organization. No cross-tenant data leakage.

### Dataset Identification

Within an organization, records are grouped by `spec`. The combination `orgID + spec` uniquely identifies a dataset.

### Document Identity

```
hubID = {orgID}_{spec}_{localID}
```

The `hubID` is the primary key for detail lookups.

---

## 8. Design Items for Future Work

These items are acknowledged and scoped out of v1. The design accommodates them without changes.

| Item | Description | Depends On |
|------|-------------|------------|
| Virtual Datasets | Persistent query context with hidden base query, served as standalone collection | Query context (designed) |
| External Label Resolution | Lookup service for labels on resources not inlined in the graph | Label resolution pipeline (designed) |
| Geospatial Filters | Bounding box, distance, polygon filters | ES geo_point mapping (exists) |
| SPARQL Backend | Alternative SearchStore implementation | Backend agnostic interface (exists) |
| Schema Discovery | Automatic detection of application profiles from indexed data patterns | Introspection (designed) |
| Facet Exploration Endpoint | Dedicated endpoint for deep facet value browsing with cursor pagination | Facets + query context (designed) |
| Configuration Refactoring | Move hardcoded EDM configs to TOML files | POST config override makes this less urgent |

---

## 9. Relationship to Existing Code

### What Stays

- `SearchStore` interface -- correct abstraction, backend agnostic
- `V2SearchAdapter` -- pragmatic bridge for initial implementation
- Hydra Collection response format -- already aligned with this design
- `SemanticView` / semantic format -- produces the JSON-LD content documents
- FragmentGraph indexing pipeline -- provides the data
- ES v2 mapping with nested resources/entries -- supports path queries

### What Changes

- `ConfigRegistry` with hardcoded EDM configs -- replaced by POST config + introspection
- Base64 scroll tokens -- replaced by query context resources
- Static `/docs` endpoint -- enhanced with capabilities from code
- ResourceEntry -- enhanced with `resolvedFrom` and `resolvedLevel` for label resolution

### What's New

- Introspection endpoints (classes, properties, fields, paths)
- Query context CRUD
- Index-time label resolution
- Path query parsing and execution
- Error suggestions (typo detection)

---

## 10. Glossary

| Term | Definition |
|------|------------|
| **Structure** | The API's operations, query language, response shapes, and self-documentation |
| **Content** | The JSON-LD documents served by the API, opaque to the structure layer |
| **Query Context** | A short-lived server-side resource capturing search state for pagination and follow-up |
| **Introspection** | Examining indexed data to discover classes, properties, paths, and value distributions |
| **SearchLabel** | Normalized predicate name (e.g., `dc_creator`) used for filtering and faceting |
| **Path** | Slash-separated predicate chain for multi-hop queries (e.g., `dc_creator/foaf_name`) |
| **RecDef** | Record-definition -- an ingestion-time schema that describes how source data was transformed |
| **Virtual Dataset** | A persistent query context with hidden base query, serving a data slice as a standalone collection |
| **Label Resolution** | Index-time process of resolving resource links to human-readable labels with provenance tracking |
