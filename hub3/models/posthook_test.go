package models

import (
	"io/ioutil"
	"strings"

	"github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getRDFString(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

var subject = "http://data.brabantcloud.nl/resource/aggregation/enb-83-beeldmateriaal/enb-83.beeldmateriaal-620b3fa2-a2d8-796c-eae1-b8b9ca6947b7-14b9d8fd-a7f5-c901-2e2d-ae6d0966bd25"

var _ = Describe("Posthook", func() {

	Describe("when creating", func() {

		Context("from an RDF string", func() {

			It("should populate a graph", func() {
				content, err := getRDFString("test_data/enb_test_1.nt")
				Expect(err).ToNot(HaveOccurred())
				g := rdf2go.NewGraph(subject)
				err = g.Parse(strings.NewReader(content), "text/turtle")
				Expect(err).ToNot(HaveOccurred())
				posthook := NewPostHookJob(g, "enb-83-beeldmateriaal", false, subject)
				Expect(posthook).ToNot(BeNil())
				Expect(posthook.Graph.Len()).ToNot(Equal(0))
				jsonld, err := posthook.String()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonld).To(ContainSubstring("brabant"))
				Expect(jsonld).To(ContainSubstring("{\"@id\":"))
			})

		})

	})

})
