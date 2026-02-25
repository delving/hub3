# V2 API Feature Freeze

> **Status:** Feature-frozen as of 2026-02-25. No new parameters will be added to the V2 search API.
> All new development targets the Semantic V1 API (`/api/semantic/v1`).

## Overview

This document catalogs every V2 search API parameter, its Semantic V1 equivalent, and migration status.
Frontend teams should use this as a reference when migrating from V2 to Semantic V1.

## Parameter Mapping

### Text Search

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `q`, `query` | `query=` | Mapped |

### Filters

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `qf=field:value` | `filter[field][eq]=value` | Mapped |
| `hqf=field:value` | `hfilter[field][eq]=value` | Mapped |
| `qf.exist=field` | `filter[field][exists]=true` | Mapped |
| `qf.dateRange=field:[a TO b]` | `filter[field][gte]=a&filter[field][lte]=b` | Mapped |
| `qf.date=field:value` | `filter[field][eq]=value` | Mapped |

### Facets

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `facet.field=name` | `facet[name]=20` | Mapped |
| `facet.size` / `facet.limit` | `facet[name]=50` (limit as value) | Mapped |
| `facet.sort` | `facet[name].sort=count` | Mapped |
| `facetBoolType=and` | `facetBool=and` | Mapped |
| `facet.full=field` | `facet[field]=2000` | Expressible |
| `facet.mergeFilter` | N/A | Dropped (rarely used) |

### Pagination

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `page=N` | `page=N` | Mapped |
| `rows` / `limit` / `size` | `size=N` | Mapped |
| `start=N` | `page=ceil(start/size)+1` | Mapped |
| `scrollID` / `qs` | `cursor=` | Replaced |

### Sorting

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `sortBy=field` | `sort=field` | Mapped |
| `sortAsc=false` | `sort=-field` | Mapped |

### Collapse / Grouping

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `collapseOn=field` | `collapse=field` | Mapped |
| `collapseSize=N` | `collapse.size=N` | Mapped |
| `collapseSort=field` | `collapse.sort=field` | Mapped |

### Peek / Debug

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `peek=f1,f2` | `peek=f1,f2` | Mapped |
| `echo=searchResponse` | `debug=query` | Mapped |

### Language / Format

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `lang=en,nl` | `languages=en,nl` | Mapped |
| `itemFormat` / `format` | JSON-LD only | Dropped by design |

### More Like This (MLT)

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `mlt=true&mlt.id=X` | `GET /resource/{X}?include=relatedItems` | Mapped (detail endpoint) |

### Cross-Index Search

| V2 Parameter | Semantic V1 Equivalent | Status |
|---|---|---|
| `contextIndex=idx` | `contextIndex=idx` | New in D3 |

### Dropped / Out of Scope

| V2 Parameter | Notes |
|---|---|
| `rq` (query refinement) | Use additional filters instead |
| `scrollID` / `qs` | Replaced by `cursor=` |
| `facet.mergeFilter` | Rarely used, not implemented |
| `itemFormat` / `format` | Semantic V1 returns JSON-LD only |
| `v1.mode` | Internal parameter, dropped |
| Tree/EAD params (`byLeaf`, `fillTree`, etc.) | Out of scope for semantic search |

## URL Mapping

| V2 Endpoint | Semantic V1 Endpoint |
|---|---|
| `/api/search/v2` | `/api/semantic/v1/search` |
| `/v2/search` | `/api/semantic/v1/search` |
| `/api/v3/search` | `/api/semantic/v1/search` |
| `/api/search/v2/{id}` | `/api/semantic/v1/resource/{id}` |

## Notes

- Field names use underscores in URLs (`dc_creator`) and colons internally (`dc:creator`).
  The semantic API handles conversion automatically.
- The V2 API will remain available during the migration period but receives no new features or bug fixes.
- See `docs/v2-to-semantic-migration-guide.md` for the full migration guide with examples.
