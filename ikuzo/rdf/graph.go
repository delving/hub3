package rdf

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"sort"
	"sync"

	"github.com/kiivihal/rdf2go"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/namespace"
	"github.com/delving/hub3/ikuzo/validator"
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

	// store an error
	err error

	// support for collections
	collections map[Subject][]*Triple

	//
	index            *GraphIndex
	stats            *GraphStats
	resources        map[Subject]*Resource
	UseIndex         bool
	UseResource      bool
	GraphName        string
	NamespaceManager NamespaceManager
	Subject          Subject
	firstErr         error
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
		Subject:          nullSubject{},
	}

	return g
}

// Err returns the first non-EOF error that was encountered by the [Scanner].
func (g *Graph) Err() error {
	// TODO: implement me
	return nil
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
// Note: there is deduplication. Only the first triple is added.
// Subsequent identical triples are ignored.
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

	sort.SliceStable(ns, func(i, j int) bool {
		return ns[i].Prefix < ns[j].Prefix
	})

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
		rsc.order = len(g.resources) + 1
	}

	rsc.Add(t)
	g.resources[t.Subject] = rsc
}

func (g *Graph) Get(s Subject) (rsc *Resource, ok bool) {
	rsc, ok = g.resources[s]
	return
}

func (g *Graph) GetByType(iri IRI) (subjects []Subject) {
	for _, t := range g.Triples() {
		if !t.Predicate.Equal(IsA) {
			continue
		}

		if t.Object.RawValue() != iri.RawValue() {
			continue
		}

		subjects = append(subjects, t.Subject)
	}

	return subjects
}

func (g *Graph) Resources() map[Subject]*Resource {
	return g.resources
}

func (g *Graph) ResourcesOrdered() []*Resource {
	// TODO: build resources when it is empty
	var resources []*Resource
	for _, rsc := range g.resources {
		resources = append(resources, rsc)
	}
	slices.SortStableFunc(resources, func(a, b *Resource) int {
		return cmp.Compare(a.order, b.order)
	})

	return resources
}

type TripleFilter interface {
	Filter(t *Triple) bool
}

func (g *Graph) FilterTriples(f TripleFilter) []*Triple {
	triples := []*Triple{}
	for _, t := range g.Triples() {
		if f.Filter(t) {
			triples = append(triples, t)
		}
	}

	return triples
}

type ResourceFilter interface {
	Filter(rsc *Resource) bool
}

func (g *Graph) FilterResources(f ResourceFilter) []*Resource {
	var resources []*Resource

	if len(g.resources) == 0 {
		return resources
	}

	for _, rsc := range g.resources {
		if f.Filter(rsc) {
			resources = append(resources, rsc)
		}
	}

	return resources
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

// SubjectSafe returns the first Subject in the Graph if it is not explicitely set already.
func (g *Graph) SubjectSafe() Subject {
	if g.Subject.RawValue() != "" {
		return g.Subject
	}

	if len(g.triples) == 0 {
		return g.Subject
	}

	g.Subject = g.triples[0].Subject

	return g.Subject
}

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

// AsLegacyGraph converts the Graph into an *rdf2go.Graph
//
// NOTE: this function can be removed when rdf2go package
// is no longer used as a dependency.
func (g *Graph) AsLegacyGraph() (*rdf2go.Graph, error) {
	legacyGraph := rdf2go.NewGraph("")
	for _, triple := range g.Triples() {
		legacyTriple, convErr := triple.AsLegacyTriple()
		if convErr != nil {
			return nil, convErr
		}

		legacyGraph.Add(legacyTriple)
	}
	return legacyGraph, nil
}

// GetAboutURI returns the first matching rdf.type object from the graph.
// It returns an error when the aboutRDFType is invalid or the aboutRDFType cannot be found
// in the rdf.Graph.
func (g *Graph) GetAboutURI(aboutRDFTypes []string) (string, error) {
	for _, aboutRDFType := range aboutRDFTypes {
		aboutType, err := NewIRI(aboutRDFType)
		if err != nil {
			return "", fmt.Errorf("unable to create rdf AboutTypeURI; %w", err)
		}

		aboutURIs := g.GetByType(aboutType)
		if len(aboutURIs) == 0 {
			continue
		}

		for _, s := range aboutURIs {
			if s.Type() != TermIRI {
				continue
			}
			return s.RawValue(), nil
		}
	}

	return "", fmt.Errorf("unable to retrieve aboutType %q from graph", aboutRDFTypes)
}

// Internal null subject implementation
type nullSubject struct{}

func (n nullSubject) ValidAsSubject()                {}
func (n nullSubject) Equal(t Term) bool              { return false }
func (n nullSubject) RawValue() string               { return "" }
func (n nullSubject) String() string                 { return "" }
func (n nullSubject) Type() TermType                 { return TermIRI }
func (n nullSubject) Validate() *validator.Validator { return validator.New() }
func (n nullSubject) Valid() error                   { return nil }
