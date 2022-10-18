package rdf

import (
	"fmt"
	"log"
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
	}

	return g
}

// AddTriple appends triple to the Graph triples
// Note: there is no deduplication. The same triple can be added multiple times
func (g *Graph) Add(triples ...*Triple) {
	if g.export {
		g.addAfterExport = true
	}

	for _, t := range triples {
		hash := getHash(t)

		_, ok := g.seen[hash]
		if ok {
			continue
		}

		g.lock.Lock()
		if g.UseIndex {
			// TODO(kiivihal): what to do with this error
			g.index.update(t)
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

// // Remove removes a triples from the SortedGraph
// func (g *Graph) Remove(t *r.Triple) {
// triples := []*r.Triple{}
// for _, tt := range g.triples {
// if t != tt {
// triples = append(triples, tt)
// }
// }
// g.triples = triples
// }

// func containsString(s []string, e string) bool {
// for _, a := range s {
// if a == e {
// return true
// }
// }
// return false
// }
