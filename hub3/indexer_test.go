package hub3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexer", func() {

	Describe("when initialised", func() {

		It("should have a search client", func() {
			Expect(client).ToNot(BeNil())
		})

		It("should have a bulk-indexer", func() {
			Expect(processor).ToNot(BeNil())
		})

	})
})
