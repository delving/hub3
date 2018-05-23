package fragments_test

import (
	. "github.com/delving/rapid-saas/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {

	Describe("when creating a resource map", func() {

		It("return an empty map when the graph is empty", func() {
			rm, err := CreateResourceMap(r.NewGraph(""))
			Expect(err).To(HaveOccurred())
			Expect(rm).To(BeEmpty())
		})

		It("return an non empty map when the graph is not empty", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := CreateResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).ToNot(BeEmpty())
			Expect(rm).To(HaveLen(12))
			Expect(rm).To(HaveKey("http://data.jck.nl/resource/aggregation/jhm-foto/F900893"))
		})
	})

	Describe("when appending a triple", func() {

		It("should add the subject to the resource map", func() {
			rm := make(map[string]*FragmentResource)
			Expect(rm).To(BeEmpty())
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).To(HaveLen(1))
			Expect(rm).To(HaveKey(t.GetSubjectID()))
			fr, ok := rm[t.GetSubjectID()]
			Expect(ok).To(BeTrue())
			Expect(fr.Types).To(HaveLen(1))
		})

		It("should add the subject only once", func() {
			rm := make(map[string]*FragmentResource)
			Expect(rm).To(BeEmpty())
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())
			err = (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).To(HaveLen(1))
		})

	})

})
