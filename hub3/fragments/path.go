package fragments

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/index"
	elastic "github.com/olivere/elastic/v7"
)

// PathHop is one hop of a ResourcePath: the search label of a predicate and
// the classes of the resource it starts from, one of which it must have
// (none: any class).
type PathHop struct {
	SearchLabel    string
	SubjectClasses []string
}

// ResourcePath is a route through a record's resources in the path syntax of
// ikuzo/rdf/index (path_query_syntax.md): [class]predicate pairs joined by
// ->, where a class names the resource that has the predicate, [] is any
// class and alternatives are separated by |. For example
//
//	[]schema:creator->[schema:Role]schema:creator->[schema:Person|schema:Organization]schema:name
//
// is the name of a person or organisation that is the creator of a role that
// is the object's creator. A path that ends with a predicate ends at a
// literal (Property) of the last resource; one that ends with [class] ends at
// that resource itself.
//
// The index keeps every linked resource as a nested resources element with
// its route from the root in resources.context: an element per hop with its
// Level, the predicate's SearchLabel and the SubjectClass of the resource the
// hop starts from, which is what [class]predicate says. So the resource at
// the end of a path is one resources element whose context matches the hops,
// which is what lets several conditions hold for the same linked resource.
// This generalises QueryFilter's single Level2 hop.
type ResourcePath struct {
	Hops     []PathHop
	Classes  []string // of the resource at the end
	Property string   // a literal of it, or "" for the resource itself
}

// ParseResourcePath reads a path in the syntax of ikuzo/rdf/index.
func ParseResourcePath(s string) (*ResourcePath, error) {
	entries, err := index.ParsePath(s, true)
	if err != nil {
		return nil, fmt.Errorf("path %q: %w", s, err)
	}
	p := &ResourcePath{}
	for i, e := range entries {
		last := i == len(entries)-1
		if !last && e.Predicate == "" {
			return nil, fmt.Errorf("path %q: step %d has a class but no predicate", s, i+1)
		}
		var label string
		if e.Predicate != "" {
			if label, err = rdf.DefaultNamespaceManager.GetSearchLabel(e.Predicate); err != nil {
				return nil, fmt.Errorf("path %q: predicate %s: %w", s, e.Predicate, err)
			}
		}
		if !last {
			p.Hops = append(p.Hops, PathHop{SearchLabel: label, SubjectClasses: e.Classes()})
			continue
		}
		p.Classes, p.Property = e.Classes(), label
	}
	return p, nil
}

// Depth is the number of hops.
func (p *ResourcePath) Depth() int { return len(p.Hops) }

// ResourceQuery matches records with a resource at the end of the hops for
// which all inner queries hold, such as EntryQuery for its literals at level
// Depth()+1. Without hops that resource is the record's root.
func (p *ResourcePath) ResourceQuery(inner ...elastic.Query) elastic.Query {
	var must []elastic.Query
	for i, h := range p.Hops {
		hop := []elastic.Query{
			elastic.NewTermQuery("resources.context.level", i+1),
			elastic.NewTermQuery("resources.context.searchLabel", h.SearchLabel),
		}
		if len(h.SubjectClasses) > 0 {
			hop = append(hop, termsQuery("resources.context.subjectClass", h.SubjectClasses))
		}
		must = append(must, elastic.NewNestedQuery("resources.context", elastic.NewBoolQuery().Must(hop...)))
	}
	if len(p.Classes) > 0 {
		must = append(must, termsQuery("resources.types", p.Classes))
	}
	must = append(must, inner...)
	return elastic.NewNestedQuery("resources", elastic.NewBoolQuery().Must(must...))
}

// ValueQuery matches records whose path ends in one of values: a literal of
// Property, or for a path that ends at a resource, its URI. Without values it
// matches records where the path exists.
func (p *ResourcePath) ValueQuery(values ...string) elastic.Query {
	if p.Property != "" {
		return p.ResourceQuery(EntryQuery(p.Property, p.Depth()+1, values...))
	}
	if len(values) == 0 {
		return p.ResourceQuery()
	}
	return p.ResourceQuery(termsQuery("resources.id", values))
}

// EntryQuery matches a resource with a literal of the predicate searchLabel
// at level (0: any level) that is one of values (none: any value).
func EntryQuery(searchLabel string, level int, values ...string) elastic.Query {
	must := []elastic.Query{elastic.NewTermQuery(entriesSearchLabel, searchLabel)}
	if level > 0 {
		must = append(must, elastic.NewTermQuery("resources.entries.level", level))
	}
	if len(values) > 0 {
		must = append(must, termsQuery("resources.entries.@value.keyword", values))
	}
	return elastic.NewNestedQuery(resourcesEntries, elastic.NewBoolQuery().Must(must...))
}

func termsQuery(field string, values []string) elastic.Query {
	vs := make([]any, len(values))
	for i, v := range values {
		vs[i] = v
	}
	return elastic.NewTermsQuery(field, vs...)
}
