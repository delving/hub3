// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

import (
	"encoding/hex"
	fmt "fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	c "github.com/delving/rapid-saas/config"
	proto "github.com/golang/protobuf/proto"
	elastic "gopkg.in/olivere/elastic.v5"
)

// DefaultSearchRequest takes an Config Objects and sets the defaults
func DefaultSearchRequest(c *c.RawConfig) *SearchRequest {
	sr := &SearchRequest{
		ResponseSize: int32(16),
	}
	return sr
}

// SearchRequestToHex converts the SearchRequest to a hex string
func SearchRequestToHex(sr *SearchRequest) (string, error) {
	output, err := proto.Marshal(sr)
	if err != nil {
	}
	return fmt.Sprintf("%x", output), nil
}

// SearchRequestFromHex creates a SearchRequest object from a string
func SearchRequestFromHex(s string) (*SearchRequest, error) {
	decoded, err := hex.DecodeString(s)
	newSr := &SearchRequest{}
	if err != nil {
		return newSr, nil
	}
	err = proto.Unmarshal(decoded, newSr)
	if err != nil {
		return newSr, nil
	}
	return newSr, nil
}

// NewSearchRequest builds a search request object from URL Parameters
func NewSearchRequest(params url.Values) (*SearchRequest, error) {
	hexRequest := params.Get("scrollID")
	if hexRequest == "" {
		hexRequest = params.Get("qs")
	}
	if hexRequest != "" {
		sr, err := SearchRequestFromHex(hexRequest)
		sr.Paging = true
		if err != nil {
			log.Println("Unable to parse search request from scrollID")
			return nil, err
		}
		return sr, nil
	}

	sr := DefaultSearchRequest(&c.Config)
	for p, v := range params {
		switch p {
		case "q", "query":
			sr.Query = params.Get(p)
		//case "qf", "qf[]":
		//sr.QueryFilter = append(sr.QueryFilter, v)
		case "facet.field":
			for _, field := range v {
				sr.FacetField = append(sr.FacetField, field)
			}
		case "format":
			switch params.Get(p) {
			case "protobuf":
				sr.ResponseFormatType = ResponseFormatType_PROTOBUF
			}
		case "rows":
			size, err := strconv.Atoi(params.Get(p))
			if err != nil {
				log.Printf("unable to convert %v to int", v)
				return sr, err
			}
			sr.ResponseSize = int32(size)
		}
	}
	return sr, nil
}

// ElasticQuery creates an ElasticSearch query from the Search Request
// This query can be passed into an elastic Search Object.
func (sr *SearchRequest) ElasticQuery() (elastic.Query, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("meta.docType", FragmentGraphDocType))
	query = query.Must(elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID))

	if sr.GetQuery() != "" {
		rawQuery := strings.Replace(sr.GetQuery(), "delving_spec:", "meta.spec:", 1)
		qs := elastic.NewQueryStringQuery(rawQuery)
		qs = qs.DefaultField("resources.entries.@value")
		nq := elastic.NewNestedQuery("resources.entries", qs)
		query = query.Must(nq)

	}

	return query, nil
}

// Aggregations returns the aggregations for the SearchRequest
func (sr *SearchRequest) Aggregations() (map[string]elastic.Aggregation, error) {

	aggs := map[string]elastic.Aggregation{}

	for _, facetField := range sr.FacetField {
		agg, err := sr.CreateAggregationBySearchLabel("resources.entries", facetField, false, 10)
		if err != nil {
			return nil, err
		}
		aggs[facetField] = agg
	}
	return aggs, nil
}

// CreateAggregationBySearchLabel creates Elastic aggregations for the nested fragment resources
func (sr *SearchRequest) CreateAggregationBySearchLabel(path, searchLabel string, byId bool, size int) (elastic.Aggregation, error) {
	nestedPath := fmt.Sprintf("%s.searchLabel", path)
	fieldQuery := elastic.NewTermQuery(nestedPath, searchLabel)

	entryKey := "@value.keyword"
	if byId {
		entryKey = "@id"
	}

	termAggPath := fmt.Sprintf("%s.%s", path, entryKey)

	labelAgg := elastic.NewTermsAggregation().Field(termAggPath).Size(size).OrderByCountDesc()

	filterAgg := elastic.NewFilterAggregation().Filter(fieldQuery).SubAggregation("value", labelAgg)

	testAgg := elastic.NewNestedAggregation().Path(path)
	testAgg = testAgg.SubAggregation("inner", filterAgg)

	return testAgg, nil
}

// ElasticSearchService creates the elastic SearchService for execution
func (sr *SearchRequest) ElasticSearchService(client *elastic.Client) (*elastic.SearchService, error) {
	idSort := elastic.NewFieldSort("meta.hubID")
	scoreSort := elastic.NewFieldSort("_score")
	s := client.Search().
		Index(c.Config.ElasticSearch.IndexName).
		Size(int(sr.GetResponseSize())).
		SortBy(scoreSort, idSort)

	//if sr.SearchAfter != "" {
	//s = s.SearchAfter()

	//}

	query, err := sr.ElasticQuery()
	if err != nil {
		log.Println("Unable to build the query result.")
		return s, err
	}

	s = s.Query(query)

	// Add aggregations
	if sr.Paging {
		return s.Query(query), err
	}

	aggs, err := sr.Aggregations()
	if err != nil {
		log.Println("Unable to build the Aggregations.")
		return s, err
	}
	for facetField, agg := range aggs {
		s = s.Aggregation(facetField, agg)
	}

	return s.Query(query), err
}

// NewScrollPager returns a ScrollPager with defaults set
func NewScrollPager() *ScrollPager {
	sp := &ScrollPager{}
	sp.Total = 0
	sp.Cursor = 0
	return sp

}

// Echo returns a json version of the request object for introspection
func (sr *SearchRequest) Echo(echoType string, total int64) (interface{}, error) {
	switch echoType {
	case "es":
		query, err := sr.ElasticQuery()
		if err != nil {
			return nil, err
		}
		source, _ := query.Source()
		return source, nil
	case "aggs":
		aggs, err := sr.Aggregations()
		if err != nil {
			return nil, err
		}
		sourceMap := map[string]interface{}{}
		for k, v := range aggs {
			source, _ := v.Source()
			sourceMap[k] = source
		}
		return sourceMap, nil
	case "nextScrollID":
		pager, err := sr.NextScrollID(total)
		if err != nil {
			return nil, err
		}
		next, err := SearchRequestFromHex(pager.GetScrollID())
		if err != nil {
			return nil, err
		}
		return next, nil
	case "searchRequest":
		return sr, nil
	case "searchService", "searchResponse", "request":
		return nil, nil
	}
	return nil, fmt.Errorf("unknown echoType: %s", echoType)

}

// NextScrollID creates a ScrollPager from a SearchRequest
// This is used to provide a scrolling pager for returning SearchItems
func (sr *SearchRequest) NextScrollID(total int64) (*ScrollPager, error) {

	sp := NewScrollPager()

	// if no results return empty pager
	if total == 0 {
		return sp, nil
	}
	sp.Cursor = sr.GetStart()

	// set the next cursor
	sr.Start = sr.GetStart() + sr.GetResponseSize()

	sp.Rows = sr.GetResponseSize()
	sp.Total = total

	// return empty ScrollID if there is no next page
	if sr.GetStart() >= int32(total) {
		return sp, nil
	}

	hex, err := SearchRequestToHex(sr)
	if err != nil {
		return nil, err
	}
	sp.ScrollID = hex
	return sp, nil
}

// NewQueryFilter parses the filter string and creates a QueryFilter object
func NewQueryFilter(filter string) (*QueryFilter, error) {
	// split once on the first :
	// split on first part and ]. This should give one or two
	// determine the levels of nesting for the filter
	// assign to values of the QueryFilter struct

	return &QueryFilter{}, nil
}

// ElasticFilter creates an elasticsearch filter from the QueryFilter
func (qf *QueryFilter) ElasticFilter() elastic.Query {
	return nil
}

// AddQueryFilter adds a QueryFilter to the SearchRequest
// The raw query from the QueryString are added here. This function converts
// this string to a QueryFilter.
func (sr *SearchRequest) AddQueryFilter(filter string) error {
	return nil
}

// RemoveQueryFilter removes a QueryFilter from the SearchRequest
// The raw query from the QueryString are added here.
func (sr *SearchRequest) RemoveQueryFilter(filter string) error {
	return nil
}

// DecodeFacets decodes the elastic aggregations in the SearchResult to fragments.QueryFacets
func (sr SearchRequest) DecodeFacets(res *elastic.SearchResult) ([]*QueryFacet, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil
	}

	var aggs []*QueryFacet
	for k := range res.Aggregations {
		facet, ok := res.Aggregations.Nested(k)
		if ok {
			inner, ok := facet.Filter("inner")
			if ok {
				value, ok := inner.Terms("value")
				if ok {
					qf := &QueryFacet{
						Name:      k,
						Total:     inner.DocCount,
						OtherDocs: value.SumOfOtherDocCount,
						Links:     []*FacetLink{},
					}
					for _, b := range value.Buckets {
						key := fmt.Sprintf("%s", b.Key)
						fl := &FacetLink{
							Value:         key,
							Count:         b.DocCount,
							DisplayString: fmt.Sprintf("%s (%d)", key, b.DocCount),
						}
						qf.Links = append(qf.Links, fl)
					}
					aggs = append(aggs, qf)
				}
			}
		}

	}
	return aggs, nil
}
