package semantic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// GenerateAPIReference generates a human-readable Markdown API reference
// from the same code structures that power the /docs endpoint.
// This ensures documentation never drifts from the implementation.
func GenerateAPIReference(baseURL string) string {
	if baseURL == "" {
		baseURL = "/api/semantic/v1"
	}

	var b strings.Builder

	writeHeader(&b, baseURL)
	writeEndpoints(&b, baseURL)
	writeQueryParameters(&b)
	writeOperators(&b)
	writeFilterSyntax(&b, baseURL)
	writeFacetSyntax(&b, baseURL)
	writeSortSyntax(&b, baseURL)
	writePagination(&b)
	writeCollapse(&b)
	writeCrossIndex(&b, baseURL)
	writeSearchContext(&b, baseURL)
	writeResponseFormats(&b)
	writeHydraVocabulary(&b)
	writeJSONLDContext(&b)
	writeCapabilities(&b)
	writeErrorHandling(&b)

	return b.String()
}

func writeHeader(b *strings.Builder, baseURL string) {
	b.WriteString("# Hub3 Semantic Search API Reference\n\n")
	b.WriteString("> This document is generated from code. Do not edit manually.\n")
	b.WriteString("> Regenerate with: `go run ./tools/cmd/gendocs`\n\n")
	b.WriteString("## Overview\n\n")
	b.WriteString("The Hub3 Semantic Search API provides a Hydra-compatible JSON-LD interface\n")
	b.WriteString("for searching, filtering, and navigating cultural heritage metadata.\n\n")
	fmt.Fprintf(b, "- **Base URL:** `%s`\n", baseURL)
	b.WriteString("- **Content-Type:** `application/ld+json`\n")
	b.WriteString("- **Version:** 1.0.0\n\n")
	b.WriteString("---\n\n")
}

func writeEndpoints(b *strings.Builder, baseURL string) {
	b.WriteString("## Endpoints\n\n")
	b.WriteString("| Method | Path | Description |\n")
	b.WriteString("|--------|------|-------------|\n")

	endpoints := []struct{ method, path, desc string }{
		{"GET", baseURL + "/", "API entry point with navigation links"},
		{"GET", baseURL + "/docs", "Machine-readable API documentation (Hydra ApiDocumentation)"},
		{"GET", baseURL + "/search", "Search resources with URL query parameters"},
		{"POST", baseURL + "/search", "Search resources with JSON-LD query body"},
		{"GET", baseURL + "/resource/{id}", "Get a single resource by ID"},
		{"GET", baseURL + "/resource/{id}?include=relatedItems", "Get resource with related items (MLT)"},
		{"GET", baseURL + "/type/{type}/search", "Search within a specific resource type"},
		{"POST", baseURL + "/type/{type}/search", "Type-scoped search with JSON-LD query body"},
		{"GET", baseURL + "/type/{type}/docs", "Documentation for a specific resource type"},
		{"GET", baseURL + "/introspect/classes", "List all classes in the index"},
		{"GET", baseURL + "/introspect/classes/{class}/properties", "List properties for a class"},
		{"GET", baseURL + "/introspect/fields/{field}", "Get details for a specific field"},
		{"GET", baseURL + "/introspect/paths", "List all field paths in the index"},
		{"POST", baseURL + "/contexts/query/", "Create a search context for detail navigation"},
		{"GET", baseURL + "/contexts/query/{id}", "Get a saved search context"},
		{"DELETE", baseURL + "/contexts/query/{id}", "Delete a search context"},
	}

	for _, e := range endpoints {
		fmt.Fprintf(b, "| `%s` | `%s` | %s |\n", e.method, e.path, e.desc)
	}

	b.WriteString("\n---\n\n")
}

func writeQueryParameters(b *strings.Builder) {
	b.WriteString("## Query Parameters\n\n")
	b.WriteString("All search parameters are passed as URL query parameters on `GET` requests.\n\n")
	b.WriteString("| Parameter | Type | Description | Example |\n")
	b.WriteString("|-----------|------|-------------|---------|\n")

	params := []struct{ name, typ, desc, example string }{
		{"query", "string", "Full-text search query", "`query=rembrandt`"},
		{"filter[field][op]", "string", "Property filter (see Operators)", "`filter[dc_creator][eq]=Rembrandt`"},
		{"hfilter[field][op]", "string", "Hidden filter (not reflected in facet counts)", "`hfilter[dc_type][eq]=painting`"},
		{"facet[field]", "int", "Request facet aggregation with limit", "`facet[dc_type]=20`"},
		{"facet[field].sort", "string", "Facet sort order: `count` or `index`", "`facet[dc_type].sort=index`"},
		{"facet[field].cursor", "string", "Cursor for paginating facet values", "`facet[dc_type].cursor=abc`"},
		{"facetBool", "string", "Facet logic: `and` or `or` (default: `or`)", "`facetBool=and`"},
		{"page", "int", "Page number (default: 1)", "`page=2`"},
		{"size", "int", fmt.Sprintf("Results per page (default: %d)", semantic.DefaultPagination().Size), "`size=50`"},
		{"cursor", "string", "Opaque cursor for deep pagination", "`cursor=eyJhZnRl...`"},
		{"sort", "string", "Sort field; prefix `-` for descending", "`sort=-dc_date`"},
		{"collapse", "string", "Group results by field", "`collapse=edm_dataProvider`"},
		{"collapse.size", "int", "Inner hits per group (default: 1)", "`collapse.size=3`"},
		{"collapse.sort", "string", "Sort inner hits; prefix `-` for desc", "`collapse.sort=-dc_date`"},
		{"peek", "string", "Comma-separated fields for facet-only response (size=0)", "`peek=dc_type,dc_creator`"},
		{"languages", "string", "Comma-separated language preferences", "`languages=en,nl`"},
		{"fields", "string", "Comma-separated field selection", "`fields=dc_title,dc_creator`"},
		{"expand", "string", "Relations to expand inline", "`expand=relatedItems`"},
		{"detailLevel", "string", "Response detail: `minimal`, `standard`, `full`", "`detailLevel=full`"},
		{"debug", "string", "Diagnostic mode (e.g., `query` shows ES query)", "`debug=query`"},
		{"contextIndex", "string", "Additional org IDs for cross-index search (repeatable)", "`contextIndex=org-b`"},
		{"include", "string", "Sections to include on detail endpoint", "`include=relatedItems`"},
	}

	for _, p := range params {
		fmt.Fprintf(b, "| `%s` | %s | %s | %s |\n", p.name, p.typ, p.desc, p.example)
	}

	b.WriteString("\n---\n\n")
}

func writeOperators(b *strings.Builder) {
	b.WriteString("## Filter Operators\n\n")
	b.WriteString("Used in `filter[field][operator]=value` syntax.\n\n")
	b.WriteString("| Operator | Description | Example |\n")
	b.WriteString("|----------|-------------|---------|\n")

	for _, op := range semantic.AllOperators() {
		example := operatorExample(op)
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", op.String(), op.Description(), example)
	}

	b.WriteString("\n---\n\n")
}

func operatorExample(op semantic.Operator) string {
	switch op {
	case semantic.OpEqual:
		return "`filter[dc_creator][eq]=Rembrandt`"
	case semantic.OpNotEqual:
		return "`filter[dc_type][neq]=painting`"
	case semantic.OpIn:
		return "`filter[dc_type][in]=painting&filter[dc_type][in]=drawing`"
	case semantic.OpNotIn:
		return "`filter[dc_type][nin]=sketch`"
	case semantic.OpGreaterThan:
		return "`filter[dc_date][gt]=1600`"
	case semantic.OpGreaterEqual:
		return "`filter[dc_date][gte]=1600`"
	case semantic.OpLessThan:
		return "`filter[dc_date][lt]=1700`"
	case semantic.OpLessEqual:
		return "`filter[dc_date][lte]=1700`"
	case semantic.OpContains:
		return "`filter[dc_title][contains]=night`"
	case semantic.OpStartsWith:
		return "`filter[dc_title][startsWith]=Night`"
	case semantic.OpExists:
		return "`filter[nave_thumbnail][exists]=true`"
	case semantic.OpBBox:
		return "`filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4`"
	default:
		return ""
	}
}

func writeFilterSyntax(b *strings.Builder, baseURL string) {
	b.WriteString("## Filter Syntax\n\n")
	b.WriteString("Filters use bracket notation: `filter[field][operator]=value`\n\n")
	b.WriteString("Field names use underscores in URLs (`dc_creator`) which map to colons internally (`dc:creator`).\n\n")
	b.WriteString("### Regular filters\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "%s/search?filter[dc_creator][eq]=Rembrandt\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("### Hidden filters\n\n")
	b.WriteString("Hidden filters narrow results without affecting facet counts:\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "%s/search?hfilter[dc_type][eq]=painting&facet[dc_type]=10\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("### Range filters\n\n")
	b.WriteString("Combine `gte` and `lte` operators for range queries:\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "%s/search?filter[dc_date][gte]=1600&filter[dc_date][lte]=1700\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("### Geospatial filters\n\n")
	b.WriteString("Bounding box filter with west,south,east,north coordinates:\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "%s/search?filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("---\n\n")
}

func writeFacetSyntax(b *strings.Builder, baseURL string) {
	b.WriteString("## Facet Syntax\n\n")
	b.WriteString("Request facet aggregations with `facet[field]=limit`.\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "# Request dc_type facet with top 20 values\n")
	fmt.Fprintf(b, "%s/search?facet[dc_type]=20\n\n", baseURL)
	fmt.Fprintf(b, "# Multiple facets\n")
	fmt.Fprintf(b, "%s/search?facet[dc_type]=10&facet[dc_creator]=20\n\n", baseURL)
	fmt.Fprintf(b, "# Facet with cursor pagination\n")
	fmt.Fprintf(b, "%s/search?facet[dc_creator]=50&facet[dc_creator].cursor=abc123\n\n", baseURL)
	fmt.Fprintf(b, "# Sort facet alphabetically instead of by count\n")
	fmt.Fprintf(b, "%s/search?facet[dc_type]=10&facet[dc_type].sort=index\n", baseURL)
	b.WriteString("```\n\n")

	b.WriteString("### Facet Bool Type\n\n")
	b.WriteString("Controls how multiple facet selections combine:\n\n")
	b.WriteString("| Value | Behavior |\n")
	b.WriteString("|-------|----------|\n")
	b.WriteString("| `or` (default) | Selected values broaden results |\n")
	b.WriteString("| `and` | Selected values narrow results |\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "%s/search?facetBool=and&filter[dc_type][eq]=painting\n", baseURL)
	b.WriteString("```\n\n")

	b.WriteString("### Supported Facet Types\n\n")
	for _, ft := range supportedFacetTypes() {
		fmt.Fprintf(b, "- `%s`\n", ft)
	}
	b.WriteString("\n---\n\n")
}

func writeSortSyntax(b *strings.Builder, baseURL string) {
	b.WriteString("## Sort Syntax\n\n")
	b.WriteString("Use the `sort` parameter with an optional `-` prefix for descending order.\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "# Sort ascending (default)\n")
	fmt.Fprintf(b, "%s/search?sort=dc_date\n\n", baseURL)
	fmt.Fprintf(b, "# Sort descending\n")
	fmt.Fprintf(b, "%s/search?sort=-dc_date\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("---\n\n")
}

func writePagination(b *strings.Builder) {
	defaults := semantic.DefaultPagination()
	caps := semantic.DefaultStoreCapabilities()

	b.WriteString("## Pagination\n\n")
	b.WriteString("| Parameter | Default | Description |\n")
	b.WriteString("|-----------|---------|-------------|\n")
	fmt.Fprintf(b, "| `page` | %d | Page number (1-based) |\n", defaults.Page)
	fmt.Fprintf(b, "| `size` | %d | Results per page (max: %d) |\n", defaults.Size, caps.MaxPageSize)
	b.WriteString("| `cursor` | - | Opaque cursor for deep pagination |\n\n")
	b.WriteString("The response includes a `hydra:PartialCollectionView` with navigation links:\n\n")
	b.WriteString("```json\n")
	b.WriteString("{\n")
	b.WriteString("  \"view\": {\n")
	b.WriteString("    \"@type\": \"PartialCollectionView\",\n")
	b.WriteString("    \"first\": \"/api/semantic/v1/search?query=...&page=1\",\n")
	b.WriteString("    \"next\": \"/api/semantic/v1/search?query=...&page=3\",\n")
	b.WriteString("    \"previous\": \"/api/semantic/v1/search?query=...&page=1\"\n")
	b.WriteString("  }\n")
	b.WriteString("}\n")
	b.WriteString("```\n\n")
	b.WriteString("---\n\n")
}

func writeCollapse(b *strings.Builder) {
	b.WriteString("## Result Collapsing\n\n")
	b.WriteString("Group results by a field value (e.g., deduplicate by data provider).\n\n")
	b.WriteString("| Parameter | Description |\n")
	b.WriteString("|-----------|-------------|\n")
	b.WriteString("| `collapse=field` | Field to group by |\n")
	b.WriteString("| `collapse.size=N` | Inner hits per group (default: 1) |\n")
	b.WriteString("| `collapse.sort=field` | Sort order for inner hits (prefix `-` for desc) |\n\n")
	b.WriteString("---\n\n")
}

func writeCrossIndex(b *strings.Builder, baseURL string) {
	b.WriteString("## Cross-Index Search\n\n")
	b.WriteString("Search across multiple organization indices in a single query using `contextIndex`.\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "# Search primary org plus org-b\n")
	fmt.Fprintf(b, "%s/search?query=rembrandt&contextIndex=org-b\n\n", baseURL)
	fmt.Fprintf(b, "# Search across three organizations\n")
	fmt.Fprintf(b, "%s/search?query=*&contextIndex=org-a&contextIndex=org-b\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("The parameter is repeatable. The primary organization (resolved from the request domain) is always included.\n\n")
	b.WriteString("---\n\n")
}

func writeSearchContext(b *strings.Builder, baseURL string) {
	b.WriteString("## Search Context & Detail Navigation\n\n")
	b.WriteString("The Search Context is a core concept that enables **stateful navigation through\n")
	b.WriteString("search results at the item level**. It bridges the gap between a search result list\n")
	b.WriteString("and individual resource detail views, allowing users to step through results with\n")
	b.WriteString("next/previous links without losing their place.\n\n")

	b.WriteString("### How It Works\n\n")
	b.WriteString("```\n")
	b.WriteString("1. User performs a search\n")
	b.WriteString("   GET /api/semantic/v1/search?query=rembrandt\n")
	b.WriteString("   \u2192 Response includes hub3:searchContext with a token\n\n")
	b.WriteString("2. User clicks a result to view details\n")
	b.WriteString("   GET /api/semantic/v1/resource/{id}?context={token}\n")
	b.WriteString("   \u2192 Response includes hub3:navigation with next/previous/first/last links\n\n")
	b.WriteString("3. User navigates to next result\n")
	b.WriteString("   GET /api/semantic/v1/resource/{next-id}?context={token}\n")
	b.WriteString("   \u2192 Navigation updates to reflect new position\n")
	b.WriteString("```\n\n")

	b.WriteString("### Lifecycle\n\n")
	b.WriteString("| Phase | What Happens |\n")
	b.WriteString("|-------|-------------|\n")
	fmt.Fprintf(b, "| **Creation** | Automatically created when a search returns results. "+
		"The context captures the ordered list of result IDs and the original query. |\n")
	fmt.Fprintf(b, "| **Usage** | Pass the context token via `?context={token}` on resource detail requests "+
		"to activate navigation. |\n")
	fmt.Fprintf(b, "| **Expiration** | Contexts expire after %s of inactivity. "+
		"Accessing the context extends its TTL. |\n", "15 minutes")
	b.WriteString("| **Deletion** | Contexts can be explicitly deleted or are cleaned up after expiry. |\n\n")

	b.WriteString("### Search Response with Context\n\n")
	b.WriteString("When a search returns results, the response includes a context reference:\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
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
`)
	b.WriteString("```\n\n")

	b.WriteString("### Resource Detail with Navigation\n\n")
	b.WriteString("When you fetch a resource with a context token, the response includes navigation:\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "GET %s/resource/{id}?context=a1b2c3d4\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
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
`)
	b.WriteString("```\n\n")

	b.WriteString("### Navigation Fields\n\n")
	b.WriteString("| Field | Type | Description |\n")
	b.WriteString("|-------|------|-------------|\n")
	b.WriteString("| `position` | int | Current 1-based position in the result set |\n")
	b.WriteString("| `totalResults` | int | Total number of results in the search |\n")
	b.WriteString("| `hasNext` | bool | Whether there is a next result |\n")
	b.WriteString("| `hasPrevious` | bool | Whether there is a previous result |\n")
	b.WriteString("| `hydra:first` | URL | Link to the first result in the set |\n")
	b.WriteString("| `hydra:previous` | URL | Link to the previous result |\n")
	b.WriteString("| `hydra:next` | URL | Link to the next result |\n")
	b.WriteString("| `hydra:last` | URL | Link to the last result in the set |\n")
	b.WriteString("| `hub3:backToSearch` | URL | Link back to the original search results page |\n\n")

	b.WriteString("### Context Management API\n\n")
	b.WriteString("Contexts are usually created automatically during search, but can also be managed explicitly.\n\n")

	b.WriteString("#### Create a Context\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "POST %s/contexts/query/\n", baseURL)
	b.WriteString("Content-Type: application/json\n\n")
	b.WriteString(`{
  "query": { "text": "rembrandt" },
  "totalResults": 150
}
`)
	b.WriteString("```\n\n")
	b.WriteString("Response (`201 Created`):\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
  "@type": "hub3:QueryContext",
  "id": "a1b2c3d4",
  "totalResults": 150,
  "expiresAt": "2026-02-26T15:30:00Z"
}
`)
	b.WriteString("```\n\n")

	b.WriteString("#### Retrieve a Context\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "GET %s/contexts/query/{contextID}\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("Returns the context metadata, or `404` if expired/not found.\n\n")

	b.WriteString("#### Delete a Context\n\n")
	b.WriteString("```\n")
	fmt.Fprintf(b, "DELETE %s/contexts/query/{contextID}\n", baseURL)
	b.WriteString("```\n\n")
	b.WriteString("Returns `204 No Content`.\n\n")

	b.WriteString("### Frontend Integration Pattern\n\n")
	b.WriteString("A typical frontend integration follows this pattern:\n\n")
	b.WriteString("1. **Search page:** Execute search, store the `hub3:searchContext.token` from the response\n")
	b.WriteString("2. **Detail page:** When navigating to a result, append `?context={token}` to the resource URL\n")
	b.WriteString("3. **Navigation UI:** Use `hub3:navigation` fields to render next/previous buttons\n")
	b.WriteString("4. **Back button:** Use `hub3:backToSearch` to return to the search results\n")
	b.WriteString("5. **Expiration:** If the context is not found (expired), gracefully degrade to a detail\n")
	b.WriteString("   view without navigation\n\n")

	b.WriteString("---\n\n")
}

func writeResponseFormats(b *strings.Builder) {
	b.WriteString("## Response Formats\n\n")
	b.WriteString("All responses use `application/ld+json` content type.\n\n")

	b.WriteString("### Search Response (`hydra:Collection`)\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
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
`)
	b.WriteString("```\n\n")

	b.WriteString("### Resource Detail Response\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
  "@context": { "...": "..." },
  "@id": "http://example.org/resource/1",
  "@type": ["edm:ProvidedCHO"],
  "dc_title": "The Night Watch",
  "dc_creator": "Rembrandt van Rijn",
  "dc_date": "1642"
}
`)
	b.WriteString("```\n\n")

	b.WriteString("### Error Response (`hydra:Error`)\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
  "@type": "Error",
  "hydra:title": "Bad Request",
  "hydra:description": "invalid operator 'invalid' for field 'creator'",
  "hub3:statusCode": 400
}
`)
	b.WriteString("```\n\n")
	b.WriteString("---\n\n")
}

func writeHydraVocabulary(b *strings.Builder) {
	b.WriteString("## Hydra Vocabulary\n\n")
	b.WriteString("The API uses the [Hydra Core Vocabulary](http://www.w3.org/ns/hydra/core) for hypermedia controls.\n\n")

	b.WriteString("### Standard Hydra Types\n\n")
	b.WriteString("| Type | Usage |\n")
	b.WriteString("|------|-------|\n")
	b.WriteString("| `hydra:EntryPoint` | Root endpoint (`/`) |\n")
	b.WriteString("| `hydra:ApiDocumentation` | API docs (`/docs`) |\n")
	b.WriteString("| `hydra:Collection` | Search result sets |\n")
	b.WriteString("| `hydra:PartialCollectionView` | Pagination links |\n")
	b.WriteString("| `hydra:IriTemplate` | URL template for search |\n")
	b.WriteString("| `hydra:Operation` | Supported HTTP operations |\n")
	b.WriteString("| `hydra:Class` | Resource type definitions |\n")
	b.WriteString("| `hydra:SupportedProperty` | Filterable properties |\n")
	b.WriteString("| `hydra:Error` | Error responses |\n\n")

	b.WriteString("### Hub3 Custom Types\n\n")
	b.WriteString("| Type | Description |\n")
	b.WriteString("|------|-------------|\n")

	hub3Types := []struct{ name, desc string }{
		{"hub3:SearchQuery", "Structured search query (POST body)"},
		{"hub3:TextQuery", "Full-text search query component"},
		{"hub3:PropertyFilter", "Property equality/comparison filter"},
		{"hub3:RangeFilter", "Numeric/date range filter"},
		{"hub3:ExistsFilter", "Field existence filter"},
		{"hub3:GeoBBoxFilter", "Geographic bounding box filter"},
		{"hub3:GeoDistanceFilter", "Geographic distance filter"},
		{"hub3:GeoPolygonFilter", "Geographic polygon filter"},
		{"hub3:FacetRequest", "Facet aggregation request"},
		{"hub3:Facet", "Facet result in response"},
		{"hub3:FacetValue", "Individual facet value with count"},
		{"hub3:GeoCluster", "Geographic cluster aggregation result"},
		{"hub3:SearchResultNavigation", "Detail-level navigation context"},
		{"hub3:NavigateOperation", "Navigation operation (next/previous)"},
		{"hub3:FilterDefinition", "Filter definition in type documentation"},
		{"hub3:FacetDefinition", "Facet definition in type documentation"},
		{"hub3:SortField", "Sort field definition in type documentation"},
		{"hub3:Example", "API usage example"},
	}

	for _, t := range hub3Types {
		fmt.Fprintf(b, "| `%s` | %s |\n", t.name, t.desc)
	}

	b.WriteString("\n---\n\n")
}

func writeJSONLDContext(b *strings.Builder) {
	b.WriteString("## JSON-LD Context\n\n")
	b.WriteString("The API provides JSON-LD contexts for semantic interoperability.\n\n")

	b.WriteString("### Available Contexts\n\n")
	b.WriteString("| Context | URL |\n")
	b.WriteString("|---------|-----|\n")
	b.WriteString("| Hub3 Search Vocabulary | `/contexts/hub3/1.0/context.jsonld` |\n")
	b.WriteString("| Hub3 Search (latest) | `/contexts/hub3/latest/context.jsonld` |\n")
	b.WriteString("| EDM (Europeana Data Model) | `/contexts/edm/1.0/context.jsonld` |\n")
	b.WriteString("| EDM (latest) | `/contexts/edm/latest/context.jsonld` |\n\n")

	b.WriteString("### Namespace Prefixes\n\n")

	ctx := semantic.DefaultContext()
	prefixes := make([]struct{ prefix, uri string }, 0)

	for k, v := range ctx {
		uri, ok := v.(string)
		if !ok {
			continue
		}

		// Only include namespace prefixes (URIs), not shortcuts
		if strings.HasPrefix(uri, "http") {
			prefixes = append(prefixes, struct{ prefix, uri string }{k, uri})
		}
	}

	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].prefix < prefixes[j].prefix
	})

	b.WriteString("| Prefix | Namespace URI |\n")
	b.WriteString("|--------|---------------|\n")

	for _, p := range prefixes {
		fmt.Fprintf(b, "| `%s` | `%s` |\n", p.prefix, p.uri)
	}

	b.WriteString("\n---\n\n")
}

func writeCapabilities(b *strings.Builder) {
	caps := buildCapabilities()

	b.WriteString("## API Capabilities\n\n")
	b.WriteString("These values are derived from code and always reflect the running system.\n\n")
	fmt.Fprintf(b, "- **Default page size:** %v\n", caps["defaultPageSize"])
	fmt.Fprintf(b, "- **Maximum page size:** %v\n", caps["maxPageSize"])
	fmt.Fprintf(b, "- **Sort options:** %v\n", caps["sortOptions"])
	fmt.Fprintf(b, "- **Search context TTL:** %v\n", caps["contextTTL"])
	fmt.Fprintf(b, "- **Path separator:** `%v`\n\n", caps["pathSeparator"])
	b.WriteString("---\n\n")
}

func writeErrorHandling(b *strings.Builder) {
	b.WriteString("## Error Handling\n\n")
	b.WriteString("All errors return a `hydra:Error` response with appropriate HTTP status codes.\n\n")
	b.WriteString("| Status | Description |\n")
	b.WriteString("|--------|-------------|\n")
	b.WriteString("| 400 | Bad Request — invalid parameters, operators, or malformed query |\n")
	b.WriteString("| 404 | Not Found — unknown resource, type, or search context |\n")
	b.WriteString("| 500 | Internal Server Error — unexpected backend failure |\n\n")
	b.WriteString("Error responses include `hydra:title` (short label) and `hydra:description` (detailed message).\n")
}
