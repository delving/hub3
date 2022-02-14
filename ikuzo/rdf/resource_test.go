package rdf

import (
	"testing"

	"github.com/matryer/is"
)

// nolint:gocritic
func TestResource(t *testing.T) {
	is := is.New(t)
	subject, err := NewIRI("urn:subject/123")
	is.NoErr(err)
	is.True(subject.String() != "")

	t.Run("NewResource", func(t *testing.T) {
		is = is.New(t)
		rsc := NewResource(&subject)
		is.True(rsc.subject == &subject)
		is.Equal(len(rsc.predicates), 0)
	})

	t.Run("Add triple", func(t *testing.T) {
		is = is.New(t)

		p, err := DC.IRI("title")
		is.NoErr(err)

		o, err := NewLiteralWithLang("test", "nl")
		is.NoErr(err)

		triple := NewTriple(Subject(subject), Predicate(p), Object(o))
		rsc := NewResource(&subject)
		rsc.Add(triple)
		is.Equal(len(rsc.predicates), 1)
	})

	t.Run("AddSimpleLiteral() without PredicateURIBase", func(t *testing.T) {
		is = is.New(t)
		rsc := NewResource(&subject)

		rsc.AddSimpleLiteral("title", "simple title")
		is.True(rsc.HasErrors())
		is.Equal(len(rsc.predicates), 0)
	})

	t.Run("AddSimpleLiteral() with PredicateURIBase", func(t *testing.T) {
		is = is.New(t)

		rsc := NewResource(&subject)
		rsc.PredicateURIBase = DC
		rsc.AddSimpleLiteral("title", "simple title")
		is.True(!rsc.HasErrors())
		is.Equal(len(rsc.predicates), 1)
	})

	t.Run("AddSimpleLiteral() with language option", func(t *testing.T) {
		is = is.New(t)

		rsc := NewResource(&subject)
		rsc.PredicateURIBase = DC
		rsc.AddSimpleLiteral("title", "simple title", WithLanguage("nl"))
		is.True(!rsc.HasErrors())
		is.Equal(len(rsc.predicates), 1)
		triples := rsc.Triples()
		is.Equal(len(triples), 1)
		is.Equal(triples[0].String(), "<urn:subject/123> <http://purl.org/dc/elements/1.1/title> \"simple title\"@nl .")
	})

	t.Run("AddSimpleLiteral() with DataType option", func(t *testing.T) {
		is = is.New(t)

		rsc := NewResource(&subject)
		rsc.PredicateURIBase = DC
		dt, err := XSD.IRI("dateTime")
		is.NoErr(err)
		rsc.AddSimpleLiteral("title", "simple title", WithDataType(dt))
		// no duplicates are stored
		rsc.AddSimpleLiteral("title", "simple title", WithDataType(dt))
		for _, err := range rsc.errors {
			t.Logf("errors: %s", err)
		}
		is.True(!rsc.HasErrors())
		is.Equal(len(rsc.predicates), 1)
		triples := rsc.Triples()
		is.Equal(len(triples), 1)
		is.Equal(
			triples[0].String(),
			"<urn:subject/123> <http://purl.org/dc/elements/1.1/title> \"simple title\"^^<http://www.w3.org/2001/XMLSchema#dateTime> .",
		)
	})
}
