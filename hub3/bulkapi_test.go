package hub3

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bulkapi", func() {

	Describe("When reading a io.Reader", func() {

		Context("and lines are counted", func() {

			It("it should count all lines", func() {
				fourLines := "1\n2\n3\n4\n"
				Expect(lineCounter(strings.NewReader(fourLines))).To(Equal(4))
			})

			It("it should count last line with no linefeed", func() {
				fourLines := "1\n2\n3\n4"
				Expect(lineCounter(strings.NewReader(fourLines))).To(Equal(4))
			})

		})
	})

})
