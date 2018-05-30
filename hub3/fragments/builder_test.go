package fragments_test

import (
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "github.com/delving/rapid-saas/config"
	. "github.com/delving/rapid-saas/hub3/fragments"
)

var _ = Describe("Builder", func() {

	Describe("FragmentBuilder", func() {
		fb, err := testDataGraph(false)

		Context("when creating a new builder", func() {

			It("should have empty graph when parse is called", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fb.Graph.Len()).ToNot(Equal(0))
			})

			It("should have an empty map of resource labels", func() {
				Expect(fb.ResourceLabels).To(BeEmpty())
			})

			It("should have an empty map of resources", func() {
				Expect(fb.Resources).To(BeNil())
			})

		})

		Context("when giving access", func() {

			It("should give back a pointer to the FragmentGraph", func() {
				Expect(fb.FragmentGraph()).ToNot(BeNil())
			})

			It("should give back a pointer to the FragmentGraph that can be converted to JSON", func() {
				Expect(fb.Doc()).ToNot(BeNil())
			})

		})

		Context("when a graphName is present", func() {
			spec := "test-spec"
			rev := int32(1)
			ng := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893/graph"
			fg := testFragmentGraph(spec, rev, ng)

			It("should extract the about or source uri", func() {
				sourceURI := fg.GetAboutURI()
				Expect(sourceURI).ToNot(BeEmpty())
				Expect(sourceURI).ToNot(HaveSuffix("/graph"))
			})
		})

		Describe("When receiving a Triple", func() {

			spec := "test-spec"
			rev := int32(1)
			ng := "urn:1/graph"
			fg := testFragmentGraph(spec, rev, ng)
			fb := NewFragmentBuilder(fg)

			//Context("with an object resource", func() {

			//t := r.NewTriple(URIRef("http://example.com/resource/1"), URIRef("urn:subject"), URIRef("urn:target"))
			//f, err := fb.CreateFragment(t)

			//It("should not have a header without calling AddHeader", func() {
			//Expect(t).ToNot(BeNil())
			//Expect(err).ToNot(HaveOccurred())
			//Expect(f).ToNot(BeNil())
			////Expect(f.GetSpec()).ToNot(Equal(spec))

			//err := f.AddHeader(fb)
			//Expect(err).ToNot(HaveOccurred())
			//Expect(f.GetSpec()).To(Equal(spec))
			//})

			//It("should have a  lodKey", func() {
			//Expect(f.GetLodKey()).To(Equal("/1"))
			//})

			//It("should have a spec", func() {
			//Expect(f.GetSpec()).To(Equal(spec))

			//Expect(f.GetPredicate()).To(Equal("urn:subject"))
			//})

			//It("should have a revision number", func() {
			//Expect(f.GetRevision()).To(Equal(rev))
			//})

			//It("should have a NamedGraphURI", func() {
			//Expect(f.GetNamedGraphURI()).To(Equal(ng))
			//})

			//It("should have an n-triple", func() {
			//t := f.GetTriple()
			//Expect(t).ToNot(BeEmpty())
			//})

			//It("should have a quad with the NamedGraphURI", func() {
			//q := f.Quad()
			//Expect(q).ToNot(BeEmpty())
			//Expect(q).To(HaveSuffix("<urn:target> <%s> .", f.GetNamedGraphURI()))
			//})

			//It("should have an id that is a hashed version of the Quad", func() {
			//id := f.ID()
			//Expect(id).ToNot(BeEmpty())
			//hash := CreateHash(f.Quad())
			//Expect(id).To(Equal(hash))
			//})

			//It("should have a subject without <>", func() {
			//r := f.GetSubject()
			//Expect(r).To(Equal("http://example.com/resource/1"))
			//Expect(r).ToNot(HaveSuffix("%s", ">"))
			//Expect(r).ToNot(HavePrefix("%s", "<"))
			//})

			//It("should have predicate without <>", func() {
			//r := f.GetPredicate()
			//Expect(r).To(Equal("urn:subject"))
			//Expect(r).ToNot(HaveSuffix("%s", ">"))
			//Expect(r).ToNot(HavePrefix("%s", "<"))
			//})

			//It("should have an object", func() {
			//r := f.GetObject()
			//Expect(r).To(Equal("urn:target"))
			//Expect(r).ToNot(HaveSuffix("%s", ">"))
			//Expect(r).ToNot(HavePrefix("%s", "<"))
			//})

			//It("should have resource as objecttype", func() {
			//t := f.GetObjectType()
			//Expect(t).ToNot(BeNil())
			//Expect(t).To(Equal(ObjectType_RESOURCE))
			//})

			//It("should create a lodKey with a url fragment", func() {
			//t := r.NewTriple(URIRef("http://example.com/resource/1#a"), URIRef("urn:subject"), URIRef("urn:target"))
			//f, err := fb.CreateFragment(t)
			//Expect(err).ToNot(HaveOccurred())

			//key, err := f.CreateLodKey()
			//Expect(key).ToNot(BeEmpty())
			//Expect(err).ToNot(HaveOccurred())

			//Expect(key).To(HaveSuffix("#a"))
			//})

			//It("should give back on empty lodKey when the path does not start with Config.Lod.Resource HavePrefix", func() {
			//t := r.NewTriple(URIRef("http://example.com/bresource/1#a"), URIRef("urn:subject"), URIRef("urn:target"))
			//f, err := fb.CreateFragment(t)
			//Expect(err).ToNot(HaveOccurred())

			//key, err := f.CreateLodKey()
			//Expect(key).To(BeEmpty())
			//Expect(err).ToNot(HaveOccurred())

			//})

			//})

			Context("When receiving a object resource link", func() {

				g := r.NewGraph(c.Config.RDF.BaseURL)
				t1 := r.NewTriple(NSRef("1"), NSRef("subject"), NSRef("2"))
				t2 := r.NewTriple(NSRef("1"), NSRef("subject"), NSRef("3"))
				t3 := r.NewTriple(
					NSRef("1"),
					NSRef("subject"),
					r.NewResource("https://data.cultureelerfgoed.nl/term/id/cht/99efdcca-cce0-4629-adfb-becab8381183"),
				)
				t4 := r.NewTriple(
					NSRef("1"),
					r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
					r.NewResource("http://www.europeana.eu/schemas/edm/WebResource"),
				)
				g.Add(t1)
				g.Add(t2)
				g.Add(t3)
				g.Add(t4)
				g.AddTriple(NSRef("2"), NSRef("prefLabel"), r.NewLiteral("subject of 2"))
				g.AddTriple(NSRef("1"), NSRef("title"), r.NewLiteral("2"))
				fb.Graph = g

				It("should mark the fragment not graphExternal if subject present in graph", func() {
					//frag, err := fb.CreateFragment(t1)
					//Expect(t1).ToNot(BeNil())
					//Expect(err).ToNot(HaveOccurred())
					//Expect(frag).ToNot(BeNil())
					//external := fb.IsGraphExternal(t1.Object)
					//Expect(external).To(BeFalse())
					// todo reimplement with tags
					//Expect(frag.GraphExternalLink).To(BeFalse())

				})

				It("should mark the fragment not graphExternal if subject present in graph", func() {
					//frag, err := fb.CreateFragment(t2)
					//Expect(t2).ToNot(BeNil())
					//Expect(err).ToNot(HaveOccurred())
					//Expect(frag).ToNot(BeNil())
					//external := fb.IsGraphExternal(t2.Object)
					//Expect(external).To(BeTrue())
					// todo reimplement with tags
					// Expect(frag.GraphExternalLink).To(BeTrue())
				})

				It("should mark the fragment as domainExternal when the host differs from the RDF base url", func() {
					//frag, err := fb.CreateFragment(t3)
					//Expect(t3).ToNot(BeNil())
					//Expect(err).ToNot(HaveOccurred())
					//Expect(frag).ToNot(BeNil())
					//external, err := fb.IsDomainExternal(frag.Object)
					//Expect(err).ToNot(HaveOccurred())
					//Expect(external).To(BeTrue())
					// todo reimplement with tags
					//Expect(frag.DomainExternalLink).To(BeTrue())
				})

				It("should mark the fragment as not domainExternal when the host equals the RDF base url", func() {
					//frag, err := fb.CreateFragment(t2)
					//Expect(t2).ToNot(BeNil())
					//Expect(err).ToNot(HaveOccurred())
					//Expect(frag).ToNot(BeNil())
					//external, err := fb.IsDomainExternal(frag.Object)
					//Expect(err).ToNot(HaveOccurred())
					//Expect(external).To(BeFalse())
					// todo reimplement with tags
					//Expect(frag.DomainExternalLink).To(BeFalse())
				})

				It("should not make type links as external", func() {
					//frag, err := fb.CreateFragment(t4)
					//Expect(t4).ToNot(BeNil())
					//Expect(err).ToNot(HaveOccurred())
					//Expect(frag).ToNot(BeNil())
					// todo reimplement with tags
					//Expect(frag.IsTypeLink()).To(BeTrue())
					//Expect(frag.GetTypeLink()).To(BeTrue())
				})

			})

			Context("when getting the ObjectXSDType", func() {

				It("should return the XSD label", func() {

				})

				It("should trim <>", func() {
					t, err := GetObjectXSDType("<http://www.w3.org/2001/XMLSchema#date>")
					Expect(err).ToNot(HaveOccurred())
					Expect(t).ToNot(BeNil())
				})
			})

			Context("when receiving a triple with a literal object", func() {

				//t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "", ObjectXSDType_STRING))
				//f, err := fb.CreateFragment(t)

				//It("should have literal as objecttype", func() {
				//Expect(err).ToNot(HaveOccurred())
				//t := f.GetObjectType()
				//Expect(t).ToNot(BeNil())
				//Expect(t).To(Equal(ObjectType_LITERAL))
				//})

				//It("should have no language", func() {
				//Expect(f.Language).To(Equal(""))
				//})

				//It("should have string as datatype", func() {
				//Expect(f.DataType).To(Equal(ObjectXSDType_STRING))
				//})

				//It("should have http://www.w3.org/2001/XMLSchema#string as default xsdRaw", func() {
				//Expect(f.GetXSDRaw()).To(Equal("xsd:string"))
				//})
			})

			Context("when receiving a triple with a literal and language", func() {
				//t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "en", ObjectXSDType_STRING))
				//f, err := fb.CreateFragment(t)

				//It("should have a language", func() {
				//Expect(err).ToNot(HaveOccurred())
				//Expect(f.Language).To(Equal("en"))
				//})

				//It("should have string as datatype", func() {
				//Expect(f.DataType).To(Equal(ObjectXSDType_STRING))
				//})

				//It("should have xsd:string as default xsdRaw", func() {
				//Expect(f.GetXSDRaw()).To(Equal("xsd:string"))
				//})
			})

			Context("when receiving a triple with literal and type", func() {

				It("should have the custom dataType", func() {
					//t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "", ObjectXSDType_DATE))
					//f, err := fb.CreateFragment(t)
					//Expect(err).ToNot(HaveOccurred())
					//Expect(f.GetDataType()).To(Equal(ObjectXSDType_DATE))
					//Expect(f.GetXSDRaw()).To(Equal("xsd:date"))
				})
			})
		})

	})

	Describe("when creating a fragment", func() {

		//fg := testFragmentGraph("test", int32(1), "urn:1/graph")
		//fb := NewFragmentBuilder(fg)

		Context("and converting it to a BulkIndexRequest", func() {

			//t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "en", ObjectXSDType_STRING))
			//f, err := fb.CreateFragment(t)

			//It("the fragment should be valid", func() {
			//Expect(err).ToNot(HaveOccurred())
			//Expect(f).ToNot(BeNil())
			//})

			//It("should create the BulkIndexRequest", func() {
			//bir, err := f.CreateBulkIndexRequest()
			//Expect(err).ToNot(HaveOccurred())
			//Expect(bir).ToNot(BeNil())
			//})

			//It("should have a valid header", func() {
			//bir, err := f.CreateBulkIndexRequest()
			//Expect(err).ToNot(HaveOccurred())
			//lines, err := bir.Source()
			//Expect(err).ToNot(HaveOccurred())
			//header := lines[0]
			////body := lines[1]
			//var h interface{}
			//err = json.Unmarshal([]byte(header), &h)
			//Expect(err).ToNot(HaveOccurred())
			//m := h.(map[string]interface{})
			//Expect(m["index"]).To(HaveKeyWithValue("_id", f.ID()))
			//Expect(m["index"]).To(HaveKeyWithValue("_type", DocType))
			//Expect(m["index"]).To(HaveKeyWithValue("_index", c.Config.ElasticSearch.IndexName))
			//})

			//It("should have a valid body", func() {
			//bir, err := f.CreateBulkIndexRequest()
			//Expect(err).ToNot(HaveOccurred())
			//lines, err := bir.Source()
			//Expect(err).ToNot(HaveOccurred())
			//body := lines[1]
			//Expect(body).ToNot(BeEmpty())
			//var b interface{}
			//err = json.Unmarshal([]byte(body), &b)
			//Expect(err).ToNot(HaveOccurred())
			//m := b.(map[string]interface{})
			//Expect(m).To(HaveKeyWithValue("subject", "urn:1"))
			//Expect(m).To(HaveKeyWithValue("XSDRaw", "xsd:string"))
			//Expect(m).To(HaveKeyWithValue("language", "en"))
			//})
		})
	})

})
