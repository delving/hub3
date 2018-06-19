package fragments_test

import (
	"github.com/delving/rapid-saas/config"
	. "github.com/delving/rapid-saas/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {

	Describe("when creating a resource map", func() {

		It("return an empty map when the graph is empty", func() {
			rm, err := NewResourceMap(r.NewGraph(""))
			Expect(err).To(HaveOccurred())
			Expect(rm.Resources()).To(BeEmpty())
		})

		It("return an non empty map when the graph is not empty", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := NewResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).ToNot(BeNil())
			rs := rm.Resources()
			Expect(rs).ToNot(BeEmpty())
			Expect(rs).To(HaveLen(12))
			Expect(rs).To(HaveKey("http://data.jck.nl/resource/aggregation/jhm-foto/F900893"))
		})

		It("should have a FragmentResource for each map key", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := NewResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm.Resources()).ToNot(BeEmpty())

			subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"
			fr, ok := rm.GetResource(subject)
			Expect(ok).To(BeTrue())
			Expect(fr.ID).To(Equal(subject))
			Expect(fr.Types).To(ContainElement("http://www.openarchives.org/ore/terms/Aggregation"))
			Expect(fr.Types).To(HaveLen(1))
			Expect(fr.ObjectIDs()).To(HaveLen(6))
			// todo properly check for not referring to itself
			Expect(fr.ObjectIDs()).ToNot(ContainElement(subject))
			Expect(fr.Predicates()).To(HaveLen(6))
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
			Expect(entry.ObjectIDs()).To(HaveLen(0))
		})

		It("should add objectIDS for resources", func() {
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
			Expect(entry.ObjectIDs()).To(HaveLen(1))
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
			Expect(entry.Triple).ToNot(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Resource"))
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
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Bnode"))
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
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Literal"))
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
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Language).To(Equal("en"))
			Expect(entry.EntryType).To(Equal("Literal"))
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
			Expect(entry.DataType).ToNot(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Literal"))
		})
	})

	Describe("when creating FragmentReferrerContext", func() {

		Context("and determining the level", func() {
			fb, _ := testDataGraph(false)
			rm, _ := NewResourceMap(fb.Graph)
			subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"

			It("should not have 0 as level", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(fb).ToNot(BeNil())
				rm, err := NewResourceMap(fb.Graph)
				Expect(err).ToNot(HaveOccurred())
				Expect(rm.Resources()).ToNot(BeEmpty())

				fr, ok := rm.GetResource(subject)
				Expect(ok).To(BeTrue())

				level := fr.GetLevel()
				Expect(level).To(Equal(int32(1)))
			})
			It("should throw an error when the subject is unknown", func() {
				Expect(rm).ToNot(BeNil())
				err := rm.SetContextLevels("urn:unknown")
				Expect(err).To(HaveOccurred())
			})

			config.InitConfig()
			It("should determine its level by the number of context is has", func() {
				Expect(rm).ToNot(BeNil())
				err := rm.SetContextLevels(subject)
				Expect(err).ToNot(HaveOccurred())

				providedCHO, ok := rm.GetResource("http://data.jck.nl/resource/document/jhm-foto/F900893")
				Expect(providedCHO).ToNot(BeNil())
				Expect(ok).To(BeTrue())
				Expect(providedCHO.Context).To(HaveLen(1))
				Expect(providedCHO.Context[0].GetSubjectClass()).To(HaveLen(1))
				Expect(providedCHO.Context[0].Level).To(Equal(int32(1)))
				Expect(providedCHO.GetLevel()).To(Equal(int32(2)))
				label, lang := providedCHO.GetLabel()
				Expect(label).To(Equal(""))
				Expect(lang).To(Equal(""))

				skosConcept, ok := rm.GetResource("http://data.jck.nl/resource/skos/thesau/90000072")
				Expect(skosConcept).ToNot(BeNil())
				Expect(ok).To(BeTrue())
				Expect(skosConcept.Context).To(HaveLen(2))
				Expect(skosConcept.GetLevel()).To(Equal(int32(3)))
				Expect(skosConcept.Context[1].Level).To(Equal(int32(2)))
				Expect(skosConcept.Context[1].GetSubjectClass()).To(HaveLen(1))
				Expect(skosConcept.Context[0].Level).To(Equal(int32(1)))
				Expect(skosConcept.Context[0].GetSubjectClass()).To(HaveLen(1))
				Expect(config.Config.RDFTag.Label).To(HaveLen(2))
				label, lang = skosConcept.GetLabel()
				Expect(label).To(Equal("grafsteen"))
				Expect(lang).To(Equal("nl"))
			})
		})
	})

	Describe("when creating a ResultSummary", func() {

		Context("from a resource entry", func() {

			It("should only set a field once", func() {
				entry1 := &ResourceEntry{
					Value: "test1",
					Tags:  []string{"label"},
				}
				entry2 := &ResourceEntry{
					Value: "test2",
					Tags:  []string{"label"},
				}
				sum := &ResultSummary{}
				Expect(sum.Title).To(BeEmpty())
				Expect(sum.Thumbnail).To(BeEmpty())
				sum.AddEntry(entry1)
				Expect(sum.Title).To(Equal("test1"))
				sum.AddEntry(entry2)
				Expect(sum.Thumbnail).To(BeEmpty())
				Expect(sum.Title).To(Equal("test1"))
				Expect(sum.Thumbnail).To(BeEmpty())
			})
		})
	})

	Describe("when creating a Header", func() {

		fb, _ := testDataGraph(false)
		//rm, _ := NewResourceMap(fb.Graph)
		//subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"

		Context("from a FragmentGraph", func() {

			//skosConcept, _ := rm.GetResource("http://data.jck.nl/resource/skos/thesau/90000072")
			//entry := skosConcept.Predicates["http://www.w3.org/2004/02/skos/core#prefLabel"]
			header := fb.FragmentGraph().CreateHeader("fragment")

			It("should set the OrgID", func() {
				Expect(header.OrgID).To(Equal("rapid"))
			})

			It("should set the spec", func() {
				Expect(header.Spec).To(Equal("test-spec"))
			})

			It("should set the Revision", func() {
				Expect(header.Revision).To(Equal(int32(1)))
			})

			It("should set the hubID", func() {
				Expect(header.HubID).ToNot(Equal(""))
			})

			It("should have no tags", func() {
				Expect(header.Tags).To(BeEmpty())
			})

			It("should have a docType", func() {
				Expect(header.GetDocType()).To(Equal("fragment"))
			})

		})

		Context("and adding Tags", func() {

			It("should only add a tag", func() {
				header := fb.FragmentGraph().CreateHeader("")
				Expect(header.Tags).To(BeEmpty())
				header.AddTags("tag1")
				Expect(header.Tags).ToNot(BeEmpty())
				Expect(header.Tags).To(HaveLen(1))

			})

			It("should not add a tag twice", func() {
				header := fb.FragmentGraph().CreateHeader("")
				Expect(header.Tags).To(BeEmpty())
				header.AddTags("tag1", "tag2")
				header.AddTags("tag1")
				Expect(header.Tags).ToNot(BeEmpty())
				Expect(header.Tags).To(HaveLen(2))

			})

		})

	})

})
