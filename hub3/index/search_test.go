package index

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Search", func() {

	Context("When creating a RDFObject", func() {

		It("should have a subject", func() {
			o := RdfObject{}
			Expect(o).ToNot(BeNil())
		})
	})

})
