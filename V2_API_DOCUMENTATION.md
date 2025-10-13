# Hub3 V2 API Documentation

## Overview

The Hub3 V2 API is a modern, feature-rich search API that provides advanced search, filtering, faceting, and data retrieval capabilities. It offers flexible response formats, dynamic configuration, and advanced features like geospatial search and "More Like This" functionality.

## Table of Contents

1. [Base Endpoints](#base-endpoints)
2. [Search Functionality](#search-functionality)
3. [Record Detail View](#record-detail-view)
4. [More Like This](#more-like-this)
5. [Filtering](#filtering)
6. [Faceting](#faceting)
7. [Map Search](#map-search)
8. [Pagination](#pagination)
9. [Sorting](#sorting)
10. [Response Formats](#response-formats)
11. [Advanced Features](#advanced-features)
12. [Common Use Cases](#common-use-cases)

## Base Endpoints

### Search Endpoints
```
GET /api/search/v2
GET /v2/search
GET /api/v3/search
```
All three endpoints provide the same functionality with identical parameters.

### Record Detail Endpoints
```
GET /api/search/v2/{id}
GET /v2/search/{id}
GET /api/v3/search/{id}
```
Retrieve specific records by identifier with optional "More Like This" functionality.

## Search Functionality

### Basic Search
**Parameters:**
- `q` (string): Main search query using Lucene/Elasticsearch syntax

**Examples:**
```
GET /api/search/v2?q=rembrandt
GET /api/search/v2?q=title:"night watch"
GET /api/search/v2?q=creator:rembrandt AND type:painting
```

### Advanced Query Syntax
- Simple terms: `rembrandt`
- Phrases: `"night watch"`
- Field searches: `dc.creator:rembrandt`
- Wildcards: `rem*` or `?embrandt`
- Boolean: `AND`, `OR`, `NOT`
- Grouping: `(rembrandt OR vermeer) AND painting`
- Range: `date:[1600 TO 1700]`
- Fuzzy: `rembrant~2`
- Proximity: `"night watch"~5`

## Record Detail View

### Retrieving Individual Records

**Endpoint:**
```
GET /api/search/v2/{id}
```

**Parameters:**
- `{id}` (path): Unique record identifier
- `format` (query): Response format
- `itemFormat` (query): Item display format
- `lang` (query): Language preference

**Example:**
```
GET /api/search/v2/museum-123?itemFormat=fragmentGraph&lang=en,nl
```

### Response Formats for Details
- `summary`: Basic metadata
- `fragmentGraph`: Full RDF graph structure
- `jsonld`: JSON-LD format
- `flat`: Flattened structure
- `tree`: Hierarchical structure
- `grouped`: Grouped by predicate

## More Like This

The "More Like This" feature finds similar records based on content analysis.

### In Search Results
**Parameters:**
- `moreLikeThis` (boolean): Enable MLT for search results
- `moreLikeThisCount` (integer): Number of similar items per result

**Example:**
```
GET /api/search/v2?q=rembrandt&moreLikeThis=true&moreLikeThisCount=5
```

### For Individual Records
When retrieving a record detail, MLT can be activated:

```
GET /api/search/v2/museum-123?moreLikeThis=true&moreLikeThisCount=10
```

**Response includes:**
```json
{
  "record": { /* main record data */ },
  "moreLikeThis": [
    {
      "_id": "museum-456",
      "score": 0.95,
      "fields": { /* similar record */ }
    }
  ]
}
```

## Filtering

### Filter Types

#### 1. Standard Query Filters (qf[])
**Format:** `qf[]=field:value`

```
GET /api/search/v2?qf[]=dc.type:painting&qf[]=dc.creator:rembrandt
```

#### 2. ID-based Filters (qf.id[])
For filtering by identifier fields:
```
GET /api/search/v2?qf.id[]=collection:rijksmuseum-paintings
```

#### 3. Existence Filters (qf.exist[])
Filter for records where a field exists:
```
GET /api/search/v2?qf.exist[]=dc.description&qf.exist[]=edm.preview
```

#### 4. Date Filters
**Single date (qf.date[]):**
```
GET /api/search/v2?qf.date[]=dc.date:1642
```

**Date range (qf.dateRange[]):**
```
GET /api/search/v2?qf.dateRange[]=dc.date:1600~1700
```

#### 5. Tree Filters (qf.tree[])
For hierarchical/archival content:
```
GET /api/search/v2?qf.tree[]=archive.fond:NL-HaNA_2.21.281
```

### Combining Filters
All filters use AND logic by default:
```
GET /api/search/v2?q=portrait&qf[]=dc.type:painting&qf.exist[]=edm.preview&qf.dateRange[]=dc.date:1600~1700
```

## Faceting

### Dynamic Facet Configuration

**Parameters:**
- `facet.field[]`: Fields to generate facets for
- `facet.limit`: Max facet values (default: 10)
- `facet.boolType`: Boolean logic (`and`/`or`)
- `facet.orBetween`: Use OR between different facets

**Example:**
```
GET /api/search/v2?q=*&facet.field[]=dc.type&facet.field[]=dc.creator&facet.limit=20
```

### Advanced Faceting

#### Facet Pagination
For facets with many values:
```
GET /api/search/v2?facet.field[]=dc.creator&facet.limit=50&facet.cursor=abc123
```

#### Facet Expansion
Get all values for a specific facet without search results:
```
GET /api/search/v2?facet.expand=dc.creator&facet.limit=100
```

#### Facet Sorting
Control sort order for facet expansion:
```
GET /api/search/v2?facet.expand=dc.creator&facet.sort=count&facet.limit=100
GET /api/search/v2?facet.expand=dc.creator&facet.sort=alpha&facet.limit=100
```

**Sort Options**:
- `facet.sort=count` (default): Sort by frequency (most common first), limited to ~10,000 unique values
- `facet.sort=alpha`: Sort alphabetically, supports deep pagination via `facet.cursor`

**Note**: `facet.cursor` pagination only works with `facet.sort=alpha`

#### Facet Filtering
Filter facet values (supports wildcards):
```
GET /api/search/v2?facet.expand=dc.creator&facet.filter=rem
GET /api/search/v2?facet.expand=dc.creator&facet.filter=*van*
GET /api/search/v2?facet.expand=dc.creator&facet.filter=rem*&facet.sort=count
```

**Filter Behavior**:
- Case-insensitive matching
- Implicit "contains" search: `rem` matches "Rembrandt"
- Wildcard support: `*van*` matches "Rembrandt van Rijn"
- Works with both sort modes

#### Merge Filters
Merge certain facets into the main query:
```
GET /api/search/v2?facet.mergeFilter[]=dc.type&facet.mergeFilter[]=dc.date
```

### Facet Response Structure
```json
{
  "result": {
    "facets": [
      {
        "field": "dc.type",
        "name": "Type",
        "size": 10,
        "total": 156,
        "values": [
          {
            "value": "painting",
            "count": 1234,
            "displayValue": "Painting",
            "searchLink": "?qf[]=dc.type:painting"
          }
        ],
        "nextCursor": "xyz789"
      }
    ]
  }
}
```

## Map Search

### Bounding Box Search
Search within a geographic area:

**Parameters:**
- `min_x`: Western longitude
- `min_y`: Southern latitude  
- `max_x`: Eastern longitude
- `max_y`: Northern latitude

**Example:**
```
GET /api/search/v2?min_x=4.8&min_y=52.3&max_x=4.9&max_y=52.4
```

### Distance Search
Search within radius of a point:

**Parameters:**
- `pt`: Center point (lat,lon format)
- `d`: Distance in kilometers

**Example:**
```
GET /api/search/v2?pt=52.370216,4.895168&d=10
```

## Pagination

### Offset-based Pagination
**Parameters:**
- `start`: Starting position (0-based)
- `rows`: Results per page (max: 1000)
- `page`: Page number (1-based, alternative to start)

**Examples:**
```
GET /api/search/v2?q=*&start=0&rows=20
GET /api/search/v2?q=*&page=2&rows=20
```

### Cursor-based Pagination
For deep pagination beyond 10,000 results:

**Parameters:**
- `searchAfter`: Cursor from previous response

**Example:**
```
GET /api/search/v2?q=*&rows=20&searchAfter=WzE2NDI1MDAwMDAsImRvYy0xMjM0NSJd
```

### Pagination Response
```json
{
  "result": {
    "pagination": {
      "numFound": 152340,
      "start": 20,
      "rows": 20,
      "previousPage": 1,
      "currentPage": 2,
      "nextPage": 3,
      "lastPage": 7617,
      "searchAfter": "WzE2NDI1MDAwMDAsImRvYy0xMjM0NSJd",
      "links": {
        "first": "/api/search/v2?q=*&page=1",
        "previous": "/api/search/v2?q=*&page=1",
        "next": "/api/search/v2?q=*&page=3",
        "last": "/api/search/v2?q=*&page=7617"
      }
    }
  }
}
```

## Sorting

**Parameters:**
- `sortBy`: Field to sort by (prefix with `^` for ascending)
- `sortAsc`: Boolean for ascending order

**Examples:**
```
GET /api/search/v2?q=*&sortBy=dc.date&sortAsc=false
GET /api/search/v2?q=*&sortBy=^dc.title
```

**Common Sort Fields:**
- `_score`: Relevance score (default)
- `dc.date`: Date field
- `dc.title.keyword`: Exact title sorting
- `meta.modified`: Modification timestamp

## Response Formats

### Output Format
**Parameter:** `format`
- `default`: Standard JSON response
- `jsonld`: JSON-LD with @context
- `bulkaction`: Bulk action format

### Item Format
**Parameter:** `itemFormat`
- `summary`: Minimal fields
- `fragmentGraph`: Complete RDF graph
- `jsonld`: Individual items as JSON-LD
- `flat`: Flattened structure
- `tree`: Hierarchical view
- `grouped`: Grouped by predicate

**Example:**
```
GET /api/search/v2?q=painting&format=jsonld&itemFormat=fragmentGraph
```

## Advanced Features

### Result Collapsing
Group results by a field:

**Parameters:**
- `collapseOn`: Field to collapse on
- `collapseSize`: Inner hits per group
- `collapseSort`: Sort for inner hits
- `collapseFormat`: Format for collapsed results

**Example:**
```
GET /api/search/v2?q=*&collapseOn=dc.creator&collapseSize=3
```

### Language Preferences
**Parameter:** `lang` (comma-separated)

```
GET /api/search/v2?q=kunst&lang=nl,en,de
```

### Field Selection
**Parameter:** `fl[]` (field list)

```
GET /api/search/v2?q=*&fl[]=dc.title&fl[]=dc.creator&fl[]=dc.date
```

### Peek Mode
Preview facet values without full search:

```
GET /api/search/v2?peek=dc.creator
```

### Tree Navigation
For hierarchical content:

**Parameters:**
- `tree.cLevel`: Current level
- `tree.cPath`: Current path
- `tree.childCount`: Include child counts

## Common Use Cases

### 1. Advanced Search with Multiple Criteria
```
GET /api/search/v2?q=landscape&qf[]=dc.type:painting&qf.dateRange[]=dc.date:1850~1900&qf[]=dc.creator:"Van Gogh"&facet.field[]=dc.subject&facet.field[]=meta.collection
```

### 2. Geographic Search with Facets
```
GET /api/search/v2?min_x=4.8&min_y=52.3&max_x=4.9&max_y=52.4&facet.field[]=dc.type&facet.field[]=dc.date
```

### 3. Similar Items Discovery
```
GET /api/search/v2/museum-12345?moreLikeThis=true&moreLikeThisCount=20&itemFormat=summary
```

### 4. Hierarchical Archive Browsing
```
GET /api/search/v2?qf.tree[]=archive.fond:NL-HaNA_2.21.281&tree.cLevel=2&tree.childCount=true
```

### 5. Faceted Browse with OR Logic
```
GET /api/search/v2?facet.field[]=dc.type&facet.field[]=dc.creator&facet.boolType=or&facet.orBetween=true
```

### 6. Export Large Result Set
```
GET /api/search/v2?q=meta.collection:manuscripts&format=bulkaction&rows=1000&searchAfter=cursor
```

### 7. Multilingual Search
```
GET /api/search/v2?q=kunst OR art OR arte&lang=nl,en,es&facet.field[]=dc.language
```

## Error Handling

### Common HTTP Status Codes
- `200 OK`: Successful request
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Record not found (detail endpoints)
- `500 Internal Server Error`: Server error

### Error Response Format
```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Invalid date range format",
    "details": {
      "parameter": "qf.dateRange",
      "value": "dc.date:invalid"
    }
  }
}
```

## Performance Tips

1. **Use specific queries** instead of wildcards when possible
2. **Limit facet fields** to those actually needed
3. **Use cursor-based pagination** for large result sets
4. **Enable result collapsing** to reduce duplicate content
5. **Specify itemFormat** to reduce response size
6. **Use field selection** (fl[]) to retrieve only needed fields

## Migration from V1

Key differences from V1:
- Dynamic facet configuration
- Multiple filter types
- Geospatial search support
- More Like This functionality
- Advanced pagination options
- Multiple response formats
- Result collapsing
- Language preferences
- Tree navigation

Field naming: V2 uses dots instead of underscores:
- V1: `dc_title` → V2: `dc.title`
- V1: `meta_spec` → V2: `meta.spec`