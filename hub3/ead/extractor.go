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

package ead

import (
	"regexp"
)

type NLPType int

const (
	Unknown NLPType = iota
	Person
	GeoLocation
	DateText
	DateIso
	Organization
)

type NLPToken struct {
	Type NLPType
	Text string
}

type Extractor struct {
	tokens []NLPToken
}

func NewExtractor(b []byte) (*Extractor, error) {
	e := Extractor{tokens: []NLPToken{}}

	re := regexp.MustCompile(`<(geogname|persname)>(.*?)</(geogname|persname)>`)
	matches := re.FindAllSubmatch(b, -1)

	for _, m := range matches {
		switch string(m[1]) {
		case "geogname":
			e.tokens = append(e.tokens, NLPToken{Text: string(m[2]), Type: GeoLocation})
		case "persname":
			e.tokens = append(e.tokens, NLPToken{Text: string(m[2]), Type: Person})
		}
	}

	return &e, nil
}

func (e *Extractor) Tokens() []NLPToken {
	return e.tokens
}
