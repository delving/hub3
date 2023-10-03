package rdf

import (
	"fmt"
	"strings"

	"github.com/kiivihal/rdf2go"
)

const (
	RDFType            = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
	RDFCollectionFirst = "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"
	RDFCollectionRest  = "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"
	RDFCollectionNil   = "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"
)

var IsA = Predicate(IRI{str: RDFType})

// Subject interface distiguishes which Terms are valid as a Subject of a Triple.
type Subject interface {
	Term
	ValidAsSubject()
}

// Predicate interface distiguishes which Terms are valid as a Predicate of a Triple.
type Predicate interface {
	Term
	ValidAsPredicate()
}

// Object interface distiguishes which Terms are valid as a Object of a Triple.
type Object interface {
	Term
	ValidAsObject()
}

// Triple represents a RDF triple.
type Triple struct {
	Subject   Subject
	Predicate Predicate
	Object    Object
}

// TODO(kiivihal): add triple validator

// NewTriple returns a new triple with the given subject, predicate and object.
func NewTriple(subject Subject, predicate Predicate, object Object) (triple *Triple) {
	return &Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}
}

// ID returns a content-based ID.
// Each triple has a unique identifier that can be used to check for uniqueness or used
// for dedupliplication.
//
// BlankNodes get a hash not on the BlankNode ID but on a combination of the resource and Predicate
// that point to the BlankNode. This ensures that the ID remains the same regardless of the BlankNode ID.
func (triple Triple) ID() string {
	var sb strings.Builder
	return sb.String()
}

// Equal returns this triple is equivalent to the argument.
func (triple Triple) Equal(other *Triple) bool {
	return triple.Subject.Equal(other.Subject) &&
		triple.Predicate.Equal(other.Predicate) &&
		triple.Object.Equal(other.Object)
}

// String returns the NTriples representation of this triple.
func (triple Triple) String() string {
	var subj string
	if triple.Subject != nil {
		subj = triple.Subject.String()
	}

	var pred string
	if triple.Predicate != nil {
		pred = triple.Predicate.String()
	}

	var obj string
	if triple.Object != nil {
		obj = triple.Object.String()
	}

	return fmt.Sprintf("%s %s %s .", subj, pred, obj)
}

// needed for refactor remove later
func (triple Triple) GetRDFType() (string, bool) {
	switch triple.Predicate.RawValue() {
	case RDFType:
		return triple.Object.RawValue(), true
	default:
		return "", false
	}
}

// asLegacyTriple converts a rdf.Triple to legacy package Triple
//
// NOTE: This function should be removed when the rdf2go package is
// no longer used
func (triple Triple) asLegacyTriple() (*rdf2go.Triple, error) {
	var s, p, o rdf2go.Term
	switch subj := triple.Subject; subj.Type() {
	case TermBlankNode:
		s = rdf2go.NewBlankNode(subj.RawValue())
	default:
		s = rdf2go.NewResource(subj.RawValue())
	}

	switch pred := triple.Predicate; pred.Type() {
	case TermBlankNode:
		p = rdf2go.NewBlankNode(pred.RawValue())
	default:
		p = rdf2go.NewResource(pred.RawValue())
	}

	switch obj := triple.Object; obj.Type() {
	case TermBlankNode:
		o = rdf2go.NewBlankNode(obj.RawValue())
	case TermIRI:
		o = rdf2go.NewResource(obj.RawValue())
	case TermLiteral:
		lit, ok := obj.(Literal)
		if !ok {
			return nil, fmt.Errorf("unable to convert literal")
		}
		switch {
		case lit.Lang() != "":
			o = rdf2go.NewLiteralWithLanguage(lit.RawValue(), lit.Lang())
		case lit.DataType.RawValue() != "":
			o = rdf2go.NewLiteralWithDatatype(lit.RawValue(), rdf2go.NewResource(lit.DataType.RawValue()))
		default:
			o = rdf2go.NewLiteral(lit.RawValue())
		}
	}

	newTriple := rdf2go.NewTriple(s, p, o)
	return newTriple, nil
}
