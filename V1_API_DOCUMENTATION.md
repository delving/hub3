# Hub3 V1 API Documentation

## Overview

The Hub3 V1 API is the legacy search API that provides basic search, filtering, and record retrieval functionality. This API follows traditional REST patterns and returns results in a fixed format optimized for backward compatibility.

## Table of Contents

1. [Base Endpoints](#base-endpoints)
2. [Search Functionality](#search-functionality)
3. [Record Detail View](#record-detail-view)
4. [Filtering](#filtering)
5. [Faceting](#faceting)
6. [Pagination](#pagination)
7. [Sorting](#sorting)
8. [Response Format](#response-format)
9. [Common Use Cases](#common-use-cases)

## Base Endpoints

### Search Endpoint
```
GET /api/search/v1
```
Search across all records with filtering and faceting capabilities.

### Record Detail Endpoint
```
GET /api/search/v1/{id}
```
Retrieve a specific record by its identifier.

## Search Functionality

### Basic Search Query
The main search parameter is `query` which accepts any valid Lucene query syntax.

**Parameters:**
- `query` (string): The search query using Lucene syntax
- `id` (string): When provided in the query parameters, retrieves a specific record

**Example:**
```
GET /api/search/v1?query=rembrandt
GET /api/search/v1?query=title:"night watch"
GET /api/search/v1?query=date:[1600 TO 1700]
```

### Query Syntax
- Simple terms: `rembrandt`
- Phrases: `"night watch"`
- Field searches: `dc_creator:rembrandt`
- Wildcards: `rem*` or `?embrandt`
- Boolean operators: `rembrandt AND painting NOT sketch`
- Range queries: `date:[1600 TO 1700]`

## Record Detail View

### Retrieving Individual Records

**Endpoint:**
```
GET /api/search/v1/{id}
```

**Parameters:**
- `{id}` (path): The unique identifier of the record
- `format` (query): Response format (default: v1)

**Example:**
```
GET /api/search/v1/museum-collection-12345
```

### Alternative Method
You can also retrieve a record using the search endpoint with an id parameter:
```
GET /api/search/v1?id=museum-collection-12345
```

**Response:** Returns the complete record in V1 format with all metadata fields.

## Filtering

### Query Filters (qf[])
Filters allow you to narrow down search results by specific field values.

**Format:** `qf[]=field:value`

**Common Filter Fields:**
- `meta.spec`: Dataset/collection identifier
- `dc_type`: Type of object
- `dc_creator`: Creator/artist
- `dc_subject`: Subject matter
- `dc_date`: Date fields
- `edm_dataProvider`: Data provider institution

**Examples:**
```
GET /api/search/v1?qf[]=meta.spec:paintings&qf[]=dc_creator:rembrandt
GET /api/search/v1?qf[]=dc_type:painting&query=landscape
```

### Multiple Filters
Multiple filters can be combined. By default, filters use AND logic.

```
GET /api/search/v1?qf[]=dc_type:painting&qf[]=dc_date:1650&qf[]=dc_creator:rembrandt
```

## Faceting

In V1 API, facets are configured server-side and automatically included based on the organization's configuration. Common facet fields include:

- `meta.spec` - Collections/datasets
- `dc_type` - Object types
- `dc_creator` - Creators/artists
- `dc_subject` - Subjects
- `dc_date` - Dates
- `edm_dataProvider` - Providing institutions
- `dc_language` - Languages

### Facet Response Structure
```json
{
  "result": {
    "facets": [
      {
        "name": "dc_type",
        "displayName": "Type",
        "items": [
          {
            "value": "painting",
            "count": 1234,
            "url": "?qf[]=dc_type:painting"
          },
          {
            "value": "drawing",
            "count": 567,
            "url": "?qf[]=dc_type:drawing"
          }
        ]
      }
    ]
  }
}
```

## Pagination

V1 API uses offset-based pagination with `start` and `rows` parameters.

**Parameters:**
- `start` (integer): Starting position (0-based), default: 0
- `rows` (integer): Number of results per page, default: 20, max: 1000

**Examples:**
```
GET /api/search/v1?query=*&start=0&rows=20    # First page
GET /api/search/v1?query=*&start=20&rows=20   # Second page
GET /api/search/v1?query=*&start=40&rows=20   # Third page
```

### Pagination Response
```json
{
  "result": {
    "pagination": {
      "numFound": 15234,
      "start": 20,
      "rows": 20,
      "hasNext": true,
      "hasPrevious": true,
      "nextStart": 40,
      "previousStart": 0
    }
  }
}
```

## Sorting

**Parameters:**
- `sortBy` (string): Field to sort by
- `sortOrder` (string): Sort direction (`asc` or `desc`)

**Common Sort Fields:**
- `dc_date` - Date
- `dc_title` - Title
- `timestamp` - Index timestamp

**Examples:**
```
GET /api/search/v1?query=*&sortBy=dc_date&sortOrder=desc
GET /api/search/v1?query=painting&sortBy=dc_title&sortOrder=asc
```

## Response Format

### Search Response Structure
```json
{
  "result": {
    "query": {
      "query": "rembrandt",
      "filters": ["dc_type:painting"]
    },
    "pagination": {
      "numFound": 142,
      "start": 0,
      "rows": 20
    },
    "items": [
      {
        "_id": "museum-123",
        "fields": {
          "dc_title": ["The Night Watch"],
          "dc_creator": ["Rembrandt van Rijn"],
          "dc_date": ["1642"],
          "dc_type": ["painting"]
        }
      }
    ],
    "facets": [...]
  }
}
```

### Detail Response Structure
```json
{
  "_id": "museum-123",
  "fields": {
    "dc_title": ["The Night Watch"],
    "dc_creator": ["Rembrandt van Rijn"],
    "dc_date": ["1642"],
    "dc_description": ["A militia group portrait..."],
    "dc_type": ["painting"],
    "edm_dataProvider": ["Rijksmuseum"],
    "edm_isShownAt": ["https://www.rijksmuseum.nl/..."],
    "edm_preview": ["https://media.rijksmuseum.nl/..."]
  },
  "_source": {
    // Original source data
  }
}
```

## Common Use Cases

### 1. Search within a Collection
```
GET /api/search/v1?query=landscape&qf[]=meta.spec:paintings
```

### 2. Browse by Creator
```
GET /api/search/v1?qf[]=dc_creator:"van Gogh"&rows=50
```

### 3. Date Range Search
```
GET /api/search/v1?query=portrait&qf[]=dc_date:[1600 TO 1700]
```

### 4. Get All Records from a Provider
```
GET /api/search/v1?qf[]=edm_dataProvider:Rijksmuseum&rows=100
```

### 5. Search and Filter by Type
```
GET /api/search/v1?query=flowers&qf[]=dc_type:drawing
```

### 6. Retrieve Specific Record
```
GET /api/search/v1/rijksmuseum-sk-a-1234
```

## Field Naming Conventions

V1 API uses underscores in field names:
- `dc_title` - Dublin Core title
- `dc_creator` - Dublin Core creator
- `meta_spec` - Metadata specification/collection
- `edm_dataProvider` - Europeana Data Model data provider

## Limitations

- Fixed facet configuration (cannot dynamically request facets)
- Limited to 1000 results per page
- No support for advanced features like:
  - Geospatial search
  - Dynamic facet configuration
  - Result collapsing
  - Custom response formats
  - More Like This functionality

## Migration Note

For new implementations, consider using the V2 API which offers:
- Dynamic facet configuration
- Advanced filtering options
- Geospatial search
- Multiple response formats
- Better pagination options
- More Like This functionality