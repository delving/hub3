// Copyright 2017 Delving B.V.
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

package config

import (
	"sync"
)

// RDFTag holds tag information how to tag predicate values
type RDFTag struct {
	Title            []string `json:"title"`
	Label            []string `json:"label"`
	Owner            []string `json:"owner"`
	Thumbnail        []string `json:"thumbnail"`
	LandingPage      []string `json:"landingPage"`
	Description      []string `json:"description"`
	Subject          []string `json:"subject"`
	Date             []string `json:"date"`
	Collection       []string `json:"collection"`
	SubCollectection []string `json:"subCollectection"`
	ObjectType       []string `json:"objectType"`
	ObjectID         []string `json:"objectID"`
	Creator          []string `json:"creator"`
	LatLong          []string `json:"latLong"`
	IsoDate          []string `json:"isoDate"`
	DateRange        []string `json:"dateRange"`
	Integer          []string `json:"integer"`
	IntegerRange     []string `json:"integerRange"`
}

// RDFTagMap contains all the URIs that trigger indexing labels
type RDFTagMap struct {
	sync.RWMutex
	TagMap map[string][]string
}

type tagPair struct {
	tag  string
	uris []string
}

// NewRDFTagMap return
func NewRDFTagMap(c *RawConfig) *RDFTagMap {
	pairs := []tagPair{
		{"label", c.RDFTag.Label},
		{"title", c.RDFTag.Title},
		{"owner", c.RDFTag.Owner},
		{"thumbnail", c.RDFTag.Thumbnail},
		{"landingPage", c.RDFTag.LandingPage},
		{"latLong", c.RDFTag.LatLong},
		{"isoDate", c.RDFTag.IsoDate},
		{"date", c.RDFTag.Date},
		{"description", c.RDFTag.Description},
		{"subject", c.RDFTag.Subject},
		{"collection", c.RDFTag.Collection},
		{"subCollection", c.RDFTag.SubCollectection},
		{"objectType", c.RDFTag.ObjectType},
		{"objectID", c.RDFTag.ObjectID},
		{"creator", c.RDFTag.Creator},
		{"dateRange", c.RDFTag.DateRange},
	}

	tagMap := make(map[string][]string)

	for _, pair := range pairs {
		for _, uri := range pair.uris {
			values, ok := tagMap[uri]
			if ok {
				tagMap[uri] = append(values, pair.tag)
				continue
			}

			tagMap[uri] = []string{pair.tag}
		}
	}

	return &RDFTagMap{
		TagMap: tagMap,
	}
}

// Len return number of URIs in the RDFTagMap
func (rtm *RDFTagMap) Len() int {
	return len(rtm.TagMap)
}

// Get returns the indexType label for a given URI
func (rtm *RDFTagMap) Get(uri string) ([]string, bool) {
	rtm.RLock()
	label, ok := rtm.TagMap[uri]
	rtm.RUnlock()

	return label, ok
}

func NewDataSetTagMap(c *RawConfig) *RDFTagMap {
	tagMap := make(map[string][]string)

	for tag, specs := range c.DataSetTag {
		for _, spec := range specs.Specs {
			values, ok := tagMap[spec]
			if ok {
				tagMap[spec] = append(values, tag)
				continue
			}

			tagMap[spec] = []string{tag}
		}
	}

	return &RDFTagMap{
		TagMap: tagMap,
	}
}
