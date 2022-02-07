package resource

import "fmt"

// Graph is a collection of triples where the order of insertion is remembered
type Graph struct {
	// simple implementation first
	triples []*Triple
	BaseURI *IRI
	// lock    sync.Mutex
	// order uint64
	export         bool // set when all triples read from the graph
	addAfterExport bool

	//
	index    *index
	stats    *GraphStats
	resouces map[*IRI]*Resource
}

func NewGraph() *Graph {
	g := &Graph{}
	return g
}

// AddTriple appends triple to the Graph triples
// Note: there is no deduplication. The same triple can be added multiple times
func (g *Graph) Add(t ...*Triple) {
	if g.export {
		g.addAfterExport = true
	}

	g.triples = append(g.triples, t...)
}

// AddTriple is used to add a triple made of individual S, P, O objects
// func (g *Graph) AddTriple(s Subject, p Predicate, o Object) {
// g.AddTriple()
// g.triples = append(g.triples, r.NewTriple(s, p, o))
// }

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
