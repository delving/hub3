package schema

import "github.com/delving/hub3/ikuzo/rdf/schema/shacl"

type Label struct {
	Value    string
	Language string
}

type Class struct {
	URI        string
	Label      []Label
	Comment    string
	SubClassOf []string
	Inferred   struct {
		SuperClass []string
		SubClassOf []string
		Properties []string
	}
	Property  []Property
	NodeShape shacl.NodeShape
	Example   []Example
}

type Example struct {
	SourceSystem string
	Mimetype     string
	Data         string
}
