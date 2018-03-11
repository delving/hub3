package harvesting_test

import (
	//. "github.com/delving/rapid/hub3/harvesting"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sparql", func() {

	Describe("When setting up the initial request", func() {

		Context("and retrieving the original count", func() {

			It("should parse the sparql response", func() {
				Expect("test").ToNot(BeEmpty())
			})

			It("should extract the count", func() {

			})

			It("should fill the channel with sparql harvest pages", func() {
			})
		})

	})

	Describe("A sparql harvest worker", func() {

		Context("when receiving a sparql harvest page", func() {

			It("should build a request", func() {
			})

			It("should execute a request", func() {

			})
		})

		Context("when receving a sparql response", func() {

			It("should store the request in a file", func() {
			})

			It("should set the file name to offset_limit.json", func() {
			})

			It("should be able to create an rdf2go Graph", func() {
			})
		})
	})
})
