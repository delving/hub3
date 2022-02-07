// Package jsonld provides tools to parse and serialize RDF data in the JSON-LD format.
//
// For more information about JSON-LD, see - https://json-ld.org.
package jsonld

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/delving/hub3/ikuzo/resource"
	jsonld "github.com/kiivihal/gojsonld"
	"github.com/piprate/json-gold/ld"
)

func Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	if g == nil {
		g = resource.NewGraph()
	}

	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(r)
	if err != nil {
		return g, err
	}

	jsonData, err := jsonld.ReadJSON(buf.Bytes())
	if err != nil {
		return g, err
	}

	options := &jsonld.Options{}
	options.Base = ""
	options.ProduceGeneralizedRdf = false

	dataSet, err := jsonld.ToRDF(jsonData, options)
	if err != nil {
		return g, err
	}

	for triple := range dataSet.IterTriples() {
		t, err := jtriple2triple(triple)
		if err != nil {
			return g, err
		}

		g.Add(t)
	}

	return g, nil
}

func ParseWithContext(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	if g == nil {
		g = resource.NewGraph()
	}

	doc, err := ld.DocumentFromReader(r)
	if err != nil {
		return g, err
	}

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	client := &http.Client{}
	nl := ld.NewDefaultDocumentLoader(client)

	// testing caching
	cdl := ld.NewCachingDocumentLoader(nl)
	// cdl.PreloadWithMapping(map[string]string{
	// "https://schema.org/": "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// "http://schema.org/":  "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// "https://schema.org":  "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// "http://schema.org":   "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// })

	options.DocumentLoader = cdl
	// options.Format = "application/nquads"

	rdf, err := proc.ToRDF(doc, options)
	if err != nil {
		return g, err
	}

	dataset, ok := rdf.(*ld.RDFDataset)
	if !ok {
		return g, fmt.Errorf("*ld.RDFDataset should have been returned")
	}

	for graph, quads := range dataset.Graphs {
		if graph != "@default" {
			continue
		}

		for _, quad := range quads {
			t, err := quad2triple(quad)
			if err != nil {
				return g, err
			}

			g.Add(t)
		}
	}

	return g, nil
}

func quad2triple(quad *ld.Quad) (*resource.Triple, error) {
	s, err := ldnode2term(quad.Subject)
	if err != nil {
		return nil, err
	}

	p, err := ldnode2term(quad.Predicate)
	if err != nil {
		return nil, err
	}

	o, err := ldnode2term(quad.Object)
	if err != nil {
		return nil, err
	}

	return resource.NewTriple(s.(resource.Subject), p.(resource.Predicate), o.(resource.Object)), nil
}

func ldnode2term(node ld.Node) (resource.Term, error) {
	switch term := node.(type) {
	case *ld.BlankNode:
		return resource.NewBlankNode(term.GetValue())
	case *ld.Literal:
		if len(term.Language) > 0 {
			return resource.NewLiteralWithLang(term.GetValue(), term.Language)
		}

		if term.Datatype != "" {
			dt, err := resource.NewIRI(term.Datatype)
			if err != nil {
				return nil, err
			}

			return resource.NewLiteralWithType(term.Value, &dt)
		}

		return resource.NewLiteral(term.Value)
	case *ld.IRI:
		return resource.NewIRI(term.GetValue())
	}

	return nil, fmt.Errorf("unknown resource.TermType")
}

func jtriple2triple(triple *jsonld.Triple) (*resource.Triple, error) {
	s, err := jterm2term(triple.Subject)
	if err != nil {
		return nil, err
	}

	p, err := jterm2term(triple.Predicate)
	if err != nil {
		return nil, err
	}

	o, err := jterm2term(triple.Object)
	if err != nil {
		return nil, err
	}

	return resource.NewTriple(s.(resource.Subject), p.(resource.Predicate), o.(resource.Object)), nil
}

func jterm2term(term jsonld.Term) (resource.Term, error) {
	switch term := term.(type) {
	case *jsonld.BlankNode:
		return resource.NewBlankNode(term.RawValue())
	case *jsonld.Literal:
		if len(term.Language) > 0 {
			return resource.NewLiteralWithLang(term.RawValue(), term.Language)
		}

		if term.Datatype != nil && len(term.Datatype.String()) > 0 {
			dt, err := resource.NewIRI(term.Datatype.RawValue())
			if err != nil {
				return nil, err
			}

			return resource.NewLiteralWithType(term.Value, &dt)
		}

		return resource.NewLiteral(term.Value)
	case *jsonld.Resource:
		return resource.NewIRI(term.RawValue())
	}

	return nil, fmt.Errorf("unknown resource.TermType")
}

// compile time check of interface
var _ resource.Parser = (*p)(nil)

type p struct{}

func (p p) Parse(r io.Reader, g *resource.Graph) (*resource.Graph, error) {
	return Parse(r, g)
}
