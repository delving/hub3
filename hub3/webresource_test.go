package hub3_test

import (
	"bitbucket.org/delving/rapid/hub3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webresource", func() {

	Describe("hasher", func() {

		Context("When given a string", func() {

			It("should return a short hash", func() {
				hash := hub3.CreateHash("rapid rocks.")
				Expect(hash).To(Equal("a5b3be36c0f378a1"))
			})
		})

	})
})
