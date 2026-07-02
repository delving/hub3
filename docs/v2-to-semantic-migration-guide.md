# V2 to Semantic V1 Migration Guide

This guide helps frontend teams migrate from the V2 search API to the Semantic V1 API.

## Quick Reference

| Aspect | V2 | Semantic V1 |
|---|---|---|
| Base URL | `/api/search/v2` | `/api/semantic/v1/search` |
| Detail URL | `/api/search/v2/{id}` | `/api/semantic/v1/resource/{id}` |
| Response format | Flat JSON | Hydra JSON-LD collections |
| Filter syntax | `qf=field:value` | `filter[field][eq]=value` |
| Facet syntax | `facet.field=name` | `facet[name]=20` |

---

## 1. URL Structure

### Search

```
# V2
GET /api/search/v2?q=rembrandt

# Semantic V1
GET /api/semantic/v1/search?query=rembrandt
```

### Detail / Resource

```
# V2
GET /api/search/v2/{id}

# Semantic V1
GET /api/semantic/v1/resource/{id}
```

---

## 2. Before/After Examples

### Example 1: Simple text search

```
# V2
/api/search/v2?q=rembrandt&rows=20

# Semantic V1
/api/semantic/v1/search?query=rembrandt&size=20
```

### Example 2: Filter by creator

```
# V2
/api/search/v2?q=*&qf=dc_creator:Rembrandt

# Semantic V1
/api/semantic/v1/search?query=*&filter[dc_creator][eq]=Rembrandt
```

### Example 3: Hidden filter (not shown in facets)

```
# V2
/api/search/v2?hqf=dc_type:painting

# Semantic V1
/api/semantic/v1/search?hfilter[dc_type][eq]=painting
```

### Example 4: Date range filter

```
# V2
/api/search/v2?qf.dateRange=dc_date:[1600 TO 1700]

# Semantic V1
/api/semantic/v1/search?filter[dc_date][gte]=1600&filter[dc_date][lte]=1700
```

### Example 5: Exists filter

```
# V2
/api/search/v2?qf.exist=nave_thumbnail

# Semantic V1
/api/semantic/v1/search?filter[nave_thumbnail][exists]=true
```

### Example 6: Facets with limit

```
# V2
/api/search/v2?facet.field=dc_type&facet.size=50

# Semantic V1
/api/semantic/v1/search?facet[dc_type]=50
```

### Example 7: Sort descending by date

```
# V2
/api/search/v2?sortBy=dc_date&sortAsc=false

# Semantic V1
/api/semantic/v1/search?sort=-dc_date
```

### Example 8: Collapse / group by provider

```
# V2
/api/search/v2?collapseOn=edm_dataProvider&collapseSize=3

# Semantic V1
/api/semantic/v1/search?collapse=edm_dataProvider&collapse.size=3
```

### Example 9: Peek (facets only, no results)

```
# V2
/api/search/v2?peek=dc_creator,dc_type

# Semantic V1
/api/semantic/v1/search?peek=dc_creator,dc_type
```

### Example 10: More Like This

```
# V2
/api/search/v2?mlt=true&mlt.id=doc-123

# Semantic V1
GET /api/semantic/v1/resource/doc-123?include=relatedItems
```

---

## 3. Filter Syntax

V2 uses `qf=field:value` syntax. Semantic V1 uses bracket notation with explicit operators.

| V2 | Semantic V1 |
|---|---|
| `qf=dc_creator:Rembrandt` | `filter[dc_creator][eq]=Rembrandt` |
| `hqf=dc_type:painting` | `hfilter[dc_type][eq]=painting` |
| `qf.exist=nave_thumbnail` | `filter[nave_thumbnail][exists]=true` |
| `qf.dateRange=dc_date:[1600 TO 1700]` | `filter[dc_date][gte]=1600&filter[dc_date][lte]=1700` |
| `qf.date=dc_date:1650` | `filter[dc_date][eq]=1650` |

### Available operators

| Operator | Description | Example |
|---|---|---|
| `eq` | Exact match | `filter[dc_creator][eq]=Rembrandt` |
| `neq` | Not equal | `filter[dc_type][neq]=painting` |
| `in` | One of (multi-value) | `filter[dc_type][in]=painting&filter[dc_type][in]=drawing` |
| `contains` | Full-text match | `filter[dc_title][contains]=night` |
| `startsWith` | Prefix match | `filter[dc_title][startsWith]=night` |
| `gt`, `gte`, `lt`, `lte` | Range | `filter[dc_date][gte]=1600&filter[dc_date][lte]=1700` |
| `exists` | Field exists | `filter[nave_thumbnail][exists]=true` |
| `bbox` | Geo bounding box | `filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4` |

---

## 4. Facet Syntax

| V2 | Semantic V1 |
|---|---|
| `facet.field=dc_type` | `facet[dc_type]=20` |
| `facet.field=dc_type&facet.size=50` | `facet[dc_type]=50` |
| `facet.sort=index` | `facet[dc_type].sort=index` |
| `facetBoolType=and` | `facetBool=and` |
| `facet.full=dc_type` (all values) | `facet[dc_type]=2000` |

### Facet sub-parameters

- `facet[field].sort=count` — sort by count (default) or `index` (alphabetical)
- `facet[field].cursor=abc123` — paginate through facet values

---

## 5. Sort Syntax

| V2 | Semantic V1 |
|---|---|
| `sortBy=dc_date` | `sort=dc_date` |
| `sortBy=dc_date&sortAsc=false` | `sort=-dc_date` |
| `sortBy=dc_date&sortAsc=true` | `sort=dc_date` |

The `-` prefix means descending. No prefix (or `+`) means ascending.

---

## 6. Pagination

| V2 | Semantic V1 |
|---|---|
| `rows=20` / `limit=20` / `size=20` | `size=20` |
| `page=2` | `page=2` |
| `start=40` (offset-based) | `page=3` (with `size=20`) |
| `scrollID=abc` | `cursor=abc` |

Default: `page=1`, `size=20`.

---

## 7. Response Format Differences

### V2 Response

```json
{
  "result": [...],
  "pagination": {"numFound": 100, "start": 0, "rows": 20},
  "facets": [...]
}
```

### Semantic V1 Response (Hydra JSON-LD)

```json
{
  "@context": "https://hub3.org/schemas/semantic/v1",
  "@type": "hydra:Collection",
  "hydra:totalItems": 100,
  "hydra:member": [...],
  "hydra:view": {
    "@type": "hydra:PartialCollectionView",
    "hydra:first": "/api/semantic/v1/search?query=...&page=1",
    "hydra:next": "/api/semantic/v1/search?query=...&page=2",
    "hydra:last": "/api/semantic/v1/search?query=...&page=5"
  },
  "hydra:facets": [
    {
      "field": "dc_type",
      "values": [
        {"value": "painting", "count": 50},
        {"value": "sculpture", "count": 30}
      ]
    }
  ]
}
```

Key differences:
- Results are in `hydra:member` (not `result`)
- Total count is `hydra:totalItems` (not `pagination.numFound`)
- Pagination links are provided as Hydra views (no need to compute offsets)
- Facets include `field` and `values` with `value`/`count` pairs

---

## 8. Error Handling

### V2

```json
{"error": "invalid query parameter", "status": 400}
```

### Semantic V1

```json
{
  "@type": "hydra:Error",
  "hydra:title": "Bad Request",
  "hydra:description": "invalid operator 'invalid' for field 'creator'",
  "statusCode": 400
}
```

Errors follow the Hydra error format with structured `@type`, `hydra:title`, and `hydra:description`.

---

## 9. Cross-Index Search

To search across multiple organizations (e.g., aggregated portals):

```
# Semantic V1
/api/semantic/v1/search?query=rembrandt&contextIndex=org-a&contextIndex=org-b
```

This searches the primary organization's index plus the specified context indices.
The context indices are additional Elasticsearch indices that will be included in the search.

---

## 10. Backend Scope

Semantic V1 currently wraps the existing V2 search implementation and returns
JSON-LD/Hydra responses. Clients do not select a backend per request.

```
/api/semantic/v1/search?query=rembrandt
```

Native Elasticsearch backend work is future development. It must preserve the
same Semantic V1 request and response contract before it can replace the V2
adapter.

---

## 11. Resource Detail View

### By Hub ID (Elasticsearch `_id`)

```
GET /api/semantic/v1/resource/abc123-def456
```

### By Full URI (URL-encoded)

```
GET /api/semantic/v1/resource/http%3A%2F%2Fexample.org%2Fobject%2F123
```

The detail endpoint supports both lookup methods. Full URIs must be URL-encoded
and are resolved via the `meta.entryURI` field. This is natural for JSON-LD consumers
that use the `@id` value from search results.

### With Navigation Context

```
GET /api/semantic/v1/resource/abc123?context=ctx_abc12345
```

When a search context token is provided, the response includes `hub3:navigation`
with previous/next links for browsing through search results.

### With Related Items (MLT)

```
GET /api/semantic/v1/resource/abc123?include=relatedItems&relatedItems.count=5
```

---

## 12. Full Parameter Mapping

See [v2-feature-freeze.md](v2-feature-freeze.md) for the comprehensive parameter mapping table.

### Dropped Parameters

These V2 parameters have no Semantic V1 equivalent:

| V2 Parameter | Migration Path |
|---|---|
| `rq` (query refinement) | Use additional `filter[...]` parameters |
| `facet.mergeFilter` | Not implemented (rarely used) |
| `facet.expand` | Use `facet[field].cursor=token` for paging through facet values |
| `facetOrBetween` | Not implemented (rarely used) |
| `itemFormat` / `format` | Semantic V1 returns JSON-LD only |
| `v1.mode` | Internal parameter, removed |
| Tree/EAD params | Out of scope for semantic search |

---

## 13. Field Name Convention

Field names use underscores in URLs and colons internally:

- URL: `dc_creator`, `edm_dataProvider`
- Internal: `dc:creator`, `edm:dataProvider`

The API handles this conversion automatically. Use the underscore form in query parameters.

---

## Appendix: Future Native Backend Notes

When all frontend applications have been migrated to the Semantic V1 API, a later
native-backend project can decide whether to remove the V2 adapter. That work is
outside the current wrapper delivery scope.

### Future Preconditions

- [ ] All frontend apps confirmed migrated to Semantic V1
- [ ] Zero V2 API traffic for 2+ weeks (verify via access logs)
- [ ] Native backend implements the accepted Semantic V1 contract
- [ ] Native backend stable in production for 2+ weeks
- [ ] Cross-index search (`contextIndex`) working on Semantic V1

### Packages to Remove

| Package | Files | Description |
|---|---|---|
| `ikuzo/storage/x/v2adapter/` | 8 files | V2-to-semantic bridge adapter |
| `ikuzo/storage/x/elasticsearch/semantic_store.go` | 1 file | Old ES semantic store |
| `ikuzo/storage/x/elasticsearch/facets.go` | 1 file | Old ES facet handling |

### Configuration to Update

| File | Change |
|---|---|
| `ikuzo/ikuzoctl/cmd/config/semantic.go` | Remove V2 adapter creation, `olivere` client setup, and `UseV2Adapter` flag once a native backend has replaced it. |

### Routes to Remove

| File | Routes |
|---|---|
| `hub3/server/http/handlers/search.go` | `/api/search/v2`, `/v2/search`, `/api/v3/search` |

### Code to Clean Up

| File | Change |
|---|---|
| `ikuzo/service/x/semantic/service.go` | Remove internal alternate-store hooks if future native-backend experiments no longer need them |
| `ikuzo/service/x/semantic/options.go` | Remove `WithAlternateStore()` if unused |
| `ikuzo/domain/semantic/query.go` | Remove `Backend` field from `QueryOptions` |
| `ikuzo/domain/semantic/pagination.go` | Remove `Backend` field from `SearchContext` |

### Dependencies

| Dependency | Impact |
|---|---|
| `olivere/elastic/v7` | 32+ files import it; full removal requires broader legacy cleanup beyond the V2 adapter. The V2 adapter removal eliminates the semantic API dependency, but the bulk indexing and other legacy paths still use olivere. |

### Removal Procedure

1. Verify all pre-removal conditions are met
2. Remove V2 adapter package (`ikuzo/storage/x/v2adapter/`)
3. Remove old ES store files
4. Update `semantic.go` config to use the replacement backend
5. Clean up any internal backend experiment hooks that are no longer needed
6. Remove V2 search routes from `hub3/server/http/handlers/search.go`
7. Run `go build ./...` and `go test ./...` to verify
8. Deploy to staging and verify no regressions
9. Deploy to production
