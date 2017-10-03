package models

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

	Context("when a new store is initialised with a dbName", func() {
		It("should use that dbName", func() {
			Expect("test.db").To(BeAnExistingFile())
		})
	})

})
