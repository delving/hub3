package index

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/delving/hub3/ikuzo/rdf"
)

// ResourceLabelPredicates are used to find the label for a Resource
// This is also used for presenting labels for linked resources
var ResourceLabelPredicates = []string{
	"http://purl.org/dc/elements/1.1/title",
	"http://www.w3.org/2004/02/skos/core#prefLabel",
	"http://www.w3.org/2000/01/rdf-schema#label",
	"http://www.w3.org/2004/02/skos/core#altLabel",
	"http://xmlns.com/foaf/0.1/name",
	"http://www.geonames.org/ontology#name",
	"http://schemas.delving.eu/narthex/terms/proxyLiteralValue",
	"http://dbpedia.org/ontology/name",
}

// Resource holds all the context information for a RDF resource
type Resource struct {
	// ID contains the IRI/URI of the subject of the resource, i.e. the rdf.Subject
	ID string `json:"id"`
	// Types contains the IRI/URI of the rdf Classes, i.e. the rdf.Type
	Types []string `json:"types"`
	// Entries contain the nested []*Entry that contain the wrapped rdf.Predicate and rdf.Object information
	Entries []*Entry `json:"entries"`
	// Context contains an ordered list of Referrers from the root Resource of the graph to the current Resource
	Context []*ContextRef `json:"context"`

	// GraphExternalContext []*ContextRef `json:"graphExternalContext"`
	// Tags                 []string      `json:"tags,omitempty"`
	// predicates           map[string][]*FragmentEntry
	// objectIDs []*ContextRef

	mu sync.RWMutex
}

// Add will add an Entry to Resource.Entries
//
// It is save for concurrent use and will deduplicate on insert.
// When the fingerprint is known, it will replace and otherwise append.
func (rsc *Resource) Add(entry *Entry) {
	entry.processTags()
	if entry.Predicate == "" && entry.SearchLabel != "" {
		p, err := getPredicate(entry.SearchLabel)
		if err != nil {
			slog.Warn("unable to create predicate", "searchLabel", searchLabel)
		}
		entry.Predicate = p
	}
	hash := entry.Fingerprint()

	rsc.mu.Lock()
	defer rsc.mu.Unlock()

	pos := slices.IndexFunc(rsc.Entries, func(e *Entry) bool {
		return e.Fingerprint() == hash
	})

	if pos != -1 {
		rsc.Entries[pos] = entry
		return
	}

	rsc.Entries = append(rsc.Entries, entry)
}

// AddTo converts the Resource.Entries to triples and adds them to the rdf.Graph
func (rsc *Resource) AddTo(g *rdf.Graph) error {
	subject, err := rdf.NewIRI(rsc.ID)
	if err != nil {
		return err
	}

	for _, rdfType := range rsc.Types {
		rdfType, err := rdf.NewIRI(rdfType)
		if err != nil {
			return err
		}

		g.AddTriple(subject, rdf.IsA, rdfType)
	}

	for _, entry := range rsc.Entries {
		triple, err := entry.AsTriple(subject)
		if err != nil {
			return err
		}

		g.Add(triple)
	}

	return nil
}

// appendContext adds the referrerContext to the FragmentResource
//
// Calling this method increments the context level/depth of the resource
func (rsc *Resource) AppendContext(ctxs ...*ContextRef) {
	for _, ctx := range ctxs {
		if !containsContext(rsc.Context, ctx) {
			rsc.Context = append(rsc.Context, ctx)
		}
	}
}

// AddTypes adds rdf:Type to the Types list if it is unique
func (rsc *Resource) AddTypes(types ...string) {
	rsc.Types = appendUnique(rsc.Types, types...)
}

func (rsc *Resource) GetLabel() (label, language string) {
	if rsc.ID == "" {
		return "", ""
	}

	for _, p := range ResourceLabelPredicates {
		o := rsc.predicate(p)
		if len(o) > 0 {
			return o[0].Value, o[0].Language
		}
	}

	return "", ""
}

// GetLevel returns the relative level that this resource has from the root
// or parent resource
func (rsc *Resource) GetLevel() int32 {
	highestLevel := int32(0)
	for _, ctx := range rsc.Context {
		if ctx.Level > highestLevel {
			highestLevel = ctx.Level
		}
	}

	return highestLevel + 1
}

// IsEmpty returns true when no triples are part of the Resource
func (rsc *Resource) IsEmpty() bool {
	return len(rsc.Entries) == 0 && len(rsc.Types) == 0
}

func (rsc *Resource) newContext(predicate, objectID string) ContextRef {
	searchLabel, err := rdf.DefaultNamespaceManager.GetSearchLabel(predicate)
	if err != nil {
		slog.Warn("unable to find search label", "predicate", predicate, "objectID", objectID, "error", err)
		searchLabel = ""
	}

	label, _ := rsc.GetLabel()

	return ContextRef{
		Subject:      rsc.ID,
		SubjectClass: rsc.Types,
		Predicate:    predicate,
		SearchLabel:  searchLabel,
		Level:        rsc.GetLevel(),
		ObjectID:     objectID,
		Label:        label,
	}
}

// objectIDs return a list of all the outward pointing object resources
//
// These are used to set the ContextRefs and inline the resources in the Graph
func (rsc *Resource) objectIDs() (ids []ContextRef) {
	for _, e := range rsc.Entries {
		if e.EntryType == Bnode || e.EntryType == ResourceType {
			ids = append(ids, rsc.newContext(e.Predicate, e.ID))
		}
	}

	return ids
}

func (rsc *Resource) predicate(p string) (entries []*Entry) {
	for _, e := range rsc.Entries {
		if e.Predicate == p {
			entries = append(entries, e)
		}
	}

	return entries
}

func getEntries(r *rdf.Resource) (entries []*Entry, err error) {
	for p, rp := range r.Predicates() {
		for _, obj := range rp.Objects() {

			e := entryFromObject(p, obj)
			entries = append(entries, e)
		}
	}
	return
}

func entryFromObject(pred rdf.Predicate, obj rdf.Object) *Entry {
	e := &Entry{
		Predicate:   pred.RawValue(),
		SearchLabel: searchLabel(pred.RawValue()),
	}

	switch obj.Type() {
	case rdf.TermBlankNode:
		e.ID = obj.RawValue()
		e.EntryType = Bnode
	case rdf.TermIRI:
		e.ID = obj.RawValue()
		e.EntryType = ResourceType
	case rdf.TermLiteral:
		l := obj.(rdf.Literal)
		e.Value = l.RawValue()
		e.Language = l.Lang()
		e.DataType = l.DataType.RawValue()
		e.EntryType = Literal
	}
	return e
}

func searchLabel(predicate string) string {
	searchLabel, err := rdf.DefaultNamespaceManager.GetSearchLabel(predicate)
	if err != nil {
		slog.Warn("unable to find search label", "predicate", predicate, "error", err)
		return ""
	}

	return searchLabel
}
