package api_test

import (
	fmt "fmt"
	"testing"

	c "github.com/delving/rapid/config"
	. "github.com/delving/rapid/hub3/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apiutils", func() {

	Describe("SearchRequest", func() {
		c.InitConfig()
		Context("When create a new SearchRequest", func() {

			sr := DefaultSearchRequest(&c.Config)

			It("should not be empty", func() {
				Expect(sr).ToNot(BeNil())
			})

			It("should start at 0", func() {
				Expect(sr.GetStart()).To(Equal(int32(0)))
			})

		})

		Context("When Serializing a SearchRequest", func() {

			sr := DefaultSearchRequest(&c.Config)

			It("should marshal to a string", func() {
				output, err := SearchRequestToHex(sr)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())
				Expect(output).To(Equal("1810"))
			})

			It("should marshal from a string", func() {
				sr := &SearchRequest{
					Query:        "Rapid Rocks Gööd",
					ResponseSize: int32(20),
					FacetLimit:   int32(100),
				}
				output, err := SearchRequestToHex(sr)
				fmt.Println(output)
				Expect(err).ToNot(HaveOccurred())
				input := "0a12526170696420526f636b732047c3b6c3b66418145864"
				Expect(output).To(Equal(input))
				newSr, err := SearchRequestFromHex(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(newSr.GetResponseSize()).To(Equal(int32(20)))
				Expect(newSr.GetQuery()).To(Equal("Rapid Rocks Gööd"))
			})

		})

	})

})

func BenchmarkSearchRequestToHex(b *testing.B) {
	sr := &SearchRequest{Query: "TestQuery"}
	for n := 0; n < b.N; n++ {
		SearchRequestToHex(sr)
	}
}

func BenchmarkSearchRequestFromHex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		SearchRequestFromHex("1810")
	}
}
