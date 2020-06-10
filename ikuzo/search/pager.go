// Copyright 2020 Delving B.V.
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

package search

// ScrollPager holds all paging information for a search result.
type ScrollPager struct {
	// scrollID is serialized version SearchRequest
	ScrollID string `json:"scrollID"`
	Cursor   int32  `json:"cursor"`
	Total    int64  `json:"total"`
	Rows     int32  `json:"rows"`
}

// NewScrollPager returns a ScrollPager with defaults set
func NewScrollPager() ScrollPager {
	return ScrollPager{}
}
