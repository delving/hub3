package index

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Elasticsearch", func() {

	Describe("CreateClient", func() {

		Context("When initialised", func() {

			It("Should return an elastic client", func() {
				client := ESClient()
				Expect(client).ToNot(BeNil())
				Expect(fmt.Sprintf("%T", client)).To(Equal("*elastic.Client"))
			})
		})
	})

	Describe("CustomRetrier", func() {

		Context("When initialised", func() {

			It("should return a Retrier", func() {
				Expect(NewCustomRetrier()).ToNot(BeNil())
			})
		})
	})
})
