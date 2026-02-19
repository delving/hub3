package semantic

import (
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/go-chi/chi/v5"
)

// handleEntryPoint handles the root API entry point.
func (s *Service) handleEntryPoint(w http.ResponseWriter, r *http.Request) {
	entryPoint := s.buildEntryPointResponse(r)
	s.respondJSON(w, r, entryPoint, http.StatusOK)
}

// handleAPIDocumentation handles API documentation requests.
func (s *Service) handleAPIDocumentation(w http.ResponseWriter, r *http.Request) {
	doc := s.buildAPIDocumentation(r)
	s.respondJSON(w, r, doc, http.StatusOK)
}

// handleTypeDocumentation handles type-specific documentation requests.
func (s *Service) handleTypeDocumentation(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")

	config, ok := s.registry.Get(resourceType)
	if !ok {
		s.respondError(w, r, semantic.NewHydraError(
			"Unknown Resource Type",
			fmt.Sprintf("Resource type '%s' is not configured", resourceType),
			http.StatusNotFound,
		))
		return
	}

	doc := s.buildTypeDocumentation(r, config)
	s.respondJSON(w, r, doc, http.StatusOK)
}

// buildEntryPointResponse builds the lightweight API entry point.
func (s *Service) buildEntryPointResponse(r *http.Request) map[string]any {
	baseURL := s.buildBaseURL(r)
	searchURL := baseURL + "/search"
	docsURL := baseURL + "/docs"

	entryPoint := map[string]any{
		"@context":          semantic.DefaultContext(),
		"@id":               baseURL,
		"@type":             "hydra:EntryPoint",
		"hydra:title":       "Hub3 Semantic Search API",
		"hydra:description": "Entry point for the semantic search API. Visit /docs for full documentation.",
		"hub3:version":      "1.0.0",
	}

	// Main operations
	entryPoint["hub3:search"] = map[string]any{
		"@id":               searchURL,
		"@type":             "hydra:Link",
		"hydra:title":       "Search Endpoint",
		"hydra:description": "Search resources using GET or POST requests",
		"hub3:methods":      []string{"GET", "POST"},
	}

	entryPoint["hub3:documentation"] = map[string]any{
		"@id":               docsURL,
		"@type":             "hydra:Link",
		"hydra:title":       "Full API Documentation",
		"hydra:description": "Complete API documentation with all resource types, filters, and examples",
	}

	// List available resource types with their endpoints
	types := s.registry.ListResourceTypes()
	resourceTypes := make([]map[string]any, 0, len(types))

	for _, resourceType := range types {
		config, ok := s.registry.Get(resourceType)
		if !ok {
			continue
		}

		resourceTypes = append(resourceTypes, map[string]any{
			"@id":               resourceType,
			"@type":             "hydra:Class",
			"hydra:title":       config.Label,
			"hydra:description": config.Description,
			"hub3:searchURL":    fmt.Sprintf("%s/type/%s/search", baseURL, resourceType),
			"hub3:docsURL":      fmt.Sprintf("%s/type/%s/docs", baseURL, resourceType),
		})
	}
	entryPoint["hub3:resourceTypes"] = resourceTypes

	// Quick examples
	entryPoint["hub3:quickExamples"] = []map[string]any{
		{
			"hydra:title":       "Simple search",
			"hub3:url":          searchURL + "?query=rembrandt",
			"hydra:description": "Full-text search across all resources",
		},
		{
			"hydra:title":       "Search specific type",
			"hub3:url":          baseURL + "/type/ore:Aggregation/search?query=painting",
			"hydra:description": "Search within a specific resource type",
		},
		{
			"hydra:title":       "View documentation",
			"hub3:url":          docsURL,
			"hydra:description": "Complete API documentation with all capabilities",
		},
	}

	return entryPoint
}

// buildAPIDocumentation builds the main API documentation.
func (s *Service) buildAPIDocumentation(r *http.Request) map[string]any {
	baseURL := s.buildBaseURL(r)

	doc := map[string]any{
		"@context": semantic.DefaultContext(),
		"@id":      baseURL + "/docs",
		"@type":    "hydra:ApiDocumentation",
		"hydra:title": "Hub3 Semantic Search API - Complete Documentation",
		"hydra:description": "Complete API documentation including all resource types, supported filters, " +
			"facets, operators, and detailed examples. For a quick overview, visit the root endpoint.",
		"hub3:version": "1.0.0",
		"hub3:baseURL": baseURL,
	}

	// Link back to entry point
	doc["hub3:entrypoint"] = map[string]any{
		"@id":               baseURL,
		"@type":             "hydra:Link",
		"hydra:title":       "API Entry Point",
		"hydra:description": "Lightweight API entry point with navigation links",
	}

	// Add supported classes (resource types) with full details
	supportedClasses := s.buildSupportedClasses(r)
	if len(supportedClasses) > 0 {
		doc["hydra:supportedClass"] = supportedClasses
	}

	// Add supported operations
	doc["hub3:supportedOperations"] = s.buildSupportedOperations(r)

	// Add detailed examples
	doc["hub3:examples"] = s.buildExamples(r)

	// Add code-generated capabilities
	doc["hub3:capabilities"] = buildCapabilities()

	// Add introspection link
	doc["hub3:introspection"] = map[string]any{
		"@id":               baseURL + "/introspect",
		"@type":             "hydra:Link",
		"hydra:title":       "Data Introspection",
		"hydra:description": "Discover classes, properties, and paths in the indexed data",
	}

	return doc
}

// buildSupportedClasses builds the list of supported resource types.
func (s *Service) buildSupportedClasses(r *http.Request) []map[string]any {
	types := s.registry.ListResourceTypes()
	classes := make([]map[string]any, 0, len(types))

	for _, resourceType := range types {
		config, ok := s.registry.Get(resourceType)
		if !ok {
			continue
		}

		class := map[string]any{
			"@id":                resourceType,
			"@type":              "hydra:Class",
			"hydra:title":        config.Label,
			"hydra:description":  config.Description,
			"hub3:searchEndpoint": fmt.Sprintf("%s%s/type/%s/search", getScheme(r), r.Host, resourceType),
			"hub3:docsEndpoint":   fmt.Sprintf("%s%s/type/%s/docs", getScheme(r), r.Host, resourceType),
		}

		// Add supported properties (filterable fields)
		supportedProperties := make([]map[string]any, 0, len(config.Filters))
		for _, filter := range config.Filters {
			if !filter.AllowedInQuery {
				continue
			}

			prop := map[string]any{
				"@type":             "hydra:SupportedProperty",
				"hydra:property":    filter.PropertyURI,
				"hydra:title":       filter.Label,
				"hydra:description": filter.Description,
				"hydra:required":    false,
				"hub3:filterable":   true,
			}

			// Add allowed operators
			operators := make([]string, len(filter.AllowedOperators))
			for i, op := range filter.AllowedOperators {
				operators[i] = string(op)
			}
			prop["hub3:allowedOperators"] = operators
			prop["hub3:valueType"] = filter.ValueType

			if len(filter.Examples) > 0 {
				prop["hub3:examples"] = filter.Examples
			}

			supportedProperties = append(supportedProperties, prop)
		}

		if len(supportedProperties) > 0 {
			class["hydra:supportedProperty"] = supportedProperties
		}

		classes = append(classes, class)
	}

	return classes
}

// buildSupportedOperations builds the list of supported operations.
func (s *Service) buildSupportedOperations(r *http.Request) []map[string]any {
	baseURL := fmt.Sprintf("%s://%s%s", getScheme(r), r.Host, s.baseURL)

	return []map[string]any{
		{
			"@type":             "hydra:Operation",
			"hydra:method":      "GET",
			"hydra:title":       "List Resources",
			"hydra:description": "Search and list resources with optional filters and facets",
			"hydra:expects":     nil,
			"hydra:returns":     "hydra:Collection",
			"hub3:endpoint":     baseURL + "/search",
		},
		{
			"@type":             "hydra:Operation",
			"hydra:method":      "POST",
			"hydra:title":       "Advanced Search",
			"hydra:description": "Execute complex searches using structured JSON-LD queries",
			"hydra:expects":     "hub3:SearchQuery",
			"hydra:returns":     "hydra:Collection",
			"hub3:endpoint":     baseURL + "/search",
		},
		{
			"@type":             "hydra:Operation",
			"hydra:method":      "GET",
			"hydra:title":       "Get Resource",
			"hydra:description": "Retrieve a single resource by ID",
			"hydra:expects":     nil,
			"hydra:returns":     "schema:Thing",
			"hub3:endpoint":     baseURL + "/resource/{id}",
		},
	}
}

// buildExamples builds example queries.
func (s *Service) buildExamples(r *http.Request) []map[string]any {
	baseURL := fmt.Sprintf("%s://%s%s/search", getScheme(r), r.Host, s.baseURL)

	return []map[string]any{
		{
			"@type":             "hub3:Example",
			"hydra:title":       "Simple Text Search",
			"hydra:description": "Search for 'rembrandt' across all fields",
			"hub3:url":          baseURL + "?query=rembrandt",
			"hub3:method":       "GET",
		},
		{
			"@type":             "hub3:Example",
			"hydra:title":       "Filter by Creator",
			"hydra:description": "Find works by Rembrandt van Rijn",
			"hub3:url":          baseURL + "?filter[creator][eq]=Rembrandt van Rijn",
			"hub3:method":       "GET",
		},
		{
			"@type":             "hub3:Example",
			"hydra:title":       "Faceted Search",
			"hydra:description": "Search with type and creator facets",
			"hub3:url":          baseURL + "?query=painting&facet[type]=10&facet[creator]=20",
			"hub3:method":       "GET",
		},
		{
			"@type":             "hub3:Example",
			"hydra:title":       "Geospatial Search",
			"hydra:description": "Find resources within Amsterdam bounding box",
			"hub3:url":          baseURL + "?filter[spatialCoverage][bbox]=4.8,52.3,4.9,52.4",
			"hub3:method":       "GET",
		},
		{
			"@type":             "hub3:Example",
			"hydra:title":       "Advanced POST Search",
			"hydra:description": "Complex search with multiple filters using JSON-LD",
			"hub3:url":          baseURL,
			"hub3:method":       "POST",
			"hub3:body": map[string]any{
				"@context": "https://hub3.delving.org/contexts/search",
				"@type":    "hub3:SearchQuery",
				"query": map[string]any{
					"value":  "night watch",
					"fields": []string{"title", "description"},
					"fuzzy":  true,
				},
				"filters": []map[string]any{
					{
						"@type":    "hub3:PropertyFilter",
						"field":    "creator",
						"operator": "eq",
						"value":    "Rembrandt van Rijn",
					},
					{
						"@type":    "hub3:PropertyFilter",
						"field":    "type",
						"operator": "in",
						"value":    []string{"Painting", "Drawing"},
					},
				},
				"pagination": map[string]any{
					"page": 1,
					"size": 50,
				},
			},
		},
	}
}

// buildTypeDocumentation builds documentation for a specific resource type.
func (s *Service) buildTypeDocumentation(r *http.Request, config *semantic.ResourceConfig) map[string]any {
	baseURL := fmt.Sprintf("%s://%s%s/type/%s", getScheme(r), r.Host, s.baseURL, config.ResourceType)

	doc := map[string]any{
		"@context":          semantic.DefaultContext(),
		"@id":               baseURL + "/docs",
		"@type":             "hydra:ApiDocumentation",
		"hydra:title":       config.Label + " API",
		"hydra:description": config.Description,
		"hub3:resourceType": config.ResourceType,
	}

	// Add search template
	template := semantic.BuildTemplateFromConfigs(baseURL+"/search", config.Filters, config.Facets)
	doc["hydra:search"] = template

	// Add filterable fields
	filters := make([]map[string]any, 0, len(config.Filters))
	for _, filter := range config.Filters {
		f := map[string]any{
			"@type":             "hub3:FilterDefinition",
			"hydra:property":    filter.PropertyURI,
			"hub3:field":        filter.Field,
			"hydra:title":       filter.Label,
			"hydra:description": filter.Description,
			"hub3:valueType":    filter.ValueType,
			"hub3:multiValue":   filter.MultiValue,
		}

		// Add allowed operators
		operators := make([]map[string]any, len(filter.AllowedOperators))
		for i, op := range filter.AllowedOperators {
			operators[i] = map[string]any{
				"operator":    string(op),
				"description": op.Description(),
			}
		}
		f["hub3:allowedOperators"] = operators

		if len(filter.Examples) > 0 {
			f["hub3:examples"] = filter.Examples
		}

		filters = append(filters, f)
	}
	doc["hub3:filters"] = filters

	// Add facets
	facets := make([]map[string]any, 0, len(config.Facets))
	for _, facet := range config.Facets {
		f := map[string]any{
			"@type":             "hub3:FacetDefinition",
			"hydra:property":    facet.PropertyURI,
			"hub3:field":        facet.Field,
			"hydra:title":       facet.Label,
			"hydra:description": facet.Description,
			"hub3:facetType":    facet.FacetType,
			"hub3:defaultSize":  facet.Size,
			"hub3:sortBy":       facet.SortBy,
			"hub3:sortOrder":    facet.SortOrder,
		}

		// Add ranges for range facets
		if facet.FacetType == "range" && len(facet.Ranges) > 0 {
			ranges := make([]map[string]any, len(facet.Ranges))
			for i, r := range facet.Ranges {
				ranges[i] = map[string]any{
					"key":   r.Key,
					"from":  r.From,
					"to":    r.To,
					"label": r.Label,
				}
			}
			f["hub3:ranges"] = ranges
		}

		facets = append(facets, f)
	}
	doc["hub3:facets"] = facets

	// Add sort fields
	sortFields := make([]map[string]any, 0, len(config.SortFields))
	for _, sort := range config.SortFields {
		sortFields = append(sortFields, map[string]any{
			"@type":            "hub3:SortField",
			"hub3:field":       sort.Field,
			"hydra:property":   sort.PropertyURI,
			"hydra:title":      sort.Label,
			"hub3:defaultDir":  sort.DefaultDir,
		})
	}
	if len(sortFields) > 0 {
		doc["hub3:sortFields"] = sortFields
	}

	// Add configuration
	doc["hub3:configuration"] = map[string]any{
		"defaultSize":   config.DefaultSize,
		"maxSize":       config.MaxSize,
		"defaultSort":   config.DefaultSort,
		"allowFullText": config.AllowFullText,
	}

	return doc
}

// buildCapabilities returns a map of API capabilities generated from code constants.
// This ensures the documentation always reflects the actual system capabilities.
func buildCapabilities() map[string]any {
	return map[string]any{
		"operators":       supportedOperators(),
		"facetTypes":      supportedFacetTypes(),
		"sortOptions":     []string{"asc", "desc"},
		"maxPageSize":     semantic.DefaultStoreCapabilities().MaxPageSize,
		"defaultPageSize": semantic.DefaultPagination().Size,
		"pathSeparator":   "/",
		"contextTTL":      "15m",
	}
}

// supportedOperators returns the list of supported filter operators.
// This is derived from the domain operator constants to ensure consistency.
func supportedOperators() []string {
	allOps := semantic.AllOperators()
	ops := make([]string, len(allOps))

	for i, op := range allOps {
		ops[i] = op.String()
	}

	return ops
}

// supportedFacetTypes returns the list of supported facet types.
// Update this list when new facet types are added to the system.
func supportedFacetTypes() []string {
	return []string{"enum", "range", "date", "boolean"}
}
