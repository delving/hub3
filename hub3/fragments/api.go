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
		log.Println("Unable to uild the query result.")
		return s, err
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
