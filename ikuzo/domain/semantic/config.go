package semantic

import (
	"fmt"
)

// FilterConfig defines the configuration for a filterable field.
// This is the single source of truth for which fields can be filtered,
// which operators are allowed, and how values should be validated.
type FilterConfig struct {
	Field            string     `json:"field"`
	PropertyURI      string     `json:"propertyURI"`
	Label            string     `json:"label"`
	Description      string     `json:"description"`
	AllowedOperators []Operator `json:"allowedOperators"`
	AllowedInQuery   bool       `json:"allowedInQuery"`   // Can be used in GET requests
	AllowedInPost    bool       `json:"allowedInPost"`    // Can be used in POST requests
	ValueType        string     `json:"valueType"`        // string, integer, float, boolean, date, uri
	MultiValue       bool       `json:"multiValue"`       // Whether multiple values are allowed
	ValidationRegex  string     `json:"validationRegex,omitempty"`
	Examples         []string   `json:"examples,omitempty"`
}

// NewFilterConfig creates a basic filter configuration.
func NewFilterConfig(field, propertyURI, label string) *FilterConfig {
	return &FilterConfig{
		Field:            field,
		PropertyURI:      propertyURI,
		Label:            label,
		AllowedOperators: []Operator{OpEqual},
		AllowedInQuery:   true,
		AllowedInPost:    true,
		ValueType:        "string",
		MultiValue:       false,
	}
}

// WithOperators sets the allowed operators.
func (fc *FilterConfig) WithOperators(operators ...Operator) *FilterConfig {
	fc.AllowedOperators = operators
	return fc
}

// WithDescription sets the description.
func (fc *FilterConfig) WithDescription(description string) *FilterConfig {
	fc.Description = description
	return fc
}

// WithValueType sets the value type.
func (fc *FilterConfig) WithValueType(valueType string) *FilterConfig {
	fc.ValueType = valueType
	return fc
}

// WithMultiValue sets whether multiple values are allowed.
func (fc *FilterConfig) WithMultiValue(multiValue bool) *FilterConfig {
	fc.MultiValue = multiValue
	return fc
}

// WithQueryAccess sets whether this filter can be used in GET requests.
func (fc *FilterConfig) WithQueryAccess(allowed bool) *FilterConfig {
	fc.AllowedInQuery = allowed
	return fc
}

// WithExamples sets example values.
func (fc *FilterConfig) WithExamples(examples ...string) *FilterConfig {
	fc.Examples = examples
	return fc
}

// IsOperatorAllowed checks if an operator is allowed for this field.
func (fc *FilterConfig) IsOperatorAllowed(op Operator) bool {
	for _, allowed := range fc.AllowedOperators {
		if allowed == op {
			return true
		}
	}
	return false
}

// Validate validates a filter against this configuration.
func (fc *FilterConfig) Validate(filter Filter) error {
	if filter.Field() != fc.Field {
		return NewValidationError(filter.Field(), "field does not match config", filter.Field())
	}

	if !fc.IsOperatorAllowed(filter.Operator()) {
		return NewValidationError(
			filter.Field(),
			fmt.Sprintf("operator '%s' not allowed for field '%s'", filter.Operator(), fc.Field),
			filter.Operator(),
		)
	}

	return nil
}

// FacetConfig defines the configuration for a facet field.
type FacetConfig struct {
	Field         string            `json:"field"`
	PropertyURI   string            `json:"propertyURI"`
	Label         string            `json:"label"`
	Description   string            `json:"description"`
	FacetType     string            `json:"facetType"` // enum, range, date, boolean
	Size          int               `json:"size"`      // Number of values to return
	MinDocCount   int64             `json:"minDocCount"`
	SortBy        string            `json:"sortBy"`        // count, term, index
	SortOrder     string            `json:"sortOrder"`     // asc, desc
	Missing       bool              `json:"missing"`       // Include missing values
	Ranges        []FacetRange      `json:"ranges,omitempty"`
	DateIntervals []DateInterval    `json:"dateIntervals,omitempty"`
	URLTemplate   string            `json:"urlTemplate,omitempty"` // Template for facet URLs
	Metadata      map[string]any    `json:"metadata,omitempty"`    // Additional metadata
}

// FacetRange defines a range for range facets.
type FacetRange struct {
	Key   string  `json:"key"`
	From  float64 `json:"from,omitempty"`
	To    float64 `json:"to,omitempty"`
	Label string  `json:"label,omitempty"`
}

// DateInterval defines an interval for date facets.
type DateInterval struct {
	Interval string `json:"interval"` // year, quarter, month, week, day
	Format   string `json:"format,omitempty"`
}

// NewFacetConfig creates a basic facet configuration.
func NewFacetConfig(field, propertyURI, label, facetType string) *FacetConfig {
	return &FacetConfig{
		Field:       field,
		PropertyURI: propertyURI,
		Label:       label,
		FacetType:   facetType,
		Size:        10,
		MinDocCount: 1,
		SortBy:      "count",
		SortOrder:   "desc",
		Missing:     false,
	}
}

// WithDescription sets the description.
func (fc *FacetConfig) WithDescription(description string) *FacetConfig {
	fc.Description = description
	return fc
}

// WithSize sets the maximum number of facet values.
func (fc *FacetConfig) WithSize(size int) *FacetConfig {
	fc.Size = size
	return fc
}

// WithMinDocCount sets the minimum document count.
func (fc *FacetConfig) WithMinDocCount(count int64) *FacetConfig {
	fc.MinDocCount = count
	return fc
}

// WithSort sets the sort field and order.
func (fc *FacetConfig) WithSort(sortBy, sortOrder string) *FacetConfig {
	fc.SortBy = sortBy
	fc.SortOrder = sortOrder
	return fc
}

// WithRanges sets the ranges for a range facet.
func (fc *FacetConfig) WithRanges(ranges ...FacetRange) *FacetConfig {
	fc.Ranges = ranges
	return fc
}

// WithDateIntervals sets the intervals for a date facet.
func (fc *FacetConfig) WithDateIntervals(intervals ...DateInterval) *FacetConfig {
	fc.DateIntervals = intervals
	return fc
}

// WithMissing sets whether to include missing values.
func (fc *FacetConfig) WithMissing(missing bool) *FacetConfig {
	fc.Missing = missing
	return fc
}

// WithMetadata adds metadata to the facet config.
func (fc *FacetConfig) WithMetadata(key string, value any) *FacetConfig {
	if fc.Metadata == nil {
		fc.Metadata = make(map[string]any)
	}
	fc.Metadata[key] = value
	return fc
}

// ResourceConfig defines the complete search configuration for a resource type.
// This includes all filterable fields, facets, and resource-specific settings.
type ResourceConfig struct {
	ResourceType   string          `json:"resourceType"`   // e.g., "schema:CreativeWork"
	Label          string          `json:"label"`
	Description    string          `json:"description"`
	Filters        []FilterConfig  `json:"filters"`
	Facets         []FacetConfig   `json:"facets"`
	SortFields     []SortConfig    `json:"sortFields"`
	DefaultSort    string          `json:"defaultSort"`
	DefaultSize    int             `json:"defaultSize"`
	MaxSize        int             `json:"maxSize"`
	AllowFullText  bool            `json:"allowFullText"`
	Metadata       map[string]any  `json:"metadata,omitempty"`
}

// SortConfig defines a sortable field.
type SortConfig struct {
	Field       string `json:"field"`
	PropertyURI string `json:"propertyURI"`
	Label       string `json:"label"`
	DefaultDir  string `json:"defaultDir"` // asc, desc
}

// NewResourceConfig creates a new resource configuration.
func NewResourceConfig(resourceType, label string) *ResourceConfig {
	return &ResourceConfig{
		ResourceType:  resourceType,
		Label:         label,
		Filters:       []FilterConfig{},
		Facets:        []FacetConfig{},
		SortFields:    []SortConfig{},
		DefaultSize:   20,
		MaxSize:       100,
		AllowFullText: true,
	}
}

// WithDescription sets the description.
func (rc *ResourceConfig) WithDescription(description string) *ResourceConfig {
	rc.Description = description
	return rc
}

// AddFilter adds a filter configuration.
func (rc *ResourceConfig) AddFilter(filter *FilterConfig) *ResourceConfig {
	rc.Filters = append(rc.Filters, *filter)
	return rc
}

// AddFacet adds a facet configuration.
func (rc *ResourceConfig) AddFacet(facet *FacetConfig) *ResourceConfig {
	rc.Facets = append(rc.Facets, *facet)
	return rc
}

// AddSortField adds a sortable field.
func (rc *ResourceConfig) AddSortField(field, propertyURI, label, defaultDir string) *ResourceConfig {
	rc.SortFields = append(rc.SortFields, SortConfig{
		Field:       field,
		PropertyURI: propertyURI,
		Label:       label,
		DefaultDir:  defaultDir,
	})
	return rc
}

// WithDefaultSort sets the default sort field.
func (rc *ResourceConfig) WithDefaultSort(sortField string) *ResourceConfig {
	rc.DefaultSort = sortField
	return rc
}

// WithDefaultSize sets the default page size.
func (rc *ResourceConfig) WithDefaultSize(size int) *ResourceConfig {
	rc.DefaultSize = size
	return rc
}

// WithMaxSize sets the maximum page size.
func (rc *ResourceConfig) WithMaxSize(size int) *ResourceConfig {
	rc.MaxSize = size
	return rc
}

// GetFilterConfig gets the filter configuration for a field.
func (rc *ResourceConfig) GetFilterConfig(field string) (*FilterConfig, bool) {
	for i := range rc.Filters {
		if rc.Filters[i].Field == field {
			return &rc.Filters[i], true
		}
	}
	return nil, false
}

// GetFacetConfig gets the facet configuration for a field.
func (rc *ResourceConfig) GetFacetConfig(field string) (*FacetConfig, bool) {
	for i := range rc.Facets {
		if rc.Facets[i].Field == field {
			return &rc.Facets[i], true
		}
	}
	return nil, false
}

// IsFieldFilterable checks if a field can be filtered.
func (rc *ResourceConfig) IsFieldFilterable(field string) bool {
	_, ok := rc.GetFilterConfig(field)
	return ok
}

// IsFieldFacetable checks if a field can be faceted.
func (rc *ResourceConfig) IsFieldFacetable(field string) bool {
	_, ok := rc.GetFacetConfig(field)
	return ok
}

// ValidateFilter validates a filter against this resource configuration.
func (rc *ResourceConfig) ValidateFilter(filter Filter) error {
	config, ok := rc.GetFilterConfig(filter.Field())
	if !ok {
		return NewValidationError(
			filter.Field(),
			fmt.Sprintf("field '%s' is not filterable for resource type '%s'", filter.Field(), rc.ResourceType),
			filter.Field(),
		)
	}

	return config.Validate(filter)
}

// ValidateQueryOptions validates query options against this resource configuration.
func (rc *ResourceConfig) ValidateQueryOptions(opts *QueryOptions) error {
	// Validate filters
	for _, filter := range opts.Filters {
		if err := rc.ValidateFilter(filter); err != nil {
			return err
		}
	}

	// Validate facets
	for _, facetReq := range opts.Facets {
		if !rc.IsFieldFacetable(facetReq.Field) {
			return NewValidationError(
				facetReq.Field,
				fmt.Sprintf("field '%s' is not facetable for resource type '%s'", facetReq.Field, rc.ResourceType),
				facetReq.Field,
			)
		}
	}

	// Validate pagination
	if opts.Pagination != nil {
		if opts.Pagination.Size > rc.MaxSize {
			return NewValidationError(
				"size",
				fmt.Sprintf("page size %d exceeds maximum of %d", opts.Pagination.Size, rc.MaxSize),
				opts.Pagination.Size,
			)
		}
		if opts.Pagination.Size < 1 {
			return NewValidationError(
				"size",
				"page size must be at least 1",
				opts.Pagination.Size,
			)
		}
	}

	return nil
}

// ConfigRegistry manages resource configurations for different resource types.
type ConfigRegistry struct {
	configs map[string]*ResourceConfig
}

// NewConfigRegistry creates a new configuration registry.
func NewConfigRegistry() *ConfigRegistry {
	return &ConfigRegistry{
		configs: make(map[string]*ResourceConfig),
	}
}

// Register registers a resource configuration.
func (cr *ConfigRegistry) Register(config *ResourceConfig) {
	cr.configs[config.ResourceType] = config
}

// Get retrieves a resource configuration by type.
func (cr *ConfigRegistry) Get(resourceType string) (*ResourceConfig, bool) {
	config, ok := cr.configs[resourceType]
	return config, ok
}

// GetDefault retrieves the default resource configuration.
// This is typically the first registered config or a designated default.
func (cr *ConfigRegistry) GetDefault() *ResourceConfig {
	// For now, return the first config if any exist
	for _, config := range cr.configs {
		return config
	}
	return nil
}

// ListResourceTypes returns all registered resource types.
func (cr *ConfigRegistry) ListResourceTypes() []string {
	types := make([]string, 0, len(cr.configs))
	for t := range cr.configs {
		types = append(types, t)
	}
	return types
}
