package semantic

// This file contains EDM (Europeana Data Model) resource configurations.
// These demonstrate how to configure filters and facets for EDM-based resources.

// AggregationConfig returns a configuration for ore:Aggregation resources.
// This is the main entry point in EDM that aggregates all information about a cultural heritage object.
func AggregationConfig() *ResourceConfig {
	config := NewResourceConfig("ore:Aggregation", "Aggregation").
		WithDescription("EDM Aggregations that bundle cultural heritage objects with their digital representations and metadata").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("_score")

	// Core aggregation relationships
	config.AddFilter(
		NewFilterConfig("edm.aggregatedCHO", "edm:aggregatedCHO", "Aggregated CHO").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by aggregated cultural heritage object URI").
			WithMultiValue(true),
	)

	// Provider information
	config.AddFilter(
		NewFilterConfig("edm.dataProvider", "edm:dataProvider", "Data Provider").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by data provider").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("edm.provider", "edm:provider", "Provider").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by aggregator/provider").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("edm.intermediateProvider", "edm:intermediateProvider", "Intermediate Provider").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by intermediate provider").
			WithMultiValue(true),
	)

	// Rights
	config.AddFilter(
		NewFilterConfig("edm.rights", "edm:rights", "Rights").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by rights statement URI").
			WithMultiValue(true).
			WithExamples("http://creativecommons.org/publicdomain/mark/1.0/"),
	)

	config.AddFilter(
		NewFilterConfig("dc.rights", "dc:rights", "DC Rights").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by rights description").
			WithMultiValue(true),
	)

	// Digital objects
	config.AddFilter(
		NewFilterConfig("edm.isShownAt", "edm:isShownAt", "Shown At").
			WithOperators(OpEqual, OpExists).
			WithDescription("Filter by landing page URL").
			WithValueType("uri"),
	)

	config.AddFilter(
		NewFilterConfig("edm.isShownBy", "edm:isShownBy", "Shown By").
			WithOperators(OpEqual, OpExists).
			WithDescription("Filter by direct access URL").
			WithValueType("uri"),
	)

	config.AddFilter(
		NewFilterConfig("edm.object", "edm:object", "Object").
			WithOperators(OpEqual, OpExists).
			WithDescription("Filter by thumbnail/preview URL").
			WithValueType("uri"),
	)

	config.AddFilter(
		NewFilterConfig("edm.hasView", "edm:hasView", "Has View").
			WithOperators(OpEqual, OpIn, OpExists).
			WithDescription("Filter by additional view URLs").
			WithMultiValue(true).
			WithValueType("uri"),
	)

	// Content type and classification
	config.AddFilter(
		NewFilterConfig("edm.ugc", "edm:ugc", "User Generated Content").
			WithOperators(OpEqual).
			WithDescription("Filter by user-generated content flag").
			WithValueType("boolean"),
	)

	// Inherit common DC fields from CHO via aggregatedCHO relationship
	config.AddFilter(
		NewFilterConfig("dc.title", "dc:title", "Title").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by title (from aggregated CHO)").
			WithMultiValue(false),
	)

	config.AddFilter(
		NewFilterConfig("dc.creator", "dc:creator", "Creator").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by creator (from aggregated CHO)").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("dc.type", "dc:type", "Type").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by type (from aggregated CHO)").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("dc.subject", "dc:subject", "Subject").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by subject (from aggregated CHO)").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("dc.date", "dc:date", "Date").
			WithOperators(OpEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual).
			WithDescription("Filter by date (from aggregated CHO)").
			WithValueType("date"),
	)

	config.AddFilter(
		NewFilterConfig("edm.datasetName", "edm:datasetName", "Dataset").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by dataset or collection").
			WithMultiValue(true),
	)

	// Facets - Provider information
	config.AddFacet(
		NewFacetConfig("edm.dataProvider", "edm:dataProvider", "Data Provider", "enum").
			WithDescription("Facet by data provider").
			WithSize(50).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("edm.provider", "edm:provider", "Provider", "enum").
			WithDescription("Facet by provider").
			WithSize(30),
	)

	config.AddFacet(
		NewFacetConfig("edm.datasetName", "edm:datasetName", "Dataset", "enum").
			WithDescription("Facet by dataset or collection").
			WithSize(30).
			WithSort("count", "desc"),
	)

	// Facets - Rights
	config.AddFacet(
		NewFacetConfig("edm.rights", "edm:rights", "Rights", "enum").
			WithDescription("Facet by rights statement").
			WithSize(15),
	)

	config.AddFacet(
		NewFacetConfig("edm.ugc", "edm:ugc", "User Generated", "enum").
			WithDescription("Facet by user-generated content").
			WithSize(2),
	)

	// Facets - Content from aggregated CHO
	config.AddFacet(
		NewFacetConfig("dc.type", "dc:type", "Type", "enum").
			WithDescription("Facet by object type").
			WithSize(20).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("dc.creator", "dc:creator", "Creator", "enum").
			WithDescription("Facet by creator").
			WithSize(50).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("dc.subject", "dc:subject", "Subject", "enum").
			WithDescription("Facet by subject").
			WithSize(50).
			WithSort("count", "desc"),
	)

	// Date range facet
	config.AddFacet(
		NewFacetConfig("dc.date", "dc:date", "Date", "range").
			WithDescription("Facet by date ranges").
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

	// Sort fields
	config.AddSortField("dc.title", "dc:title", "Title", "asc")
	config.AddSortField("dc.date", "dc:date", "Date", "desc")
	config.AddSortField("edm.dataProvider", "edm:dataProvider", "Data Provider", "asc")
	config.AddSortField("_score", "hub3:score", "Relevance", "desc")

	return config
}

// CHOConfig returns a configuration for edm:ProvidedCHO (Cultural Heritage Object) resources.
// This represents the cultural heritage object being described in the aggregation.
func CHOConfig() *ResourceConfig {
	config := NewResourceConfig("edm:ProvidedCHO", "Cultural Heritage Object").
		WithDescription("Cultural heritage objects following the Europeana Data Model").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("_score")

	// Text filters - Dublin Core elements
	config.AddFilter(
		NewFilterConfig("dc.title", "dc:title", "Title").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by object title").
			WithMultiValue(false).
			WithExamples("The Night Watch", "Rembrandt"),
	)

	config.AddFilter(
		NewFilterConfig("dc.description", "dc:description", "Description").
			WithOperators(OpContains).
			WithDescription("Filter by description content").
			WithQueryAccess(false), // Only allow in POST for complex queries
	)

	config.AddFilter(
		NewFilterConfig("dc.creator", "dc:creator", "Creator").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by creator/artist name").
			WithMultiValue(true).
			WithExamples("Rembrandt van Rijn", "Johannes Vermeer"),
	)

	config.AddFilter(
		NewFilterConfig("dc.contributor", "dc:contributor", "Contributor").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by contributor").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("dc.publisher", "dc:publisher", "Publisher").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by publisher or institution").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("dc.subject", "dc:subject", "Subject").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by subject").
			WithMultiValue(true),
	)

	// Date filters
	config.AddFilter(
		NewFilterConfig("dc.date", "dc:date", "Date").
			WithOperators(OpEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual).
			WithDescription("Filter by date").
			WithValueType("date").
			WithExamples("1642", "1600-1700"),
	)

	config.AddFilter(
		NewFilterConfig("dcterms.created", "dcterms:created", "Date Created").
			WithOperators(OpEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual).
			WithDescription("Filter by creation date").
			WithValueType("date"),
	)

	config.AddFilter(
		NewFilterConfig("dcterms.issued", "dcterms:issued", "Date Issued").
			WithOperators(OpEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual).
			WithDescription("Filter by issue date").
			WithValueType("date"),
	)

	// Type and categorization
	config.AddFilter(
		NewFilterConfig("dc.type", "dc:type", "Type").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by object type").
			WithMultiValue(true).
			WithExamples("Painting", "Photograph", "Manuscript"),
	)

	config.AddFilter(
		NewFilterConfig("dc.format", "dc:format", "Format").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by format").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("edm.type", "edm:type", "EDM Type").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by EDM type").
			WithMultiValue(true).
			WithExamples("IMAGE", "TEXT", "VIDEO", "SOUND", "3D"),
	)

	// Language
	config.AddFilter(
		NewFilterConfig("dc.language", "dc:language", "Language").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by language").
			WithMultiValue(true).
			WithExamples("nl", "en", "de", "fr"),
	)

	// Geospatial filters
	config.AddFilter(
		NewFilterConfig("dcterms.spatial", "dcterms:spatial", "Location").
			WithOperators(OpEqual, OpIn, OpContains).
			WithDescription("Filter by geographic location").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("geo", "geo:lat_long", "Geographic Coordinates").
			WithOperators(OpBBox, OpWithin).
			WithDescription("Filter by geographic coordinates").
			WithValueType("geo"),
	)

	// Rights and access - EDM specific
	config.AddFilter(
		NewFilterConfig("dc.rights", "dc:rights", "Rights").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by rights statement").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("edm.rights", "edm:rights", "EDM Rights").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by EDM rights").
			WithMultiValue(true).
			WithExamples("http://creativecommons.org/publicdomain/mark/1.0/"),
	)

	// Collection/Dataset
	config.AddFilter(
		NewFilterConfig("edm.datasetName", "edm:datasetName", "Dataset").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by dataset or collection").
			WithMultiValue(true),
	)

	// Provider information
	config.AddFilter(
		NewFilterConfig("edm.dataProvider", "edm:dataProvider", "Data Provider").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by data provider").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("edm.provider", "edm:provider", "Provider").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by provider").
			WithMultiValue(true),
	)

	// Facets - Dublin Core
	config.AddFacet(
		NewFacetConfig("dc.type", "dc:type", "Type", "enum").
			WithDescription("Facet by object type").
			WithSize(20).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("dc.creator", "dc:creator", "Creator", "enum").
			WithDescription("Facet by creator/artist").
			WithSize(50).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("dc.subject", "dc:subject", "Subject", "enum").
			WithDescription("Facet by subject").
			WithSize(50).
			WithSort("count", "desc"),
	)

	config.AddFacet(
		NewFacetConfig("dc.publisher", "dc:publisher", "Publisher", "enum").
			WithDescription("Facet by publisher or institution").
			WithSize(30),
	)

	config.AddFacet(
		NewFacetConfig("dc.language", "dc:language", "Language", "enum").
			WithDescription("Facet by language").
			WithSize(20),
	)

	// Facets - EDM specific
	config.AddFacet(
		NewFacetConfig("edm.type", "edm:type", "EDM Type", "enum").
			WithDescription("Facet by EDM type").
			WithSize(10),
	)

	config.AddFacet(
		NewFacetConfig("edm.dataProvider", "edm:dataProvider", "Data Provider", "enum").
			WithDescription("Facet by data provider").
			WithSize(30),
	)

	config.AddFacet(
		NewFacetConfig("edm.datasetName", "edm:datasetName", "Dataset", "enum").
			WithDescription("Facet by dataset or collection").
			WithSize(20),
	)

	config.AddFacet(
		NewFacetConfig("edm.rights", "edm:rights", "Rights", "enum").
			WithDescription("Facet by rights statement").
			WithSize(15),
	)

	// Date range facet
	config.AddFacet(
		NewFacetConfig("dc.date", "dc:date", "Date", "range").
			WithDescription("Facet by date ranges").
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

	// Sort fields
	config.AddSortField("dc.title", "dc:title", "Title", "asc")
	config.AddSortField("dc.date", "dc:date", "Date", "desc")
	config.AddSortField("dcterms.created", "dcterms:created", "Date Created", "desc")
	config.AddSortField("_score", "hub3:score", "Relevance", "desc")

	return config
}

// AgentConfig returns a configuration for edm:Agent resources.
// Agents represent people, organizations, or groups.
func AgentConfig() *ResourceConfig {
	config := NewResourceConfig("edm:Agent", "Agent").
		WithDescription("People, organizations, and groups in the Europeana Data Model").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("_score")

	// Name filters
	config.AddFilter(
		NewFilterConfig("skos.prefLabel", "skos:prefLabel", "Preferred Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by preferred name").
			WithExamples("Rembrandt van Rijn", "Rijksmuseum"),
	)

	config.AddFilter(
		NewFilterConfig("skos.altLabel", "skos:altLabel", "Alternative Name").
			WithOperators(OpEqual, OpContains).
			WithDescription("Filter by alternative name").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("foaf.name", "foaf:name", "Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by name"),
	)

	// Dates
	config.AddFilter(
		NewFilterConfig("edm.begin", "edm:begin", "Begin Date").
			WithOperators(OpEqual, OpGreaterThan, OpLessThan).
			WithDescription("Filter by begin date (birth, founding)").
			WithValueType("date"),
	)

	config.AddFilter(
		NewFilterConfig("edm.end", "edm:end", "End Date").
			WithOperators(OpEqual, OpGreaterThan, OpLessThan).
			WithDescription("Filter by end date (death, dissolution)").
			WithValueType("date"),
	)

	// Type
	config.AddFilter(
		NewFilterConfig("edm.hasMet", "edm:hasMet", "Associated With").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by associations").
			WithMultiValue(true),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("skos.prefLabel", "skos:prefLabel", "Agent Name", "enum").
			WithDescription("Facet by agent name").
			WithSize(50),
	)

	// Birth/begin date range facet
	config.AddFacet(
		NewFacetConfig("edm.begin", "edm:begin", "Begin Period", "range").
			WithDescription("Facet by begin period").
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
	config.AddSortField("skos.prefLabel", "skos:prefLabel", "Name", "asc")
	config.AddSortField("edm.begin", "edm:begin", "Begin Date", "asc")
	config.AddSortField("edm.end", "edm:end", "End Date", "asc")

	return config
}

// PlaceConfig returns a configuration for edm:Place resources.
func PlaceConfig() *ResourceConfig {
	config := NewResourceConfig("edm:Place", "Place").
		WithDescription("Geographic locations in the Europeana Data Model").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("_score")

	// Name filters
	config.AddFilter(
		NewFilterConfig("skos.prefLabel", "skos:prefLabel", "Preferred Name").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by place name").
			WithExamples("Amsterdam", "Paris", "Rome"),
	)

	config.AddFilter(
		NewFilterConfig("skos.altLabel", "skos:altLabel", "Alternative Name").
			WithOperators(OpEqual, OpContains).
			WithDescription("Filter by alternative or historical names").
			WithMultiValue(true),
	)

	// Geographic containment
	config.AddFilter(
		NewFilterConfig("dcterms.isPartOf", "dcterms:isPartOf", "Part Of").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by containing place (e.g., country, region)").
			WithMultiValue(true),
	)

	// Geospatial filters
	config.AddFilter(
		NewFilterConfig("geo", "geo:lat_long", "Geographic Coordinates").
			WithOperators(OpBBox, OpWithin, OpIntersects).
			WithDescription("Filter by geographic location").
			WithValueType("geo"),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("skos.prefLabel", "skos:prefLabel", "Place Name", "enum").
			WithDescription("Facet by place name").
			WithSize(50),
	)

	config.AddFacet(
		NewFacetConfig("dcterms.isPartOf", "dcterms:isPartOf", "Part Of", "enum").
			WithDescription("Facet by containing place").
			WithSize(30),
	)

	// Sort fields
	config.AddSortField("skos.prefLabel", "skos:prefLabel", "Name", "asc")

	return config
}

// ConceptConfig returns a configuration for skos:Concept resources.
func ConceptConfig() *ResourceConfig {
	config := NewResourceConfig("skos:Concept", "Concept").
		WithDescription("Controlled vocabulary concepts and subject headings").
		WithDefaultSize(20).
		WithMaxSize(100).
		WithDefaultSort("_score")

	// Name filters
	config.AddFilter(
		NewFilterConfig("skos.prefLabel", "skos:prefLabel", "Preferred Label").
			WithOperators(OpEqual, OpContains, OpStartsWith).
			WithDescription("Filter by preferred label").
			WithMultiValue(false),
	)

	config.AddFilter(
		NewFilterConfig("skos.altLabel", "skos:altLabel", "Alternative Label").
			WithOperators(OpEqual, OpContains).
			WithDescription("Filter by alternative labels").
			WithMultiValue(true),
	)

	// Hierarchical relationships
	config.AddFilter(
		NewFilterConfig("skos.broader", "skos:broader", "Broader Concept").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by broader concepts").
			WithMultiValue(true),
	)

	config.AddFilter(
		NewFilterConfig("skos.narrower", "skos:narrower", "Narrower Concept").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by narrower concepts").
			WithMultiValue(true),
	)

	// Scheme
	config.AddFilter(
		NewFilterConfig("skos.inScheme", "skos:inScheme", "In Scheme").
			WithOperators(OpEqual, OpIn).
			WithDescription("Filter by concept scheme").
			WithMultiValue(true),
	)

	// Facets
	config.AddFacet(
		NewFacetConfig("skos.prefLabel", "skos:prefLabel", "Concept", "enum").
			WithDescription("Facet by concept").
			WithSize(50),
	)

	config.AddFacet(
		NewFacetConfig("skos.inScheme", "skos:inScheme", "Scheme", "enum").
			WithDescription("Facet by concept scheme").
			WithSize(20),
	)

	// Sort fields
	config.AddSortField("skos.prefLabel", "skos:prefLabel", "Label", "asc")

	return config
}

// DefaultRegistry creates a registry with all EDM resource configurations.
func DefaultRegistry() *ConfigRegistry {
	registry := NewConfigRegistry()
	registry.Register(AggregationConfig()) // Main entry point for EDM
	registry.Register(CHOConfig())
	registry.Register(AgentConfig())
	registry.Register(PlaceConfig())
	registry.Register(ConceptConfig())
	return registry
}
