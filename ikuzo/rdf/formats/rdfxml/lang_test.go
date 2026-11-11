package rdfxml

import (
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
)

// xml:lang is inherited by every literal inside the element that sets it
// (RDF/XML syntax §2.7), unless an inner element sets its own.
func TestParseXMLLangInheritance(t *testing.T) {
	const head = `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/"`

	tests := []struct {
		name string
		doc  string
		want map[string]string // literal value -> language
	}{
		{
			"on rdf:RDF",
			head + ` xml:lang="nl"><rdf:Description rdf:about="http://x/1"><dc:title>t</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"t": "nl"},
		},
		{
			"on rdf:RDF, overridden by node",
			head + ` xml:lang="nl"><rdf:Description rdf:about="http://x/1" xml:lang="en"><dc:title>t</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"t": "en"},
		},
		{
			"on node element",
			head + `><rdf:Description rdf:about="http://x/1" xml:lang="nl"><dc:title>t</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"t": "nl"},
		},
		{
			"does not leak to sibling node",
			head + `><rdf:Description rdf:about="http://x/1" xml:lang="nl"><dc:title>a</dc:title></rdf:Description><rdf:Description rdf:about="http://x/2"><dc:title>b</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"a": "nl", "b": ""},
		},
		{
			"on outer node, nested node",
			head + `><rdf:Description rdf:about="http://x/1" xml:lang="nl"><dc:creator><rdf:Description rdf:about="http://x/2"><dc:title>t</dc:title></rdf:Description></dc:creator></rdf:Description></rdf:RDF>`,
			map[string]string{"t": "nl"},
		},
		{
			"on property element, nested node",
			head + `><rdf:Description rdf:about="http://x/1"><dc:creator xml:lang="nl"><rdf:Description rdf:about="http://x/2"><dc:title>a</dc:title><dc:subject>b</dc:subject></rdf:Description></dc:creator></rdf:Description></rdf:RDF>`,
			map[string]string{"a": "nl", "b": "nl"},
		},
		{
			"on property element, nested node with own lang",
			head + `><rdf:Description rdf:about="http://x/1"><dc:creator xml:lang="nl"><rdf:Description rdf:about="http://x/2" xml:lang="en"><dc:title>t</dc:title></rdf:Description></dc:creator></rdf:Description></rdf:RDF>`,
			map[string]string{"t": "en"},
		},
		{
			"on parseType Resource",
			head + `><rdf:Description rdf:about="http://x/1"><dc:creator xml:lang="nl" rdf:parseType="Resource"><dc:title>a</dc:title><dc:subject>b</dc:subject></dc:creator><dc:description>c</dc:description></rdf:Description></rdf:RDF>`,
			map[string]string{"a": "nl", "b": "nl", "c": ""},
		},
		{
			"property lang does not leak to sibling property",
			head + `><rdf:Description rdf:about="http://x/1" xml:lang="nl"><dc:subject xml:lang="en">s</dc:subject><dc:title>t</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"s": "en", "t": "nl"},
		},
		{
			"explicit empty lang resets",
			head + ` xml:lang="nl"><rdf:Description rdf:about="http://x/1"><dc:title xml:lang="">t</dc:title></rdf:Description></rdf:RDF>`,
			map[string]string{"t": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(strings.NewReader(tt.doc), nil, "seed")
			if err != nil {
				t.Fatal(err)
			}

			got := map[string]string{}
			for _, triple := range g.Triples() {
				if l, ok := triple.Object.(rdf.Literal); ok {
					got[l.RawValue()] = l.Lang()
				}
			}

			for value, lang := range tt.want {
				if got[value] != lang {
					t.Errorf("literal %q: lang %q, want %q", value, got[value], lang)
				}
			}
		})
	}
}
