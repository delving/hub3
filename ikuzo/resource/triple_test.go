package resource

import (
	"testing"

	"github.com/matryer/is"
)

func TestTriple(t *testing.T) {
	t.Run("NewTriple", func(t *testing.T) {
		// nolint: gocritic
		is := is.New(t)

		s, err := NewIRI("urn:s/123")
		is.NoErr(err)

		p, err := DC.IRI("subject")
		is.NoErr(err)

		o, err := NewLiteralWithLang("some text", "en")
		is.NoErr(err)

		triple := NewTriple(s, p, o)
		is.True(triple.Subject.Equal(s))
		is.True(triple.Predicate.Equal(p))
		is.True(triple.Object.Equal(o))

		is.Equal(triple.String(), `<urn:s/123> <http://purl.org/dc/elements/1.1/subject> "some text"@en .`)

		otherO, err := NewLiteralWithLang("wat tekst", "nl")
		is.NoErr(err)

		other := NewTriple(s, p, otherO)
		is.True(!triple.Equal(other))
	})
}
