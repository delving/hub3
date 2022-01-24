// Package jsonld provides tools to parse and serialize RDF data in the JSON-LD format.
//
// For more information about JSON-LD, see - https://json-ld.org.
package jsonld

import (
	"bytes"
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/resource"
	jsonld "github.com/kiivihal/gojsonld"
)

func Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	if g == nil {
		g = resource.NewGraph()
	}

	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(r)
	if err != nil {
		return g, err
	}

	jsonData, err := jsonld.ReadJSON(buf.Bytes())
	if err != nil {
		return g, err
	}

	options := &jsonld.Options{}
	options.Base = ""
	options.ProduceGeneralizedRdf = false

	dataSet, err := jsonld.ToRDF(jsonData, options)
	if err != nil {
		return g, err
	}

	for triple := range dataSet.IterTriples() {
		t, err := jtriple2triple(triple)
		if err != nil {
			return g, err
		}

		g.Add(t)
	}

	return g, nil
}

func jtriple2triple(triple *jsonld.Triple) (*resource.Triple, error) {
	s, err := jterm2term(triple.Subject)
	if err != nil {
		return nil, err
	}

	p, err := jterm2term(triple.Predicate)
	if err != nil {
		return nil, err
	}

	o, err := jterm2term(triple.Object)
	if err != nil {
		return nil, err
	}

	return resource.NewTriple(s.(resource.Subject), p.(resource.Predicate), o.(resource.Object)), nil
}

func jterm2term(term jsonld.Term) (resource.Term, error) {
	switch term := term.(type) {
	case *jsonld.BlankNode:
		return resource.NewBlankNode(term.RawValue())
	case *jsonld.Literal:
		if len(term.Language) > 0 {
			return resource.NewLiteralWithLang(term.RawValue(), term.Language)
		}

		if term.Datatype != nil && len(term.Datatype.String()) > 0 {
			dt, err := resource.NewIRI(term.Datatype.RawValue())
			if err != nil {
				return nil, err
			}

			return resource.NewLiteralWithType(term.Value, &dt)
		}

		return resource.NewLiteral(term.Value)
	case *jsonld.Resource:
		return resource.NewIRI(term.RawValue())
	}

	return nil, fmt.Errorf("unknown resource.TermType")
}

// compile time check of interface
var _ resource.Parser = (*p)(nil)

type p struct{}

func (p p) Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	return Parse(r, g)
}
