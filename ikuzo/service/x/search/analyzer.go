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

import "strings"

const (
	trimCharacters = "\".,;:[]()?'"
)

// Analyzer is the default analyzer for Search actions.
// It folds unicode to ASCII characters and lowercases them all.
//
// The goal is to have this analyzer behave similarly to the ElasticSearch
// Analyzer that Ikuzo comes preconfigured with.
type Analyzer struct{}

func (a *Analyzer) Transform(text string) string {
	return strings.Trim(
		strings.ToLower(
			LuceneASCIIFolding(text),
		),
		trimCharacters,
	)
}

func (a *Analyzer) TransformPhrase(text string) string {
	cleanWords := []string{}

	for _, word := range strings.Fields(text) {
		cleanWords = append(cleanWords, a.Transform(word))
	}

	return strings.Join(cleanWords, " ")
}
