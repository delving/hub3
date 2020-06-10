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

// Facet is used in the search response to render Facet information.
type Facet struct {
	Name        string       `json:"name"`
	Field       string       `json:"field"`
	IsSelected  bool         `json:"isSelected"`
	I18n        string       `json:"i18N,omitempty"`
	Total       int64        `json:"total"`
	MissingDocs int64        `json:"missingDocs"`
	OtherDocs   int64        `json:"otherDocs"`
	Min         string       `json:"min,omitempty"`
	Max         string       `json:"max,omitempty"`
	Type        string       `json:"type,omitempty"`
	Links       []*FacetLink `json:"links"`
}

// FacetLink is used to build filter URIs in the search response.
type FacetLink struct {
	URL           string `json:"url"`
	IsSelected    bool   `json:"isSelected"`
	Value         string `json:"value"`
	DisplayString string `json:"displayString"`
	Count         int64  `json:"count"`
}

type Response struct {
	//Pager      *ScrollPager       `json:"pager"`
	//Query      *Query             `json:"query"`
	//Items      []*FragmentGraph   `json:"items,omitempty"`
	//Collapsed  []*Collapsed       `json:"collapse,omitempty"`
	//Peek       map[string]int64   `json:"peek,omitempty"`
	//Facets     []*QueryFacet      `json:"facets,omitempty"`
	//TreeHeader *TreeHeader        `json:"treeHeader,omitempty"`
	//Tree       []*Tree            `json:"tree,omitempty"`
	//TreePage   map[string][]*Tree `json:"treePage,omitempty"`
	//ProtoBuf   *ProtoBuf          `json:"protobuf,omitempty"`
}
