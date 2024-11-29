package rdfxml

import (
	"os"
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
