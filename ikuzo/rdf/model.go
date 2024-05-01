package rdf

// Model can be used as a base for structs that will implement
// UnmarshalRDF.
type Model struct {
	Type []string `json:"@type,omitempty" rdf:"@types"`
	ID   string   `json:"@id,omitempty" rdf:"@id"`
}

// LiteralOrResource can be used as datatype for
type LiteralOrResource struct {
	ID       string
	Value    string
	Language string
	DataType string
}

func (lor *LiteralOrResource) String() string {
	if lor.ID != "" {
		return lor.ID
	}

	return lor.Value
}

func (lor *LiteralOrResource) IsEmpty() bool {
	return lor.String() == ""
}
