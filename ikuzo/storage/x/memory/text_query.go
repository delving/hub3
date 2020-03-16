package memory

import (
	"errors"
	"log"
	"strings"

	"github.com/delving/hub3/ikuzo/service/x/search"
)

const (
	startTag = "<em class=\"dchl\">"
	endTag   = "</em>"
)

type TextQuery struct {
	ti         *TextIndex
	q          *search.QueryTerm
	Hits       *SearchHits
	EmStartTag string
	EmEndTag   string
}

func NewTextQuery(q *search.QueryTerm) *TextQuery {
	return &TextQuery{
		q:          q,
		ti:         NewTextIndex(),
		Hits:       newSearchHits(),
		EmStartTag: startTag,
		EmEndTag:   endTag,
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

func (tq *TextQuery) Highlight(text string) (string, bool) {
	tq.ti.reset()

	err := tq.ti.AppendString(text)
	if err != nil {
		log.Printf("index error: %#v", err)
		return "", false
	}

	hits, err := tq.ti.Search(tq.q)
	if errors.Is(err, ErrSearchNoMatch) {
		return text, false
	}

	tq.Hits.Merge(hits)

	return tq.hightlightWithVectors(text, hits.matchPositions), true
}

func (tq *TextQuery) hightlightWithVectors(text string, positions map[int]bool) string {
	words := []string{}

	for idx, word := range strings.Fields(text) {
		if _, ok := positions[idx]; !ok {
			words = append(words, word)
			continue
		}

		words = append(words, tq.EmStartTag+word+tq.EmEndTag)
	}

	return strings.Join(words, " ")
}

// Highlight(text string) (string, bool)
