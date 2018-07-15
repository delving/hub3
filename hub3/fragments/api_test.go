package fragments

import (
	fmt "fmt"
	"testing"

	c "github.com/delving/rapid-saas/config"
	"github.com/google/go-cmp/cmp"
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

		Context("When echoing a protobuf entry", func() {

			params := make(map[string][]string)
			params["rows"] = []string{"10"}
			params["query"] = []string{"1930"}
			sr, err := NewSearchRequest(params)

			It("should show the Elastic Query", func() {
				Expect(err).ToNot(HaveOccurred())
				query, err := sr.ElasticQuery()
				Expect(err).ToNot(HaveOccurred())
				Expect(query).ToNot(BeNil())

				echo, err := sr.Echo("es", 20)
				Expect(err).ToNot(HaveOccurred())
				Expect(echo).ToNot(BeNil())
				Expect(echo).To(HaveKey("bool"))

			})

			It("should return an error on unknown echoType", func() {
				echo, err := sr.Echo("unknown", 20)
				Expect(err).To(HaveOccurred())
				Expect(echo).To(BeNil())

			})

			//It("should show the scrollID", func() {
			//echo, err := sr.Echo("nextScrollID", int64(30))
			//Expect(err).ToNot(HaveOccurred())
			//Expect(echo).ToNot(BeNil())
			//Expect(echo.(*SearchRequest).GetQuery()).To(ContainSubstring("1930"))
			//Expect(echo.(*SearchRequest).GetResponseSize()).To(Equal(int32(10)))
			//})

			It("should show the Search Request", func() {
				echo, err := sr.Echo("searchRequest", 20)
				Expect(err).ToNot(HaveOccurred())
				Expect(echo).ToNot(BeNil())
				Expect(echo.(*SearchRequest).GetQuery()).To(ContainSubstring("1930"))

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

func Test_qfSplit(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"start bracket", args{'['}, true},
		{"end bracket", args{']'}, true},
		{"no bracket", args{'a'}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qfSplit(tt.args.r); got != tt.want {
				defer GinkgoRecover()
				Fail(fmt.Sprintf("qfSplit() = %v, want %v", got, tt.want))
			}
		})
	}
}

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

func TestNewQueryFilter(t *testing.T) {

	tt := []struct {
		name   string
		input  string
		err    error
		output *QueryFilter
	}{
		{
			"simple filter", "dc_subject:boerderij", nil,
			&QueryFilter{SearchLabel: "dc_subject", Value: "boerderij"},
		},
		{
			"bad filter", "dc_subjectboerderij", fmt.Errorf("no query field specified in: dc_subjectboerderij"),
			&QueryFilter{SearchLabel: "dc_subject", Value: "boerderij"},
		},
		{
			"class filter", "[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place"},
		},
		{
			"empty class filter", "[]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: ""},
		},
		{
			"context level 1 with class filter", "dcterms_spatial[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial"}},
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			new, err := NewQueryFilter(tc.input)
			if tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("%s should not throw error %v: got %v", tc.name, tc.err, err)
			}
			if !cmp.Equal(new, tc.output) && tc.err == nil {
				defer GinkgoRecover()
				Fail(fmt.Sprintf("%s should be converted to %v; got %v", tc.input, tc.output, new))
			}
		})
	}

}
