package rdf

import "testing"

func TestNormaliseSchemaOrg(t *testing.T) {
	iri := func(s string) IRI { i, _ := NewIRI(s); return i }
	lit, _ := NewLiteral("http://schema.org/name")
	bn, _ := NewBlankNode("b1")

	got := NormaliseSchemaOrg(NewTriple(iri("http://schema.org/Thing/1"), iri("http://schema.org/about"), iri("http://schema.org/CreativeWork")))
	if got.Subject.RawValue() != "https://schema.org/Thing/1" || got.Predicate.RawValue() != "https://schema.org/about" || got.Object.RawValue() != "https://schema.org/CreativeWork" {
		t.Errorf("got %s %s %s", got.Subject.RawValue(), got.Predicate.RawValue(), got.Object.RawValue())
	}

	// literals, blank nodes and other IRIs are left alone
	got = NormaliseSchemaOrg(NewTriple(bn, iri("http://schema.orgx/name"), lit))
	if got.Subject != Subject(bn) || got.Predicate.RawValue() != "http://schema.orgx/name" || got.Object != Object(lit) {
		t.Errorf("got %#v", got)
	}
}
