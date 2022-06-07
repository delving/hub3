package rdf

import "io"

// Parser is an interface for parsing RDF in io.Reader and storing it into the Graph.
//
// This should be implemement by RDF parsing packages
type Parser interface {
	Parse(r io.Reader, g *Graph) (*Graph, error)
}
