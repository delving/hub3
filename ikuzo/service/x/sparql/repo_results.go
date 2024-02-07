package sparql

import (
	"encoding/json"
	fmt "fmt"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
)

type TermType string

const (
	TypeURI          TermType = "uri"
	TypeBnode        TermType = "bnode"
	TypeLiteral      TermType = "literal"
	TypeTypedLiteral TermType = "typed-literal"
)

// Results holds the parsed results of a application/sparql-results+json response.
type Results struct {
	Head    Header
	Results results
}

type Header struct {
	Link []string
	Vars []string
}

type results struct {
	Distinct bool
	Ordered  bool
	Bindings []map[string]*Entry
}

type Entry struct {
	Type     TermType
	XMLLang  string `json:"xml:lang"`
	Value    string
	DataType string
}

func (e *Entry) asSubject() (rdf.Subject, error) {
	switch e.Type {
	case TypeURI:
		return rdf.NewIRI(e.Value)
	case TypeBnode:
		return rdf.NewBlankNode(e.Value)
	}

	return nil, fmt.Errorf("entry is invalid subject: %#v", e)
}

func (e *Entry) asPredicate() (rdf.Predicate, error) {
	if e.Type != TypeURI {
		return nil, fmt.Errorf("invalid entry.Type for predicate: %#v", e)
	}

	return rdf.NewIRI(e.Value)
}

func (e *Entry) asObject() (rdf.Object, error) {
	switch e.Type {
	case TypeURI:
		return rdf.NewIRI(e.Value)
	case TypeBnode:
		return rdf.NewBlankNode(e.Value)
	case TypeLiteral, TypeTypedLiteral:
		switch {
		case e.DataType != "":
			dt, err := rdf.NewIRI(e.DataType)
			if err != nil {
				return nil, err
			}

			return rdf.NewLiteralWithType(e.Value, dt)
		case e.XMLLang != "":
			return rdf.NewLiteralWithLang(e.Value, e.XMLLang)
		default:
			return rdf.NewLiteral(e.Value)
		}
	}

	return nil, fmt.Errorf("invalid entry.Type for object: %#v", e)
}

// parseJSON takes an application/sparql-results+json response and parses it
// into a Results struct.
func parseJSON(r io.Reader) (*Results, error) {
	var res Results
	err := json.NewDecoder(r).Decode(&res)

	return &res, err
}

// Bindings returns a map of the bound variables in the SPARQL response, where
// each variable points to one or more RDF terms.
func (r *Results) Bindings() map[string][]rdf.Term {
	rb := make(map[string][]rdf.Term)

	for _, v := range r.Head.Vars {
		for _, b := range r.Results.Bindings {
			t, err := b[v].asObject()
			if err == nil {
				rb[v] = append(rb[v], t)
			}
		}
	}

	return rb
}

// Solutions returns a slice of the query solutions, each containing a map
// of all bindings to RDF terms.
func (r *Results) Solutions() []map[string]rdf.Term {
	var rs []map[string]rdf.Term

	for _, s := range r.Results.Bindings {
		solution := make(map[string]rdf.Term)

		for k, v := range s {
			term, err := v.asObject()
			if err == nil {
				solution[k] = term
			}
		}

		rs = append(rs, solution)
	}

	return rs
}
