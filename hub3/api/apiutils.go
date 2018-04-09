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

package api

import (
	"encoding/hex"
	fmt "fmt"
	"net/url"

	c "github.com/delving/rapid-saas/config"
	proto "github.com/golang/protobuf/proto"
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
		return "", err
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
func NewSearchRequest(params url.Values) (SearchRequest, error) {
	sr := SearchRequest{}
	//for p, v := range params {
	//switch p {
	//case "q", "query":
	//sr.Query = v
	//case "qf", "qf[]":
	//sr.QueryFilter = append(sr, v)
	//case "rows":
	//size, err := strconv.Atoi(v)
	//if err != nil {
	//log.Printf("unable to convert %v to int", v)
	//return sr, err
	//}
	//sr.ResponseSize = size
	//}
	//}
	return sr, nil
}
