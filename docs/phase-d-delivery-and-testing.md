# Phase D Delivery: V2 Deprecation & Migration Support

## What Was Delivered

Phase D adds three capabilities to the Hub3 platform:

### 1. V2 Feature Freeze Documentation

The V2 search API is now feature-frozen. A complete parameter mapping document
(`docs/v2-feature-freeze.md`) catalogs every V2 parameter, its Semantic V1
equivalent, and whether it is mapped, replaced, or dropped. This serves as the
authoritative reference for the migration.

### 2. Frontend Migration Guide

A step-by-step migration guide (`docs/v2-to-semantic-migration-guide.md`) helps
frontend developers switch from the V2 API to the Semantic V1 API. It includes:

- 10 before/after query examples
- URL structure changes
- Filter, facet, sort, and pagination syntax mappings
- Response format differences (flat JSON to Hydra JSON-LD)
- Error handling changes

### 3. Cross-Index Search

A new `contextIndex` parameter enables searching across multiple organization
indices in a single query. This supports aggregated portals and cross-collection
discovery.

**How it works:** When `contextIndex` is provided, the search expands to include
documents from the specified organizations in addition to the primary organization.
Both the Elasticsearch query filter and the index targeting are updated to cover
all specified organizations.

---

## How to Test

### Prerequisites

1. Hub3 is running with the Semantic API enabled:

   ```toml
   [semantic]
   enabled       = true
   useES8Backend = true
   ```

2. At least one organization has indexed data (the examples below assume the
   default organization is accessible at `localhost:3001`).

3. For cross-index testing, at least two organizations with indexed data are
   needed.

### Test 1: Basic Search

Verify the Semantic V1 search endpoint returns results.

```bash
# Search for all documents
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=5" | jq .

# Expected: JSON-LD response with hydra:totalItems > 0 and hydra:member array
```

**Check:**
- Response contains `"@type": "hydra:Collection"`
- `hydra:totalItems` shows a number greater than 0
- `hydra:member` contains up to 5 result objects

### Test 2: Text Search with Filters

```bash
# Search with a text query and an equality filter
curl -s "http://localhost:3001/api/semantic/v1/search?query=schilderij&filter[dc_type][eq]=painting" | jq .

# Expected: Results where dc_type matches "painting"
```

**Check:**
- Results are filtered to only include documents matching the filter
- `hydra:totalItems` is less than or equal to an unfiltered search

### Test 3: Facets

```bash
# Request facets for dc_type with a limit of 10 values
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&facet[dc_type]=10&size=0" | jq .

# Expected: Facet results showing top 10 dc_type values with counts
```

**Check:**
- Response includes a `facets` array
- Each facet has a `field` name and `values` array with `value`/`count` pairs

### Test 4: Sorting

```bash
# Sort results by date descending
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&sort=-dc_date&size=5" | jq .

# Expected: Results ordered by dc_date in descending order
```

### Test 5: Pagination

```bash
# Get page 1
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&page=1&size=2" | jq .

# Get page 2
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&page=2&size=2" | jq .

# Expected: Different results on each page, hydra:view contains navigation links
```

**Check:**
- `hydra:view` contains `hydra:first`, `hydra:next`, and/or `hydra:last` links
- Page 2 results differ from page 1

### Test 6: Collapse / Grouping

```bash
# Group results by data provider, showing 2 inner hits per group
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&collapse=edm_dataProvider&collapse.size=2&size=5" | jq .

# Expected: Results grouped by edm:dataProvider
```

### Test 7: Peek (Facets Only)

```bash
# Get only facets, no search results
curl -s "http://localhost:3001/api/semantic/v1/search?peek=dc_type,dc_creator" | jq .

# Expected: size=0 results with facet data for dc_type and dc_creator
```

**Check:**
- `hydra:member` is empty or absent
- Facets are returned for the specified fields

### Test 8: Resource Detail

```bash
# First, get an ID from search results
ID=$(curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=1" | jq -r '.["hydra:member"][0]["@id"]')

# Then fetch the full resource
curl -s "http://localhost:3001/api/semantic/v1/resource/$ID" | jq .

# Expected: Full document with all fields
```

### Test 9: More Like This (Related Items)

```bash
# Get similar items for a specific resource
ID=$(curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=1" | jq -r '.["hydra:member"][0]["@id"]')

curl -s "http://localhost:3001/api/semantic/v1/resource/$ID?include=relatedItems" | jq .

# Expected: Resource detail with a relatedItems section
```

### Test 10: Cross-Index Search

This test requires two organizations with indexed data. Replace `org-a` and
`org-b` with actual organization IDs.

```bash
# Search only the primary organization (baseline)
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=0" | jq '.["hydra:totalItems"]'

# Search across primary + another organization
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=0&contextIndex=org-b" | jq '.["hydra:totalItems"]'

# Expected: The second query returns equal or higher totalItems
```

**Check:**
- The total with `contextIndex` is greater than or equal to the baseline
- Results from both organizations appear in the response

```bash
# Multiple context indices
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&size=5&contextIndex=org-a&contextIndex=org-b" | jq .

# Expected: Results from primary org, org-a, and org-b
```

### Test 11: Hidden Filters

```bash
# Apply a hidden filter (not reflected in facet counts)
curl -s "http://localhost:3001/api/semantic/v1/search?query=*&hfilter[dc_type][eq]=painting&facet[dc_type]=10" | jq .

# Expected: Results filtered to paintings, but facet counts show all types
```

### Test 12: Debug Mode

```bash
# Enable debug to see the generated Elasticsearch query
curl -s "http://localhost:3001/api/semantic/v1/search?query=rembrandt&debug=query" | jq .

# Expected: Response includes diagnostic information about the ES query
```

### Test 13: Introspection

```bash
# List available classes
curl -s "http://localhost:3001/api/semantic/v1/introspect/classes" | jq .

# List properties for a specific class
curl -s "http://localhost:3001/api/semantic/v1/introspect/classes/edm:ProvidedCHO/properties" | jq .

# Expected: Structured list of classes/properties available in the index
```

---

## Comparing V2 and Semantic V1 Side-by-Side

To verify that the Semantic V1 API returns equivalent results to V2, run
the same logical query on both endpoints and compare totals:

```bash
# V2
curl -s "http://localhost:3001/api/search/v2?q=rembrandt&rows=0" | jq '.result.total'

# Semantic V1
curl -s "http://localhost:3001/api/semantic/v1/search?query=rembrandt&size=0" | jq '.["hydra:totalItems"]'

# Expected: Both return the same total count
```

---

## Known Limitations

- **Cross-index search** requires that all referenced organizations have their
  data indexed in Elasticsearch. An invalid `contextIndex` value will target a
  non-existent index, which may return an error.
- **Response format** is JSON-LD only. The V2 `itemFormat` and `format`
  parameters are not supported.
- **Tree/EAD parameters** (`byLeaf`, `fillTree`, etc.) are out of scope for
  the Semantic V1 API.
- **Scroll-based pagination** (`scrollID`) is replaced by cursor-based
  pagination (`cursor=`).

---

## Files Delivered

| File | Description |
|---|---|
| `docs/v2-feature-freeze.md` | Complete V2 parameter mapping and freeze notice |
| `docs/v2-to-semantic-migration-guide.md` | Frontend migration guide with examples and D4 checklist |
| `docs/phase-d-delivery-and-testing.md` | This document |
| `ikuzo/domain/semantic/query.go` | Added `ContextIndices` field to `QueryOptions` |
| `ikuzo/domain/semantic/store.go` | Added `ContextIndices` field to `SimilarOptions` |
| `ikuzo/service/x/semantic/parser.go` | Parse `contextIndex` query parameter |
| `ikuzo/service/x/semantic/parser_test.go` | Tests for contextIndex parsing |
| `ikuzo/storage/x/elasticsearch8/query_builder.go` | Multi-org filter (term/terms switch) |
| `ikuzo/storage/x/elasticsearch8/query_builder_test.go` | Tests for multi-org query building |
| `ikuzo/storage/x/elasticsearch8/store.go` | Multi-index search resolution |
| `ikuzo/storage/x/elasticsearch8/store_test.go` | Tests for index resolution |
