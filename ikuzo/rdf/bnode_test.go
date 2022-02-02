package rdf

import (
	"errors"
	"testing"

	"github.com/delving/hub3/ikuzo/validator"
	"github.com/matryer/is"
)

func TestNewBlankNode(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	bnode, err := NewBlankNode("")
	is.True(errors.Is(err, ErrEmptyBlankNode))
	is.Equal(bnode.RawValue(), "")
	is.Equal(bnode.String(), "")

	bnode, err = NewBlankNode("  ")
	is.True(errors.Is(err, ErrEmptyBlankNode))
	is.Equal(bnode.RawValue(), "")
	is.Equal(bnode.String(), "")

	bnode, err = NewBlankNode("  123\n")
	is.NoErr(err)
	is.Equal(bnode.String(), "_:123")
	is.Equal(bnode.RawValue(), "123")

	is.Equal(bnode.Type(), TermBlankNode)
}

func TestBlankNode_Equal(t *testing.T) {
	//nolint:gocritic
	is := is.New(t)

	iri, err := NewBlankNode("123")
	is.NoErr(err)
	is.Equal(iri.RawValue(), "123") // RawValue() should return str
	is.Equal(iri.id, "_:123")       // str should mirror input
	is.Equal(iri.String(), "_:123") // str should mirror input

	other, err := NewBlankNode("123")
	is.NoErr(err)
	is.True(other.Equal(iri))
	is.True(iri.Equal(other))

	// test as pointers
	otherPtr := &other
	iriPtr := &iri
	is.True(otherPtr.Equal(iriPtr))
	is.True(iriPtr.Equal(otherPtr))

	noMatch, err := NewBlankNode("urn:1234")
	is.NoErr(err)
	is.True(!noMatch.Equal(iri))
	is.True(!iri.Equal(noMatch))

	// equal returns false if other is not an BlankNode
	is.True(!iri.Equal(nonBlankNode{str: "urn123"}))
}

type nonBlankNode struct {
	str string
}

func (n nonBlankNode) String() string                 { return n.str }
func (n nonBlankNode) RawValue() string               { return n.str }
func (n nonBlankNode) Equal(Term) bool                { return false }
func (n nonBlankNode) Type() TermType                 { return TermBlankNode }
func (n nonBlankNode) Validate() *validator.Validator { return nil }
