package rdf

import (
	"fmt"
)

type Option func(obj *Literal) error

// Resource contains all the predicates linked to a Subject
type Resource struct {
	subject          Subject
	types            []IRI
	predicateLookup  map[Predicate]*resourcePredicate
	predicates       []*resourcePredicate
	errors           []error
	PredicateURIBase *IRIBuilder `json:"-"`
	inFatalError     bool
	order            int
}

func NewResource(subject Subject) *Resource {
	return &Resource{
		subject:         subject,
		predicateLookup: map[Predicate]*resourcePredicate{},
	}
}

func (r *Resource) Subject() Subject {
	return r.subject
}

func (r *Resource) Label() (Literal, bool) {
	l, _ := RDFS.IRI("label")

	o, ok := r.predicateLookup[Predicate(l)]
	if ok {
		for _, obj := range o.Objects() {
			if obj.Type() == TermLiteral {
				return obj.(Literal), true
			}
		}
	}

	return Literal{}, false
}

func (r *Resource) Types() []IRI {
	if len(r.types) == 0 {
		iri, _ := RDF.IRI("Description")

		return []IRI{iri}
	}

	return r.types
}

func (r *Resource) Predicates() map[Predicate]*resourcePredicate {
	return r.predicateLookup
}

func (r *Resource) SortedPredicates() []*resourcePredicate {
	return r.predicates
}

func (r *Resource) Triples() []*Triple {
	triples := []*Triple{}

	for _, predicate := range r.predicates {
		for _, obj := range predicate.objects {
			triples = append(triples, NewTriple(r.subject, predicate.IRI(), obj))
		}
	}

	return triples
}

func (r *Resource) Add(t *Triple) {
	rp, present := r.predicateLookup[t.Predicate]
	if !present {
		rp = &resourcePredicate{
			iri:          t.Predicate.(IRI),
			objectLookup: map[hasher]Object{},
			objects:      []Object{},
		}
		r.predicates = append(r.predicates, rp)
	}

	if rp.iri.Equal(IsA) {
		r.types = append(r.types, t.Object.(IRI))
	}

	h := getHash(t.Object)
	if _, ok := rp.objectLookup[h]; !ok {
		rp.objectLookup[h] = t.Object
		rp.objects = append(rp.objects, t.Object)

		r.predicateLookup[t.Predicate] = rp
	}
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

	rp, present := r.predicateLookup[p]
	if !present {
		rp = &resourcePredicate{
			iri:          p,
			objectLookup: map[hasher]Object{},
			objects:      []Object{},
		}
		r.predicates = append(r.predicates, rp)
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
	if _, ok := rp.objectLookup[h]; !ok {
		rp.objectLookup[h] = l
		rp.objects = append(rp.objects, l)

		r.predicateLookup[p] = rp
	}
}

func (r *Resource) HasErrors() bool {
	return len(r.errors) > 0
}

func (r *Resource) addError(err error) {
	r.errors = append(r.errors, err)
}

// Order returns the order of the Resource based on the insertion order
func (r *Resource) Order() int {
	return r.order
}

type resourcePredicate struct {
	iri          IRI
	objectLookup map[hasher]Object
	objects      []Object
}

func (rp *resourcePredicate) IRI() IRI {
	return rp.iri
}

func (rp *resourcePredicate) Objects() (objects []Object) {
	return rp.objects
}

// WithDataType is an Option to set a DataType on AddSimpleLiteral()
func WithDataType(dt IRI) Option {
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
