package ntriples

import (
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/resource"
	rdf "github.com/kiivihal/gon3"
)

func Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	if g == nil {
		g = resource.NewGraph()
	}

	// TODO(kiivihal): later replace with baseuri from the graph
	parser, err := rdf.NewParser("").Parse(r)
	if err != nil {
		return g, err
	}

	for triple := range parser.IterTriples() {
		t, err := rdftriple2triple(triple)
		if err != nil {
			return g, err
		}

		g.Add(t)
	}

	return g, nil
}

func rdftriple2triple(triple *rdf.Triple) (*resource.Triple, error) {
	s, err := rdf2term(triple.Subject)
	if err != nil {
		return nil, err
	}

	p, err := rdf2term(triple.Predicate)
	if err != nil {
		return nil, err
	}

	o, err := rdf2term(triple.Object)
	if err != nil {
		return nil, err
	}

	return resource.NewTriple(s.(resource.Subject), p.(resource.Predicate), o.(resource.Object)), nil
}

func rdf2term(term rdf.Term) (resource.Term, error) {
	switch term := term.(type) {
	case *rdf.BlankNode:
		return resource.NewBlankNode(term.RawValue())
	case *rdf.Literal:
		if len(term.LanguageTag) > 0 {
			return resource.NewLiteralWithLang(term.LexicalForm, term.LanguageTag)
		}

		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			dt, err := resource.NewIRI(term.DatatypeIRI.RawValue())
			if err != nil {
				return nil, err
			}

			return resource.NewLiteralWithType(term.LexicalForm, &dt)
		}

		return resource.NewLiteral(term.RawValue())
	case *rdf.IRI:
		return resource.NewIRI(term.RawValue())
	}

	return nil, fmt.Errorf("unknown RDF term type: %T", term)
}

// compile time check of interface
var _ resource.Parser = (*p)(nil)

type p struct{}

func (p p) Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	return Parse(r, g)
}
