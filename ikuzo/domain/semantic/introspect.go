package semantic

// IntrospectionResult contains the result of an introspection query.
type IntrospectionResult struct {
	Scope      IntrospectionScope `json:"hub3:scope"`
	Classes    []ClassInfo        `json:"hub3:classes,omitempty"`
	Schemas    []SchemaInfo       `json:"hub3:schemas,omitempty"`
	Properties []PropertyInfo     `json:"hub3:properties,omitempty"`
}

// IntrospectionScope describes what data the introspection covers.
type IntrospectionScope struct {
	Type           string `json:"type"`
	ContextID      string `json:"context,omitempty"`
	TotalDocuments int64  `json:"totalDocuments"`
}

// ClassInfo describes an RDF class found in the data.
type ClassInfo struct {
	URI            string `json:"uri"`
	Label          string `json:"label"`
	Count          int64  `json:"count"`
	PropertiesLink string `json:"properties,omitempty"`
}

// PropertyInfo describes a property found on a class.
type PropertyInfo struct {
	Field             string      `json:"field"`
	Predicate         string      `json:"predicate"`
	Label             string      `json:"label"`
	ValueTypes        []string    `json:"valueTypes"`
	Count             int64       `json:"count"`
	Languages         []string    `json:"languages,omitempty"`
	HasResolvedLabels bool        `json:"hasResolvedLabels,omitempty"`
	DataType          string      `json:"dataType,omitempty"`
	Range             *FieldRange `json:"range,omitempty"`
	Paths             []string    `json:"paths,omitempty"`
	Schema            *SchemaRef  `json:"schema,omitempty"`
}

// FieldRange describes the min/max range for numeric or date fields.
type FieldRange struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

// SchemaInfo describes a record-definition schema found in the data.
type SchemaInfo struct {
	RecDefID      string   `json:"recDefID"`
	DocumentCount int64    `json:"documentCount"`
	Specs         []string `json:"spec"`
}

// SchemaRef links a property to its record-definition documentation.
type SchemaRef struct {
	RecDefID      string `json:"recDefID"`
	Documentation string `json:"documentation,omitempty"`
}

// PathInfo describes a predicate path between classes.
type PathInfo struct {
	Path      string `json:"path"`
	FromClass string `json:"fromClass"`
	ToClass   string `json:"toClass"`
	Count     int64  `json:"count"`
}
