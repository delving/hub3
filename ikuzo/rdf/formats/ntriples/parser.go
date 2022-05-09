package ntriples

import (
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
	gonrdf "github.com/kiivihal/gon3"
)

func Parse(r io.Reader, g *rdf.Graph) (*rdf.Graph, error) {
	if g == nil {
		g = rdf.NewGraph()
	}

	// TODO(kiivihal): later replace with baseuri from the graph
	parser, err := gonrdf.NewParser("").Parse(r)
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

func rdftriple2triple(triple *gonrdf.Triple) (*rdf.Triple, error) {
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

	return rdf.NewTriple(s.(rdf.Subject), p.(rdf.Predicate), o.(rdf.Object)), nil
}

func rdf2term(term gonrdf.Term) (rdf.Term, error) {
	switch term := term.(type) {
	case *gonrdf.BlankNode:
		return rdf.NewBlankNode(term.RawValue())
	case *gonrdf.Literal:
		if len(term.LanguageTag) > 0 {
			return rdf.NewLiteralWithLang(term.LexicalForm, term.LanguageTag)
		}

		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			dt, err := rdf.NewIRI(term.DatatypeIRI.RawValue())
			if err != nil {
				return nil, err
			}

			return rdf.NewLiteralWithType(term.LexicalForm, dt)
		}

		return rdf.NewLiteral(term.RawValue())
	case *gonrdf.IRI:
		return rdf.NewIRI(term.RawValue())
	}

	return nil, fmt.Errorf("unknown RDF term type: %T", term)
}

// compile time check of interface
var _ rdf.Parser = (*p)(nil)

type p struct{}

func (p p) Parse(r io.Reader, g *rdf.Graph) (*rdf.Graph, error) {
	return Parse(r, g)
}
