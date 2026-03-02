# V2 Search API — Feature Freeze Notice

> **Status:** Feature-frozen as of 2026-02-25. No new features will be added to the V2 search API.
> All new development targets the Semantic V1 API (`/api/semantic/v1`).
> The V2 API will be removed once all frontends have migrated.

## Overview

The V2 search API (`/api/search/v2`) has been superseded by the Semantic V1 API.
This document catalogs every V2 parameter, its semantic V1 equivalent, and current status.

Frontend teams should use this as a reference when migrating to the semantic API.
See [v2-to-semantic-migration-guide.md](v2-to-semantic-migration-guide.md) for step-by-step migration instructions.

## Runtime Backend Switching

During the migration period, both backends coexist. Use `?backend=v2` or `?backend=es8`
on any semantic V1 endpoint to select the backend per-request. The default backend is
determined by server configuration. The backend choice is preserved in search contexts
for consistent navigation.

## Parameter Mapping

### Text Search

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `q=text` | `query=text` | **Mapped** |
| `query=text` | `query=text` | **Mapped** |
| `rq=refinement` | Use additional filters | **Dropped** — rarely used, filters are more expressive |
| `searchFields=field` | `fields=field` | **Mapped** |

### Filters

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `qf=field:value` | `filter[field][eq]=value` | **Mapped** |
| `qf[]=field:value` | `filter[field][eq]=value` (multiple) | **Mapped** |
| `hqf=field:value` | `hfilter[field][eq]=value` | **Mapped** — hidden filter |
| `hqf[]=field:value` | `hfilter[field][eq]=value` (multiple) | **Mapped** — hidden filter |
| `qf.id=field:value` | `filter[field][eq]=value` | **Mapped** — ID filter is implicit in V1 |
| `qf.exist=field` | `filter[field][exists]=true` | **Mapped** |
| `qf.exist[]=field` | `filter[field][exists]=true` (multiple) | **Mapped** |
| `qf.dateRange=field:[a TO b]` | `filter[field][gte]=a&filter[field][lte]=b` | **Mapped** |
| `qf.dateRange[]=field:[a TO b]` | Same as above (multiple) | **Mapped** |
| `qf.date=field:value` | `filter[field][eq]=value` | **Mapped** |
| `qf.date[]=field:value` | Same as above (multiple) | **Mapped** |
| `qf.tree=field:value` | N/A | **Dropped** — tree params out of scope |

### Facets

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `facet.field=name` | `facet[name]=20` | **Mapped** — limit as value |
| `facet.field=name` (multiple) | `facet[name1]=20&facet[name2]=20` | **Mapped** |
| `facet.size=N` | `facet[name]=N` (limit as value) | **Mapped** |
| `facet.limit=N` | `facet[name]=N` (limit as value) | **Mapped** |
| `facet.sort=count\|index` | `facet[name].sort=count\|index` | **Mapped** |
| `facetBoolType=and\|or` | `facetBool=and\|or` | **Mapped** |
| `facet.boolType=and\|or` | `facetBool=and\|or` | **Mapped** |
| `facet.cursor=token` | `facet[name].cursor=token` | **Mapped** — per-facet cursor |
| `facet.expand=field` | N/A | **Dropped** — use facet cursor instead |
| `facet.filter=expr` | Use filters | **Dropped** — use filter params |
| `facet.mergeFilter=expr` | N/A | **Dropped** — rarely used |
| `facetOrBetween=true` | N/A | **Dropped** — rarely used |
| `facet.full=field` | `facet[field]=2000` | **Expressible** — set limit to 2000 |
| `peek=f1,f2` | `peek=f1,f2` | **Mapped** |

### Pagination

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `page=N` | `page=N` | **Mapped** |
| `rows=N` | `size=N` | **Mapped** |
| `limit=N` | `size=N` | **Mapped** |
| `size=N` | `size=N` | **Mapped** |
| `start=N` | `page=ceil(start/size)+1` | **Mapped** — offset-based to page-based |
| `scrollID=hex` | `cursor=token` | **Replaced** — cursor-based pagination |
| `qs=hex` | `cursor=token` | **Replaced** — cursor-based pagination |

### Sorting

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `sortBy=field` | `sort=field` | **Mapped** |
| `sortBy=^field` | `sort=field` (ascending is default) | **Mapped** |
| `sortAsc=true\|false` | `sort=field` or `sort=-field` | **Mapped** — prefix `-` for desc |
| `sortOrder=asc\|desc` | `sort=field` or `sort=-field` | **Mapped** |

### Collapse (Field Collapsing)

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `collapseOn=field` | `collapse=field` | **Mapped** |
| `collapseSize=N` | `collapse.size=N` | **Mapped** |
| `collapseSort=field` | `collapse.sort=field` or `collapse.sort=-field` | **Mapped** |
| `collapseFormat=fmt` | N/A | **Dropped** — JSON-LD only |

### Peek / Debug

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `peek=f1,f2` | `peek=f1,f2` | **Mapped** |
| `echo=searchResponse` | `debug=query` | **Mapped** |

### Language / Format

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `lang=en,nl` | `languages=en,nl` | **Mapped** |
| `format=protobuf\|jsonld\|bulkaction` | N/A | **Dropped** — JSON-LD only |
| `itemFormat=summary\|grouped\|flat\|...` | N/A | **Dropped** — semantic format only |
| `item.format=semantic` | Default (always semantic) | **Default** |

### More Like This (MLT)

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `mlt=true&mlt.id=X` | `GET /resource/{X}?include=relatedItems` | **Mapped** — moved to detail endpoint |

### Cross-Index Search

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `contextIndex=idx` | `contextIndex=idx` | **Mapped** — identical parameter name |

### Backend Selection (New in Semantic V1)

These parameters are new in the Semantic V1 API and have no V2 equivalent:

| Semantic V1 Parameter | Description |
|---|---|
| `backend=v2\|es8` | Select search backend per-request |
| `include=relatedItems` | Include related items in detail view |
| `detailLevel=full\|summary` | Control detail level in responses |

### Resource Lookup

The Semantic V1 detail endpoint (`/resource/{id}`) supports both lookup methods:

| Method | Example | Notes |
|---|---|---|
| Hub ID (`_id`) | `/resource/abc123` | Direct Elasticsearch `_id` lookup |
| Full URI | `/resource/http%3A%2F%2Fexample.org%2Fobj%2F1` | URL-encoded URI, resolved via `meta.entryURI` |

### Tree / EAD Parameters (Out of Scope)

The following V2 tree/EAD parameters have no semantic V1 equivalent and are not planned for migration.
They are specific to archival finding aids and will be addressed separately if needed.

| V2 Parameter | Status |
|---|---|
| `byLeaf`, `fillTree` | **Out of scope** |
| `byDepth` | **Out of scope** |
| `byChildCount` | **Out of scope** |
| `byParent` | **Out of scope** |
| `byType` | **Out of scope** |
| `byLabel` | **Out of scope** |
| `byQuery` | **Out of scope** |
| `withFields` | **Out of scope** |
| `hasDigitalObject` | **Out of scope** |
| `paging`, `pageMode` | **Out of scope** |
| `hasRestriction` | **Out of scope** |
| `byUnitID`, `allParents` | **Out of scope** |
| `byMimeType` | **Out of scope** |
| `cursorHint` | **Out of scope** |
| `treePage` | **Out of scope** |

### Internal / Deprecated

| V2 Parameter | Status |
|---|---|
| `v1.mode=true` | **Dropped** — internal legacy flag |

## URL Mapping

| V2 Endpoint | Semantic V1 Endpoint |
|---|---|
| `/api/search/v2` | `/api/semantic/v1/search` |
| `/v2/search` | `/api/semantic/v1/search` |
| `/api/v3/search` | `/api/semantic/v1/search` |
| `/api/search/v2/{id}` | `/api/semantic/v1/resource/{id}` |

## Status Legend

| Status | Meaning |
|---|---|
| **Mapped** | Direct equivalent exists in semantic V1 |
| **Replaced** | Functionality exists but with different mechanism |
| **Expressible** | Can be achieved using existing V1 parameters |
| **Dropped** | No equivalent; feature was rarely used or not needed |
| **Default** | Behavior is the default in semantic V1 |
| **Out of scope** | Tree/EAD specific; not part of semantic search migration |

## Notes

- Field names use underscores in URLs (`dc_creator`) and colons internally (`dc:creator`).
  The semantic API handles conversion automatically.
- The V2 API will remain available during the migration period but receives no new features or bug fixes.
- See [v2-to-semantic-migration-guide.md](v2-to-semantic-migration-guide.md) for the full migration guide with examples.
