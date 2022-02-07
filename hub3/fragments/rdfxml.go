// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

import (
	"io"

	r "github.com/kiivihal/rdf2go"
	"github.com/knakk/rdf"
)

// DecodeRDFXML parses RDF-XML into triples
func DecodeRDFXML(r io.Reader) ([]rdf.Triple, error) {
	dec := rdf.NewTripleDecoder(r, rdf.RDFXML)
	return dec.DecodeAll()
}

// NewResourceMapFromXML creates a resource map from the triples
func NewResourceMapFromXML(orgID string, triples []rdf.Triple) (*ResourceMap, error) {
	rm := NewEmptyResourceMap(orgID)
	for idx, triple := range triples {
		newTriple := ConvertTriple(triple)
		err := rm.AppendOrderedTriple(newTriple, false, idx)
		if err != nil {
			return nil, err
		}

	}

	return rm, nil
}

// ConvertTriple converts a knakk/rdf Triple to a kiivihal/rdf2go Triple
func ConvertTriple(triple rdf.Triple) *r.Triple {
	var s r.Term
	switch triple.Subj.Type() {
	case rdf.TermBlank:
		s = r.NewBlankNode(triple.Subj.String())
	default:
		s = r.NewResource(triple.Subj.String())
	}

	p := r.NewResource(triple.Pred.String())

	var o r.Term
	switch triple.Obj.Type() {
	case rdf.TermBlank:
		o = r.NewBlankNode(triple.Obj.String())
	case rdf.TermLiteral:
		l := triple.Obj.(rdf.Literal)
		if l.Lang() != "" {
			o = r.NewLiteralWithLanguage(l.String(), l.Lang())
			break
		}
		xsdString := "http://www.w3.org/2001/XMLSchema#string"
		if l.DataType.String() != "" && l.DataType.String() != xsdString {
			o = r.NewLiteralWithDatatype(l.String(), r.NewResource(l.DataType.String()))
			break
		}
		o = r.NewLiteral(triple.Obj.String())
	case rdf.TermIRI:
		o = r.NewResource(triple.Obj.String())
	}

	return r.NewTriple(s, p, o)
}
