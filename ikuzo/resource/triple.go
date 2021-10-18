package resource

import "fmt"

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

// NewTriple returns a new triple with the given subject, predicate and object.
func NewTriple(subject Subject, predicate Predicate, object Object) (triple *Triple) {
	return &Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}
}

// Equal returns this triple is equivalent to the argument.
func (triple Triple) Equal(other *Triple) bool {
	return triple.Subject.Equal(other.Subject) &&
		triple.Predicate.Equal(other.Predicate) &&
		triple.Object.Equal(other.Object)
}

// String returns the NTriples representation of this triple.
func (triple Triple) String() (str string) {
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
