package shacl

const (
	URI    = "http://www.w3.org/ns/shacl#"
	Prefix = "shacl"
)

type (
	Class string
	IRI   string
)

type ShapeGraph struct {
	// Namespaces []
	Shapes   []NodeShape
	Property []Property // used when parsing from source
}

type NodeShape struct {
	Name            string
	Closed          bool
	TargetClass     Class
	Property        []Property
	TargetsObjectOf []Property
	Or              []Class
	Error           struct {
		Severity IRI
		Message  string
	}
}

type Property struct {
	DataType    string
	Description string
	MinCount    int
	MaxCount    int
	Name        string
	Path        IRI
	Class       Class
	Validation  struct {
		HasValues string
		Pattern   string
	}
	Error struct {
		Severity IRI
		Message  string
	}
}
