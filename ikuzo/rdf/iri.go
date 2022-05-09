package rdf

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/delving/hub3/ikuzo/validator"
)

var (
	_ Subject   = (*IRI)(nil)
	_ Predicate = (*IRI)(nil)
	_ Object    = (*IRI)(nil)
)

// IRI is an URI / IRI reference.
type IRI struct {
	str string
}

// NewIRI returns a new IRI, or an error if it's not valid.
//
// A valid IRI cannot be empty, or contain any of the disallowed characters: [\x00-\x20<>"{}|^`\].
func NewIRI(iri string) (IRI, error) {
	uri := IRI{str: iri}

	v := uri.Validate()
	if !v.Valid() {
		return IRI{}, v.ErrorOrNil()
	}

	return uri, nil
}

// Equal returns whether this resource is equal to another.
func (u IRI) Equal(other Term) bool {
	if spec, ok := other.(*IRI); ok {
		return u.str == spec.str
	}

	// support non-pointer as well
	if spec, ok := other.(IRI); ok {
		return u.str == spec.str
	}

	return false
}

// RawValue returns the string value of the a resource without brackets.
func (u IRI) RawValue() (str string) {
	return u.str
}

// String returns the NTriples representation of this resource.
func (u IRI) String() (str string) {
	return fmt.Sprintf("<%s>", u.str)
}

// Type returns the TermType of a IRI.
func (u IRI) Type() TermType {
	return TermIRI
}

func (u IRI) Validate() *validator.Validator {
	v := validator.New()

	// http://www.ietf.org/rfc/rfc3987.txt
	v.Check(strings.TrimSpace(u.str) != "", "invalid id", ErrEmptyIRI, "")

	for _, r := range u.str {
		if r >= '\x00' && r <= '\x20' {
			v.AddError("bad character", fmt.Errorf("%w: %q", ErrDisallowedCharacterInIRI, r), "")

			// only get the first one
			return v
		}

		switch r {
		case '<', '>', '"', '{', '}', '|', '^', '`', '\\':
			v.AddError("invalid character", fmt.Errorf("%w: %q", ErrDisallowedCharacterInIRI, r), "")

			// only get the first one
			return v
		}
	}

	return v
}

// Split returns the prefix and suffix of the IRI string, splitted at the first
// '/' or '#' character, in reverse order of the string.
//
// When the IRI can't be split, both the prefix and suffix are returned empty
func (u IRI) Split() (prefix, suffix string) {
	i := len(u.str)
	for i > 0 {
		r, w := utf8.DecodeLastRuneInString(u.str[0:i])
		if r == '/' || r == '#' {
			prefix, suffix = u.str[0:i], u.str[i:len(u.str)]
			break
		}

		i -= w
	}

	return prefix, suffix
}

// ValidAsSubject is a placeholder to verify that it can be used as a Subject.
func (u IRI) ValidAsSubject() {}

// ValidAsPredicate is a placeholder to verify that it can be used as a Predicate .
func (u IRI) ValidAsPredicate() {}

// ValidAsObject is a placeholder to verify that it can be used as an Object .
func (u IRI) ValidAsObject() {}
