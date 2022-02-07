package resource

import (
	"fmt"
)

type Option func(obj *Literal) error

// Resource contains all the predicates linked to a Subject
type Resource struct {
	subject          Subject
	predicates       map[Predicate]*resourcePredicate
	errors           []error
	PredicateURIBase *IRIBuilder `json:"-"`
	inFatalError     bool
}

func NewResource(subject Subject) *Resource {
	return &Resource{
		subject:    subject,
		predicates: map[Predicate]*resourcePredicate{},
	}
}

func (r *Resource) Triples() []*Triple {
	triples := []*Triple{}

	for p, objects := range r.predicates {
		for _, obj := range objects.objects {
			triples = append(triples, NewTriple(r.subject, p, obj))
		}
	}

	return triples
}

func (r *Resource) AddSimpleLiteral(predicateLabel, value string, options ...Option) {
	if r.inFatalError {
		// keep running but ignore everything
		return
	}

	if value == "" {
		// ignore empty values
		return
	}

	if r.PredicateURIBase == nil {
		r.addError(fmt.Errorf("resource.PredicateBase must not be nil"))
		r.inFatalError = true

		return
	}

	p, err := r.PredicateURIBase.IRI(predicateLabel)
	if err != nil {
		r.addError(err)
		return
	}

	rp, present := r.predicates[p]
	if !present {
		rp = &resourcePredicate{
			iri:     &p,
			objects: map[hasher]Object{},
		}
	}

	l, err := NewLiteral(value)
	if err != nil {
		r.addError(err)
		return
	}

	for _, option := range options {
		if err := option(&l); err != nil {
			r.addError(err)
			return
		}
	}

	if v := l.Validate(); !v.Valid() {
		r.addError(v.ErrorOrNil())
		return
	}

	h := getHash(l)
	if _, ok := rp.objects[h]; !ok {
		rp.objects[h] = l

		r.predicates[p] = rp
	}
}

func (r *Resource) HasErrors() bool {
	return len(r.errors) > 0
}

func (r *Resource) addError(err error) {
	r.errors = append(r.errors, err)
}

// GetType is normal call to predicate map

type resourcePredicate struct {
	iri     *IRI
	objects map[hasher]Object
}

// WithDataType is an Option to set a DataType on AddSimpleLiteral()
func WithDataType(dt *IRI) Option {
	return func(obj *Literal) error {
		obj.DataType = dt
		return nil
	}
}

// WithLanguage is an Option to set a Language on AddSimpleLiteral()
func WithLanguage(lang string) Option {
	return func(obj *Literal) error {
		obj.lang = lang
		return nil
	}
}
