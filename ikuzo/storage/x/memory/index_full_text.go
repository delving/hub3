package memory

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/service/x/search"
)

var (
	ErrSearchNoMatch = errors.New("the search query does not match the index")
)

// TestIndex is a single document full-text index.
// This means that all data you append to it will have its position incremented
// and appends to the known state. It is not replaced. To reset the index to
// an empty state you have to call the reset method.
type TextIndex struct {
	Terms    map[string]*search.Vectors
	a        search.Analyzer
	DocCount int
	Docs     map[int]bool
}

func NewTextIndex() *TextIndex {
	return &TextIndex{
		Terms: make(map[string]*search.Vectors),
		Docs:  make(map[int]bool),
	}
}

func (ti *TextIndex) reset() {
	ti.Terms = make(map[string]*search.Vectors)
	ti.Docs = make(map[int]bool)
	ti.DocCount = 0
}

func (ti *TextIndex) setDocID(docID ...int) int {
	var id int

	if len(docID) == 0 {
		ti.DocCount++
		id = ti.DocCount
		ti.Docs[id] = true

		return id
	}

	id = docID[0]
	if id == 0 {
		return ti.setDocID()
	}

	ti.DocCount = id
	ti.Docs[id] = true

	return id
}

func (ti *TextIndex) hasDocID(docID int) bool {
	_, ok := ti.Docs[docID]
	return ok
}

// AppendBytes extract words from bytes and updates the TextIndex.
func (ti *TextIndex) AppendBytes(b []byte, docID ...int) error {
	id := ti.setDocID(docID...)

	tok := search.NewTokenizer()
	for _, token := range tok.ParseBytes(b, id).Tokens() {
		if !token.Ignored {
			err := ti.addTerm(token.Normal, token.TermVector)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AppendString extract words from bytes and updates the TextIndex.
func (ti *TextIndex) AppendString(text string, docID ...int) error {
	id := ti.setDocID(docID...)

	tok := search.NewTokenizer()
	for _, token := range tok.ParseString(text, id).Tokens() {
		if !token.Ignored {
			err := ti.addTerm(token.RawText, token.TermVector)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ti *TextIndex) size() int {
	return len(ti.Terms)
}

func (ti *TextIndex) setTermVector(term string, pos int) {
	tv, ok := ti.Terms[term]
	if !ok {
		tv = search.NewVectors()
		ti.Terms[term] = tv
	}

	tv.Add(ti.DocCount, pos)
}

func (ti *TextIndex) addTerm(word string, pos int) error {
	if word == "" {
		return fmt.Errorf("cannot index empty string")
	}

	analyzedTerm := ti.a.Transform(word)

	if analyzedTerm == "" {
		return nil
	}

	ti.setTermVector(analyzedTerm, pos)

	return nil
}

func (ti *TextIndex) match(qt *search.QueryTerm, hits *search.Matches) bool {
	switch qt.Type() {
	case search.WildCardQuery:
		return ti.matchWildcard(qt, hits)
	case search.PhraseQuery:
		return ti.matchPhrase(qt, hits)
	case search.FuzzyQuery:
		return ti.matchFuzzy(qt, hits)
	default:
		// search.TermQuery is the default
		return ti.matchTerm(qt, hits)
	}
}

func (ti *TextIndex) matchPhrase(qt *search.QueryTerm, hits *search.Matches) bool {
	var nextVectors map[search.Vector]bool

	phrasePositions := map[search.Vector]string{}

	words := strings.Fields(qt.Value)

	if len(words) == 1 {
		term, ok := ti.Terms[qt.Value]
		if !ok {
			return false
		}

		hits.AppendTerm(qt.Value, term)

		return true
	}

	var previousTerm string

	for idx, word := range words {
		term, ok := ti.Terms[word]
		if !ok {
			return false
		}

		if idx != 0 {
			var wordMatch bool

			for vector := range nextVectors {
				// posMatch determines if this position can be followed
				var posMatch bool

				for _, next := range search.ValidPhrasePosition(vector, qt.Slop) {
					ok := term.HasVector(next)

					if ok {
						phrasePositions[next] = word
						posMatch = true
						wordMatch = true
					}
				}

				if posMatch {
					phrasePositions[vector] = previousTerm
				}
			}

			if !wordMatch {
				return false
			}
		}

		nextVectors = term.Locations

		previousTerm = word
	}

	matches := len(phrasePositions)

	if matches != 0 {
		for term, vectors := range sortAndCountPhrases(words, phrasePositions) {
			hits.AppendTerm(term, vectors)
		}
	}

	return matches != 0
}

func sortAndCountPhrases(words []string, phrases map[search.Vector]string) map[string]*search.Vectors {
	phraseHits := map[string]*search.Vectors{}

	vectors := []search.Vector{}
	phraseSize := len(words)

	for vector := range phrases {
		vectors = append(vectors, vector)
	}

	sort.Slice(vectors, func(i, j int) bool {
		return vectors[i].Location < vectors[j].Location
	})

	phrase := []string{}

	prevVectors := []search.Vector{}

	for idx, vector := range vectors {
		phrase = append(phrase, phrases[vector])
		if len(phrase) != phraseSize {
			if idx < len(vectors) {
				prevVectors = append(prevVectors, vector)
				continue
			}
		}

		currentPhrase := strings.Join(phrase, " ")

		tv, ok := phraseHits[currentPhrase]
		if !ok {
			tv = search.NewVectors()

			phraseHits[currentPhrase] = tv
		}

		for _, v := range prevVectors {
			tv.AddPhraseVector(v)
		}

		tv.AddVector(vector)

		phrase = []string{}
		prevVectors = []search.Vector{}
	}

	return phraseHits
}

func (ti *TextIndex) matchFuzzy(qt *search.QueryTerm, hits *search.Matches) bool {
	var hasMatch bool

	for k, tv := range ti.Terms {
		ok, _ := search.IsFuzzyMatch(k, qt.Value, float64(qt.Fuzzy), search.Levenshtein)
		if ok {
			hasMatch = true

			hits.AppendTerm(k, tv)
		}
	}

	return hasMatch
}

func (ti *TextIndex) matchWildcard(qt *search.QueryTerm, hits *search.Matches) bool {
	var matcher func(s, prefix string) bool

	switch {
	case qt.SuffixWildcard:
		matcher = strings.HasSuffix
	default:
		matcher = strings.HasPrefix
	}

	var hasMatch bool

	for k, tv := range ti.Terms {
		if matcher(k, qt.Value) {
			hasMatch = true

			hits.AppendTerm(k, tv)
		}
	}

	return hasMatch
}

func (ti *TextIndex) matchTerm(qt *search.QueryTerm, hits *search.Matches) bool {
	term, ok := ti.Terms[qt.Value]
	if ok && qt.Prohibited {
		return false
	}

	if !ok && !qt.Prohibited {
		return false
	}

	if term == nil {
		term = search.NewVectors()
	}

	hits.AppendTerm(qt.Value, term)

	return true
}

func (ti *TextIndex) Search(query *search.QueryTerm) (*search.Matches, error) {
	hits := search.NewMatches()
	err := ti.search(query, hits)

	if errors.Is(err, ErrSearchNoMatch) {
		hits.Reset()
	}

	return hits, err
}

func (ti *TextIndex) searchMustNot(query *search.QueryTerm, hits *search.Matches) error {
	for _, qt := range query.MustNot() {
		switch {
		case qt.Type() == search.BoolQuery:
			// TODO(kiivihal): find out why this is a dead branch
			if err := ti.search(qt, hits); err != nil {
				return err
			}
		default:
			if !qt.Prohibited {
				return fmt.Errorf("mustNot clauses must be marked as prohibited")
			}

			if ok := ti.match(qt, hits); !ok {
				return ErrSearchNoMatch
			}
		}
	}

	return nil
}

func (ti *TextIndex) searchMust(query *search.QueryTerm, hits *search.Matches) error {
	for _, qt := range query.Must() {
		switch {
		case qt.Type() == search.BoolQuery:
			if err := ti.search(qt, hits); err != nil {
				return err
			}
		default:
			if ok := ti.match(qt, hits); !ok {
				return ErrSearchNoMatch
			}
		}
	}

	return nil
}

func (ti *TextIndex) searchShould(query *search.QueryTerm, hits *search.Matches) error {
	var matched bool

	for _, qt := range query.Should() {
		switch {
		case qt.Type() == search.BoolQuery:
			if err := ti.search(qt, hits); err != nil {
				return err
			}
		default:
			if ok := ti.match(qt, hits); ok {
				matched = true
			}
		}
	}

	if !matched {
		return ErrSearchNoMatch
	}

	return nil
}

// recursive search function
func (ti *TextIndex) search(query *search.QueryTerm, hits *search.Matches) error {
	if len(query.MustNot()) != 0 {
		if err := ti.searchMustNot(query, hits); err != nil {
			return err
		}
	}

	if len(query.Must()) != 0 {
		if err := ti.searchMust(query, hits); err != nil {
			return err
		}
	}

	if len(query.Should()) != 0 {
		if err := ti.searchShould(query, hits); err != nil {
			return err
		}
	}

	return nil
}

func (ti *TextIndex) Encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	err := e.Encode(ti)
	if err != nil {
		return fmt.Errorf("unable to marshall TextIndex to GOB; %w", err)
	}

	return nil
}

func DecodeTextIndex(r io.Reader) (*TextIndex, error) {
	var ti TextIndex

	d := gob.NewDecoder(r)

	err := d.Decode(&ti)
	if err != nil {
		return nil, err
	}

	return &ti, nil
}
