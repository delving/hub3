// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

import (
	"encoding/json"
	fmt "fmt"
	"os"
	"reflect"
	"regexp"
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
				Expect(output).To(HavePrefix("1810"))
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
			orgID := "test"

			It("should set the query", func() {
				params := make(map[string][]string)
				params["q"] = []string{"hub3"}
				sr, err := NewSearchRequest(orgID, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())
				Expect(sr.GetQuery()).To(Equal("hub3"))
			})

			It("should set the rows param", func() {
				params := make(map[string][]string)
				params["rows"] = []string{"10"}
				sr, err := NewSearchRequest(orgID, params)

				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())
				Expect(sr.GetResponseSize()).To(Equal(int32(10)))
			})
		})

		Context("When echoing a protobuf entry", func() {
			orgID := "test"

			params := make(map[string][]string)
			params["rows"] = []string{"10"}
			params["query"] = []string{"1930"}
			sr, err := NewSearchRequest(orgID, params)

			It("should show the Elastic Query", func() {
				Expect(err).ToNot(HaveOccurred())
				query, err := sr.ElasticQuery()
				Expect(err).ToNot(HaveOccurred())
				Expect(query).ToNot(BeNil())
			})
		})

		Context("When creating a scrollID", func() {
			orgID := "test"

			It("should set defaults to zero", func() {
				sp := NewScrollPager()
				Expect(sp.Cursor).To(Equal(int32(0)))
				Expect(sp.Total).To(Equal(int64(0)))
				Expect(sp.NextScrollID).To(BeEmpty())
			})

			It("should create a scroll pager", func() {
				params := make(map[string][]string)
				sr, err := NewSearchRequest(orgID, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())

				id, err := sr.ScrollPagers(200)
				Expect(err).ToNot(HaveOccurred())
				Expect(id.Cursor).To(Equal(int32(0)))
				Expect(id.Rows).To(Equal(int32(16)))
				Expect(id.NextScrollID).ToNot(BeEmpty())
				Expect(id.PreviousScrollID).To(BeEmpty())

				srFromID, err := SearchRequestFromHex(id.NextScrollID)
				Expect(err).ToNot(HaveOccurred())
				Expect(srFromID.GetStart()).To(Equal(int32(16)))
			})

			It("should have an empty scrollID when on the last page", func() {
				params := make(map[string][]string)
				sr, err := NewSearchRequest(orgID, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(sr).ToNot(BeNil())

				id, err := sr.ScrollPagers(12)
				Expect(err).ToNot(HaveOccurred())
				Expect(id.NextScrollID).To(BeEmpty())
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
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial"},
			},
		},
		{
			"context level 1 with class empty filter", "dcterms_spatial[]nave_city:Berlicum", nil,
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial"},
			},
		},
		{
			"context level 1 with context class filter", "[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"},
			},
		},
		{
			"context level 1 with context class filter", "[]dcterms_spatial[edm_Place]nave_city:Berlicum", nil,
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: ""},
			},
		},
		{
			"context level 2 with context class filter",
			"[ore_Aggregation]edm_aggregateCHO[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum",
			nil,
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"},
				Level1: &ContextQueryFilter{SearchLabel: "edm_aggregateCHO", TypeClass: "ore_Aggregation"},
			},
		},
		{
			"context level 2 with empty context class filter",
			"[]edm_aggregateCHO[edm_ProvidedCHO]dcterms_spatial[edm_Place]nave_city:Berlicum",
			nil,
			&QueryFilter{
				SearchLabel: "nave_city", Value: "Berlicum", TypeClass: "edm_Place",
				Level2: &ContextQueryFilter{SearchLabel: "dcterms_spatial", TypeClass: "edm_ProvidedCHO"},
				Level1: &ContextQueryFilter{SearchLabel: "edm_aggregateCHO", TypeClass: ""},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer GinkgoRecover()
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
	t.Skip()

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
		{
			"simple query",
			fields{Query: "test"},
			&Query{
				Terms: "test",
				BreadCrumbs: []*BreadCrumb{
					{Href: "q=test", Display: "test", Value: "test", IsLast: true},
				},
			},
			1,
			false,
		},
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
		{
			"simple query",
			args{param: "query", qf: &QueryFilter{Value: "test"}, path: "q=test"},
			&BreadCrumb{Href: "q=test", Display: "test", Value: "test", IsLast: true}, false,
		},
		{
			"simple filter query",
			args{
				param: "qf[]", qf: &QueryFilter{Value: "boerderij", SearchLabel: "dc_subject"},
				path: "q=test&qf[]=dc_subject:boerderij",
			},
			&BreadCrumb{
				Href: "q=test&qf[]=dc_subject:boerderij", Display: "dc_subject:boerderij", Value: "boerderij",
				Field: "dc_subject", IsLast: true,
			}, false,
		},
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

func Test_isAdvancedSearch(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"simple query", args{query: "one word"}, false},
		{"AND query", args{query: "this AND that"}, true},
		{"lower case and", args{query: "this and that"}, false},
		{"OR query", args{query: "this OR that"}, true},
		{"lower case or", args{query: "this or that"}, false},
		{"exclude query", args{query: "this -that"}, true},
		{"include query", args{query: "this +that"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAdvancedSearch(tt.args.query); got != tt.want {
				t.Errorf("isAdvancedSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDateRangeFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    *QueryFilter
		wantErr bool
	}{
		{
			"full date",
			args{filter: "ead-rdf_normalDate:1600~1750"},
			&QueryFilter{
				SearchLabel: "ead-rdf_normalDate",
				Value:       "1600~1750",
				Type:        QueryFilterType_DATERANGE,
				Lte:         "1750",
				Gte:         "1600",
			},
			false,
		},
		{
			"only start date",
			args{filter: "ead-rdf_normalDate:1600~"},
			&QueryFilter{
				SearchLabel: "ead-rdf_normalDate",
				Value:       "1600~",
				Type:        QueryFilterType_DATERANGE,
				Gte:         "1600",
			},
			false,
		},
		{
			"only end date",
			args{filter: "ead-rdf_normalDate:~1750"},
			&QueryFilter{
				SearchLabel: "ead-rdf_normalDate",
				Value:       "~1750",
				Type:        QueryFilterType_DATERANGE,
				Lte:         "1750",
			},
			false,
		},
		{
			"null values should be removed",
			args{filter: "ead-rdf_normalDate:null~1750"},
			&QueryFilter{
				SearchLabel: "ead-rdf_normalDate",
				Value:       "~1750",
				Type:        QueryFilterType_DATERANGE,
				Lte:         "1750",
			},
			false,
		},
		{
			"empty range with ~ not allowed",
			args{filter: "ead-rdf_normalDate:~"},
			nil,
			true,
		},
		{
			"empty range not allowed",
			args{filter: "ead-rdf_normalDate:"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDateRangeFilter(tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDateRangeFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDateRangeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFacetURIBuilder_CreateFacetFilterQuery(t *testing.T) {
	type fields struct {
		filters []string
	}
	type args struct {
		filterField string
		andQuery    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"default query is empty bool",
			fields{
				filters: []string{},
			},
			args{
				filterField: "dc_subject",
				andQuery:    false,
			},
			`./testdata/queries/0_empty_bool_query.json`,
			false,
		},
		{
			"single same filter",
			fields{
				filters: []string{
					"dc_subject:boerderij",
				},
			},
			args{
				filterField: "dc_subject",
				andQuery:    false,
			},
			`./testdata/queries/0_empty_bool_query.json`,
			false,
		},
		{
			"single different filter",
			fields{
				filters: []string{
					"ead-rdf_cType:series",
				},
			},
			args{
				filterField: "dc_date",
				andQuery:    false,
			},
			"./testdata/queries/3_combined_nested_single_filter.json",
			false,
		},
		{
			"double different filter",
			fields{
				filters: []string{
					"ead-rdf_cType:series",
					"ead-rdf_cType:file",
				},
			},
			args{
				filterField: "dc_date",
				andQuery:    false,
			},
			"./testdata/queries/4_combined_nested_double_filter.json",
			false,
		},
		{
			"double different AND filter",
			fields{
				filters: []string{
					"ead-rdf_cType:series",
					"ead-rdf_cType:file",
				},
			},
			args{
				filterField: "dc_date",
				andQuery:    true,
			},
			"./testdata/queries/8_combined_nested_double_AND_filter.json",
			false,
		},
		{
			"combined two filter fields one selected",
			fields{
				filters: []string{
					"ead-rdf_cType:series",
					"ead-rdf_cType:file",
					"dc_date:1977",
				},
			},
			args{
				filterField: "dc_date",
				andQuery:    false,
			},
			"./testdata/queries/4_combined_nested_double_filter.json",
			false,
		},
		{
			"combined two filter fields; none selected",
			fields{
				filters: []string{
					"dc_date:1977",
					"ead-rdf_cType:series",
					"ead-rdf_cType:file",
				},
			},
			args{
				filterField: "dc_subject",
				andQuery:    false,
			},
			"./testdata/queries/5_combined_nested_mixed_query.json",
			false,
		},
		{
			"single different exclude filter",
			fields{
				filters: []string{
					"-ead-rdf_cType:series",
				},
			},
			args{
				filterField: "dc_date",
				andQuery:    false,
			},
			"./testdata/queries/6_combined_nested_single_exclude_filter.json",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fub, err := NewFacetURIBuilder("", []*QueryFilter{})
			if err != nil {
				t.Errorf("FacetURIBuilder.NewFacetURIBuilder should not throw error: %#v", err)
				return
			}

			// add filters
			for _, filter := range tt.fields.filters {
				qf, err := NewQueryFilter(filter)
				if err != nil {
					t.Errorf("NewQueryFilter should not throw error: %#v", err)
					return
				}
				err = fub.AddFilter(qf)
				if err != nil {
					t.Errorf("FacetURIBuilder.AddFilter() should not throw error: %#v", err)
					return
				}
			}

			got, err := fub.CreateFacetFilterQuery(tt.args.filterField, tt.args.andQuery)
			if (err != nil) != tt.wantErr {
				t.Errorf("FacetURIBuilder.CreateFacetFilterQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			src, _ := got.Source()
			queryMap, _ := json.MarshalIndent(src, "", "    ")
			jsonQuery, _ := os.ReadFile(tt.want)

			if diff := cmp.Diff(trimRedundantWhiteSpace(jsonQuery), trimRedundantWhiteSpace(queryMap)); diff != "" {
				t.Errorf("FacetURIBuilder.CreateFacetFilterQuery() %s mismatch (-want +got):\n%s", tt.name, diff)
			}
			//queryMap := make(map[string]interface{})
			//_ = json.Unmarshal(jsonQuery, &queryMap)

			//if diff := cmp.Diff(src, queryMap); diff != "" {
			//t.Errorf("FacetURIBuilder.CreateFacetFilterQuery() %s mismatch (-want +got):\n%s", tt.name, diff)
			//}
		})
	}
}

func trimRedundantWhiteSpace(text []byte) string {
	singleSpacePattern := regexp.MustCompile(`\s+`)
	noReturns := strings.ReplaceAll(string(text), "\n", "")
	return singleSpacePattern.ReplaceAllString(noReturns, " ")
}

func TestQueryFilter_ElasticFilter(t *testing.T) {
	type fields struct {
		qf     string
		exists bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"simple query",
			fields{qf: "ead-rdf_cType:series", exists: false},
			"./testdata/queries/1_nested_single_filter.json",
			false,
		},
		{
			"exists query",
			fields{qf: "ead-rdf_cType", exists: true},
			"./testdata/queries/2_nested_exist_query.json",
			false,
		},
		{
			"simple query",
			fields{qf: "-ead-rdf_cType:series", exists: false},
			"./testdata/queries/7_nested_single_exclude_filter.json",
			false,
		},
		// TODO add unit tests for type and nested queries
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var qf *QueryFilter
			var err error
			switch tt.fields.exists {
			case false:
				qf, err = NewQueryFilter(tt.fields.qf)
				if err != nil {
					t.Errorf("NewQueryFilter should not throw error: %#v", err)
					return
				}
			case true:
				qf = &QueryFilter{
					SearchLabel: tt.fields.qf,
					Exists:      true,
				}
			}

			got, err := qf.ElasticFilter()
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryFilter.ElasticFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			src, _ := got.Source()
			queryMap, _ := json.MarshalIndent(src, "", "    ")
			jsonQuery, _ := os.ReadFile(tt.want)
			if diff := cmp.Diff(string(jsonQuery), string(queryMap)+"\n"); diff != "" {
				t.Errorf("QueryFilter.ElasticFilter() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFacetURIBuilder_CreateFacetFilterURI(t *testing.T) {
	type fields struct {
		filters []string
	}
	type args struct {
		field string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			"no filter query",
			fields{
				filters: []string{},
			},
			args{field: "ead-rdf_cType", value: "series"},
			"qf[]=ead-rdf_cType:series",
			false,
		},
		{
			"filter query; selected",
			fields{
				filters: []string{
					"ead-rdf_cType:series",
				},
			},
			args{field: "ead-rdf_cType", value: "series"},
			"",
			true,
		},
		{
			"triple filter, same searchLabel; selected",
			fields{
				filters: []string{
					"ead-rdf_files:123 inventories",
					"ead-rdf_cType:series",
					"ead-rdf_cType:file",
				},
			},
			args{field: "ead-rdf_cType", value: "series"},
			"qf[]=ead-rdf_cType:file&qf[]=ead-rdf_files:123 inventories",
			true,
		},
		{
			"tree: double filter, same searchLabel; selected",
			fields{
				filters: []string{
					"tree.type:series",
					"tree.type:file",
					"tree.type:otherlevel",
				},
			},
			args{field: "tree.type", value: "series"},
			"qf.tree[]=type:file&qf.tree[]=type:otherlevel",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fub, err := NewFacetURIBuilder("", []*QueryFilter{})
			if err != nil {
				t.Errorf("FacetURIBuilder.NewFacetURIBuilder should not throw error: %#v", err)
				return
			}

			// add filters
			for _, filter := range tt.fields.filters {
				qf, err := NewQueryFilter(filter)
				if err != nil {
					t.Errorf("NewQueryFilter should not throw error: %#v", err)
					return
				}
				err = fub.AddFilter(qf)
				if err != nil {
					t.Errorf("FacetURIBuilder.AddFilter() should not throw error: %#v", err)
					return
				}
			}
			got, got1 := fub.CreateFacetFilterURI(tt.args.field, tt.args.value)
			if got != tt.want {
				t.Errorf("FacetURIBuilder.CreateFacetFilterURI() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FacetURIBuilder.CreateFacetFilterURI() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getCursorFromPage(t *testing.T) {
	type args struct {
		page         int32
		responseSize int32
	}

	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			"first page",
			args{page: 1, responseSize: 16},
			int32(0),
		},
		{
			"second page",
			args{page: 2, responseSize: 16},
			int32(16),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := getCursorFromPage(tt.args.page, tt.args.responseSize); got != tt.want {
				t.Errorf("getCursorFromPage() = %v, want %v", got, tt.want)
			}
		})
	}
}
