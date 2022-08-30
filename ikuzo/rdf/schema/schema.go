package rdfs

// TODO(kiivihal): toShacl, toJsonLDcontext, toRDFXML functions

// Model is a dedicated abstraction of one or more Schema instances.
// The Model can be converted into
type Model struct{}

// Schema represents a RDFS Schema compatible structure of a Linked Data Model
type Schema struct {
	// Doc is the documentation of the schema
	Doc string

	// Classes is an array of classes of the Schema
	Classes []Class

	// Properties is an array of properties of the schema
	Properties []Property

	// Namespaces that are using in schema Classes and Properties
	Namespaces map[string]string
}

type Class struct {
	// ID is the rdf IRI of the property
	ID string
	// Labels is a list of 'multilingual' literals for the label of the property
	Labels []Literal
	// Comments is a list of 'multilingual' literals for the comments on the property
	Comments []Literal
	// SubClassOf points to the SuperClass that this Class inherits from
	SubClassOf []string
	SHACL      struct {
		Closed bool
	}
}

type Property struct {
	// the rdf IRI of the property
	ID string
	// a list of 'multilangual' literals for the label of the property
	Labels []Literal
	// a list of 'multilangual' literals for the comments on the property
	Comments []Literal
	// the class where the property can be used.
	// Via SubClassOf inheritance this is applied to each sub-class of the target Class.
	Domain []string
	// the target classes of the property
	Range []string
	// SubPropertyOf defines inheritance of properties and linked classes
	SubPropertyOf []string
	// SHACL configuration of the property
	SHACL struct {
		Datatype    string
		Description string
		MaxCount    int
		MinCount    int
		Name        Literal
		Class       []string // target classes of the property
	}
}

type Literal struct {
	Value    string
	Language string
}
