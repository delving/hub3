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

package config

import "sync"

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
		tagPair{"label", c.RDFTag.Label},
		tagPair{"title", c.RDFTag.Title},
		tagPair{"owner", c.RDFTag.Owner},
		tagPair{"thumbnail", c.RDFTag.Thumbnail},
		tagPair{"landingPage", c.RDFTag.LandingPage},
		tagPair{"latLong", c.RDFTag.LatLong},
		tagPair{"isoDate", c.RDFTag.IsoDate},
		tagPair{"date", c.RDFTag.Date},
		tagPair{"description", c.RDFTag.Description},
		tagPair{"subject", c.RDFTag.Subject},
		tagPair{"collection", c.RDFTag.Collection},
		tagPair{"subCollection", c.RDFTag.SubCollectection},
		tagPair{"objectType", c.RDFTag.ObjectType},
		tagPair{"objectID", c.RDFTag.ObjectID},
		tagPair{"creator", c.RDFTag.Creator},
		tagPair{"dateRange", c.RDFTag.DateRange},
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
