package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	r "github.com/kiivihal/rdf2go"
)

type SortedGraph struct {
	triples []*r.Triple
	lock    sync.Mutex
}

func NewSortedGraph(g *r.Graph) *SortedGraph {
	sg := &SortedGraph{}

	for t := range g.IterTriples() {
		sg.Add(t)
	}
	return sg
}

// AddTriple add triple to the list of triples in the sortedGraph.
// Note: there is not deduplication
func (sg *SortedGraph) Add(t ...*r.Triple) {
	sg.triples = append(sg.triples, t...)
}

func (sg *SortedGraph) Triples() []*r.Triple {
	return sg.triples
}

// ByPredicate returns a list of triples that have the same predicate
func (sg *SortedGraph) ByPredicate(predicate r.Term) []*r.Triple {
	matches := []*r.Triple{}
	for _, t := range sg.triples {
		if t.Predicate.Equal(predicate) {
			matches = append(matches, t)
		}
	}
	return matches
}

// AddTriple is used to add a triple made of individual S, P, O objects
func (sg *SortedGraph) AddTriple(s r.Term, p r.Term, o r.Term) {
	sg.triples = append(sg.triples, r.NewTriple(s, p, o))
}

// Remove removes a triples from the SortedGraph
func (sg *SortedGraph) Remove(t *r.Triple) {
	triples := []*r.Triple{}
	for _, tt := range sg.triples {
		if t != tt {
			triples = append(triples, tt)
		}
	}
	sg.triples = triples
}

// Len returns the number of triples in the SortedGraph
func (sg *SortedGraph) Len() int {
	return len(sg.triples)
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GenerateJSONLD creates a interfaggce based model of the RDF Graph.
// This can be used to create various JSON-LD output formats, e.g.
// expand, flatten, compacted, etc.
func (sg *SortedGraph) GenerateJSONLD() ([]map[string]interface{}, error) {
	m := map[string]*r.LdEntry{}
	entries := []map[string]interface{}{}
	orderedSubjects := []string{}

	for _, t := range sg.triples {
		s := t.GetSubjectID()
		if !containsString(orderedSubjects, s) {
			orderedSubjects = append(orderedSubjects, s)
		}
		err := r.AppendTriple(m, t)
		if err != nil {
			return entries, err
		}
	}

	// log.Printf("subjects: %#v", orderedSubjects)
	// this most be sorted
	for _, v := range orderedSubjects {
		// log.Printf("v range: %s", v)
		ldEntry, ok := m[v]
		if ok {
			// log.Printf("ldentry: %#v", ldEntry.AsEntry())
			entries = append(entries, ldEntry.AsEntry())
		}
	}

	// log.Printf("graph: \n%#v", entries)

	return entries, nil
}

func (g *SortedGraph) SerializeFlatJSONLD(w io.Writer) error {
	entries, err := g.GenerateJSONLD()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	fmt.Fprint(w, string(bytes))
	return nil
}

func (sg *SortedGraph) GetRDF() ([]byte, error) {
	var b bytes.Buffer

	err := sg.SerializeFlatJSONLD(&b)
	if err != nil {
		return nil, err
	}
	log.Printf("graph: %s", b.String())
	return b.Bytes(), nil
}
