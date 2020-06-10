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

import (
	"bytes"
	"fmt"
	"index/suffixarray"
	"io"
	"sort"
)

type Autos struct {
	Term     string
	Count    int
	Metadata map[string][]string
}

type AutoComplete struct {
	sa        *suffixarray.Index
	data      []byte
	SuggestFn func(a Autos) Autos
}

func NewAutoComplete() *AutoComplete {
	return &AutoComplete{}
}

func (ac *AutoComplete) updateSuffixArray(count int, writer func(w io.Writer)) {
	if count == 0 {
		ac.data = []byte{}
		return
	}

	var buf bytes.Buffer

	buf.WriteString("\x00")
	writer(&buf)

	ac.data = buf.Bytes()
	ac.sa = suffixarray.New(ac.data)
}

func (ac *AutoComplete) FromTokenSteam(stream *TokenStream) {
	fn := func(w io.Writer) {
		for _, token := range stream.Tokens() {
			if !token.Ignored && token.Normal != "" {
				_, _ = w.Write([]byte(token.Normal + "\x00"))
			}
		}
	}

	ac.updateSuffixArray(len(stream.Tokens()), fn)
}

func (ac *AutoComplete) FromStrings(words []string) {
	fn := func(w io.Writer) {
		for _, word := range words {
			_, _ = w.Write([]byte(word + "\x00"))
		}
	}

	ac.updateSuffixArray(len(words), fn)
}

func (ac *AutoComplete) getStringFromIndex(index int) string {
	if index > len(ac.data) {
		return ""
	}

	var start, end int

	for i := index - 1; i >= 0; i-- {
		if ac.data[i] == 0 {
			start = i + 1
			break
		}
	}

	for i := index + 1; i < len(ac.data); i++ {
		if ac.data[i] == 0 {
			end = i
			break
		}
	}

	return string(ac.data[start:end])
}

func (ac *AutoComplete) Suggest(input string, limit int) ([]Autos, error) {
	if ac.sa == nil {
		return []Autos{}, fmt.Errorf("cannot suggest from empty AutoComplete")
	}

	if input == "" {
		return []Autos{}, fmt.Errorf("input cannot be empty")
	}

	indices := ac.sa.Lookup([]byte(input), -1)

	terms := map[string]int{}

	for _, idx := range indices {
		term := ac.getStringFromIndex(idx)
		terms[term]++
	}

	autos := make([]Autos, 0, len(terms))

	for term, count := range terms {
		a := Autos{
			Term:  term,
			Count: count,
		}

		if ac.SuggestFn != nil {
			a = ac.SuggestFn(a)
		}

		if a.Count != 0 {
			autos = append(autos, a)
		}
	}

	sort.Slice(autos, func(i, j int) bool { return autos[i].Count > autos[j].Count })

	if limit > 0 && limit < len(autos) {
		return autos[:limit], nil
	}

	return autos, nil
}
