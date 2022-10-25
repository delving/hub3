package rdf

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/namespace"
)

// DefaultNamespaceManager can be set at package level to
// serve as a default when no NamespaceManager is set on a
// Graph.
//
// The namespace manager is used by Resource and some RDF
// Encode/Decoder packages.
var DefaultNamespaceManager NamespaceManager

// TODO(kiivihal): replace with better solution
func init() {
	// TODO(kiivihal): implement this with new service
	svc, err := namespace.NewService(namespace.WithDefaults())
	if err != nil {
		log.Fatalf("rdf: unable to start namespace service")
	}

	DefaultNamespaceManager = svc
}

// Graph is a collection of triples where the order of insertion is remembered
type Graph struct {
	// simple implementation first
	triples []*Triple
	seen    map[hasher]bool
	BaseURI IRI
	lock    sync.Mutex
	// order uint64
	export         bool // set when all triples read from the graph
	addAfterExport bool

	// support for collections
	collections map[Subject][]*Triple

	//
	index            *GraphIndex
	stats            *GraphStats
	resources        map[Subject]*Resource
	UseIndex         bool
	UseResource      bool
	NamespaceManager NamespaceManager
}

func NewGraph() *Graph {
	g := &Graph{
		seen:             map[hasher]bool{},
		index:            newIndex(),
		stats:            &GraphStats{},
		UseIndex:         true,
		UseResource:      true,
		resources:        make(map[Subject]*Resource),
		NamespaceManager: DefaultNamespaceManager,
		collections:      make(map[Subject][]*Triple),
	}

	return g
}

func (g *Graph) extractCollection(t *Triple) bool {
	p := t.Predicate.RawValue()
	if p != RDFCollectionFirst && p != RDFCollectionRest {
		return false
	}

	triples, ok := g.collections[t.Subject]
	if !ok {
		triples = []*Triple{}
	}

	triples = append(triples, t)
	g.collections[t.Subject] = triples

	return true
}

type collectionConfig struct {
	seen        map[Subject]bool
	collections map[string][]Object
}

func (g *Graph) process(subj Subject, triples []*Triple, cfg *collectionConfig, objs []Object) []Object {
	cfg.seen[subj] = true

	sort.Slice(triples, func(i, j int) bool {
		return triples[i].Predicate.RawValue() < triples[j].Predicate.RawValue()
	})

	for _, t := range triples {
		switch t.Predicate.RawValue() {
		case RDFCollectionFirst:
			objs = append(objs, t.Object)
		case RDFCollectionRest:
			if t.Object.RawValue() == RDFCollectionNil {
				continue
			}

			nestedTriples, ok := g.collections[t.Object.(Subject)]
			if ok {
				objs = g.process(subj, nestedTriples, cfg, objs)
			}
		default:
			log.Printf("unknown collection predicate: %#v", t)
		}
	}

	return objs
}

func (g *Graph) Inline() error {
	if len(g.collections) == 0 {
		return nil
	}

	cfg := &collectionConfig{
		collections: map[string][]Object{},
		seen:        map[Subject]bool{},
	}

	for subj, triples := range g.collections {
		_, ok := cfg.seen[subj]
		if ok {
			// skip if already seen for recursion
			continue
		}

		objs := g.process(subj, triples, cfg, []Object{})
		if len(objs) != 0 {
			cfg.collections[subj.RawValue()] = objs
		}
	}

	// TODO(kiivihal): loop over subj find
	updates := []*Triple{}
	for _, t := range g.triples {
		switch t.Object.Type() {
		case TermBlankNode, TermIRI:
			_, ok := cfg.collections[t.Object.RawValue()]
			if ok {
				updates = append(updates, t)
			}
		}
	}

	g.Remove(updates...)

	for _, triple := range updates {
		objs, ok := cfg.collections[triple.Object.RawValue()]
		if !ok {
			log.Printf("WARN: object should always be part of collection: %#v", triple.Object)
		}

		for _, obj := range objs {
			newTriple := *triple
			newTriple.Object = obj
			g.Add(&newTriple)
		}
	}

	// remove collections
	g.collections = make(map[Subject][]*Triple)

	return nil
}

// AddTriple appends triple to the Graph triples
// Note: there is no deduplication. The same triple can be added multiple times
func (g *Graph) Add(triples ...*Triple) {
	if g.export {
		g.addAfterExport = true
	}

	for _, t := range triples {
		if g.extractCollection(t) {
			continue
		}

		hash := getHash(t)

		_, ok := g.seen[hash]
		if ok {
			continue
		}

		g.lock.Lock()
		if g.UseIndex {
			// TODO(kiivihal): what to do with this error
			g.index.update(t, false)
		}

		if g.UseResource {
			g.updateResources(t)
		}

		g.stats.incTriples()
		g.triples = append(g.triples, t)
		g.seen[hash] = true
		g.lock.Unlock()
	}
}

// AddTriple is used to add a triple made of individual S, P, O objects
func (g *Graph) AddTriple(s Subject, p Predicate, o Object) {
	g.Add(NewTriple(s, p, o))
}

// Len returns the number of triples in the Graph
func (g *Graph) Len() int {
	return len(g.triples)
}

// Triples returns an list based on insertion order of the triples in Graph.
func (g *Graph) Triples() []*Triple {
	g.export = true
	return g.triples
}

// TriplesOnce returns an list based on insertion order of the triples in Graph,
// an error is returned when triples have been Added after the previous read.
func (g *Graph) TriplesOnce() ([]*Triple, error) {
	g.export = true
	if g.addAfterExport {
		return []*Triple{}, fmt.Errorf("triples have been added after previous read")
	}

	return g.triples, nil
}

func (g *Graph) Namespaces() (ns []*domain.Namespace, err error) {
	for baseURI := range g.index.NamespacesURIs {
		n, retrieveErr := g.NamespaceManager.GetWithBase(baseURI)
		if retrieveErr != nil {
			return ns, fmt.Errorf("unknown baseURI: %s; %w", baseURI, retrieveErr)
		}

		ns = append(ns, n)
	}

	return ns, nil
}

func (g *Graph) Stats() *GraphStats {
	g.stats.Languages = len(g.index.Languages)
	g.stats.ObjectIRIs = len(g.index.ObjectResources)
	g.stats.Predicates = len(g.index.Predicates)
	g.stats.Resources = len(g.index.Subjects)
	g.stats.Namespaces = len(g.index.NamespacesURIs)
	// ObjectLiterals: len(idx.O),

	return g.stats
}

func (g *Graph) updateResources(t *Triple) {
	rsc, ok := g.resources[t.Subject]
	if !ok {
		rsc = NewResource(t.Subject)
	}

	rsc.Add(t)
	g.resources[t.Subject] = rsc
}

func (g *Graph) Get(s Subject) (rsc *Resource, ok bool) {
	rsc, ok = g.resources[s]
	return
}

func (g *Graph) Resources() map[Subject]*Resource {
	return g.resources
}

// // ByPredicate returns a list of triples that have the same predicate
// func (g *Graph) ByPredicate(predicate r.Term) []*r.Triple {
// matches := []*r.Triple{}
// for _, t := range sg.triples {
// if t.Predicate.Equal(predicate) {
// matches = append(matches, t)
// }
// }
// return matches
// }

// Remove removes triples from the SortedGraph
func (g *Graph) Remove(remove ...*Triple) {
	triples := []*Triple{}
	for _, tt := range g.triples {
		var exclude bool

		for _, t := range remove {
			if t.Equal(tt) {
				exclude = true
				g.stats.decrTriples()

				break
			}
		}

		if !exclude {
			triples = append(triples, tt)
		}
	}

	g.lock.Lock()
	g.triples = triples
	g.lock.Unlock()
}

// func containsString(s []string, e string) bool {
// for _, a := range s {
// if a == e {
// return true
// }
// }
// return false
// }
