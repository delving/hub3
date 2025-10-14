package semantic

import (
	"testing"
)

func TestCHOConfig(t *testing.T) {
	config := CHOConfig()

	if config.ResourceType != "edm:ProvidedCHO" {
		t.Errorf("ResourceType = %v, want edm:ProvidedCHO", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"dc.title", "dc.description", "dc.creator", "dc.contributor", "dc.publisher",
		"dc.subject", "dc.date", "dcterms.created", "dcterms.issued",
		"dc.type", "dc.format", "edm.type", "dc.language",
		"dcterms.spatial", "geo", "dc.rights", "edm.rights",
		"edm.datasetName", "edm.dataProvider", "edm.provider",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that title filter has correct operators
	titleFilter, ok := config.GetFilterConfig("dc.title")
	if !ok {
		t.Fatal("Expected to find dc.title filter")
	}

	expectedOps := []Operator{OpEqual, OpContains, OpStartsWith}
	if len(titleFilter.AllowedOperators) != len(expectedOps) {
		t.Errorf("Title filter has %d operators, want %d", len(titleFilter.AllowedOperators), len(expectedOps))
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"dc.type", "dc.creator", "dc.subject", "dc.publisher", "dc.language",
		"edm.type", "edm.dataProvider", "edm.datasetName", "edm.rights", "dc.date",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}

	// Check that date facet has ranges
	dateFacet, ok := config.GetFacetConfig("dc.date")
	if !ok {
		t.Fatal("Expected to find dc.date facet")
	}

	if dateFacet.FacetType != "range" {
		t.Errorf("dc.date facet type = %v, want range", dateFacet.FacetType)
	}

	if len(dateFacet.Ranges) == 0 {
		t.Error("Expected dc.date facet to have ranges")
	}

	// Check sort fields
	if len(config.SortFields) == 0 {
		t.Error("Expected config to have sort fields")
	}
}

func TestAgentConfig(t *testing.T) {
	config := AgentConfig()

	if config.ResourceType != "edm:Agent" {
		t.Errorf("ResourceType = %v, want edm:Agent", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"skos.prefLabel", "skos.altLabel", "foaf.name",
		"edm.begin", "edm.end", "edm.hasMet",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that prefLabel filter has correct operators
	nameFilter, ok := config.GetFilterConfig("skos.prefLabel")
	if !ok {
		t.Fatal("Expected to find skos.prefLabel filter")
	}

	if !nameFilter.IsOperatorAllowed(OpContains) {
		t.Error("Expected prefLabel filter to allow contains operator")
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"skos.prefLabel", "edm.begin",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}

	// Check that begin date facet has ranges
	beginFacet, ok := config.GetFacetConfig("edm.begin")
	if !ok {
		t.Fatal("Expected to find edm.begin facet")
	}

	if beginFacet.FacetType != "range" {
		t.Errorf("edm.begin facet type = %v, want range", beginFacet.FacetType)
	}

	if len(beginFacet.Ranges) == 0 {
		t.Error("Expected edm.begin facet to have ranges")
	}
}

func TestPlaceConfig(t *testing.T) {
	config := PlaceConfig()

	if config.ResourceType != "edm:Place" {
		t.Errorf("ResourceType = %v, want edm:Place", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"skos.prefLabel", "skos.altLabel",
		"dcterms.isPartOf", "geo", "geo.lat", "geo.long",
		"schema.addressCountry", "schema.addressRegion", "schema.addressLocality",
		"schema.postalCode", "schema.streetAddress",
		"gn.countryCode", "gn.parentCountry", "gn.parentADM1", "gn.parentADM2", "gn.parentADM3",
		"nave.city", "nave.country", "nave.province", "nave.municipality",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that geo filter has geospatial operators
	geoFilter, ok := config.GetFilterConfig("geo")
	if !ok {
		t.Fatal("Expected to find geo filter")
	}

	if !geoFilter.IsOperatorAllowed(OpBBox) {
		t.Error("Expected geo filter to allow bbox operator")
	}

	if !geoFilter.IsOperatorAllowed(OpWithin) {
		t.Error("Expected geo filter to allow within operator")
	}

	// Check that schema.org address fields are present
	addressCountryFilter, ok := config.GetFilterConfig("schema.addressCountry")
	if !ok {
		t.Fatal("Expected to find schema.addressCountry filter")
	}

	if !addressCountryFilter.IsOperatorAllowed(OpEqual) {
		t.Error("Expected schema.addressCountry filter to allow equal operator")
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"skos.prefLabel", "dcterms.isPartOf",
		"schema.addressCountry", "schema.addressRegion", "schema.addressLocality",
		"gn.countryCode",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}
}

func TestConceptConfig(t *testing.T) {
	config := ConceptConfig()

	if config.ResourceType != "skos:Concept" {
		t.Errorf("ResourceType = %v, want skos:Concept", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"skos.prefLabel", "skos.altLabel",
		"skos.broader", "skos.narrower", "skos.inScheme",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"skos.prefLabel", "skos.inScheme",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}
}

func TestAggregationConfig(t *testing.T) {
	config := AggregationConfig()

	if config.ResourceType != "ore:Aggregation" {
		t.Errorf("ResourceType = %v, want ore:Aggregation", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"edm.aggregatedCHO", "edm.dataProvider", "edm.provider", "edm.intermediateProvider",
		"edm.rights", "dc.rights", "edm.isShownAt", "edm.isShownBy", "edm.object",
		"edm.hasView", "edm.ugc", "dc.title", "dc.creator", "dc.type", "dc.subject",
		"dc.date", "edm.datasetName",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"edm.dataProvider", "edm.provider", "edm.datasetName", "edm.rights",
		"edm.ugc", "dc.type", "dc.creator", "dc.subject", "dc.date",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	// Check that all expected resource types are registered
	expectedTypes := []string{
		"ore:Aggregation",
		"edm:ProvidedCHO",
		"edm:Agent",
		"edm:Place",
		"skos:Concept",
	}

	registeredTypes := registry.ListResourceTypes()
	if len(registeredTypes) != len(expectedTypes) {
		t.Errorf("Registry has %d types, want %d", len(registeredTypes), len(expectedTypes))
	}

	for _, resourceType := range expectedTypes {
		config, ok := registry.Get(resourceType)
		if !ok {
			t.Errorf("Expected to find config for %s", resourceType)
		}
		if config == nil {
			t.Errorf("Config for %s is nil", resourceType)
		}
	}

	// Check that default config is available
	defaultConfig := registry.GetDefault()
	if defaultConfig == nil {
		t.Error("Expected to get a default config")
	}
}

func TestResourceConfig_Integration(t *testing.T) {
	// Test a complete workflow with Aggregation config (main EDM entry point)
	config := AggregationConfig()

	// Create valid query options
	validOpts := &QueryOptions{
		Query: &TextQuery{
			Value: "rembrandt",
		},
		Filters: []Filter{
			&PropertyFilter{
				FieldName:    "dc.creator",
				OperatorType: OpEqual,
				Value:        "Rembrandt van Rijn",
			},
			&PropertyFilter{
				FieldName:    "dc.type",
				OperatorType: OpIn,
				Value:        []string{"Painting", "Drawing"},
			},
		},
		Facets: []FacetRequest{
			{Field: "dc.type", Limit: 20},
			{Field: "dc.subject", Limit: 10},
		},
		Pagination: &Pagination{
			Page: 1,
			Size: 20,
		},
	}

	// Validate should pass
	if err := config.ValidateQueryOptions(validOpts); err != nil {
		t.Errorf("ValidateQueryOptions() failed for valid options: %v", err)
	}

	// Create invalid query options (non-existent field)
	invalidOpts := &QueryOptions{
		Filters: []Filter{
			&PropertyFilter{
				FieldName:    "nonexistentField",
				OperatorType: OpEqual,
				Value:        "test",
			},
		},
	}

	// Validate should fail
	if err := config.ValidateQueryOptions(invalidOpts); err == nil {
		t.Error("ValidateQueryOptions() should have failed for invalid field")
	}

	// Create invalid query options (invalid operator)
	invalidOpOpts := &QueryOptions{
		Filters: []Filter{
			&PropertyFilter{
				FieldName:    "dc.title",
				OperatorType: OpGreaterThan, // Not allowed for title
				Value:        "test",
			},
		},
	}

	// Validate should fail
	if err := config.ValidateQueryOptions(invalidOpOpts); err == nil {
		t.Error("ValidateQueryOptions() should have failed for invalid operator")
	}
}

func TestBuildTemplateFromResourceConfigs(t *testing.T) {
	config := AggregationConfig()
	baseURL := "https://example.org/api/semantic/v1/search"

	template := BuildTemplateFromConfigs(baseURL, config.Filters, config.Facets)

	if template == nil {
		t.Fatal("Expected template to be created")
	}

	// Check that template has basic mappings
	hasQuery := false
	for _, m := range template.Mapping {
		if m.Variable == "query" {
			hasQuery = true
			break
		}
	}
	if !hasQuery {
		t.Error("Expected template to have query mapping")
	}

	// Check that template has filter mappings from config
	hasCreatorFilter := false
	for _, m := range template.Mapping {
		if m.Variable == "filter[dc.creator][eq]" {
			hasCreatorFilter = true
			break
		}
	}
	if !hasCreatorFilter {
		t.Error("Expected template to have filter[dc.creator][eq] mapping")
	}

	// Check that template has facet mappings from config
	hasTypeFacet := false
	for _, m := range template.Mapping {
		if m.Variable == "facet[dc.type]" {
			hasTypeFacet = true
			break
		}
	}
	if !hasTypeFacet {
		t.Error("Expected template to have facet[dc.type] mapping")
	}

	// Check that advanced search is configured
	if template.Advanced == nil {
		t.Error("Expected template to have advanced search info")
	}
}
