// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	c "github.com/delving/hub3/config"
	. "github.com/delving/hub3/hub3/fragments"

	"os"

	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getTestGraph() (*r.Graph, error) {

	turtle, err := os.Open("testdata/test2.ttl")
	if err != nil {
		return &r.Graph{}, err
	}
	g, err := NewGraphFromTurtle(turtle)
	return g, err
}

func renderJSONLD(g *r.Graph) (string, error) {
	var b bytes.Buffer
	err := g.SerializeFlatJSONLD(&b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

var _ = Describe("V1", func() {

	c.InitConfig()

	Describe("Should be able to parse RDF", func() {

		Context("When given RDF as an io.Reader", func() {

			It("Should create a graph", func() {
				turtle, err := os.Open("testdata/test2.ttl")
				Expect(err).ToNot(HaveOccurred())
				Expect(turtle).ToNot(BeNil())
				g, err := NewGraphFromTurtle(turtle)
				Expect(err).ToNot(HaveOccurred())
				Expect(g).ToNot(BeNil())
				Expect(g.Len()).To(Equal(67))
			})

			It("Should throw an error when receiving invalid RDF", func() {
				badRDF := strings.NewReader("")
				g, err := NewGraphFromTurtle(badRDF)
				Expect(err).To(HaveOccurred())
				Expect(g.Len()).To(Equal(0))
			})
		})

	})

	Describe("indexDoc", func() {

		Context("when created from an RDF graph", func() {
			//g, err := getTestGraph()

			It("should have a valid graph", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(fb.Graph).ToNot(BeNil())
				Expect(fb.Graph.Len()).To(Equal(65))
			})

			It("should return a map", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				_ = fb.GetSortedWebResources()
				indexDoc, err := CreateV1IndexDoc(fb)
				Expect(err).ToNot(HaveOccurred())
				Expect(indexDoc).ToNot(BeEmpty())
				Expect(indexDoc).To(HaveKey("legacy"))
				Expect(indexDoc).To(HaveKey("system"))
				Expect(len(indexDoc)).To(Equal(46))
			})

			It("should return the MediaManagerUrl for a WebResource", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				urn := "urn:spec/localID"
				url := fb.MediaManagerURL(urn, "hub3")
				Expect(url).ToNot(BeEmpty())
				Expect(url).To(HaveSuffix("localID"))
				Expect(url).To(ContainSubstring("hub3"))
			})

			It("should return a list of WebResource subjects", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(fb.Graph.Len()).ToNot(Equal(0))
				wr := fb.GetSortedWebResources()
				Expect(wr).ToNot(BeNil())
				Expect(wr).To(HaveLen(3))
				var order []int
				for _, v := range wr {
					order = append(order, v.Value)
				}
				Expect(order).To(Equal([]int{1, 2, 3}))
			})

			//It("should clean-up urn: references that end with __", func() {
			//Skip("slow test")
			//urn := "urn:museum-helmond-objecten/2008-018__"
			//orgID := "brabantcloud"
			//wrb, err := testDataGraph(true)
			//wr := r.NewTriple(
			//r.NewResource(urn),
			//r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
			//GetEDMField("WebResource"),
			//)
			//wrb.Graph.Add(wr)
			//Expect(err).ToNot(HaveOccurred())
			//Expect(wrb.Graph.Len()).To(Equal(1))
			//triple := wrb.Graph.One(r.NewResource(urn), nil, nil)
			//Expect(triple).ToNot(BeNil())
			//errChan := make(chan error)
			//wrb.GetRemoteWebResource(urn, orgID, errChan)
			//Expect(errChan).To(BeEmpty())
			//Expect(wrb.Graph.Len()).ToNot(Equal(0))
			//wrList := wrb.GetSortedWebResources()
			//Expect(wrList).ToNot(BeEmpty())
			//triples := wrb.Graph.All(nil, GetEDMField("hasView"), nil)
			//Expect(triples).ToNot(BeNil())
			//object := wrb.Graph.All(nil, GetEDMField("object"), nil)
			//Expect(object).To(HaveLen(1))
			//isShownBy := wrb.Graph.All(nil, GetEDMField("isShownBy"), nil)
			//Expect(isShownBy).To(HaveLen(1))
			//})

			It("should return a list of webresources with urns", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				urns := fb.GetUrns()
				Expect(urns).To(HaveLen(3))
			})

			It("should have the ore:aggregation subject as subject for edm:hasView", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				wr := fb.GetSortedWebResources()
				Expect(wr).ToNot(BeEmpty())
				triples := fb.SortedGraph.ByPredicate(GetEDMField("hasView"))
				Expect(triples).ToNot(BeNil())
				Expect(triples).To(HaveLen(3))
				//fmt.Printf("%#v\n", triples[0].String())
				triples = fb.SortedGraph.ByPredicate(GetEDMField("isShownBy"))
				Expect(triples).ToNot(BeNil())
				Expect(triples).To(HaveLen(1))
				triple := triples[0]
				fmt.Println(triple)
				Expect(triple.Subject.(*r.Resource).RawValue()).To(HaveSuffix("F900893"))
			})

			It("should remove derivatives from BlankNodes when WebResources are urns", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				graphLength := fb.Graph.Len()
				Expect(graphLength).To(Equal(65))

			})

			It("should rerender blanknodes in cleaned up graph", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				graphLength := fb.Graph.Len()
				Expect(graphLength).To(Equal(65))
				//json, err := renderJSONLD(fb.Graph)
				//Expect(err).ToNot(HaveOccurred())
				//fmt.Println(json)

				wr := fb.GetSortedWebResources()
				Expect(wr).ToNot(BeEmpty())
				//Expect(fb.Graph.Len()).To(Equal(70))

				// have brabantcloud resource
				bType := r.NewResource("http://schemas.delving.eu/nave/terms/BrabantCloudResource")
				tRaw := fb.Graph.One(nil, nil, bType)
				Expect(tRaw).ToNot(BeNil())
				//fmt.Printf("raw resource: %#v", tRaw.String())

				//json, err = renderJSONLD(fb.Graph)
				//Expect(err).ToNot(HaveOccurred())
				//fmt.Println(json)

			})

			It("should produce valid json-ld", func() {
				g := r.NewGraph("")
				dat, err := ioutil.ReadFile("testdata/test_nave_normalised.jsonld")
				Expect(err).ToNot(HaveOccurred())
				err = g.Parse(bytes.NewReader(dat), "application/ld+json")
				Expect(err).ToNot(HaveOccurred())
				json, err := renderJSONLD(g)
				Expect(err).ToNot(HaveOccurred())
				Expect(json).ToNot(BeEmpty())
				//fmt.Printf("jsonld_1: %s\n", json)

				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				fb.Graph = g
				fb.GetSortedWebResources()
				json, err = renderJSONLD(fb.Graph)
				Expect(err).ToNot(HaveOccurred())
				Expect(json).ToNot(BeEmpty())
				//fmt.Printf("jsonld_2: %s\n", json)
				// todo add diff between two versions of the json

			})

			//It("should cleanup the dates", func() {
			//Skip("must be added to a different part of the code base")
			//fb, err := testDataGraph(false)
			//Expect(err).ToNot(HaveOccurred())
			//created := r.NewResource(GetNSField("dcterms", "created"))
			//t := fb.Graph.One(nil, created, nil)
			//Expect(t).ToNot(BeNil())
			//fb.GetSortedWebResources()
			//t = fb.Graph.One(nil, created, nil)
			//Expect(t).To(BeNil())
			//createdRaw := r.NewResource(GetNSField("dcterms", "createdRaw1"))
			//tRaw := fb.Graph.One(nil, createdRaw, nil)
			//Expect(tRaw).ToNot(BeNil())
			//})

		})

	})

	Context("when creating an IndexEntry from a blank node", func() {

		dcSubject := "http://purl.org/dc/elements/1.1/subject"
		t := r.NewTriple(
			r.NewResource("urn:1"),
			r.NewResource(dcSubject),
			r.NewBlankNode("0"),
		)
		fb, err := testDataGraph(false)

		It("should identify an resource", func() {
			Expect(err).ToNot(HaveOccurred())
			ie, err := fb.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("Bnode"))
			Expect(ie.ID).To(Equal("0"))
			Expect(ie.Value).To(Equal("0"))
			Expect(ie.Raw).To(Equal("0"))
		})

	})

	Context("when creating an IndexEntry from a resource", func() {

		dcSubject := "http://purl.org/dc/elements/1.1/subject"
		t := r.NewTriple(
			r.NewResource("urn:1"),
			r.NewResource(dcSubject),
			r.NewResource("urn:hub3"),
		)
		fb, _ := testDataGraph(false)
		err := fb.SetResourceLabels()

		It("should identify an resource", func() {
			Expect(err).ToNot(HaveOccurred())
			ie, err := fb.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("URIRef"))
			Expect(ie.ID).To(Equal("urn:hub3"))
			Expect(ie.Value).To(Equal("urn:hub3"))
			Expect(ie.Raw).To(Equal("urn:hub3"))
		})

		It("should add label when a resource object has a skos:prefLabel", func() {

			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource(dcSubject),
				r.NewResource("http://data.jck.nl/resource/skos/thesau/90073896"),
			)
			ie, err := fb.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("URIRef"))
			Expect(ie.ID).To(Equal("http://data.jck.nl/resource/skos/thesau/90073896"))
			Expect(ie.Value).To(Equal("begraafplaats"))
			Expect(ie.Raw).To(Equal("begraafplaats"))

		})

	})

	Context("when creating a context inline map", func() {

		fb, err := testDataGraph(false)

		It("should extract all prefLabels", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(fb.Graph).ToNot(BeNil())
			//Expect(fb.ResourceLabels).To(BeEmpty())
			err := fb.SetResourceLabels()
			Expect(err).ToNot(HaveOccurred())
			Expect(fb.ResourceLabels).ToNot(BeEmpty())
			Expect(fb.ResourceLabels).To(HaveLen(5))
		})

		It("should get a label if exists", func() {
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource("urn:subject"),
				r.NewResource("http://data.jck.nl/resource/skos/thesau/90073896"),
			)
			label, ok := fb.GetResourceLabel(t)
			Expect(label).ToNot(BeEmpty())
			Expect(ok).To(BeTrue())
			Expect(label).To(Equal("begraafplaats"))
		})

		It("should return not ok when no label is present", func() {
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource("urn:subject"),
				r.NewResource("http://data.jck.nl/resource/skos/thesau/none"),
			)
			label, ok := fb.GetResourceLabel(t)
			Expect(label).To(BeEmpty())
			Expect(ok).To(BeFalse())
		})

	})

	Context("when creating an IndexEntry from a literal", func() {

		dcSubject := "http://purl.org/dc/elements/1.1/subject"

		t := r.NewTriple(
			r.NewResource("urn:1"),
			r.NewResource(dcSubject),
			r.NewLiteralWithLanguage("hub3", "nl"),
		)
		fb, _ := testDataGraph(false)
		ie, err := fb.CreateV1IndexEntry(t)

		It("should identify an Literal", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("Literal"))
			Expect(ie.ID).To(BeEmpty())
			Expect(ie.Value).To(Equal(ie.Raw))
		})

		It("should limit raw to 256 characters", func() {
			rString := RandSeq(500)
			Expect(rString).To(HaveLen(500))
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource(dcSubject),
				r.NewLiteralWithLanguage(rString, "nl"),
			)
			ie, err := fb.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Raw).To(HaveLen(256))
			//
		})

		It("should limit value to 32000 characters", func() {
			rString := RandSeq(40000)
			Expect(rString).To(HaveLen(40000))
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource(dcSubject),
				r.NewLiteralWithLanguage(rString, "nl"),
			)
			ie, err := fb.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie.Raw).To(HaveLen(256))
			Expect(ie.Value).To(HaveLen(32000))
		})

		It("should add lang when present", func() {
			Expect(ie.Language).To(Equal("nl"))
		})
	})

})
