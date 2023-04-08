package rdfxml

import (
	"io"

	xmlrdf "github.com/knakk/rdf"

	"github.com/delving/hub3/ikuzo/rdf"
)

func Parse(r io.Reader, g *rdf.Graph) (*rdf.Graph, error) {
	if g == nil {
		g = rdf.NewGraph()
	}

	triples, err := decodeRDFXML(r)
	if err != nil {
		return nil, err
	}

	g.Add(triples...)

	if err := g.Inline(); err != nil {
		return g, err
	}

	return g, nil
}

// decodeRDFXML parses RDF-XML into triples
func decodeRDFXML(r io.Reader) ([]*rdf.Triple, error) {
	var newTriples []*rdf.Triple

	dec := xmlrdf.NewTripleDecoder(r, xmlrdf.RDFXML)

	triples, err := dec.DecodeAll()
	if err != nil {
		return newTriples, err
	}

	for _, t := range triples {
		newT, err := convertTriple(t)
		if err != nil {
			return newTriples, err
		}

		newTriples = append(newTriples, newT)
	}

	return newTriples, nil
}

// convertTriple converts a knakk/rdf Triple to a kiivihal/rdf2go Triple
func convertTriple(triple xmlrdf.Triple) (*rdf.Triple, error) {
	var (
		s   rdf.Subject
		err error
	)

	switch triple.Subj.Type() {
	case xmlrdf.TermBlank:
		s, err = rdf.NewBlankNode(triple.Subj.String())
		if err != nil {
			return nil, err
		}
	default:
		s, err = rdf.NewIRI(triple.Subj.String())
		if err != nil {
			return nil, err
		}
	}

	p, err := rdf.NewIRI(triple.Pred.String())
	if err != nil {
		return nil, err
	}

	var o rdf.Object

	switch triple.Obj.Type() {
	case xmlrdf.TermBlank:
		o, err = rdf.NewBlankNode(triple.Obj.String())
		if err != nil {
			return nil, err
		}
	case xmlrdf.TermLiteral:
		l := triple.Obj.(xmlrdf.Literal)
		if l.Lang() != "" {
			o, err = rdf.NewLiteralWithLang(l.String(), l.Lang())
			if err != nil {
				return nil, err
			}

			break
		}

		xsdString := "http://www.w3.org/2001/XMLSchema#string"
		if l.DataType.String() != "" && l.DataType.String() != xsdString {
			dt, iriErr := rdf.NewIRI(l.DataType.String())
			if iriErr != nil {
				return nil, iriErr
			}

			o, err = rdf.NewLiteralWithType(l.String(), dt)
			if err != nil {
				return nil, err
			}

			break
		}

		o, err = rdf.NewLiteral(triple.Obj.String())
		if err != nil {
			return nil, err
		}

	case xmlrdf.TermIRI:
		o, err = rdf.NewIRI(triple.Obj.String())
		if err != nil {
			return nil, err
		}
	}

	return rdf.NewTriple(s, p, o), nil
}
