package semantic

import (
	"testing"
)

func TestCreativeWorkConfig(t *testing.T) {
	config := CreativeWorkConfig()

	if config.ResourceType != "schema:CreativeWork" {
		t.Errorf("ResourceType = %v, want schema:CreativeWork", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"title", "description", "creator", "contributor", "publisher",
		"dateCreated", "dateModified", "type", "genre", "keywords",
		"spatialCoverage", "license", "isAccessibleForFree",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that title filter has correct operators
	titleFilter, ok := config.GetFilterConfig("title")
	if !ok {
		t.Fatal("Expected to find title filter")
	}

	expectedOps := []Operator{OpEqual, OpContains, OpStartsWith}
	if len(titleFilter.AllowedOperators) != len(expectedOps) {
		t.Errorf("Title filter has %d operators, want %d", len(titleFilter.AllowedOperators), len(expectedOps))
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"type", "creator", "publisher", "genre", "keywords", "dateCreated", "license",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}

	// Check that date facet has ranges
	dateFacet, ok := config.GetFacetConfig("dateCreated")
	if !ok {
		t.Fatal("Expected to find dateCreated facet")
	}

	if dateFacet.FacetType != "range" {
		t.Errorf("dateCreated facet type = %v, want range", dateFacet.FacetType)
	}

	if len(dateFacet.Ranges) == 0 {
		t.Error("Expected dateCreated facet to have ranges")
	}

	// Check sort fields
	if len(config.SortFields) == 0 {
		t.Error("Expected config to have sort fields")
	}
}

func TestPersonConfig(t *testing.T) {
	config := PersonConfig()

	if config.ResourceType != "schema:Person" {
		t.Errorf("ResourceType = %v, want schema:Person", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"name", "givenName", "familyName",
		"birthDate", "deathDate",
		"birthPlace", "deathPlace", "nationality",
		"hasOccupation", "gender",
	}

	for _, field := range expectedFilters {
		if !config.IsFieldFilterable(field) {
			t.Errorf("Expected field '%s' to be filterable", field)
		}
	}

	// Check that name filter has correct operators
	nameFilter, ok := config.GetFilterConfig("name")
	if !ok {
		t.Fatal("Expected to find name filter")
	}

	if !nameFilter.IsOperatorAllowed(OpContains) {
		t.Error("Expected name filter to allow contains operator")
	}

	// Check that expected facets exist
	expectedFacets := []string{
		"hasOccupation", "nationality", "birthPlace", "gender", "birthDate",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}

	// Check that birthDate facet has ranges
	birthDateFacet, ok := config.GetFacetConfig("birthDate")
	if !ok {
		t.Fatal("Expected to find birthDate facet")
	}

	if birthDateFacet.FacetType != "range" {
		t.Errorf("birthDate facet type = %v, want range", birthDateFacet.FacetType)
	}

	if len(birthDateFacet.Ranges) == 0 {
		t.Error("Expected birthDate facet to have ranges")
	}
}

func TestPlaceConfig(t *testing.T) {
	config := PlaceConfig()

	if config.ResourceType != "schema:Place" {
		t.Errorf("ResourceType = %v, want schema:Place", config.ResourceType)
	}

	// Check that expected filters exist
	expectedFilters := []string{
		"name", "alternateName", "type",
		"containedInPlace",
		"addressCountry", "addressRegion", "addressLocality",
		"geo",
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

	// Check that expected facets exist
	expectedFacets := []string{
		"type", "addressCountry", "addressRegion", "containedInPlace",
	}

	for _, field := range expectedFacets {
		if !config.IsFieldFacetable(field) {
			t.Errorf("Expected field '%s' to be facetable", field)
		}
	}

	// Check that country facet is sorted by term
	countryFacet, ok := config.GetFacetConfig("addressCountry")
	if !ok {
		t.Fatal("Expected to find addressCountry facet")
	}

	if countryFacet.SortBy != "term" {
		t.Errorf("addressCountry facet sort = %v, want term", countryFacet.SortBy)
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	// Check that all expected resource types are registered
	expectedTypes := []string{
		"schema:CreativeWork",
		"schema:Person",
		"schema:Place",
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
	// Test a complete workflow with CreativeWork config
	config := CreativeWorkConfig()

	// Create valid query options
	validOpts := &QueryOptions{
		Query: &TextQuery{
			Value: "rembrandt",
		},
		Filters: []Filter{
			&PropertyFilter{
				FieldName:    "creator",
				OperatorType: OpEqual,
				Value:        "Rembrandt van Rijn",
			},
			&PropertyFilter{
				FieldName:    "type",
				OperatorType: OpIn,
				Value:        []string{"schema:Painting", "schema:Drawing"},
			},
		},
		Facets: []FacetRequest{
			{Field: "type", Limit: 20},
			{Field: "genre", Limit: 10},
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
				FieldName:    "title",
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
	config := CreativeWorkConfig()
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
		if m.Variable == "filter[creator][eq]" {
			hasCreatorFilter = true
			break
		}
	}
	if !hasCreatorFilter {
		t.Error("Expected template to have filter[creator][eq] mapping")
	}

	// Check that template has facet mappings from config
	hasTypeFacet := false
	for _, m := range template.Mapping {
		if m.Variable == "facet[type]" {
			hasTypeFacet = true
			break
		}
	}
	if !hasTypeFacet {
		t.Error("Expected template to have facet[type] mapping")
	}

	// Check that advanced search is configured
	if template.Advanced == nil {
		t.Error("Expected template to have advanced search info")
	}
}
