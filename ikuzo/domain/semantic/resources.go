package semantic

// This file contains example resource configurations for common schema.org types.
// These demonstrate how to configure filters and facets for different resource types.

// CreativeWorkConfig returns a configuration for schema:CreativeWork resources.
// This includes books, articles, paintings, photographs, etc.
func CreativeWorkConfig() *ResourceConfig {
	config := NewResourceConfig("schema:CreativeWork", "Creative Work").
		WithDescription("Cultural heritage objects, artworks, and creative content").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("+title")

	// Text filters
	config.AddFilter(
		NewFilterConfig("title", "schema:name", "Title").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by title or name").
			WithMultiValue(false).
			WithExamples("The Night Watch", "Rembrandt"),
	)

	config.AddFilter(
		NewFilterConfig("description", "schema:description", "Description").
			WithOperators(OpContains).
			WithDescription("Filter by description content").
			WithQueryAccess(false), // Only allow in POST for complex queries
	)

	config.AddFilter(
		NewFilterConfig("creator", "schema:creator", "Creator").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by creator/artist name").
			WithMultiValue(true).
			WithExamples("Rembrandt van Rijn", "Johannes Vermeer"),
	)

	config.AddFilter(
		NewFilterConfig("contributor", "schema:contributor", "Contributor").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by contributor").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("publisher", "schema:publisher", "Publisher").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by publisher or institution").
			WithMultiValue(true),
	)

	// Date filters
	config.AddFilter(
		NewFilterConfig("dateCreated", "schema:dateCreated", "Date Created").
			WithOperators(OpEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual).
			WithDescription("Filter by creation date").
			WithValueType("date").
			WithExamples("1642", "1600-1700"),
	)

	config.AddFilter(
		NewFilterConfig("dateModified", "schema:dateModified", "Date Modified").
			WithOperators(OpGreaterThan, OpGreaterEqual).
			WithDescription("Filter by last modification date").
			WithValueType("date"),
	)

	// Type and categorization
	config.AddFilter(
		NewFilterConfig("type", "rdf:type", "Type").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by specific type").
			WithMultiValue(true).
			WithExamples("schema:Painting", "schema:Photograph"),
	)

	config.AddFilter(
		NewFilterConfig("genre", "schema:genre", "Genre").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by genre").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("keywords", "schema:keywords", "Keywords").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by keywords or tags").
			WithMultiValue(true),
	)

	// Geospatial filters
	config.AddFilter(
		NewFilterConfig("spatialCoverage", "schema:spatialCoverage", "Location").
			WithOperators(OpBBox, OpWithin).
			WithDescription("Filter by geographic coverage").
			WithValueType("geo"),
	)

	// Rights and access
	config.AddFilter(
		NewFilterConfig("license", "schema:license", "License").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by license type").
			WithMultiValue(true).
			WithExamples("CC-BY", "CC0", "Public Domain"),
	)

	config.AddFilter(
		NewFilterConfig("isAccessibleForFree", "schema:isAccessibleForFree", "Free Access").
			WithOperators(OpEqual).
			WithDescription("Filter by free access").
			WithValueType("boolean"),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("type", "rdf:type", "Type", "enum").
			WithDescription("Facet by resource type").
			WithSize(20).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("creator", "schema:creator", "Creator", "enum").
			WithDescription("Facet by creator/artist").
			WithSize(50).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("publisher", "schema:publisher", "Publisher", "enum").
			WithDescription("Facet by publisher or institution").
			WithSize(30),
	)

	config.AddFacet(
		NewFacetConfig("genre", "schema:genre", "Genre", "enum").
			WithDescription("Facet by genre").
			WithSize(20),
	)

	config.AddFacet(
		NewFacetConfig("keywords", "schema:keywords", "Keywords", "enum").
			WithDescription("Facet by keywords").
			WithSize(50),
	)

	// Date range facet
	config.AddFacet(
		NewFacetConfig("dateCreated", "schema:dateCreated", "Date Created", "range").
			WithDescription("Facet by creation date ranges").
			WithRanges(
				FacetRange{Key: "before-1500", To: 1500, Label: "Before 1500"},
				FacetRange{Key: "1500-1600", From: 1500, To: 1600, Label: "1500-1600"},
				FacetRange{Key: "1600-1700", From: 1600, To: 1700, Label: "1600-1700"},
				FacetRange{Key: "1700-1800", From: 1700, To: 1800, Label: "1700-1800"},
				FacetRange{Key: "1800-1900", From: 1800, To: 1900, Label: "1800-1900"},
				FacetRange{Key: "1900-2000", From: 1900, To: 2000, Label: "1900-2000"},
				FacetRange{Key: "2000-present", From: 2000, Label: "2000-present"},
			),
	)

	config.AddFacet(
		NewFacetConfig("license", "schema:license", "License", "enum").
			WithDescription("Facet by license type").
			WithSize(10),
	)

	// Sort fields
	config.AddSortField("title", "schema:name", "Title", "asc")
	config.AddSortField("dateCreated", "schema:dateCreated", "Date Created", "desc")
	config.AddSortField("dateModified", "schema:dateModified", "Date Modified", "desc")
	config.AddSortField("relevance", "hub3:score", "Relevance", "desc")

	return config
}

// PersonConfig returns a configuration for schema:Person resources.
func PersonConfig() *ResourceConfig {
	config := NewResourceConfig("schema:Person", "Person").
		WithDescription("People, artists, authors, and historical figures").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("+name")

	// Name filters
	config.AddFilter(
		NewFilterConfig("name", "schema:name", "Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by person name").
			WithExamples("Rembrandt van Rijn", "Vincent van Gogh"),
	)

	config.AddFilter(
		NewFilterConfig("givenName", "schema:givenName", "Given Name").
			WithOperators(OpEqual, OpContains).
			WithDescription("Filter by given name"),
	)

	config.AddFilter(
		NewFilterConfig("familyName", "schema:familyName", "Family Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by family name"),
	)

	// Dates
	config.AddFilter(
		NewFilterConfig("birthDate", "schema:birthDate", "Birth Date").
			WithOperators(OpEqual, OpGreaterThan, OpLessThan).
			WithDescription("Filter by birth date").
			WithValueType("date").
			WithExamples("1606-07-15", "1600"),
	)

	config.AddFilter(
		NewFilterConfig("deathDate", "schema:deathDate", "Death Date").
			WithOperators(OpEqual, OpGreaterThan, OpLessThan).
			WithDescription("Filter by death date").
			WithValueType("date"),
	)

	// Location
	config.AddFilter(
		NewFilterConfig("birthPlace", "schema:birthPlace", "Birth Place").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by birth place").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("deathPlace", "schema:deathPlace", "Death Place").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by death place").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("nationality", "schema:nationality", "Nationality").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by nationality").
			WithMultiValue(true).
			WithExamples("Dutch", "Flemish", "Italian"),
	)

	// Occupation and roles
	config.AddFilter(
		NewFilterConfig("hasOccupation", "schema:hasOccupation", "Occupation").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by occupation").
			WithMultiValue(true).
			WithExamples("Painter", "Sculptor", "Architect"),
	)

	// Gender
	config.AddFilter(
		NewFilterConfig("gender", "schema:gender", "Gender").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by gender").
			WithValueType("string"),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("hasOccupation", "schema:hasOccupation", "Occupation", "enum").
			WithDescription("Facet by occupation").
			WithSize(30),
	)

	config.AddFacet(
		NewFacetConfig("nationality", "schema:nationality", "Nationality", "enum").
			WithDescription("Facet by nationality").
			WithSize(50),
	)

	config.AddFacet(
		NewFacetConfig("birthPlace", "schema:birthPlace", "Birth Place", "enum").
			WithDescription("Facet by birth place").
			WithSize(50),
	)

	config.AddFacet(
		NewFacetConfig("gender", "schema:gender", "Gender", "enum").
			WithDescription("Facet by gender").
			WithSize(5),
	)

	// Birth year range facet
	config.AddFacet(
		NewFacetConfig("birthDate", "schema:birthDate", "Birth Period", "range").
			WithDescription("Facet by birth period").
			WithRanges(
				FacetRange{Key: "before-1400", To: 1400, Label: "Before 1400"},
				FacetRange{Key: "1400-1500", From: 1400, To: 1500, Label: "1400-1500"},
				FacetRange{Key: "1500-1600", From: 1500, To: 1600, Label: "1500-1600"},
				FacetRange{Key: "1600-1700", From: 1600, To: 1700, Label: "1600-1700"},
				FacetRange{Key: "1700-1800", From: 1700, To: 1800, Label: "1700-1800"},
				FacetRange{Key: "1800-1900", From: 1800, To: 1900, Label: "1800-1900"},
				FacetRange{Key: "1900-present", From: 1900, Label: "1900-present"},
			),
	)

	// Sort fields
	config.AddSortField("name", "schema:name", "Name", "asc")
	config.AddSortField("familyName", "schema:familyName", "Family Name", "asc")
	config.AddSortField("birthDate", "schema:birthDate", "Birth Date", "asc")
	config.AddSortField("deathDate", "schema:deathDate", "Death Date", "asc")

	return config
}

// PlaceConfig returns a configuration for schema:Place resources.
func PlaceConfig() *ResourceConfig {
	config := NewResourceConfig("schema:Place", "Place").
		WithDescription("Geographic locations, cities, landmarks, and regions").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("+name")

	// Name filters
	config.AddFilter(
		NewFilterConfig("name", "schema:name", "Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by place name").
			WithExamples("Amsterdam", "Paris", "Florence"),
	)

	config.AddFilter(
		NewFilterConfig("alternateName", "schema:alternateName", "Alternate Name").
			WithOperators(OpEqual, OpContains).
			WithDescription("Filter by alternate or historical names").
			WithMultiValue(true),
	)

	// Type
	config.AddFilter(
		NewFilterConfig("type", "rdf:type", "Type").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by place type").
			WithMultiValue(true).
			WithExamples("schema:City", "schema:Museum", "schema:LandmarksOrHistoricalBuildings"),
	)

	// Geographic containment
	config.AddFilter(
		NewFilterConfig("containedInPlace", "schema:containedInPlace", "Contained In").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by containing place (e.g., country, region)").
			WithMultiValue(true).
			WithExamples("Netherlands", "Italy", "France"),
	)

	// Address components
	config.AddFilter(
		NewFilterConfig("addressCountry", "schema:addressCountry", "Country").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by country").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("addressRegion", "schema:addressRegion", "Region").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by region or state").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("addressLocality", "schema:addressLocality", "Locality").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by city or locality").
			WithMultiValue(true),
	)

	// Geospatial filters
	config.AddFilter(
		NewFilterConfig("geo", "schema:geo", "Geographic Coordinates").
			WithOperators(OpBBox, OpWithin, OpIntersects).
			WithDescription("Filter by geographic location").
			WithValueType("geo"),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("type", "rdf:type", "Type", "enum").
			WithDescription("Facet by place type").
			WithSize(20),
	)

	config.AddFacet(
		NewFacetConfig("addressCountry", "schema:addressCountry", "Country", "enum").
			WithDescription("Facet by country").
			WithSize(50).
			WithSort("term", "asc"),
	)

	config.AddFacet(
		NewFacetConfig("addressRegion", "schema:addressRegion", "Region", "enum").
			WithDescription("Facet by region").
			WithSize(50),
	)

	config.AddFacet(
		NewFacetConfig("containedInPlace", "schema:containedInPlace", "Contained In", "enum").
			WithDescription("Facet by containing place").
			WithSize(30),
	)

	// Sort fields
	config.AddSortField("name", "schema:name", "Name", "asc")
	config.AddSortField("addressCountry", "schema:addressCountry", "Country", "asc")
	config.AddSortField("addressLocality", "schema:addressLocality", "Locality", "asc")

	return config
}

// DefaultRegistry creates a registry with all standard resource configurations.
func DefaultRegistry() *ConfigRegistry {
	registry := NewConfigRegistry()
	registry.Register(CreativeWorkConfig())
	registry.Register(PersonConfig())
	registry.Register(PlaceConfig())
	return registry
}
