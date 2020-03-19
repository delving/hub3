package memory

import (
	"errors"
	"log"

	"github.com/delving/hub3/ikuzo/service/x/search"
)

const (
	startTag = "<em class=\"dchl\">"
	endTag   = "</em>"
)

type TextQuery struct {
	ti         *TextIndex
	q          *search.QueryTerm
	Hits       *search.Matches
	EmStartTag string
	EmEndTag   string
}

func NewTextQuery(q *search.QueryTerm) *TextQuery {
	return &TextQuery{
		q:          q,
		ti:         NewTextIndex(),
		Hits:       search.NewMatches(),
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

	return tq.hightlightWithVectors(text, hits.WordPositions()), true
}

func (tq *TextQuery) hightlightWithVectors(text string, positions map[int]bool) string {
	tok := search.NewTokenizer()
	tokens := tok.ParseString(text)

	return tokens.Highlight(positions, tq.EmStartTag, tq.EmEndTag)
}
