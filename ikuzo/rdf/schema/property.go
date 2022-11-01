package schema

import "github.com/delving/hub3/ikuzo/rdf/schema/shacl"

type Property struct {
	URI       string
	Label     []Label
	Comment   string
	Domain    []string
	Range     []string
	InverseOf []string
	Shacl     shacl.Property
}
