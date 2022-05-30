package mappingxml

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/delving/hub3/ikuzo/rdf"
)

type FilterConfig struct {
	RDFType         rdf.IRI
	Subject         rdf.Subject
	URIPrefixFilter string // to filter out private triples
	HubID           string
	ContextLevels   int
}

// Serialize serialize the Graph to explicit XML.
// When rootType is given it will use the rdf:Type as the
// root of the XML. When contextLevels is 0 no nested
// resources are inlined in the XML. A max of 5 levels can
// be given.
func Serialize(g *rdf.Graph, w io.Writer, cfg *FilterConfig) error {
	filtered := filterResources(g.Resources(), cfg)

	doc := etree.NewDocument()
	doc.Indent(2)
	root := doc.CreateElement("rdf:RDF")
	root.CreateAttr("xmlns:rdf", "http://www.w3.org/1999/02/22-rdf-syntax-ns#")

	namespaces, err := g.Namespaces()
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		root.CreateAttr(ns.XMLNS(), ns.Base)
	}

	for _, rsc := range filtered {
		if resourceErr := createResource(g, rsc, root, map[rdf.Subject]bool{}, cfg); resourceErr != nil {
			return resourceErr
		}
	}

	_, err = doc.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}

func filterResources(resources map[rdf.Subject]*rdf.Resource, cfg *FilterConfig) []*rdf.Resource {
	var filtered []*rdf.Resource

	hasFilter := cfg.URIPrefixFilter != ""

	for _, rsc := range resources {
		if hasFilter && strings.HasPrefix(rsc.Subject().String(), cfg.URIPrefixFilter) {
			continue
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

		if !cfg.Subject.Equal(rdf.IRI{}) {
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

func createResource(
	g *rdf.Graph, rsc *rdf.Resource, parent *etree.Element,
	seen map[rdf.Subject]bool, cfg *FilterConfig,
) error {
	rdfType := rsc.Types()[0]

	baseURI, label := rdfType.Split()

	ns, err := g.NamespaceManager.GetWithBase(baseURI)
	if err != nil {
		return err
	}
	hasFilter := cfg.URIPrefixFilter != ""

	if hasFilter && strings.HasPrefix(rsc.Subject().RawValue(), cfg.URIPrefixFilter) {
		return nil
	}

	root := parent.CreateElement(fmt.Sprintf("%s:%s", ns.Prefix, label))

	switch rsc.Subject().Type() {
	case rdf.TermIRI:
		root.CreateAttr("rdf:about", rsc.Subject().RawValue())
	case rdf.TermBlankNode:
		root.CreateAttr("rdf:nodeID", rsc.Subject().RawValue())
	}

	_, ok := seen[rsc.Subject()]
	if ok {
		return nil
	}

	seen[rsc.Subject()] = true

	if len(rsc.Types()) > 1 {
		for _, rdfType := range rsc.Types()[1:] {
			xmlType := root.CreateElement("rdf:Type")
			xmlType.CreateAttr("rdf:resource", rdfType.RawValue())
		}
	}

	for _, p := range rsc.SortedPredicates() {
		if p.IRI().Equal(rdf.IsA) {
			continue
		}

		nsLabel := namespace(g, p.IRI())
		if nsLabel == "" {
			return err
		}

		elem := root.CreateElement(nsLabel)

		for _, object := range p.Objects() {
			switch object.Type() {
			case rdf.TermLiteral:
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

	return fmt.Sprintf("%s:%s", ns.Prefix, label)
}
