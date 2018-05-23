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

		It("should have a FragmentResource for each map key", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := CreateResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).ToNot(BeEmpty())

			subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"
			fr, ok := rm[subject]
			Expect(ok).To(BeTrue())
			Expect(fr.ID).To(Equal(subject))
			Expect(fr.Types).To(ContainElement("http://www.openarchives.org/ore/terms/Aggregation"))
			Expect(fr.Types).To(HaveLen(1))
			Expect(fr.ObjectIDs).To(HaveLen(6))
			Expect(fr.ObjectIDs).ToNot(ContainElement(subject))
			Expect(fr.Predicates).To(HaveLen(6))
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

		It("should add not add objectIDS for rdfType", func() {
			rm := make(map[string]*FragmentResource)
			Expect(rm).To(BeEmpty())
			subject := NSRef("1")
			t := r.NewTriple(
				subject,
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())

			entry, ok := rm[r.GetResourceID(subject)]
			Expect(ok).To(BeTrue())
			Expect(entry.ObjectIDs).To(HaveLen(0))
		})

		It("should add add objectIDS for resources", func() {
			rm := make(map[string]*FragmentResource)
			Expect(rm).To(BeEmpty())
			subject := NSRef("1")
			t := r.NewTriple(
				subject,
				NSRef("title"),
				NSRef("myBook"),
			)
			err := (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())

			err = (AppendTriple(rm, t))
			Expect(err).ToNot(HaveOccurred())

			entry, ok := rm[r.GetResourceID(subject)]
			Expect(ok).To(BeTrue())
			Expect(entry.ObjectIDs).To(HaveLen(1))
		})

	})

	Describe("when creating a fragment entry", func() {

		It("should return an ID for a resource", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			entry, id := CreateFragmentEntry(t)
			Expect(id).ToNot(BeEmpty())
			Expect(id).To(Equal(r.GetResourceID(t.Object)))
			Expect(entry.ID).To(Equal(id))
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.Datatype).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.Entrytype).To(Equal("Resource"))
		})

		It("should return an ID for a BlankNode", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewBlankNode("book"),
			)
			entry, id := CreateFragmentEntry(t)
			Expect(id).ToNot(BeEmpty())
			Expect(id).To(Equal(r.GetResourceID(t.Object)))
			Expect(id).To(HavePrefix("_:"))
			Expect(id).To(Equal("_:book"))
			Expect(entry.ID).To(Equal(id))
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.Datatype).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.Entrytype).To(Equal("Bnode"))
		})

		It("should return no ID for a Literal", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteral("book"),
			)
			entry, id := CreateFragmentEntry(t)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("book"))
			Expect(entry.Datatype).To(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.Entrytype).To(Equal("Literal"))
		})

		It("should have a language when the triple has a language", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteralWithLanguage("book", "en"),
			)
			entry, id := CreateFragmentEntry(t)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("book"))
			Expect(entry.Datatype).To(BeEmpty())
			Expect(entry.Language).To(Equal("en"))
			Expect(entry.Entrytype).To(Equal("Literal"))
		})

		It("should have a datatype for non-string", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteralWithDatatype("1", r.NewResource("http://www.w3.org/2001/XMLSchema#decimal")),
			)
			entry, id := CreateFragmentEntry(t)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("1"))
			Expect(entry.Datatype).ToNot(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.Entrytype).To(Equal("Literal"))
		})
	})

})
