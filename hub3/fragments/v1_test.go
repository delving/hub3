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
	"math/rand"
	"strings"

	c "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/fragments"

	"os"

	r "github.com/deiu/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getTestGraph() (*r.Graph, error) {

	turtle, err := os.Open("test_data/test2.ttl")
	if err != nil {
		return &r.Graph{}, err
	}
	g, err := fragments.NewGraphFromTurtle(turtle)
	return g, err
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var _ = Describe("V1", func() {

	c.InitConfig()

	Describe("Should be able to parse RDF", func() {

		Context("When given RDF as an io.Reader", func() {

			It("Should create a graph", func() {
				turtle, err := os.Open("test_data/test2.ttl")
				Expect(err).ToNot(HaveOccurred())
				Expect(turtle).ToNot(BeNil())
				g, err := fragments.NewGraphFromTurtle(turtle)
				Expect(err).ToNot(HaveOccurred())
				Expect(g).ToNot(BeNil())
				Expect(g.Len()).To(Equal(59))
			})

			It("Should throw an error when receiving invalid RDF", func() {
				badRDF := strings.NewReader("")
				g, err := fragments.NewGraphFromTurtle(badRDF)
				Expect(err).To(HaveOccurred())
				Expect(g.Len()).To(Equal(0))
			})
		})

	})

	Describe("indexDoc", func() {

		Context("when created from an RDF graph", func() {
			//g, err := getTestGraph()
			fb, err := testDataGraph()

			It("should have a valid graph", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fb.Graph).ToNot(BeNil())
				Expect(fb.Graph.Len()).To(Equal(59))
			})

			It("should return a map", func() {
				indexDoc, err := fragments.CreateV1IndexDoc(fb)
				Expect(err).ToNot(HaveOccurred())
				Expect(indexDoc).ToNot(BeEmpty())
				Expect(indexDoc).To(HaveKey("legacy"))
				Expect(indexDoc).To(HaveKey("system"))
				Expect(len(indexDoc)).To(Equal(41))
			})
		})
	})

	Context("when creating an IndexEntry from a blank node", func() {

		dcSubject := "http://purl.org/dc/elements/1.1/subject"
		t := r.NewTriple(
			r.NewResource("urn:1"),
			r.NewResource(dcSubject),
			r.NewBlankNode(0),
		)
		It("should identify an resource", func() {
			ie, err := fragments.CreateV1IndexEntry(t)
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
			r.NewResource("urn:rapid"),
		)
		It("should identify an resource", func() {
			ie, err := fragments.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("URIRef"))
			Expect(ie.ID).To(Equal("urn:rapid"))
			Expect(ie.Value).To(Equal("urn:rapid"))
			Expect(ie.Raw).To(Equal("urn:rapid"))
		})

	})

	Context("when creating an IndexEntry from a resource", func() {

		dcSubject := "http://purl.org/dc/elements/1.1/subject"

		t := r.NewTriple(
			r.NewResource("urn:1"),
			r.NewResource(dcSubject),
			r.NewLiteralWithLanguage("rapid", "nl"),
		)
		ie, err := fragments.CreateV1IndexEntry(t)

		It("should identify an Literal", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Type).To(Equal("Literal"))
			Expect(ie.ID).To(BeEmpty())
			Expect(ie.Value).To(Equal(ie.Raw))
		})

		It("should limit raw to 256 characters", func() {
			rString := randSeq(500)
			Expect(rString).To(HaveLen(500))
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource(dcSubject),
				r.NewLiteralWithLanguage(rString, "nl"),
			)
			ie, err := fragments.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie).ToNot(BeNil())
			Expect(ie.Raw).To(HaveLen(256))
			//
		})

		It("should limit value to 32000 characters", func() {
			rString := randSeq(40000)
			Expect(rString).To(HaveLen(40000))
			t := r.NewTriple(
				r.NewResource("urn:1"),
				r.NewResource(dcSubject),
				r.NewLiteralWithLanguage(rString, "nl"),
			)
			ie, err := fragments.CreateV1IndexEntry(t)
			Expect(err).ToNot(HaveOccurred())
			Expect(ie.Raw).To(HaveLen(256))
			Expect(ie.Value).To(HaveLen(32000))
		})

		It("should add lang when present", func() {
			Expect(ie.Language).To(Equal("nl"))
		})
	})

})
