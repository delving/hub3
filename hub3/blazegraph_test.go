package hub3

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bitbucket.org/delving/rapid/config"
)

var _ = Describe("Blazegraph", func() {

	InitConfig()
	testNS := "testns"

	Describe("Initialisation", func() {

		Context("When there is no Blazegraph namespace present", func() {

			It("should have set the BlazeGraph Base namespace uri", func() {
				uri := namespaceBaseURI()
				Expect(uri).To(MatchRegexp("/blazegraph/namespace$"))
				Expect(uri).To(ContainSubstring(Config.RDF.SparqlHost))
			})

			It("should have set the Blazegraph end-point URI", func() {
				endpoint := blazegraphEndpoint("")
				Expect(endpoint).To(MatchRegexp(fmt.Sprintf("/%s$", Config.OrgID)))
			})

			It("should format the blazegraph properties with the orgId", func() {
				properties := createNameSpaceProperties("")
				Expect(properties).To(ContainSubstring(Config.OrgID))
				Expect(properties).To(ContainSubstring("com.bigdata.rdf.sail.namespace"))
			})

			It("should return deleted = false even if the namespace does not exist", func() {
				deleted, errs := deleteNameSpace(testNS)
				Expect(errs).To(BeEmpty())
				Expect(deleted).To(BeFalse())
				exists, errs := namespaceExist(testNS)
				Expect(exists).To(BeFalse())
			})

			It("should should not have our namespaces", func() {
				exists, errs := namespaceExist(testNS)
				Expect(exists).To(BeFalse())
				Expect(errs).To(BeEmpty())
			})

			It("should create a new namespace in blazegraph", func() {
				created, errs := createNameSpace(testNS)
				Expect(errs).To(BeEmpty())
				Expect(created).To(BeTrue())
				exists, errs := namespaceExist(testNS)
				Expect(exists).To(BeTrue())
			})

			It("should delete the namespace in blazegraph", func() {
				deleted, errs := deleteNameSpace(testNS)
				Expect(errs).To(BeEmpty())
				Expect(deleted).To(BeTrue())
				exists, errs := namespaceExist(testNS)
				Expect(exists).To(BeFalse())
			})

		})
	})

})
