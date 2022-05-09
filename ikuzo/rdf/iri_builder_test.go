package rdf

import (
	"testing"

	"github.com/matryer/is"
)

func TestBuilder(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	base, err := NewIRI("http://www.w3.org/2004/02/skos/core#")
	is.NoErr(err)

	builder := NewIRIBuilder(base)
	is.True(base.Equal(builder.baseIRI)) // same baseIRI should be set

	skos, err := builder.IRI("Concept")
	is.NoErr(err)
	is.Equal(skos.RawValue(), "http://www.w3.org/2004/02/skos/core#Concept")

	invalidLabel, err := builder.IRI("Concept/123")
	is.True(err != nil)
	is.Equal(invalidLabel, IRI{})

	invalidHashLabel, err := builder.IRI("Concept#123")
	is.True(err != nil)
	is.Equal(invalidHashLabel, IRI{})

	invalidSpaceLabel, err := builder.IRI("Concept label")
	is.True(err != nil)
	is.Equal(invalidSpaceLabel, IRI{})
}
