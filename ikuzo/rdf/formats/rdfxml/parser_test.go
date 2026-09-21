package rdfxml

import (
	"os"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	xmlrdf "github.com/delving/hub3/ikuzo/rdf/formats/rdfxml/internal/rdf"
	"github.com/matryer/is"
)

var testOrgID = "test"

func TestParseXMLRDF(t *testing.T) {
	is := is.New(t)
	dat, err := os.Open("testdata/1.rdf")
	is.NoErr(err)

	g, err := Parse(dat, nil, "test_seed")
	is.NoErr(err)

	is.Equal(g.Len(), 48)
	is.Equal(g.Triples()[0].Subject.RawValue(), "http://data.brabantcloud.nl/resource/aggregation/enb-10-beeldmateriaal/enb-10.beeldmateriaal-db129c8f-50bc-930c-5f0e-90bc80ecbe30-ab16f200-8232-11e5-b3dd-0741013467d3")
}

// Regression for #3548: rdf:about/rdf:resource values are URI references,
// never QNames. Absolute URIs with a non-hierarchical scheme (urn:, mailto:)
// match the prefix:suffix shape and used to panic the decoder with
// `no name space found for prefix: "urn"`, rejecting every record with an
// urn: subject (museum-helmond-objecten, museum-klok-en-peel).
func TestParseURNAbout(t *testing.T) {
	is := is.New(t)

	record := `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	         xmlns:edm="http://www.europeana.eu/schemas/edm/"
	         xmlns:dc="http://purl.org/dc/elements/1.1/">
	  <edm:WebResource rdf:about="urn:museum-helmond-objecten/93-026__">
	    <dc:title>Zonder titel</dc:title>
	  </edm:WebResource>
	  <edm:ProvidedCHO rdf:about="http://example.org/cho/93-026">
	    <dc:creator>Eerste</dc:creator>
	    <dc:creator>Tweede</dc:creator>
	    <edm:isShownBy rdf:resource="urn:museum-helmond-objecten/93-026__"/>
	    <dc:relation rdf:resource="mailto:info@example.org"/>
	  </edm:ProvidedCHO>
	</rdf:RDF>`

	g, err := Parse(strings.NewReader(record), nil, "test_seed")
	is.NoErr(err)

	var sawURNSubject, sawURNObject, sawMailto bool
	var creators []string
	for _, t := range g.Triples() {
		if t.Subject.RawValue() == "urn:museum-helmond-objecten/93-026__" {
			sawURNSubject = true
		}
		if t.Object.RawValue() == "urn:museum-helmond-objecten/93-026__" {
			sawURNObject = true
		}
		if t.Object.RawValue() == "mailto:info@example.org" {
			sawMailto = true
		}
		if t.Predicate.RawValue() == "http://purl.org/dc/elements/1.1/creator" &&
			t.Subject.RawValue() == "http://example.org/cho/93-026" {
			creators = append(creators, t.Object.RawValue())
		}
	}
	is.True(sawURNSubject) // urn: rdf:about must survive as-is
	is.True(sawURNObject)  // urn: rdf:resource must survive as-is
	is.True(sawMailto)     // mailto: rdf:resource must survive as-is
	// declared prefixes in values keep expanding is NOT claimed here; but
	// insertion order of repeated properties must hold (the #3548 point).
	is.Equal(creators, []string{"Eerste", "Tweede"})
}

func TestParseNestedXMLRDF(t *testing.T) {
	is := is.New(t)
	dat, err := os.Open("../../index/testdata/rdf_brocade.rdf.xml")
	is.NoErr(err)

	g, err := Parse(dat, nil, "test_seed")
	is.NoErr(err)

	is.Equal(g.Len(), 105)
	is.Equal(g.Triples()[0].Subject.RawValue(), "https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:538499")

	var found bool
	for _, t := range g.Triples() {
		if t.Subject.RawValue() == "https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1" {
			found = true
		}
	}

	is.True(found) // the subject for the nested about should be found
}

func TestTripleConversion(t *testing.T) {
	tr := func(s xmlrdf.Subject, p xmlrdf.Predicate, o xmlrdf.Object) xmlrdf.Triple {
		return xmlrdf.Triple{
			Subj: s,
			Pred: p,
			Obj:  o,
		}
	}

	s := "http://example.com/subject"
	p := "http://example.com/predicate"
	b := "b1"
	o := "hello"
	oLang := "en"
	oTyped := "1"

	iS, _ := xmlrdf.NewIRI(s)
	oS, _ := rdf.NewIRI(s)
	iP, _ := xmlrdf.NewIRI(p)
	oP, _ := rdf.NewIRI(p)
	iB, _ := xmlrdf.NewBlank(b)
	oB, _ := rdf.NewBlankNode(b)
	iL, _ := xmlrdf.NewLiteral(o)
	oL, _ := rdf.NewLiteral(o)
	intType := "http://www.w3.org/2001/XMLSchema#integer"
	intTypeIRINew, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#integer")
	intTypeIRI, _ := xmlrdf.NewIRI(intType)
	iTL := xmlrdf.NewTypedLiteral(oTyped, intTypeIRI)
	oTL, _ := rdf.NewLiteralWithType(oTyped, intTypeIRINew)

	iLL, _ := xmlrdf.NewLangLiteral(o, oLang)
	oLL, _ := rdf.NewLiteralWithLang(o, oLang)

	tt := []struct {
		name   string
		input  xmlrdf.Triple
		output *rdf.Triple
	}{
		{"bnode object", tr(iS, iP, iB), rdf.NewTriple(oS, oP, oB)},
		{"bnode subject", tr(iB, iP, iB), rdf.NewTriple(oB, oP, oB)},
		{"bnode subject with Literal", tr(iB, iP, iL), rdf.NewTriple(oB, oP, oL)},
		{"literal object", tr(iS, iP, iL), rdf.NewTriple(oS, oP, oL)},
		{"literal language object", tr(iS, iP, iLL), rdf.NewTriple(oS, oP, oLL)},
		{"typed literal object", tr(iS, iP, iTL), rdf.NewTriple(oS, oP, oTL)},
		{"resource object", tr(iS, iP, iS), rdf.NewTriple(oS, oP, oS)},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			newTriple, err := convertTriple(tc.input)
			is.NoErr(err)
			if newTriple.String() != tc.output.String() {
				t.Fatalf("%s conversion of %v to new triple should be %v; got %v", tc.name, tc.input, tc.output, newTriple)
			}
		})
	}
}
