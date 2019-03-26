package fragments

import (
	fmt "fmt"
	"reflect"
	"strings"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
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
				output, err := sr.SearchRequestToHex()
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())
				Expect(output).To(Equal("1810"))
			})

			It("should marshal from a string", func() {
				sr := &SearchRequest{
					Query:        "hub3 Rocks Gööd",
					ResponseSize: int32(20),
					FacetLimit:   int32(100),
				}
				output, err := sr.SearchRequestToHex()
				Expect(err).ToNot(HaveOccurred())
				input := "0a116875623320526f636b732047c3b6c3b66418145864"
				Expect(output).To(Equal(input))
				newSr, err := SearchRequestFromHex(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(newSr.GetResponseSize()).To(Equal(int32(20)))
				Expect(newSr.GetQuery()).To(Equal("hub3 Rocks Gööd"))
			})

		})

		Context("When parsing url parameters", func() {

			It("should set the query", func() {
				params := make(map[string][]string)
				params["q"] = []string{"hub3"}
				sr, err := NewSearchRequest(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())
				Expect(sr.GetQuery()).To(Equal("hub3"))

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

				srFromID, err := SearchRequestFromHex(id.GetScrollID())
				Expect(err).ToNot(HaveOccurred())
				Expect(srFromID.GetStart()).To(Equal(int32(16)))
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

func Test_validateTypeClass(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty TypeClass", args{"a"}, ""},
		{"valid TypeClass", args{"edm_Place"}, "edm_Place"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTypeClass(tt.args.s); got != tt.want {
				defer GinkgoRecover()
				Fail(fmt.Sprintf("validateTypeClass() = %v, want %v", got, tt.want))
			}
		})
	}
}

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
		sr.SearchRequestToHex()
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
		{
			"context level 1 with class empty filter", "dcterms_spatial[]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial"}},
		},
		{
			"context level 1 with context class filter", "[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"}},
		},
		{
			"context level 1 with context class filter", "[]dcterms_spatial[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: ""}},
		},
		{
			"context level 2 with context class filter",
			"[ore_Aggregation]edm_aggregateCHO[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum",
			nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"},
				Level1: &ContextQueryFilter{SearchLabel: "edm_aggregateCHO", TypeClass: "ore_Aggregation"},
			},
		},
		{
			"context level 2 with empty context class filter",
			"[]edm_aggregateCHO[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum",
			nil,
			&QueryFilter{SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"},
				Level1: &ContextQueryFilter{SearchLabel: "edm_aggregateCHO", TypeClass: ""},
			},
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {

			new, err := NewQueryFilter(tc.input)
			if tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("%s should not throw error %v: got %v", tc.name, tc.err, err)
			}
			if !cmp.Equal(new, tc.output) && tc.err == nil {
				Fail(fmt.Sprintf("%s should be converted to %v; got %v", tc.input, tc.output, new))
			}
			normalisedInput := tc.input
			if !strings.HasPrefix(normalisedInput, "[") {
				normalisedInput = "[]" + normalisedInput
			}
			if normalisedInput != new.AsString() && tc.err == nil {
				Fail(fmt.Sprintf("%s is not converted back correctly ; got %v", normalisedInput, new.AsString()))
			}
		})
	}

}

func Test_TypeClassAsURI(t *testing.T) {
	tests := []struct {
		name    string
		given   string
		want    string
		wantErr bool
	}{
		{"correct namespace", "edm_Place", "http://www.europeana.eu/schemas/edm/Place", false},
		{"bad shorthand", "navePlace", "", true},
		{"unknown prefix", "example_Place", "", true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()

			got, err := TypeClassAsURI(tt.given)
			if (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("TypeClassAsURI() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if got != tt.want {
				Fail(fmt.Sprintf("TypeClassAsURI() = %v, want %v", got, tt.want))
			}
		})
	}
}

func TestSearchRequest_NewUserQuery(t *testing.T) {
	type fields struct {
		Query       string
		QueryFilter []*QueryFilter
	}
	tests := []struct {
		name     string
		fields   fields
		want     *Query
		crumbLen int
		wantErr  bool
	}{
		{"match all query", fields{}, &Query{}, 0, false},
		{"simple query", fields{Query: "test"}, &Query{Terms: "test",
			BreadCrumbs: []*BreadCrumb{&BreadCrumb{Href: "q=test", Display: "test", Value: "test", IsLast: true}}},
			1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()

			sr := &SearchRequest{
				Query:       tt.fields.Query,
				QueryFilter: tt.fields.QueryFilter,
			}
			got, bcb, err := sr.NewUserQuery()
			if (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("SearchRequest.NewUserQuery() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				Fail(fmt.Sprintf("SearchRequest.NewUserQuery() = %v, want %v", got, tt.want))
			}
			if len(bcb.crumbs) != tt.crumbLen {
				Fail(fmt.Sprintf("SearchRequest.NewUserQuery() length = %v, want %v", len(bcb.crumbs), tt.crumbLen))
			}
		})
	}
}

func TestAppendBreadCrumb(t *testing.T) {
	type args struct {
		param string
		qf    *QueryFilter
		path  string
	}
	bcb := &BreadCrumbBuilder{}

	tests := []struct {
		name     string
		args     args
		wantLast *BreadCrumb
		wantErr  bool
	}{
		{"empty query", args{param: "query", qf: &QueryFilter{Value: ""}}, &BreadCrumb{IsLast: true}, false},
		{"simple query", args{param: "query", qf: &QueryFilter{Value: "test"}, path: "q=test"},
			&BreadCrumb{Href: "q=test", Display: "test", Value: "test", IsLast: true}, false},
		{"simple filter query", args{param: "qf[]", qf: &QueryFilter{Value: "boerderij", SearchLabel: "dc_subject"},
			path: "q=test&qf[]=dc_subject:boerderij"},
			&BreadCrumb{Href: "q=test&qf[]=dc_subject:boerderij", Display: "dc_subject:boerderij", Value: "boerderij",
				Field: "dc_subject", IsLast: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()
			bcb.AppendBreadCrumb(tt.args.param, tt.args.qf)
			got := bcb.GetLast()
			if !reflect.DeepEqual(got, tt.wantLast) {
				Fail(fmt.Sprintf("NewBreadCrumb() = %v, want %v", got, tt.wantLast))
			}
			if bcb.GetPath() != tt.args.path {
				Fail(fmt.Sprintf("NewBreadCrumb() Path = %v, want %v", bcb.GetPath(), tt.args.path))
			}
		})
	}
}

func TestFacetURIBuilder_AddFilter(t *testing.T) {
	type fields struct {
		query string
	}
	type args struct {
		qf *QueryFilter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"empty start",
			fields{""},
			args{&QueryFilter{SearchLabel: "dc_subject", Value: "boerderij"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()

			fub := &FacetURIBuilder{
				query:   tt.fields.query,
				filters: make(map[string]map[string]*QueryFilter),
			}
			if err := fub.AddFilter(tt.args.qf); (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("FacetURIBuilder.AddFilter() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if !assert.True(t, fub.hasQueryFilter(tt.args.qf.SearchLabel, tt.args.qf.Value)) {
				Fail("Key not added.")
			}
		})
	}
}
