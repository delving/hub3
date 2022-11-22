package hextuples

import (
	"errors"

	"github.com/delving/hub3/ikuzo/rdf"
)

type Graph struct {
	tuples    []HexTuple
	resources map[string]*Resource
	// subject     string
	// subjectType string
	seen map[string]bool
}

func NewGraph() *Graph {
	return &Graph{
		tuples:    []HexTuple{},
		resources: map[string]*Resource{},
		seen:      map[string]bool{},
	}
}

func (g *Graph) Resource(subject string, withInlining bool) (*Resource, error) {
	rsc, ok := g.resources[subject]
	if !ok {
		return rsc, ErrResourceNotFound
	}

	if withInlining {
		err := g.inlineEntries(rsc, map[string]bool{})
		if err != nil {
			return nil, err
		}
	}

	return rsc, nil
}

func (g *Graph) inlineEntries(rsc *Resource, seen map[string]bool) error {
	for _, entries := range rsc.Predicates {
		for _, entry := range entries {
			switch entry.DataType {
			case namedNode, blankNode:
				nestedRsc, err := g.Resource(entry.Value, false)
				if !errors.Is(err, ErrResourceNotFound) {
					if err := g.inlineEntries(nestedRsc, seen); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (g *Graph) Add(hts ...HexTuple) error {
	for _, ht := range hts {
		hash := ht.Hash()
		if _, ok := g.seen[hash]; !ok {
			continue
		}

		g.tuples = append(g.tuples, ht)

		if err := g.addToResource(ht); err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) addToResource(ht HexTuple) error {
	rsc, ok := g.resources[ht.Subject]
	if !ok {
		rsc = newResource(ht.Subject)
	}

	if ht.Predicate == rdf.RDFType {
		rsc.Type = append(rsc.Type, ht.Value)
		g.resources[ht.Subject] = rsc

		return nil
	}

	entries, ok := rsc.Predicates[ht.Predicate]
	if !ok {
		entries = []Entry{}
	}

	entries = append(entries, ht.entry())
	rsc.Predicates[ht.Predicate] = entries

	return nil
}

func newResource(subject string) *Resource {
	return &Resource{
		Subject:      subject,
		Type:         []string{},
		Predicates:   map[string][]Entry{},
		inlinedPaths: map[string]bool{},
	}
}

type Resource struct {
	Subject      string
	Type         []string
	Predicates   map[string][]Entry
	inlinedPaths map[string]bool
}

type Entry struct {
	Value    string
	DataType string
	Language string
	Inline   *Resource
}
