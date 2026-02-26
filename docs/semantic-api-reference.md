# Hub3 Semantic Search API Reference

> This document is generated from code. Do not edit manually.
> Regenerate with: `go run ./tools/cmd/gendocs`

## Overview

The Hub3 Semantic Search API provides a Hydra-compatible JSON-LD interface
for searching, filtering, and navigating cultural heritage metadata.

- **Base URL:** `/api/semantic/v1`
- **Content-Type:** `application/ld+json`
- **Version:** 1.0.0

---

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/semantic/v1/` | API entry point with navigation links |
| `GET` | `/api/semantic/v1/docs` | Machine-readable API documentation (Hydra ApiDocumentation) |
| `GET` | `/api/semantic/v1/search` | Search resources with URL query parameters |
| `POST` | `/api/semantic/v1/search` | Search resources with JSON-LD query body |
| `GET` | `/api/semantic/v1/resource/{id}` | Get a single resource by ID |
| `GET` | `/api/semantic/v1/resource/{id}?include=relatedItems` | Get resource with related items (MLT) |
| `GET` | `/api/semantic/v1/type/{type}/search` | Search within a specific resource type |
| `POST` | `/api/semantic/v1/type/{type}/search` | Type-scoped search with JSON-LD query body |
| `GET` | `/api/semantic/v1/type/{type}/docs` | Documentation for a specific resource type |
| `GET` | `/api/semantic/v1/introspect/classes` | List all classes in the index |
| `GET` | `/api/semantic/v1/introspect/classes/{class}/properties` | List properties for a class |
| `GET` | `/api/semantic/v1/introspect/fields/{field}` | Get details for a specific field |
| `GET` | `/api/semantic/v1/introspect/paths` | List all field paths in the index |
| `POST` | `/api/semantic/v1/contexts/query/` | Create a search context for detail navigation |
| `GET` | `/api/semantic/v1/contexts/query/{id}` | Get a saved search context |
| `DELETE` | `/api/semantic/v1/contexts/query/{id}` | Delete a search context |

---

## Query Parameters

All search parameters are passed as URL query parameters on `GET` requests.

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `query` | string | Full-text search query | `query=rembrandt` |
| `filter[field][op]` | string | Property filter (see Operators) | `filter[dc_creator][eq]=Rembrandt` |
| `hfilter[field][op]` | string | Hidden filter (not reflected in facet counts) | `hfilter[dc_type][eq]=painting` |
| `facet[field]` | int | Request facet aggregation with limit | `facet[dc_type]=20` |
| `facet[field].sort` | string | Facet sort order: `count` or `index` | `facet[dc_type].sort=index` |
| `facet[field].cursor` | string | Cursor for paginating facet values | `facet[dc_type].cursor=abc` |
| `facetBool` | string | Facet logic: `and` or `or` (default: `or`) | `facetBool=and` |
| `page` | int | Page number (default: 1) | `page=2` |
| `size` | int | Results per page (default: 20) | `size=50` |
| `cursor` | string | Opaque cursor for deep pagination | `cursor=eyJhZnRl...` |
| `sort` | string | Sort field; prefix `-` for descending | `sort=-dc_date` |
| `collapse` | string | Group results by field | `collapse=edm_dataProvider` |
| `collapse.size` | int | Inner hits per group (default: 1) | `collapse.size=3` |
| `collapse.sort` | string | Sort inner hits; prefix `-` for desc | `collapse.sort=-dc_date` |
| `peek` | string | Comma-separated fields for facet-only response (size=0) | `peek=dc_type,dc_creator` |
| `languages` | string | Comma-separated language preferences | `languages=en,nl` |
| `fields` | string | Comma-separated field selection | `fields=dc_title,dc_creator` |
| `expand` | string | Relations to expand inline | `expand=relatedItems` |
| `detailLevel` | string | Response detail: `minimal`, `standard`, `full` | `detailLevel=full` |
| `debug` | string | Diagnostic mode (e.g., `query` shows ES query) | `debug=query` |
| `contextIndex` | string | Additional org IDs for cross-index search (repeatable) | `contextIndex=org-b` |
| `include` | string | Sections to include on detail endpoint | `include=relatedItems` |

---

## Filter Operators

Used in `filter[field][operator]=value` syntax.

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Exact match | `filter[dc_creator][eq]=Rembrandt` |
| `neq` | Not equal to | `filter[dc_type][neq]=painting` |
| `in` | Matches any value in list | `filter[dc_type][in]=painting&filter[dc_type][in]=drawing` |
| `nin` | Does not match any value in list | `filter[dc_type][nin]=sketch` |
| `gt` | Greater than | `filter[dc_date][gt]=1600` |
| `gte` | Greater than or equal to | `filter[dc_date][gte]=1600` |
| `lt` | Less than | `filter[dc_date][lt]=1700` |
| `lte` | Less than or equal to | `filter[dc_date][lte]=1700` |
| `contains` | Contains substring | `filter[dc_title][contains]=night` |
| `startswith` | Starts with prefix | `filter[dc_title][startsWith]=Night` |
| `exists` | Field has any value | `filter[nave_thumbnail][exists]=true` |
| `bbox` | Within bounding box | `filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4` |
| `within` | Within distance of point |  |
| `polygon` | Within polygon |  |
| `intersects` | Intersects with geometry |  |

---

## Filter Syntax

Filters use bracket notation: `filter[field][operator]=value`

Field names use underscores in URLs (`dc_creator`) which map to colons internally (`dc:creator`).

### Regular filters

```
/api/semantic/v1/search?filter[dc_creator][eq]=Rembrandt
```

### Hidden filters

Hidden filters narrow results without affecting facet counts:

```
/api/semantic/v1/search?hfilter[dc_type][eq]=painting&facet[dc_type]=10
```

### Range filters

Combine `gte` and `lte` operators for range queries:

```
/api/semantic/v1/search?filter[dc_date][gte]=1600&filter[dc_date][lte]=1700
```

### Geospatial filters

Bounding box filter with west,south,east,north coordinates:

```
/api/semantic/v1/search?filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4
```

---

## Facet Syntax

Request facet aggregations with `facet[field]=limit`.

```
# Request dc_type facet with top 20 values
/api/semantic/v1/search?facet[dc_type]=20

# Multiple facets
/api/semantic/v1/search?facet[dc_type]=10&facet[dc_creator]=20

# Facet with cursor pagination
/api/semantic/v1/search?facet[dc_creator]=50&facet[dc_creator].cursor=abc123

# Sort facet alphabetically instead of by count
/api/semantic/v1/search?facet[dc_type]=10&facet[dc_type].sort=index
```

### Facet Bool Type

Controls how multiple facet selections combine:

| Value | Behavior |
|-------|----------|
| `or` (default) | Selected values broaden results |
| `and` | Selected values narrow results |

```
/api/semantic/v1/search?facetBool=and&filter[dc_type][eq]=painting
```

### Supported Facet Types

- `enum`
- `range`
- `date`
- `boolean`

---

## Sort Syntax

Use the `sort` parameter with an optional `-` prefix for descending order.

```
# Sort ascending (default)
/api/semantic/v1/search?sort=dc_date

# Sort descending
/api/semantic/v1/search?sort=-dc_date
```

---

## Pagination

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | 1 | Page number (1-based) |
| `size` | 20 | Results per page (max: 1000) |
| `cursor` | - | Opaque cursor for deep pagination |

The response includes a `hydra:PartialCollectionView` with navigation links:

```json
{
  "view": {
    "@type": "PartialCollectionView",
    "first": "/api/semantic/v1/search?query=...&page=1",
    "next": "/api/semantic/v1/search?query=...&page=3",
    "previous": "/api/semantic/v1/search?query=...&page=1"
  }
}
```

---

## Result Collapsing

Group results by a field value (e.g., deduplicate by data provider).

| Parameter | Description |
|-----------|-------------|
| `collapse=field` | Field to group by |
| `collapse.size=N` | Inner hits per group (default: 1) |
| `collapse.sort=field` | Sort order for inner hits (prefix `-` for desc) |

---

## Cross-Index Search

Search across multiple organization indices in a single query using `contextIndex`.

```
# Search primary org plus org-b
/api/semantic/v1/search?query=rembrandt&contextIndex=org-b

# Search across three organizations
/api/semantic/v1/search?query=*&contextIndex=org-a&contextIndex=org-b
```

The parameter is repeatable. The primary organization (resolved from the request domain) is always included.

---

## Search Context & Detail Navigation

The Search Context is a core concept that enables **stateful navigation through
search results at the item level**. It bridges the gap between a search result list
and individual resource detail views, allowing users to step through results with
next/previous links without losing their place.

### How It Works

```
1. User performs a search
   GET /api/semantic/v1/search?query=rembrandt
   → Response includes hub3:searchContext with a token

2. User clicks a result to view details
   GET /api/semantic/v1/resource/{id}?context={token}
   → Response includes hub3:navigation with next/previous/first/last links

3. User navigates to next result
   GET /api/semantic/v1/resource/{next-id}?context={token}
   → Navigation updates to reflect new position
```

### Lifecycle

| Phase | What Happens |
|-------|-------------|
| **Creation** | Automatically created when a search returns results. The context captures the ordered list of result IDs and the original query. |
| **Usage** | Pass the context token via `?context={token}` on resource detail requests to activate navigation. |
| **Expiration** | Contexts expire after 15 minutes of inactivity. Accessing the context extends its TTL. |
| **Deletion** | Contexts can be explicitly deleted or are cleaned up after expiry. |

### Search Response with Context

When a search returns results, the response includes a context reference:

```json
{
  "@type": ["Collection"],
  "totalItems": 150,
  "member": [ "..." ],
  "hub3:searchContext": {
    "id": "ctx_a1b2c3",
    "token": "a1b2c3d4",
    "totalResults": 150,
    "expiresAt": "2026-02-26T15:30:00Z"
  }
}
```

### Resource Detail with Navigation

When you fetch a resource with a context token, the response includes navigation:

```
GET /api/semantic/v1/resource/{id}?context=a1b2c3d4
```

```json
{
  "@id": "http://example.org/resource/42",
  "@type": ["edm:ProvidedCHO"],
  "dc_title": "The Night Watch",
  "hub3:navigation": {
    "@type": "hub3:SearchResultNavigation",
    "position": 5,
    "totalResults": 150,
    "hasNext": true,
    "hasPrevious": true,
    "hydra:first": "/api/semantic/v1/resource/doc-001?context=a1b2c3d4",
    "hydra:previous": "/api/semantic/v1/resource/doc-041?context=a1b2c3d4",
    "hydra:next": "/api/semantic/v1/resource/doc-043?context=a1b2c3d4",
    "hydra:last": "/api/semantic/v1/resource/doc-150?context=a1b2c3d4",
    "hub3:backToSearch": "/api/semantic/v1/search?query=rembrandt&page=1"
  }
}
```

### Navigation Fields

| Field | Type | Description |
|-------|------|-------------|
| `position` | int | Current 1-based position in the result set |
| `totalResults` | int | Total number of results in the search |
| `hasNext` | bool | Whether there is a next result |
| `hasPrevious` | bool | Whether there is a previous result |
| `hydra:first` | URL | Link to the first result in the set |
| `hydra:previous` | URL | Link to the previous result |
| `hydra:next` | URL | Link to the next result |
| `hydra:last` | URL | Link to the last result in the set |
| `hub3:backToSearch` | URL | Link back to the original search results page |

### Context Management API

Contexts are usually created automatically during search, but can also be managed explicitly.

#### Create a Context

```
POST /api/semantic/v1/contexts/query/
Content-Type: application/json

{
  "query": { "text": "rembrandt" },
  "totalResults": 150
}
```

Response (`201 Created`):

```json
{
  "@type": "hub3:QueryContext",
  "id": "a1b2c3d4",
  "totalResults": 150,
  "expiresAt": "2026-02-26T15:30:00Z"
}
```

#### Retrieve a Context

```
GET /api/semantic/v1/contexts/query/{contextID}
```

Returns the context metadata, or `404` if expired/not found.

#### Delete a Context

```
DELETE /api/semantic/v1/contexts/query/{contextID}
```

Returns `204 No Content`.

### Frontend Integration Pattern

A typical frontend integration follows this pattern:

1. **Search page:** Execute search, store the `hub3:searchContext.token` from the response
2. **Detail page:** When navigating to a result, append `?context={token}` to the resource URL
3. **Navigation UI:** Use `hub3:navigation` fields to render next/previous buttons
4. **Back button:** Use `hub3:backToSearch` to return to the search results
5. **Expiration:** If the context is not found (expired), gracefully degrade to a detail
   view without navigation

---

## Property Value Types

Resource properties in the API follow JSON-LD conventions. Every property value
can appear in one of six shapes. Your parser **must** handle all of them.

### Value Type Summary

| Type | JSON Shape | When |
|------|-----------|------|
| Plain literal | `"text"` | No language tag, no datatype |
| Typed literal | `{"@value": 42, "@type": "xsd:integer"}` | Has XSD datatype |
| Language-tagged literal | `{"@value": "naam", "@language": "nl"}` | Has language tag |
| Language map | `{"nl": "naam", "en": "name"}` | Multiple language variants |
| Resource reference | `{"@id": "uri", "@type": [...], ...}` | IRI / linked resource |
| Array | `[value, value, ...]` | Multiple values for one property |

### 1. Plain Literal

A simple string with no additional metadata.

```json
"dc_title": "The Night Watch"
```

### 2. Typed Literal

A value with an explicit XSD datatype. The `@value` is converted to the
appropriate JSON type (number, boolean).

```json
"nave_productionYear": { "@value": 1642, "@type": "xsd:integer" }
"geo_lat":             { "@value": 52.3676, "@type": "xsd:double" }
"nave_hasDigitalObject": { "@value": true, "@type": "xsd:boolean" }
```

Supported XSD types: `xsd:integer`, `xsd:long`, `xsd:float`, `xsd:double`,
`xsd:boolean`, `xsd:dateTime`, `xsd:date`, `xsd:gYear`.

### 3. Language-Tagged Literal

A string with an ISO 639 language tag.

```json
"dc_description": { "@value": "Een schilderij van de nachtwacht", "@language": "nl" }
```

### 4. Language Map

When a property has values in multiple languages, they may appear as a map keyed
by language code. Use the `languages` query parameter to indicate your preference.

```json
"dc_title": {
  "nl": "De Nachtwacht",
  "en": "The Night Watch",
  "fr": "La Ronde de nuit"
}
```

### 5. Resource Reference (Nested Object)

A linked resource contains `@id` (its URI), optionally `@type`, and may include
inline properties. Use `skos_prefLabel` for display text.

```json
"dc_creator": {
  "@id": "https://example.org/agents/rembrandt",
  "@type": ["edm:Agent"],
  "skos_prefLabel": "Rembrandt van Rijn",
  "skos_altLabel": ["Rembrandt Harmensz. van Rijn"],
  "foaf_name": "Rembrandt"
}
```

Resource references can nest further — a Place may contain a broader Place, an
Agent may reference an Organization, etc.

#### Reference Types (`hub3:referenceType`)

Not every resource reference is fully expanded inline. The API annotates
ID-only references with `hub3:referenceType` so your application knows
**why** a resource was not expanded and what to do about it.

| `hub3:referenceType` | Meaning | Action |
|----------------------|---------|--------|
| *(absent)* | Fully expanded — all properties are inline | Use the data directly |
| `"cycle"` | Resource exists in this graph but was already expanded higher in the tree (circular reference) | Navigate up the tree to find the full version, or fetch via detail endpoint |
| `"external"` | Resource is not part of this graph — it lives elsewhere on the web | Fetch via Linked Open Data (LOD) request to the `@id` URI |

**Cycle example** — a collection references its items, and an item references
back to the collection:

```json
"dcterms_isPartOf": {
  "@id": "https://example.org/collection/paintings",
  "@type": ["ore:Aggregation"],
  "dc_title": "Dutch Masters Collection",
  "dcterms_hasPart": {
    "@id": "https://example.org/resource/night-watch",
    "hub3:referenceType": "cycle"
  }
}
```

The inner `dcterms_hasPart` points back to the resource being described.
Instead of creating infinite nesting, the API marks it as a `cycle` reference.

**External example** — a resource links to an authority on another server:

```json
"owl_sameAs": {
  "@id": "http://www.wikidata.org/entity/Q17593201",
  "hub3:referenceType": "external"
}
```

This resource is not part of the current graph. To retrieve its properties,
perform a LOD request to the `@id` URI, or use the API detail endpoint if
the URI belongs to this system:

```
GET /api/semantic/v1/resource/{id}
```

**As a consumer, you benefit from cycle detection automatically** — you will
never receive infinitely nested responses. The `hub3:referenceType` field
tells you exactly why a reference was not expanded so you can decide whether
to fetch the full resource or simply display its `@id`.

### 6. Arrays

Any property can have multiple values. When there is more than one value,
the property is an array. When there is exactly one value, it is **unwrapped**
(not an array).

```json
"dcterms_alternative": [
  "H. Annakerk",
  "St. Annakerk",
  "Heilige Annakerk"
]
```

Array elements can be any of the above types — plain literals, typed literals,
language-tagged literals, or resource references. A single array can contain
a **mix** of types:

```json
"edm_rights": [
  "Creative Commons Attribution-Share Alike 3.0",
  { "@id": "https://creativecommons.org/licenses/by-sa/3.0" }
]
```

### Parsing Rules for Consumers

Follow these rules to robustly parse any resource from the API:

1. **Every property can be a single value or an array.** Always normalize to an
   array before iterating. If the value is not an array, wrap it in one.

2. **Every value can be a string, a number, a boolean, or an object.** Check the
   type before accessing nested fields.

3. **If the value is an object with `@value`**, it is a literal. Read `@value` for
   the content, `@language` for the language, or `@type` for the datatype.

4. **If the value is an object with `@id`**, it is a resource reference. Use
   `skos_prefLabel` (or `rdfs_label`) for display text. Check `hub3:referenceType`
   to understand why a reference was not expanded:
   - `"cycle"` — already expanded higher in the tree; look up or fetch separately
   - `"external"` — not in this graph; fetch via LOD request to the `@id` URI
   - *(absent)* — fully expanded, all properties are inline

5. **If the value is an object with only language-code keys** (`nl`, `en`, `de`, ...),
   it is a language map. Select the language matching your user's preference.

6. **GPS coordinates are strings**, not numbers. Convert `wgspos_lat` and
   `wgspos_long` to floating-point values in your code.

7. **Dates may appear as ISO 8601 strings** (`2024-01-01T00:00:00Z`) or typed
   literals with `xsd:dateTime` / `xsd:date`. A companion `*Label` field
   (e.g., `nave_dateCreationLabel`) often provides a human-readable version.

### Realistic Resource Detail Example

This example shows the variety of value types in a single resource:

```json
{
  "@context": { "...": "..." },
  "@id": "http://example.org/resource/document/buildings/Q2450",
  "@type": ["edm:ProvidedCHO"],

  "dc_title": "Sint-Annakerk",

  "dcterms_alternative": [
    "H. Annakerk",
    "St. Annakerk",
    "Heilige Annakerk"
  ],

  "dc_creator": {
    "@id": "https://example.org/agents/van-langelaar",
    "@type": ["edm:Agent"],
    "skos_prefLabel": "Jan Jurien van Langelaar",
    "skos_altLabel": ["J.J. van Langelaar"]
  },

  "dc_description": { "@value": "Een neogotische kruisbasiliek", "@language": "nl" },

  "nave_productionYear": { "@value": 1887, "@type": "xsd:integer" },

  "nave_location": {
    "@id": "https://example.org/places/molenschot",
    "@type": ["edm:Place"],
    "skos_prefLabel": "Kapelstraat 1, Molenschot",
    "wgspos_lat": "51.573",
    "wgspos_long": "4.8821"
  },

  "edm_rights": [
    "Creative Commons Attribution-Share Alike 3.0",
    { "@id": "https://creativecommons.org/licenses/by-sa/3.0" }
  ],

  "dcterms_isPartOf": {
    "@id": "http://example.org/resource/document/buildings/collection",
    "hub3:referenceType": "cycle"
  },

  "owl_sameAs": {
    "@id": "http://www.wikidata.org/entity/Q17593201",
    "hub3:referenceType": "external"
  }
}
```

In this response:
- `dc_title` — plain literal
- `dcterms_alternative` — array of plain literals
- `dc_creator` — nested resource (fully expanded, no `hub3:referenceType`)
- `dc_description` — language-tagged literal
- `nave_productionYear` — typed literal (integer)
- `nave_location` — nested resource with GPS coordinates (strings!)
- `edm_rights` — mixed array (string + resource reference)
- `dcterms_isPartOf` — `hub3:referenceType: "cycle"` (exists in graph but already expanded above)
- `owl_sameAs` — `hub3:referenceType: "external"` (not in this graph, fetch via LOD)

---

## Response Formats

All responses use `application/ld+json` content type.

### Search Response (`hydra:Collection`)

```json
{
  "@context": { "...": "..." },
  "@type": ["Collection"],
  "totalItems": 1234,
  "member": [
    {
      "@id": "http://example.org/resource/1",
      "dc_title": "The Night Watch",
      "dc_creator": "Rembrandt van Rijn"
    }
  ],
  "view": {
    "@type": "PartialCollectionView",
    "first": "...?page=1",
    "next": "...?page=2"
  },
  "search": {
    "@type": "IriTemplate",
    "template": "/api/semantic/v1/search{?query,filter*,facet*,page,size,sort}",
    "mapping": [ "..." ]
  },
  "hub3:facets": [
    {
      "field": "dc_type",
      "values": [
        { "value": "painting", "count": 500 },
        { "value": "drawing", "count": 200 }
      ],
      "sumOther": 50
    }
  ]
}
```

### Resource Detail Response

```json
{
  "@context": { "...": "..." },
  "@id": "http://example.org/resource/1",
  "@type": ["edm:ProvidedCHO"],
  "dc_title": "The Night Watch",
  "dc_creator": "Rembrandt van Rijn",
  "dc_date": "1642"
}
```

### Error Response (`hydra:Error`)

```json
{
  "@type": "Error",
  "hydra:title": "Bad Request",
  "hydra:description": "invalid operator 'invalid' for field 'creator'",
  "hub3:statusCode": 400
}
```

---

## Hydra Vocabulary

The API uses the [Hydra Core Vocabulary](http://www.w3.org/ns/hydra/core) for hypermedia controls.

### Standard Hydra Types

| Type | Usage |
|------|-------|
| `hydra:EntryPoint` | Root endpoint (`/`) |
| `hydra:ApiDocumentation` | API docs (`/docs`) |
| `hydra:Collection` | Search result sets |
| `hydra:PartialCollectionView` | Pagination links |
| `hydra:IriTemplate` | URL template for search |
| `hydra:Operation` | Supported HTTP operations |
| `hydra:Class` | Resource type definitions |
| `hydra:SupportedProperty` | Filterable properties |
| `hydra:Error` | Error responses |

### Hub3 Custom Types

| Type | Description |
|------|-------------|
| `hub3:SearchQuery` | Structured search query (POST body) |
| `hub3:TextQuery` | Full-text search query component |
| `hub3:PropertyFilter` | Property equality/comparison filter |
| `hub3:RangeFilter` | Numeric/date range filter |
| `hub3:ExistsFilter` | Field existence filter |
| `hub3:GeoBBoxFilter` | Geographic bounding box filter |
| `hub3:GeoDistanceFilter` | Geographic distance filter |
| `hub3:GeoPolygonFilter` | Geographic polygon filter |
| `hub3:FacetRequest` | Facet aggregation request |
| `hub3:Facet` | Facet result in response |
| `hub3:FacetValue` | Individual facet value with count |
| `hub3:GeoCluster` | Geographic cluster aggregation result |
| `hub3:SearchResultNavigation` | Detail-level navigation context |
| `hub3:NavigateOperation` | Navigation operation (next/previous) |
| `hub3:FilterDefinition` | Filter definition in type documentation |
| `hub3:FacetDefinition` | Facet definition in type documentation |
| `hub3:SortField` | Sort field definition in type documentation |
| `hub3:Example` | API usage example |

---

## JSON-LD Context

The API provides JSON-LD contexts for semantic interoperability.

### Available Contexts

| Context | URL |
|---------|-----|
| Hub3 Search Vocabulary | `/contexts/hub3/1.0/context.jsonld` |
| Hub3 Search (latest) | `/contexts/hub3/latest/context.jsonld` |
| EDM (Europeana Data Model) | `/contexts/edm/1.0/context.jsonld` |
| EDM (latest) | `/contexts/edm/latest/context.jsonld` |

### Namespace Prefixes

| Prefix | Namespace URI |
|--------|---------------|
| `cc` | `https://creativecommons.org/ns#` |
| `dc` | `http://purl.org/dc/elements/1.1/` |
| `dcterms` | `http://purl.org/dc/terms/` |
| `ebucore` | `http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#` |
| `edm` | `http://www.europeana.eu/schemas/edm/` |
| `foaf` | `http://xmlns.com/foaf/0.1/` |
| `geo` | `http://www.w3.org/2003/01/geo/wgs84_pos#` |
| `geojson` | `https://purl.org/geojson/vocab#` |
| `gn` | `http://www.geonames.org/ontology#` |
| `hub3` | `https://hub3.delving.org/vocab/` |
| `hydra` | `http://www.w3.org/ns/hydra/core#` |
| `nave` | `http://schemas.delving.eu/nave/terms/` |
| `ore` | `http://www.openarchives.org/ore/terms/` |
| `owl` | `http://www.w3.org/2002/07/owl#` |
| `rdaGr2` | `http://rdvocab.info/ElementsGr2/` |
| `rdf` | `http://www.w3.org/1999/02/22-rdf-syntax-ns#` |
| `rdfs` | `http://www.w3.org/2000/01/rdf-schema#` |
| `schema` | `https://schema.org/` |
| `skos` | `http://www.w3.org/2004/02/skos/core#` |
| `vcard` | `http://www.w3.org/2006/vcard/ns#` |

---

## API Capabilities

These values are derived from code and always reflect the running system.

- **Default page size:** 20
- **Maximum page size:** 1000
- **Sort options:** [asc desc]
- **Search context TTL:** 15m
- **Path separator:** `/`

---

## Error Handling

All errors return a `hydra:Error` response with appropriate HTTP status codes.

| Status | Description |
|--------|-------------|
| 400 | Bad Request — invalid parameters, operators, or malformed query |
| 404 | Not Found — unknown resource, type, or search context |
| 500 | Internal Server Error — unexpected backend failure |

Error responses include `hydra:title` (short label) and `hydra:description` (detailed message).
