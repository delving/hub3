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

type Matches struct {
	termFrequency map[string]int
	termVectors   *Vectors
}

func NewMatches() *Matches {
	return &Matches{
		termFrequency: make(map[string]int),
		termVectors:   NewVectors(),
	}
}

// Reset is used when already gathered matches must be reset when ErrSearchNoMatch is returned.
func (m *Matches) Reset() {
	m.termFrequency = make(map[string]int)
	m.termVectors = NewVectors()
}

func (m *Matches) AppendTerm(term string, tv *Vectors) {
	if tv.Size() == 0 {
		return
	}

	m.termFrequency[term] = tv.Size()
	m.mergeVectors(tv)
}

func (m *Matches) DocCount() int {
	return m.termVectors.DocCount()
}

func (m *Matches) HasDocID(docID int) bool {
	return m.termVectors.HasDoc(docID)
}

func (m *Matches) Merge(matches *Matches) {
	for key, count := range matches.termFrequency {
		v, ok := m.termFrequency[key]
		if ok {
			m.termFrequency[key] = v + count
			continue
		}

		m.termFrequency[key] = count
	}

	m.mergeVectors(matches.termVectors)
}

func (m *Matches) mergeVectors(tv *Vectors) {
	m.termVectors.Merge(tv)
}

func (m *Matches) TermFrequency() map[string]int {
	return m.termFrequency
}

func (m *Matches) TermCount() int {
	return len(m.termFrequency)
}

func (m *Matches) Total() int {
	var total int
	for _, hitCount := range m.termFrequency {
		total += hitCount
	}

	return total
}

func (m *Matches) Vectors() *Vectors {
	return m.termVectors
}
