package fragments_test

import (
	"testing"

	c "github.com/delving/rapid-saas/config"
	. "github.com/delving/rapid-saas/hub3/fragments"

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
				Expect(err).ToNot(HaveOccurred())
				input := "0a12526170696420526f636b732047c3b6c3b66418145864"
				Expect(output).To(Equal(input))
				newSr, err := SearchRequestFromHex(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(newSr.GetResponseSize()).To(Equal(int32(20)))
				Expect(newSr.GetQuery()).To(Equal("Rapid Rocks Gööd"))
			})

		})

		Context("When parsing url parameters", func() {

			It("should set the query", func() {
				params := make(map[string][]string)
				params["q"] = []string{"rapid"}
				sr, err := NewSearchRequest(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())
				Expect(sr.GetQuery()).To(Equal("rapid"))

			})

			It("should set the rows param", func() {
				params := make(map[string][]string)
				params["rows"] = []string{"10"}
				sr, err := NewSearchRequest(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())
				Expect(sr.GetResponseSize()).To(Equal(int32(10)))
			})

			It("should prioritize scroll_id above other parameters", func() {

			})
		})

		Context("When creating a scrollID", func() {

			It("should set defaults to zero", func() {
				sp := NewScrollPager()
				Expect(sp.Cursor).To(Equal(int32(0)))
				Expect(sp.Total).To(Equal(int64(0)))
				Expect(sp.ScrollID).To(BeEmpty())
			})

			It("should create a scroll pager", func() {
				params := make(map[string][]string)
				sr, err := NewSearchRequest(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())

				id, err := sr.NextScrollID(200)
				Expect(err).ToNot(HaveOccurred())
				Expect(id.GetCursor()).To(Equal(int32(0)))
				Expect(id.GetRows()).To(Equal(int32(16)))
				Expect(id.GetScrollID()).ToNot(BeEmpty())

				srFromId, err := SearchRequestFromHex(id.GetScrollID())
				Expect(err).ToNot(HaveOccurred())
				Expect(srFromId.GetStart()).To(Equal(int32(16)))
			})

			It("should have an empty scroldlID when on the last page", func() {
				params := make(map[string][]string)
				sr, err := NewSearchRequest(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())

				id, err := sr.NextScrollID(12)
				Expect(err).ToNot(HaveOccurred())
				Expect(id.GetScrollID()).To(BeEmpty())

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
