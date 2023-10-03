package rdf

import (
	"fmt"
	"sync/atomic"
)

// GraphStats returns counts for unique values in Graph
type GraphStats struct {
	Languages      int
	ObjectIRIs     int
	ObjectLiterals int
	Predicates     int
	Resources      int
	Namespaces     int
	Triples        uint64
}

func (gs *GraphStats) incTriples() {
	atomic.AddUint64(&gs.Triples, 1)
}

func (gs *GraphStats) decrTriples() {
	atomic.AddUint64(&gs.Triples, ^uint64(0))
}

// NOTE: only used for existence checks
type GraphIndex struct {
	Subjects        map[hasher]uint64
	Predicates      map[hasher]uint64
	RDFTypes        map[hasher]uint64
	ObjectResources map[hasher]uint64
	Languages       map[hasher]uint64
	DataTypes       map[hasher]uint64
	NamespacesURIs  map[string]uint64
}

func newIndex() *GraphIndex {
	return &GraphIndex{
		Subjects:        map[hasher]uint64{},
		Predicates:      map[hasher]uint64{},
		RDFTypes:        map[hasher]uint64{},
		ObjectResources: map[hasher]uint64{},
		Languages:       map[hasher]uint64{},
		DataTypes:       map[hasher]uint64{},
		NamespacesURIs:  map[string]uint64{},
	}
}

func (gi *GraphIndex) update(t *Triple, remove bool) error {
	gi.updateSubject(t, remove)
	gi.updatePredicate(t, remove)
	gi.updateRDFType(t, remove)

	if err := gi.updateObject(t, remove); err != nil {
		return err
	}

	return nil
}

func (gi *GraphIndex) updateNamespaceURI(iri IRI, remove bool) {
	prefix, _ := iri.Split()

	count, ok := gi.NamespacesURIs[prefix]
	if !ok {
		count = 0
	}

	gi.NamespacesURIs[prefix] = getCount(count, remove)
}

func (gi *GraphIndex) updateSubject(t *Triple, remove bool) {
	// switch term := t.Subject.(type) {
	// case *IRI:
	// gi.updateNamespaceURI(term)
	// }
	s := getHash(t.Subject)

	count, ok := gi.Subjects[s]
	if !ok {
		count = 0
	}

	gi.Subjects[s] = getCount(count, remove)
}

func getCount(seed uint64, remove bool) uint64 {
	if remove {
		return seed - 1
	}
	return seed + 1
}

func (gi *GraphIndex) updatePredicate(t *Triple, remove bool) {
	switch term := t.Predicate.(type) {
	case *IRI:
		gi.updateNamespaceURI(*term, remove)
	case IRI:
		gi.updateNamespaceURI(term, remove)
	}

	p := getHash(t.Predicate)

	count, ok := gi.Predicates[p]
	if !ok {
		count = 0
	}

	gi.Predicates[p] = getCount(count, remove)
}

func (gi *GraphIndex) updateRDFType(t *Triple, remove bool) {
	if t.Predicate.RawValue() == RDFType {
		switch term := t.Object.(type) {
		case *IRI:
			gi.updateNamespaceURI(*term, remove)
		case IRI:
			gi.updateNamespaceURI(term, remove)
		}
		o := getHash(t.Object)

		count, ok := gi.RDFTypes[o]
		if !ok {
			count = 0
		}

		gi.RDFTypes[o] = getCount(count, remove)
	}
}

func (gi *GraphIndex) updateObject(t *Triple, remove bool) error {
	switch t.Object.Type() {
	case TermBlankNode, TermIRI:
		o := getHash(t.Object)

		count, ok := gi.ObjectResources[o]
		if !ok {
			count = 0
		}
		gi.ObjectResources[o] = getCount(count, remove)
	case TermLiteral:
		o, ok := t.Object.(Literal)
		if !ok {
			return fmt.Errorf("invalid literal: %#v", t.Object)
		}

		if o.Lang() != "" {
			l := hasher(hash(o.Lang()))

			count, ok := gi.Languages[l]
			if !ok {
				count = 0
			}
			gi.Languages[l] = getCount(count, remove)
		}

		if o.DataType.Equal(IRI{}) {
			dt := getHash(o.DataType)

			count, ok := gi.DataTypes[dt]
			if !ok {
				count = 0
			}
			gi.DataTypes[dt] = getCount(count, remove)
		}
	}

	return nil
}
