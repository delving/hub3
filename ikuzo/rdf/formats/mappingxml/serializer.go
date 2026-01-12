package mappingxml

import (
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/delving/hub3/ikuzo/rdf"
)

type FilterConfig struct {
	RDFType               rdf.IRI
	Subject               rdf.Subject
	URIPrefixFilter       string // to filter out private triples
	HubID                 string
	ContextLevels         int
	WikiBaseTypes         []string
	WikiBaseTypePredicate rdf.Predicate
	AllowedRDFTypes       []rdf.IRI
	SkipPredicatNamespace []string
	ExcludePrefixes       []string // prefixes to exclude for both subjects and predicates
	ExcludeTypePrefixes   []string // prefixes to exclude for rdf:type - skip whole resource if any type matches
}

func (fc *FilterConfig) HasNoFilters() bool {
	if fc.RDFType.RawValue() != "" {
		return false
	}

	if fc.Subject != nil && fc.Subject.RawValue() != "" {
		return false
	}

	if fc.URIPrefixFilter != "" {
		return false
	}

	if len(fc.ExcludePrefixes) > 0 {
		return false
	}

	if len(fc.ExcludeTypePrefixes) > 0 {
		return false
	}

	// TODO: add other filter options

	return true
}

// Serialize serialize the Graph to explicit XML.
// When rootType is given it will use the rdf:type as the
// root of the XML. When contextLevels is 0 no nested
// resources are inlined in the XML. A max of 5 levels can
// be given.
func Serialize(g *rdf.Graph, w io.Writer, cfg *FilterConfig) error {
	if cfg == nil {
		cfg = &FilterConfig{}
	}

	filtered := filterResources(g.Resources(), cfg)

	if cfg.ContextLevels == 0 {
		cfg.ContextLevels = 5
	}

	doc := etree.NewDocument()
	doc.Indent(4)
	root := doc.CreateElement("rdf:RDF")
	root.CreateAttr("xmlns:rdf", "http://www.w3.org/1999/02/22-rdf-syntax-ns#")

	namespaces, err := g.Namespaces()
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		root.CreateAttr(ns.XMLNS(), ns.URI)
	}

	for _, rsc := range filtered {
		if resourceErr := createResource(g, rsc, root, []string{}, cfg); resourceErr != nil {
			return resourceErr
		}
	}

	doc.Indent(4)

	_, err = doc.WriteTo(w)
	if err != nil {
		return err
	}

	slog.Info("finished serialize", "resources", len(filtered))

	return nil
}

// hasExcludedPrefix checks if a value starts with any of the excluded prefixes
func hasExcludedPrefix(value string, excludePrefixes []string) bool {
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func filterResources(resources map[rdf.Subject]*rdf.Resource, cfg *FilterConfig) []*rdf.Resource {
	var filtered []*rdf.Resource
	if cfg.HasNoFilters() {
		for _, rsc := range resources {
			filtered = append(filtered, rsc)
		}
		return filtered
	}

	hasFilter := cfg.URIPrefixFilter != ""
	hasExcludePrefixes := len(cfg.ExcludePrefixes) > 0
	hasExcludeTypePrefixes := len(cfg.ExcludeTypePrefixes) > 0

	for _, rsc := range resources {
		// Skip if subject matches URIPrefixFilter
		if hasFilter && strings.HasPrefix(rsc.Subject().RawValue(), cfg.URIPrefixFilter) {
			continue
		}

		// Skip if subject matches any excluded prefix
		if hasExcludePrefixes && hasExcludedPrefix(rsc.Subject().RawValue(), cfg.ExcludePrefixes) {
			continue
		}

		// Skip if any rdf:type matches excluded type prefixes
		if hasExcludeTypePrefixes {
			hasExcludedType := false
			for _, rdfType := range rsc.Types() {
				if hasExcludedPrefix(rdfType.RawValue(), cfg.ExcludeTypePrefixes) {
					hasExcludedType = true
					break
				}
			}
			if hasExcludedType {
				continue
			}
		}

		if !cfg.RDFType.Equal(rdf.IRI{}) {
			var found bool

			for _, srcType := range rsc.Types() {
				if cfg.RDFType.Equal(srcType) {
					found = true
				}
			}

			if found {
				filtered = append(filtered, rsc)
			}

			continue
		}

		if cfg.Subject != nil && !cfg.Subject.Equal(rdf.IRI{}) {
			if rsc.Subject().Equal(cfg.Subject) {
				filtered = append(filtered, rsc)
				return filtered
			}

			continue
		}

		filtered = append(filtered, rsc)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Subject().RawValue() < filtered[j].Subject().RawValue()
	})

	return filtered
}

func isElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func createResource(
	g *rdf.Graph, rsc *rdf.Resource, parent *etree.Element,
	seen []string, cfg *FilterConfig,
) error {
	hasFilter := cfg.URIPrefixFilter != ""
	hasExcludePrefixes := len(cfg.ExcludePrefixes) > 0
	hasExcludeTypePrefixes := len(cfg.ExcludeTypePrefixes) > 0

	if hasFilter && strings.HasPrefix(rsc.Subject().RawValue(), cfg.URIPrefixFilter) {
		return nil
	}

	if hasExcludePrefixes && hasExcludedPrefix(rsc.Subject().RawValue(), cfg.ExcludePrefixes) {
		return nil
	}

	// Skip if any rdf:type matches excluded type prefixes
	if hasExcludeTypePrefixes {
		for _, rdfType := range rsc.Types() {
			if hasExcludedPrefix(rdfType.RawValue(), cfg.ExcludeTypePrefixes) {
				return nil
			}
		}
	}

	ok := isElementExist(seen, rsc.Subject().RawValue())
	if ok || len(seen) > cfg.ContextLevels {
		switch rsc.Subject().Type() {
		case rdf.TermIRI:
			parent.CreateAttr("rdf:resource", rsc.Subject().RawValue())
		case rdf.TermBlankNode:
			parent.CreateAttr("rdf:nodeID", rsc.Subject().RawValue())
		}

		return nil
	}

	seen = append(seen, rsc.Subject().RawValue())

	rdfType := rsc.Types()[0]

	baseURI, label := rdfType.Split()

	ns, err := g.NamespaceManager.GetWithBase(baseURI)
	if err != nil {
		return err
	}

	root := parent.CreateElement(fmt.Sprintf("%s:%s", ns.Prefix, label))

	switch rsc.Subject().Type() {
	case rdf.TermIRI:
		root.CreateAttr("rdf:about", rsc.Subject().RawValue())
	case rdf.TermBlankNode:
		root.CreateAttr("rdf:nodeID", rsc.Subject().RawValue())
	}

	if len(rsc.Types()) > 1 {
		for _, rdfType := range rsc.Types()[1:] {
			xmlType := root.CreateElement("rdf:type")
			xmlType.CreateAttr("rdf:resource", rdfType.RawValue())
		}
	}

	for _, p := range rsc.SortedPredicates() {
		if p.IRI().Equal(rdf.IsA) {
			continue
		}

		// Skip predicates that match excluded prefixes
		if len(cfg.ExcludePrefixes) > 0 && hasExcludedPrefix(p.IRI().RawValue(), cfg.ExcludePrefixes) {
			continue
		}

		nsLabel := namespace(g, p.IRI())
		if nsLabel == "" {
			return err
		}

		elem := root.CreateElement(nsLabel)

		for _, object := range p.Objects() {
			if object.RawValue() == "" {
				continue
			}

			switch object.Type() {
			case rdf.TermLiteral:
				if elem.Text() != "" {
					elem = root.CreateElement(nsLabel)
				}

				l := object.(rdf.Literal)

				if l.Lang() != "" {
					elem.CreateAttr("xml:lang", l.Lang())
				}

				if !l.HasImpliedDataType() {
					elem.CreateAttr("rdf:dataType", l.DataType.RawValue())
				}

				elem.CreateText(l.RawValue())
			case rdf.TermBlankNode:
				bnode := object.(rdf.BlankNode)

				nestedRsc, ok := g.Get(rdf.Subject(bnode))
				if !ok {
					elem.CreateAttr("rdf:nodeID", bnode.RawValue())
					continue
				}

				if err := createResource(g, nestedRsc, elem, seen, cfg); err != nil {
					return err
				}
			case rdf.TermIRI:
				iri := object.(rdf.IRI)

				if hasFilter && strings.HasPrefix(iri.RawValue(), cfg.URIPrefixFilter) {
					continue
				}

				if hasExcludePrefixes && hasExcludedPrefix(iri.RawValue(), cfg.ExcludePrefixes) {
					continue
				}

				nestedRsc, ok := g.Get(rdf.Subject(iri))
				if !ok {
					elem.CreateAttr("rdf:resource", iri.RawValue())
					continue
				}

				if err := createResource(g, nestedRsc, elem, seen, cfg); err != nil {
					return err
				}

			default:
				return fmt.Errorf("unknown termtype: %s", object.Type())
			}
		}
	}

	return nil
}

func namespace(g *rdf.Graph, p rdf.Predicate) string {
	baseURI, label := p.(rdf.IRI).Split()

	ns, err := g.NamespaceManager.GetWithBase(baseURI)
	if err != nil {
		return ""
	}

	// Sanitize the label: XML element local names cannot contain colons
	// Replace colons with underscores to produce valid XML
	label = strings.ReplaceAll(label, ":", "_")

	return fmt.Sprintf("%s:%s", ns.Prefix, label)
}
