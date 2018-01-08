package hub3

import (
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var nt string = `<http://rapid.org/123> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.europeana.eu/schemas/edm/Place> .
<http://rapid.org/document/123> <http://www.europeana.eu/schemas/edm/type> "IMAGE" .`

var graphName string = "http://rapid.org/123/graph"

var _ = Describe("Rdf", func() {

	Describe("Converting to nquads", func() {

		Context("from ntriples", func() {

			It("Should replace end markers with graph uri", func() {
				Expect(len(strings.Split(nt, "\n"))).To(Equal(2))
				Expect(nt).ToNot(ContainSubstring("/graph>"))
				Expect(nt).To(HaveSuffix("."))
				graphUri, _ := url.Parse(graphName)
				nquads := Ntriples2Nquads(nt, graphUri)
				Expect(nquads).ToNot(BeNil())
				Expect(nquads).To(ContainSubstring("/graph>"))
			})
		})
	})

})
