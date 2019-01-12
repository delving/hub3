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
	fmt "fmt"
	"io/ioutil"
	"net/url"

	c "github.com/delving/rapid-saas/config"
	. "github.com/delving/rapid-saas/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// testGraph creates a dummy graph for testing
func testGraph() *r.Graph {
	baseUri := "http://rapid.org/resource"

	g := r.NewGraph(baseUri)
	g.Add(r.NewTriple(r.NewResource("a"), r.NewResource("b"), r.NewResource("c")))
	//r.NewTriple(r.NewResource("a"), r.NewResource("title"), r.NewLiteral("title")),
	//r.NewTriple(r.NewResource("a"), r.NewResource("subject"), r.NewLiteralWithLanguage("subject", "nl")),
	return g
}

// URIRef is a function to create an RDFLiteal resource
func URIRef(uri string) r.Term {
	return r.NewResource(uri)
}

// NSRef creates a URIRef with the RDF Base URL
func NSRef(uri string) r.Term {
	return r.NewResource(fmt.Sprintf("%s/%s", c.Config.RDF.BaseURL, uri))
}

// Literal is a utility function to create a RDF literal
//func Literal(value string, language string, dataType ObjectXSDType) r.Term {
//if language != "" {
//return r.NewLiteralWithLanguage(value, language)
//}
//if dataType != ObjectXSDType_STRING {
//t, err := dataType.GetLabel()
//if err != nil {
//log.Println("Unable to get label for this type")
//}
//return r.NewLiteralWithDatatype(value, r.NewResource(t))
//}
//return r.NewLiteral(value)
//}

func testFragmentGraph(spec string, rev int32, ng string) *FragmentGraph {
	fg := NewFragmentGraph()
	fg.Meta.OrgID = "rapid"
	fg.Meta.Spec = spec
	fg.Meta.Revision = rev
	fg.Meta.NamedGraphURI = ng
	fg.Meta.HubID = fmt.Sprintf("%s_%s_1", fg.Meta.OrgID, fg.Meta.Spec)
	return fg
}

func testDataGraph(empty bool) (*FragmentBuilder, error) {
	spec := "test-spec"
	rev := int32(1)
	ng := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893/graph"
	fg := testFragmentGraph(spec, rev, ng)
	fg.Meta.EntryURI = "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"
	fb := NewFragmentBuilder(fg)
	dat, err := ioutil.ReadFile("test_data/enb_test_2.jsonld")
	if err != nil {
		return fb, err
	}
	if !empty {
		fb.ParseGraph(bytes.NewReader(dat), "application/ld+json")
	}
	return fb, nil
}

var _ = Describe("Fragments", func() {

	Describe("creating a new FragmentRequest", func() {

		Context("directly", func() {

			It("should have no triple pattern set", func() {
				fr := NewFragmentRequest()
				Expect(fr).ToNot(BeNil())
				Expect(fr.GetSubject()).To(BeEmpty())
				Expect(fr.GetPredicate()).To(BeEmpty())
				Expect(fr.GetObject()).To(BeEmpty())
				Expect(fr.GetLanguage()).To(BeEmpty())
			})

			It("should have a non-zero page start", func() {
				fr := NewFragmentRequest()
				Expect(fr.GetPage()).To(Equal(int32(1)))
			})
		})

		Context("parsing from url.Values", func() {

			It("should ignore empty values", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("subject", "urn:1")
				v.Add("predicate", "")
				v.Add("object", "")
				v.Add("language", "")
				v.Add("page", "2")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetSubject()).To(Equal([]string{"urn:1"}))
				Expect(fr.GetPredicate()).To(BeEmpty())
				Expect(fr.GetObject()).To(BeEmpty())
				Expect(fr.GetLanguage()).To(BeEmpty())
			})

			It("should throw an error when the page is not an int", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("page", "error")
				err := fr.ParseQueryString(v)
				Expect(err).To(HaveOccurred())
			})

			It("should set the page when it is an int", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("page", "10")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetPage()).To(Equal(int32(10)))
			})

			It("should set all the non-empty values", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("subject", "urn:1")
				v.Add("predicate", "urn:subject")
				v.Add("object", "mountain")
				v.Add("language", "nl")
				v.Add("page", "3")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetSubject()).To(Equal([]string{"urn:1"}))
				Expect(fr.GetPredicate()).To(Equal("urn:subject"))
				Expect(fr.GetObject()).To(Equal("mountain"))
				Expect(fr.GetLanguage()).To(Equal("nl"))
				Expect(fr.GetPage()).To(Equal(int32(3)))
			})
		})
	})

	//Describe("ObjectXSDType conversions", func() {

	//Context("when converting to label", func() {

	//It("should return the xsd label when found", func() {
	//label, err := ObjectXSDType_BOOLEAN.GetLabel()
	//Expect(err).ToNot(HaveOccurred())
	//Expect(label).ToNot(BeEmpty())
	//Expect(label).To(Equal("http://www.w3.org/2001/XMLSchema#boolean"))
	//})

	//It("should return an error when no label could be found", func() {
	//const ObjectXSDType_ERROR ObjectXSDType = 100
	//label, err := ObjectXSDType_ERROR.GetLabel()
	//Expect(err).To(HaveOccurred())
	//Expect(label).To(BeEmpty())
	//})
	//})

	//Context("when requesting a prefix label", func() {

	//It("should shorten the namespace to xsd", func() {
	//label, err := ObjectXSDType_BOOLEAN.GetPrefixLabel()
	//Expect(err).ToNot(HaveOccurred())
	//Expect(label).ToNot(BeEmpty())
	//Expect(label).To(Equal("xsd:boolean"))

	//})
	//})

	//Context("when converting from a label", func() {

	//It("should return the ObjectXSDType", func() {
	//t, err := GetObjectXSDType("http://www.w3.org/2001/XMLSchema#boolean")
	//Expect(err).ToNot(HaveOccurred())
	//Expect(t).ToNot(BeNil())
	//Expect(t).To(Equal(ObjectXSDType_BOOLEAN))
	//})
	//})

	//})

	Describe("hasher", func() {

		Context("When given a string", func() {

			It("should return a short hash", func() {
				hash := CreateHash("rapid rocks.")
				Expect(hash).To(Equal("a5b3be36c0f378a1"))
			})
		})

	})

	Describe("FragmentRequest", func() {

		Context("when assiging an object", func() {

			It("should strip double quotes", func() {
				fr := NewFragmentRequest()
				fr.Object = `1982`
				fr.AssignObject()
				Expect(fr.GetObject()).To(Equal("1982"))
			})

			It("should set the language when the string contains @ annotation", func() {
				fr := NewFragmentRequest()
				fr.Object = `"door"@en`
				fr.AssignObject()
				Expect(fr.GetObject()).To(Equal("door"))
				Expect(fr.GetLanguage()).To(Equal("en"))
			})

		})
	})

})
