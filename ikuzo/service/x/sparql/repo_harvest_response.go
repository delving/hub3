package sparql

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

type responseWithContext struct {
	Head    Head           `json:"head"`
	Results contextResults `json:"results"`
}

type Head struct {
	Vars []string `json:"vars"`
}

type contextResults struct {
	Bindings []*contextBinding `json:"bindings"`
}

type contextBinding struct {
	S  *Entry `json:"s"`
	P  *Entry `json:"p"`
	O  *Entry `json:"o,omitempty"`
	P2 *Entry `json:"p2,omitempty"`
	O2 *Entry `json:"o2,omitempty"`
	P3 *Entry `json:"p3,omitempty"`
	O3 *Entry `json:"o3,omitempty"`
	P4 *Entry `json:"p4,omitempty"`
	O4 *Entry `json:"o4,omitempty"`
	G  *Entry `json:"g,omitempty"`
}

func newResponse(r io.Reader) (*responseWithContext, error) {
	var resp responseWithContext

	if err := json.NewDecoder(r).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func addTriple(g *rdf.Graph, s *Entry, p *Entry, o *Entry) error {
	t, err := newTriple(s, p, o)
	if err != nil {
		return err
	}

	g.Add(t)

	return nil
}

func newTriple(s *Entry, p *Entry, o *Entry) (*rdf.Triple, error) {
	subj, err := s.asSubject()
	if err != nil {
		return nil, err
	}

	pred, err := p.asPredicate()
	if err != nil {
		return nil, err
	}

	obj, err := o.asObject()
	if err != nil {
		return nil, err
	}

	return rdf.NewTriple(subj, pred, obj), nil
}

func (r *responseWithContext) Graph() (*rdf.Graph, error) {
	g := rdf.NewGraph()
	for _, b := range r.Results.Bindings {
		if err := addTriple(g, b.S, b.P, b.O); err != nil {
			return nil, err
		}

		if b.P2 != nil && b.O2 != nil {
			if err := addTriple(g, b.O, b.P2, b.O2); err != nil {
				return nil, err
			}
		}

		if b.P3 != nil && b.O3 != nil {
			if err := addTriple(g, b.O2, b.P3, b.O3); err != nil {
				return nil, err
			}
		}

		if b.P4 != nil && b.O4 != nil {
			if err := addTriple(g, b.O3, b.P4, b.O4); err != nil {
				return nil, err
			}
		}
	}

	return g, nil
}

func (r *responseWithContext) NTriples() (string, error) {
	var buf bytes.Buffer

	g, err := r.Graph()
	if err != nil {
		return "", err
	}

	if err := ntriples.Serialize(g, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *responseWithContext) MappingXML(subject rdf.Subject, wikibaseType string) (string, error) {
	g, err := r.Graph()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	cfg := mappingxml.FilterConfig{Subject: subject}

	if wikibaseType != "" {
		p, _ := rdf.NewIRI(wikibaseType)
		cfg.WikiBaseTypePredicate = rdf.Predicate(p)
	}

	if err := mappingxml.Serialize(g, &buf, &cfg); err != nil {
		return "", err
	}

	return buf.String(), nil
}
