package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/benbjohnson/immutable"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/index/embed"
)

// Graph is an indexable representation of an rdf namedgraph.
type Graph struct {
	// Header is the header of the Graph with queryable meta information
	Header Header `json:"meta,omitempty"`

	//
	// Tree       *Tree                     `json:"tree,omitempty"`

	// Resources conains a list of triples grouped by their resource Subject
	Resources []*Resource `json:"resources,omitempty"`

	// Embed contains embed.Data objects serialized as bytes
	// These can be used to store structs that should be accesible directly from the index
	Embed []embed.Raw `json:"-"`

	// contextIsSet is used to make sure the context is always set before the graph is used for indexing
	contextIsSet bool
}

// NewGraph returns a new Graph. When the header is not valid an error is returned
func NewGraph(header Header) (*Graph, error) {
	header.addDefaults()

	if validErr := header.Valid(); validErr != nil {
		return nil, validErr
	}

	g := Graph{
		Header: header,
	}

	return &g, nil
}

// AddGraph adds the triples from the rdf.Graph to the Resources list
func (g *Graph) AddGraph(graph *rdf.Graph) error {
	for subj, r := range graph.Resources() {
		rsc, created := g.Resource(subj.RawValue())
		if created {
			for _, t := range r.Types() {
				rsc.Types = append(rsc.Types, t.RawValue())
			}
		}
		entries, err := getEntries(r)
		if err != nil {
			return fmt.Errorf("unable to convert entries; %w", err)
		}

		for _, e := range entries {
			rsc.Add(e)
		}
	}
	return nil
}

// Graph returns the triples in the Graph as a *rdf.Graph
func (g *Graph) Graph() (*rdf.Graph, error) {
	if len(g.Resources) == 0 {
		return nil, fmt.Errorf("unable to create *rdf.Graph because resources is empty")
	}
	rg := rdf.NewGraph()
	for _, rsc := range g.Resources {
		if err := rsc.AddTo(rg); err != nil {
			return nil, err
		}
	}

	return rg, nil
}

// IndexMessage converts the Graph into a domainpb.IndexMessage.
//
// This is the way in which the Graph is submitted for indexing.
func (g *Graph) IndexMessage() (*domainpb.IndexMessage, error) {
	if !g.contextIsSet {
		if err := g.addContextLevels(); err != nil {
			return nil, err
		}
	}

	g.prune()

	if err := g.Header.Valid(); err != nil {
		return nil, err
	}

	b, err := g.Marshal()
	if err != nil {
		return nil, err
	}

	return &domainpb.IndexMessage{
		OrganisationID: g.Header.OrgID,
		DatasetID:      g.Header.Spec,
		RecordID:       g.Header.HubID,
		IndexType:      domainpb.IndexType_V2,
		Source:         b,
	}, nil
}

// Marshal returns the Graph as []byte marshalled as JSON
func (g *Graph) Marshal() ([]byte, error) {
	return json.Marshal(g)
}

// Reader returns an io.Reader of the Graph marshalled as indented JSON.
func (fg *Graph) Reader() (io.Reader, error) {
	b, err := json.MarshalIndent(fg, "", "    ")
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// Resource creates or returns a Resource from the Graph.
//
// When a new Resource is created true is returned.
func (g *Graph) Resource(subject string) (*Resource, bool) {
	for _, rsc := range g.Resources {
		if rsc.ID == subject {
			return rsc, false
		}
	}

	rsc := &Resource{ID: subject}
	g.Resources = append(g.Resources, rsc)
	return rsc, true
}

// SearchLabel returns the Enties with the same SearchLabel.
// When subject is not empty, only resources with that subject ID
// will be used to retrieve the matching []*Entry.
func (g *Graph) SearchLabel(subject, label string) (entries []*Entry) {
	resources := g.Resources
	if subject != "" {
		rsc, created := g.Resource(subject)
		if created {
			return // subject not found so return immediately
		}
		resources = []*Resource{rsc}
	}

	for _, rsc := range resources {
		for _, entry := range rsc.Entries {
			if entry.SearchLabel == label {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// addContextLevels recurses the root resources and sets the contextRefs
func (g *Graph) addContextLevels() error {
	if g.Header.EntryURI == "" {
		return fmt.Errorf("g.Meta.EntryURI cannot be empty when setting context levels")
	}

	if len(g.Resources) == 0 {
		return fmt.Errorf("cannot set context levels on empty Resources list")
	}

	subject, created := g.Resource(g.Header.EntryURI)
	if created {
		return fmt.Errorf("subject %s is not part of the graph", g.Header.EntryURI)
	}

	if err := g.setContextRefs(subject, immutable.NewSet(contextHasher{})); err != nil {
		return fmt.Errorf("unable to set contextLevels for Graph; %w", err)
	}

	g.contextIsSet = true

	return nil
}

// prune removes empty Resources from the graph
func (g *Graph) prune() {
	var pruned []*Resource
	for _, rsc := range g.Resources {
		if !rsc.IsEmpty() {
			pruned = append(pruned, rsc)
		}
	}

	if len(g.Resources) > len(pruned) {
		g.Resources = pruned
	}
}

// setContextRefs recurses into nested resources until it reached the end of the graph or recures on itself
func (g *Graph) setContextRefs(rsc *Resource, parents immutable.Set[ContextRef]) error {
	for _, ctxLevel := range rsc.objectIDs() {
		if parents.Has(ctxLevel) {
			var paths []string
			for _, p := range parents.Items() {
				paths = append(paths, p.ObjectID)
			}
			slog.Debug("subject cannot recurse on itself", "subject", ctxLevel.ObjectID, "resource", rsc.ID, "paths", paths)
			continue
		}
		nestedResource, ok := g.Resource(ctxLevel.ObjectID)
		if !ok {
			slog.Debug("subject is not part of the graph", "subject", ctxLevel.ObjectID, "resource", rsc.ID)
			continue
		}
		ctxLevel.Level = int32(parents.Len() + 1)
		if len(ctxLevel.SubjectClass) == 0 {
			ctxLevel.SubjectClass = rsc.Types
		}

		nestedResource.AppendContext(&ctxLevel)
		parents = parents.Add(ctxLevel)
		if err := g.setContextRefs(nestedResource, parents); err != nil {
			return err
		}
	}
	return nil
}
