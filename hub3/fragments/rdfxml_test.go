package fragments_test

import (
	"os"
	"testing"

	r "github.com/kiivihal/rdf2go"
	"github.com/knakk/rdf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/hub3/hub3/fragments"
)

var _ = Describe("Rdf", func() {

	Context("when parsing from an io.Reader", func() {

		It("should extract a list of triples", func() {
			dat, err := os.Open("test_data/1.rdf")
			Expect(err).ToNot(HaveOccurred())
			triples, err := DecodeRDFXML(dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(triples).ToNot(BeEmpty())
		})
	})

	Context("when converting a list of triples", func() {

		It("should create a fragments.ResourceMap", func() {
			dat, err := os.Open("test_data/1.rdf")
			Expect(err).ToNot(HaveOccurred())
			triples, err := DecodeRDFXML(dat)
			Expect(err).ToNot(HaveOccurred())
			rm, err := NewResourceMapFromXML(triples)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).ToNot(BeNil())
			Expect(len(rm.Resources())).To(Equal(6))
			fr, ok := rm.GetResource("http://sws.geonames.org/2759059")
			Expect(ok).To(BeTrue())
			err = fr.SetEntries(rm)
			Expect(err).ToNot(HaveOccurred())
			Expect(fr.Entries[0].Order).To(Equal(28))
			Expect(fr.Entries[3].Order).To(Equal(31))
		})
	})

})

func TestTripleConversion(t *testing.T) {

	tr := func(s rdf.Subject, p rdf.Predicate, o rdf.Object) rdf.Triple {
		return rdf.Triple{
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

	iS, _ := rdf.NewIRI(s)
	oS := r.NewResource(s)
	iP, _ := rdf.NewIRI(p)
	oP := r.NewResource(p)
	iB, _ := rdf.NewBlank(b)
	oB := r.NewBlankNode(b)
	iL, _ := rdf.NewLiteral(o)
	oL := r.NewLiteral(o)
	intType := "http://www.w3.org/2001/XMLSchema#integer"
	intTypeIRI, _ := rdf.NewIRI(intType)
	iTL := rdf.NewTypedLiteral(oTyped, intTypeIRI)
	oTL := r.NewLiteralWithDatatype(oTyped, r.NewResource(intType))

	iLL, _ := rdf.NewLangLiteral(o, oLang)
	oLL := r.NewLiteralWithLanguage(o, oLang)

	tt := []struct {
		name   string
		input  rdf.Triple
		output *r.Triple
	}{
		{"bnode object", tr(iS, iP, iB), r.NewTriple(oS, oP, oB)},
		{"bnode subject", tr(iB, iP, iB), r.NewTriple(oB, oP, oB)},
		{"bnode subject with Literal", tr(iB, iP, iL), r.NewTriple(oB, oP, oL)},
		{"literal object", tr(iS, iP, iL), r.NewTriple(oS, oP, oL)},
		{"literal language object", tr(iS, iP, iLL), r.NewTriple(oS, oP, oLL)},
		{"typed literal object", tr(iS, iP, iTL), r.NewTriple(oS, oP, oTL)},
		{"resource object", tr(iS, iP, iS), r.NewTriple(oS, oP, oS)},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			newTriple := ConvertTriple(tc.input)
			if newTriple.String() != tc.output.String() {
				t.Fatalf("%s conversion of %v to new triple should be %v; got %v", tc.name, tc.input, tc.output, newTriple)
			}
		})
	}
}
