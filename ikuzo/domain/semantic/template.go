package semantic

import (
	"fmt"
	"net/url"
	"strings"
)

// SearchTemplate represents a Hydra IRI template for search operations.
// It provides self-documenting search capabilities by exposing available
// search parameters and their constraints.
type SearchTemplate struct {
	TypeValue string             `json:"@type"`
	Template  string             `json:"hydra:template"`
	Mapping   []TemplateMapping  `json:"hydra:mapping"`
	Advanced  *AdvancedSearchInfo `json:"hub3:advancedSearch,omitempty"`
}

// NewSearchTemplate creates a new search template with default values.
func NewSearchTemplate(baseURL string) *SearchTemplate {
	return &SearchTemplate{
		TypeValue: "hydra:IriTemplate",
		Template:  baseURL + "{?query,filter*,facet*,page,size,sort}",
		Mapping:   []TemplateMapping{},
	}
}

// TemplateMapping represents a mapping between a template variable and a property.
// It includes documentation and validation information for self-documentation.
type TemplateMapping struct {
	Variable    string `json:"hydra:variable"`
	Property    string `json:"hydra:property"`
	Required    bool   `json:"hydra:required"`
	Description string `json:"schema:description,omitempty"`
	Operator    string `json:"hub3:operator,omitempty"`
	Example     string `json:"hub3:example,omitempty"`
	ValueType   string `json:"hub3:valueType,omitempty"` // string, integer, float, boolean
}

// NewTemplateMapping creates a basic template mapping.
func NewTemplateMapping(variable, property, description string) TemplateMapping {
	return TemplateMapping{
		Variable:    variable,
		Property:    property,
		Description: description,
	}
}

// WithOperator sets the operator for this mapping.
func (tm TemplateMapping) WithOperator(op Operator) TemplateMapping {
	tm.Operator = string(op)
	return tm
}

// WithExample sets an example value for this mapping.
func (tm TemplateMapping) WithExample(example string) TemplateMapping {
	tm.Example = example
	return tm
}

// WithRequired marks this mapping as required.
func (tm TemplateMapping) WithRequired(required bool) TemplateMapping {
	tm.Required = required
	return tm
}

// WithValueType sets the value type for this mapping.
func (tm TemplateMapping) WithValueType(valueType string) TemplateMapping {
	tm.ValueType = valueType
	return tm
}

// AdvancedSearchInfo provides information about the POST endpoint for complex queries.
type AdvancedSearchInfo struct {
	TypeValue   string `json:"@type"`
	Method      string `json:"hydra:method"`
	Expects     string `json:"hydra:expects"`
	Returns     string `json:"hydra:returns"`
	Description string `json:"schema:description"`
	Endpoint    string `json:"hub3:endpoint,omitempty"`
}

// NewAdvancedSearchInfo creates advanced search info for POST endpoint.
func NewAdvancedSearchInfo(endpoint string) *AdvancedSearchInfo {
	return &AdvancedSearchInfo{
		TypeValue:   "hydra:Operation",
		Method:      "POST",
		Expects:     "hub3:SearchQuery",
		Returns:     "hydra:Collection",
		Description: "Advanced search with complex filters using JSON-LD",
		Endpoint:    endpoint,
	}
}

// TemplateBuilder helps construct search templates from filter and facet configurations.
type TemplateBuilder struct {
	baseURL  string
	mappings []TemplateMapping
	advanced *AdvancedSearchInfo
}

// NewTemplateBuilder creates a new template builder.
func NewTemplateBuilder(baseURL string) *TemplateBuilder {
	return &TemplateBuilder{
		baseURL:  baseURL,
		mappings: []TemplateMapping{},
	}
}

// AddMapping adds a template mapping.
func (tb *TemplateBuilder) AddMapping(mapping TemplateMapping) *TemplateBuilder {
	tb.mappings = append(tb.mappings, mapping)
	return tb
}

// AddFilterMapping adds a filter mapping for a field.
func (tb *TemplateBuilder) AddFilterMapping(field, property, description string, operators []Operator) *TemplateBuilder {
	for _, op := range operators {
		variable := fmt.Sprintf("filter[%s][%s]", field, op)
		mapping := NewTemplateMapping(variable, property, description).
			WithOperator(op).
			WithExample(fmt.Sprintf("filter[%s][%s]=value", field, op))
		tb.mappings = append(tb.mappings, mapping)
	}
	return tb
}

// AddFacetMapping adds a facet mapping for a field.
func (tb *TemplateBuilder) AddFacetMapping(field, property, description string) *TemplateBuilder {
	variable := fmt.Sprintf("facet[%s]", field)
	mapping := NewTemplateMapping(variable, property, description).
		WithExample(fmt.Sprintf("facet[%s]=true", field)).
		WithValueType("boolean")
	tb.mappings = append(tb.mappings, mapping)
	return tb
}

// AddBasicMappings adds the basic search mappings (query, page, size, sort).
func (tb *TemplateBuilder) AddBasicMappings() *TemplateBuilder {
	tb.mappings = append(tb.mappings,
		NewTemplateMapping("query", "hydra:freetextQuery", "Full-text search query").
			WithExample("query=rembrandt").
			WithValueType("string"),
		NewTemplateMapping("page", "hydra:pageIndex", "Page number (1-based)").
			WithExample("page=1").
			WithValueType("integer"),
		NewTemplateMapping("size", "hydra:pageSize", "Number of results per page").
			WithExample("size=20").
			WithValueType("integer"),
		NewTemplateMapping("sort", "hub3:sortField", "Sort field with direction (+/-)").
			WithExample("sort=+title").
			WithValueType("string"),
	)
	return tb
}

// WithAdvancedSearch adds advanced search information.
func (tb *TemplateBuilder) WithAdvancedSearch(endpoint string) *TemplateBuilder {
	tb.advanced = NewAdvancedSearchInfo(endpoint)
	return tb
}

// Build constructs the search template.
func (tb *TemplateBuilder) Build() *SearchTemplate {
	// Build template string with all variables
	variables := []string{}
	for _, m := range tb.mappings {
		// Check if variable is already in the list (avoid duplicates)
		found := false
		for _, v := range variables {
			if v == m.Variable {
				found = true
				break
			}
		}
		if !found {
			variables = append(variables, m.Variable)
		}
	}

	template := tb.baseURL
	if len(variables) > 0 {
		template += "{?" + strings.Join(variables, ",") + "}"
	}

	return &SearchTemplate{
		TypeValue: "hydra:IriTemplate",
		Template:  template,
		Mapping:   tb.mappings,
		Advanced:  tb.advanced,
	}
}

// BuildTemplateFromConfigs builds a search template from filter and facet configurations.
func BuildTemplateFromConfigs(baseURL string, filters []FilterConfig, facets []FacetConfig) *SearchTemplate {
	builder := NewTemplateBuilder(baseURL)
	builder.AddBasicMappings()

	// Add filter mappings
	for _, fc := range filters {
		if fc.AllowedInQuery {
			builder.AddFilterMapping(fc.Field, fc.PropertyURI, fc.Description, fc.AllowedOperators)
		}
	}

	// Add facet mappings
	for _, fc := range facets {
		builder.AddFacetMapping(fc.Field, fc.PropertyURI, fc.Description)
	}

	// Add advanced search endpoint
	builder.WithAdvancedSearch(baseURL)

	return builder.Build()
}

// ExpandTemplate expands a template with the given values.
// This is useful for generating example URLs or validating template usage.
func ExpandTemplate(template *SearchTemplate, values url.Values) (string, error) {
	baseURL := template.Template[:strings.Index(template.Template, "{")]

	if len(values) == 0 {
		return baseURL, nil
	}

	// Validate that all values correspond to mapped variables
	validVariables := make(map[string]bool)
	for _, m := range template.Mapping {
		validVariables[m.Variable] = true
	}

	params := url.Values{}
	for key, vals := range values {
		if !validVariables[key] {
			return "", fmt.Errorf("unknown template variable: %s", key)
		}
		for _, v := range vals {
			params.Add(key, v)
		}
	}

	if len(params) == 0 {
		return baseURL, nil
	}

	return baseURL + "?" + params.Encode(), nil
}
