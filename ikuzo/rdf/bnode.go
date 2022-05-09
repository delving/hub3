package rdf

import (
	"strings"

	"github.com/delving/hub3/ikuzo/validator"
)

// BlankNode represents a RDF blank node; an unqualified IRI with identified by a label.
type BlankNode struct {
	id string
}

// NewBlank returns a new blank node with a given label. It returns
// an error only if the supplied label is blank.
//
// Leading and trailing whitespace is trimmed from the id.
func NewBlankNode(id string) (BlankNode, error) {
	bnode := BlankNode{id: "_:" + strings.TrimSpace(id)}

	v := bnode.Validate()
	if !v.Valid() {
		return BlankNode{}, v.ErrorOrNil()
	}

	return bnode, nil
}

// Equal returns whether this blank node is equivalent to another.
func (b BlankNode) Equal(other Term) bool {
	if spec, ok := other.(*BlankNode); ok {
		return b.id == spec.id
	}

	if spec, ok := other.(BlankNode); ok {
		return b.id == spec.id
	}

	return false
}

// RawValue returns the Blank node label
func (b BlankNode) RawValue() string {
	if len(b.id) <= 2 {
		return ""
	}

	return b.id[2:]
}

// String returns the NTriples representation of the blank node.
func (b BlankNode) String() string {
	return b.id
}

// Type returns the TermType of a blank node.
func (b BlankNode) Type() TermType {
	return TermBlankNode
}

func (b BlankNode) Validate() *validator.Validator {
	v := validator.New()
	v.Check(len(b.id) > 2, "bnode", ErrEmptyBlankNode, "")

	return v
}

// ValidAsSubject denotes that a Blank node is valid as a Triple's Subject.
func (b BlankNode) ValidAsSubject() {}

// ValidAsObject denotes that a Blank node is valid as a Triple's Object.
func (b BlankNode) ValidAsObject() {}
