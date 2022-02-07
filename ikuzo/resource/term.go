package resource

import "github.com/delving/hub3/ikuzo/validator"

// TermType describes the type of RDF term: Blank node, IRI or Literal
type TermType int

// Exported RDF term types.
const (
	TermBlankNode TermType = iota
	TermIRI
	TermLiteral
)

// A Term is the value of a subject, predicate or object,  i.e. a IRI reference, BlankNode or
// Literal.
//
// To work with the underlying concrete type,  use a type assertion or a type switch.
//
//	  t, ok := term.(IRI)
//
type Term interface {
	// Equal returns whether this term is equal to another.
	Equal(Term) bool

	// RawValue returns the raw value of this term.
	RawValue() string

	// String returns the NTriples representation of this term.
	String() string

	// Type returns the Term type.
	Type() TermType

	// Validate returns is the Term is valid
	Validate() *validator.Validator
	// TODO(kiivihal): maybe later change with validator
}
