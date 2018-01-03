package hub3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bulkapi", func() {

	Describe("When reading a io.Reader", func() {

		Context("and all lines are valid", func() {

			It("should parse all lines", func() {
				Expect("").To(BeEmpty())
			})

		})
	})

})
