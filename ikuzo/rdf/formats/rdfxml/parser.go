package rdfxml

import (
	"errors"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
	irdf "github.com/delving/hub3/ikuzo/rdf/formats/rdfxml/internal/rdf"
)

func Parse(r io.Reader, g *rdf.Graph, seed string) (*rdf.Graph, error) {
	if g == nil {
		g = rdf.NewGraph()
	}

	triples, err := decodeRDFXML(r, seed)
	if err != nil {
		return nil, err
	}

	g.Add(triples...)

	return g, nil
}

// decodeRDFXML parses RDF-XML into triples
func decodeRDFXML(r io.Reader, seed string) ([]*rdf.Triple, error) {
	var newTriples []*rdf.Triple

	dec := irdf.NewRDFXMLDecoder(r, irdf.WithSeed(seed), irdf.WithIDStrategy(irdf.DeterministicIDs))

	triples, err := dec.DecodeAll()
	if err != nil {
		return newTriples, err
	}

	for _, t := range triples {
		newT, err := convertTriple(t)
		if err != nil {
			if errors.Is(err, rdf.ErrEmptyLiteral) {
				continue
			}
			return newTriples, err
		}

		newTriples = append(newTriples, newT)
	}

	return newTriples, nil
}

// convertTriple converts a knakk/rdf Triple to a kiivihal/rdf2go Triple
func convertTriple(triple irdf.Triple) (*rdf.Triple, error) {
	var (
		s   rdf.Subject
		err error
	)

	switch triple.Subj.Type() {
	case irdf.TermBlank:
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
	case irdf.TermBlank:
		o, err = rdf.NewBlankNode(triple.Obj.String())
		if err != nil {
			return nil, err
		}
	case irdf.TermLiteral:
		l := triple.Obj.(irdf.Literal)
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

	case irdf.TermIRI:
		o, err = rdf.NewIRI(triple.Obj.String())
		if err != nil {
			return nil, err
		}
	}

	return rdf.NewTriple(s, p, o), nil
}
