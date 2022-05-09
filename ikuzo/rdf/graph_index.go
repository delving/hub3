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

func (gi *GraphIndex) update(t *Triple) error {
	gi.updateSubject(t)
	gi.updatePredicate(t)
	gi.updateRDFType(t)

	if err := gi.updateObject(t); err != nil {
		return err
	}

	return nil
}

func (gi *GraphIndex) updateNamespaceURI(iri IRI) {
	prefix, _ := iri.Split()

	count, ok := gi.NamespacesURIs[prefix]
	if !ok {
		count = 0
	}
	count++
	gi.NamespacesURIs[prefix] = count
}

func (gi *GraphIndex) updateSubject(t *Triple) {
	// switch term := t.Subject.(type) {
	// case *IRI:
	// gi.updateNamespaceURI(term)
	// }
	s := getHash(t.Subject)

	count, ok := gi.Subjects[s]
	if !ok {
		count = 0
	}
	count++
	gi.Subjects[s] = count
}

func (gi *GraphIndex) updatePredicate(t *Triple) {
	switch term := t.Predicate.(type) {
	case *IRI:
		gi.updateNamespaceURI(*term)
	case IRI:
		gi.updateNamespaceURI(term)
	}

	p := getHash(t.Predicate)

	count, ok := gi.Predicates[p]
	if !ok {
		count = 0
	}
	count++
	gi.Predicates[p] = count
}

func (gi *GraphIndex) updateRDFType(t *Triple) {
	if t.Predicate.RawValue() == RDFType {
		o := getHash(t.Object)

		count, ok := gi.RDFTypes[o]
		if !ok {
			count = 0
		}
		count++
		gi.RDFTypes[o] = count
	}
}

func (gi *GraphIndex) updateObject(t *Triple) error {
	switch t.Object.Type() {
	case TermBlankNode, TermIRI:
		o := getHash(t.Object)

		count, ok := gi.ObjectResources[o]
		if !ok {
			count = 0
		}
		count++
		gi.ObjectResources[o] = count
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
			count++
			gi.Languages[l] = count
		}

		if o.DataType.Equal(IRI{}) {
			dt := getHash(o.DataType)

			count, ok := gi.DataTypes[dt]
			if !ok {
				count = 0
			}
			count++
			gi.DataTypes[dt] = count
		}
	}

	return nil
}
