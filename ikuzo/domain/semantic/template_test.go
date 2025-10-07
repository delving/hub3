package semantic

import (
	"net/url"
	"testing"
)

func TestNewSearchTemplate(t *testing.T) {
	baseURL := "https://example.org/api/semantic/v1/search"
	template := NewSearchTemplate(baseURL)

	if template.TypeValue != "hydra:IriTemplate" {
		t.Errorf("TypeValue = %v, want hydra:IriTemplate", template.TypeValue)
	}

	expectedTemplate := baseURL + "{?query,filter*,facet*,page,size,sort}"
	if template.Template != expectedTemplate {
		t.Errorf("Template = %v, want %v", template.Template, expectedTemplate)
	}
}

func TestTemplateMapping_WithMethods(t *testing.T) {
	mapping := NewTemplateMapping("query", "hydra:freetextQuery", "Full-text search").
		WithOperator(OpEqual).
		WithExample("query=test").
		WithRequired(true).
		WithValueType("string")

	if mapping.Operator != string(OpEqual) {
		t.Errorf("Operator = %v, want %v", mapping.Operator, OpEqual)
	}

	if mapping.Example != "query=test" {
		t.Errorf("Example = %v, want query=test", mapping.Example)
	}

	if !mapping.Required {
		t.Error("Expected mapping to be required")
	}

	if mapping.ValueType != "string" {
		t.Errorf("ValueType = %v, want string", mapping.ValueType)
	}
}

func TestTemplateBuilder_Build(t *testing.T) {
	baseURL := "https://example.org/api/semantic/v1/search"
	builder := NewTemplateBuilder(baseURL)

	builder.AddBasicMappings().
		AddFilterMapping("creator", "schema:creator", "Filter by creator", []Operator{OpEqual, OpIn}).
		AddFacetMapping("type", "rdf:type", "Facet by type").
		WithAdvancedSearch(baseURL)

	template := builder.Build()

	if template.TypeValue != "hydra:IriTemplate" {
		t.Errorf("TypeValue = %v, want hydra:IriTemplate", template.TypeValue)
	}

	// Check that basic mappings were added
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

	// Check advanced search info
	if template.Advanced == nil {
		t.Fatal("Expected advanced search info")
	}

	if template.Advanced.Method != "POST" {
		t.Errorf("Advanced.Method = %v, want POST", template.Advanced.Method)
	}
}

func TestBuildTemplateFromConfigs(t *testing.T) {
	baseURL := "https://example.org/api/semantic/v1/search"

	filters := []FilterConfig{
		*NewFilterConfig("title", "schema:name", "Title").
			WithOperators(OpEqual, OpContains),
		*NewFilterConfig("creator", "schema:creator", "Creator").
			WithOperators(OpEqual, OpIn),
	}

	facets := []FacetConfig{
		*NewFacetConfig("type", "rdf:type", "Type", "enum"),
		*NewFacetConfig("year", "schema:dateCreated", "Year", "range"),
	}

	template := BuildTemplateFromConfigs(baseURL, filters, facets)

	if template == nil {
		t.Fatal("Expected template to be created")
	}

	// Check that filter mappings were created
	hasCreatorEq := false
	for _, m := range template.Mapping {
		if m.Variable == "filter[creator][eq]" {
			hasCreatorEq = true
			break
		}
	}
	if !hasCreatorEq {
		t.Error("Expected template to have filter[creator][eq] mapping")
	}

	// Check that facet mappings were created
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
}

func TestExpandTemplate(t *testing.T) {
	baseURL := "https://example.org/api/semantic/v1/search"
	template := NewSearchTemplate(baseURL)
	template.Mapping = []TemplateMapping{
		NewTemplateMapping("query", "hydra:freetextQuery", "Query"),
		NewTemplateMapping("page", "hydra:pageIndex", "Page"),
	}

	tests := []struct {
		name      string
		values    url.Values
		wantURL   string
		wantError bool
	}{
		{
			name:      "empty values",
			values:    url.Values{},
			wantURL:   baseURL,
			wantError: false,
		},
		{
			name: "valid values",
			values: url.Values{
				"query": []string{"rembrandt"},
				"page":  []string{"1"},
			},
			wantURL:   baseURL + "?page=1&query=rembrandt",
			wantError: false,
		},
		{
			name: "unknown variable",
			values: url.Values{
				"unknown": []string{"value"},
			},
			wantURL:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := ExpandTemplate(template, tt.values)
			if (err != nil) != tt.wantError {
				t.Errorf("ExpandTemplate() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && gotURL != tt.wantURL {
				t.Errorf("ExpandTemplate() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestNewAdvancedSearchInfo(t *testing.T) {
	endpoint := "https://example.org/api/semantic/v1/search"
	info := NewAdvancedSearchInfo(endpoint)

	if info.TypeValue != "hydra:Operation" {
		t.Errorf("TypeValue = %v, want hydra:Operation", info.TypeValue)
	}

	if info.Method != "POST" {
		t.Errorf("Method = %v, want POST", info.Method)
	}

	if info.Expects != "hub3:SearchQuery" {
		t.Errorf("Expects = %v, want hub3:SearchQuery", info.Expects)
	}

	if info.Returns != "hydra:Collection" {
		t.Errorf("Returns = %v, want hydra:Collection", info.Returns)
	}

	if info.Endpoint != endpoint {
		t.Errorf("Endpoint = %v, want %v", info.Endpoint, endpoint)
	}
}
