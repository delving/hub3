package hub3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {

	Describe("When initialised", func() {

		Context("the BoltDB-backed database", func() {

			It("should be available", func() {
				Expect(orm).ToNot(BeNil())
			})
		})

	})

})
