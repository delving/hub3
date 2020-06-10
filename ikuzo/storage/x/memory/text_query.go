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

package memory

import (
	"errors"
	"fmt"

	"github.com/delving/hub3/ikuzo/service/x/search"
)

const (
	startTag   = "em"
	styleClass = "dchl"
)

type TextQuery struct {
	ti           *TextIndex
	q            *search.QueryTerm
	Hits         *search.Matches
	EmStartTag   string
	EmStyleClass string
}

func NewTextQuery(q *search.QueryTerm) *TextQuery {
	return &TextQuery{
		q:            q,
		ti:           NewTextIndex(),
		EmStartTag:   startTag,
		EmStyleClass: styleClass,
	}
}

func NewTextQueryFromString(query string) (*TextQuery, error) {
	qp, err := search.NewQueryParser()
	if err != nil {
		return nil, err
	}

	q, err := qp.Parse(query)
	if err != nil {
		return nil, err
	}

	tq := NewTextQuery(q)

	return tq, nil
}

func (tq *TextQuery) Reset() {
	tq.Hits = nil
	tq.ti = NewTextIndex()
}

func (tq *TextQuery) AppendString(text string, docID int) error {
	if err := tq.ti.AppendString(text, docID); err != nil {
		return fmt.Errorf("text query index error: %w", err)
	}

	return nil
}

func (tq *TextQuery) PerformSearch() (bool, error) {
	hits, err := tq.ti.Search(tq.q)
	if errors.Is(err, ErrSearchNoMatch) {
		return false, nil
	}

	tq.Hits = hits

	return true, nil
}

func (tq *TextQuery) Highlight(text string, docID int) (string, bool) {
	if !tq.ti.hasDocID(docID) || tq.Hits == nil {
		return text, false
	}

	return tq.hightlightWithVectors(text, docID, tq.Hits.Vectors()), true
}

func (tq *TextQuery) hightlightWithVectors(text string, docID int, vectors *search.Vectors) string {
	if !tq.ti.hasDocID(docID) || vectors == nil {
		return text
	}

	tok := search.NewTokenizer()
	tokens := tok.ParseString(text, docID)

	return tokens.Highlight(vectors, tq.EmStartTag, tq.EmStyleClass)
}

func (tq *TextQuery) SetTextIndex(ti *TextIndex) {
	tq.ti = ti
}
