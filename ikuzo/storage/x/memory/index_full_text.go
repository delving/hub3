package memory

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/service/x/search"
)

var (
	ErrSearchNoMatch = errors.New("the search query does not match the index")
)

type termVector struct {
	position map[int]bool
	split    bool // replace with elasticsearch term for synonym in same position
}

func newTermVector() *termVector {
	return &termVector{
		position: make(map[int]bool),
	}
}

func (tv *termVector) size() int {
	return len(tv.position)
}

// TestIndex is a single document full-text index.
// This means that all data you append to it will have its position incremented
// and appends to the known state. It is not replaced. To reset the index to
// an empty state you have to call the reset method.
type TextIndex struct {
	terms map[string]*termVector
	a     search.Analyzer
}

func NewTextIndex() *TextIndex {
	return &TextIndex{
		terms: make(map[string]*termVector),
	}
}

// AppendBytes extract words from bytes and updates the TextIndex.
func (ti *TextIndex) AppendBytes(b []byte) error {
	words := bytes.Fields(b)
	for idx, word := range words {
		err := ti.addTerm(string(word), idx)
		if err != nil {
			return err
		}
	}

	return nil
}

// AppendBytes extract words from bytes and updates the TextIndex.
func (ti *TextIndex) AppendString(text string) error {
	words := strings.Fields(text)
	for idx, word := range words {
		err := ti.addTerm(word, idx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ti *TextIndex) reset() {
	ti.terms = make(map[string]*termVector)
}

func (ti *TextIndex) size() int {
	return len(ti.terms)
}

func (ti *TextIndex) setTermVector(term string, pos int, split bool) {
	tv, ok := ti.terms[term]
	if !ok {
		tv = newTermVector()
		ti.terms[term] = tv
		tv.split = split
	}

	tv.position[pos] = true
}

func (ti *TextIndex) addTerm(word string, pos int) error {
	if word == "" {
		return fmt.Errorf("cannot index empty string")
	}

	analyzedTerm := ti.a.Transform(word)

	if analyzedTerm == "" {
		return nil
	}

	ti.setTermVector(analyzedTerm, pos, false)

	if strings.Contains(analyzedTerm, "-") {
		for _, p := range strings.Split(analyzedTerm, "-") {
			ti.setTermVector(p, pos, true)
		}
	}

	return nil
}

type SearchHits struct {
	hits map[string]int
}

func newSearchHits() *SearchHits {
	return &SearchHits{
		hits: make(map[string]int),
	}
}

func (sh *SearchHits) appendTerm(term string, count int) {
	sh.hits[term] = count
}

func (sh *SearchHits) Total() int {
	var total int
	for _, hitCount := range sh.hits {
		total += hitCount
	}

	return total
}

func (sh *SearchHits) Hits() map[string]int {
	return sh.hits
}

func (ti *TextIndex) match(qt *search.QueryTerm, hits *SearchHits) bool {
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

func (ti *TextIndex) matchPhrase(qt *search.QueryTerm, hits *SearchHits) bool {
	var nextPositions []int

	phrasePositions := map[int]string{}

	words := strings.Fields(qt.Value)

	if len(words) == 1 {
		term, ok := ti.terms[qt.Value]
		if !ok {
			return false
		}

		hits.appendTerm(qt.Value, term.size())

		return true
	}

	var previousTerm string

	for idx, word := range words {
		term, ok := ti.terms[word]
		if !ok {
			return false
		}

		if idx != 0 {
			var wordMatch bool

			for _, pos := range nextPositions {
				// posMatch determines if this position can be followed
				var posMatch bool

				for _, nextPos := range search.ValidPhrasePosition(pos, qt.Slop) {
					_, ok := term.position[nextPos]
					if ok {
						phrasePositions[nextPos] = word
						posMatch = true
						wordMatch = true
					}
				}

				if posMatch {
					phrasePositions[pos] = previousTerm
				}
			}

			if !wordMatch {
				return false
			}
		}

		nextPositions = []int{}
		for pos := range term.position {
			nextPositions = append(nextPositions, pos)
		}

		previousTerm = word
	}

	matches := len(phrasePositions)

	if matches != 0 {
		phraseHits := sortAndCountPhrases(words, phrasePositions)

		for k, v := range phraseHits {
			hits.appendTerm(k, v)
		}
	}

	return matches != 0
}

func sortAndCountPhrases(words []string, phrases map[int]string) map[string]int {
	positions := []int{}
	phraseSize := len(words)
	phraseHits := map[string]int{}

	for k := range phrases {
		positions = append(positions, k)
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i] < positions[j]
	})

	phrase := []string{}

	for idx, p := range positions {
		phrase = append(phrase, phrases[p])
		if len(phrase) != phraseSize {
			if idx < len(positions) {
				continue
			}
		}

		currentPhrase := strings.Join(phrase, " ")
		phraseHits[currentPhrase]++

		phrase = []string{}
	}

	return phraseHits
}

func (ti *TextIndex) matchFuzzy(qt *search.QueryTerm, hits *SearchHits) bool {
	var hasMatch bool

	for k, v := range ti.terms {
		ok, _ := search.IsFuzzyMatch(k, qt.Value, float64(qt.Fuzzy), search.Levenshtein)
		if ok {
			hasMatch = true

			hits.appendTerm(k, v.size())
		}
	}

	return hasMatch
}

func (ti *TextIndex) matchWildcard(qt *search.QueryTerm, hits *SearchHits) bool {
	var matcher func(s, prefix string) bool

	switch {
	case qt.SuffixWildcard:
		matcher = strings.HasSuffix
	default:
		matcher = strings.HasPrefix
	}

	var hasMatch bool

	for k, v := range ti.terms {
		if matcher(k, qt.Value) {
			hasMatch = true

			hits.appendTerm(k, v.size())
		}
	}

	return hasMatch
}

func (ti *TextIndex) matchTerm(qt *search.QueryTerm, hits *SearchHits) bool {
	term, ok := ti.terms[qt.Value]
	if ok && qt.Prohibited {
		return false
	}

	if !ok && !qt.Prohibited {
		return false
	}

	var count int

	if ok {
		count = term.size()
	}

	hits.appendTerm(qt.Value, count)

	return true
}

func (ti *TextIndex) Search(query *search.QueryTerm) (*SearchHits, error) {
	hits := newSearchHits()
	err := ti.search(query, hits)

	return hits, err
}

func (ti *TextIndex) searchMustNot(query *search.QueryTerm, hits *SearchHits) error {
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

func (ti *TextIndex) searchMust(query *search.QueryTerm, hits *SearchHits) error {
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

func (ti *TextIndex) searchShould(query *search.QueryTerm, hits *SearchHits) error {
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
func (ti *TextIndex) search(query *search.QueryTerm, hits *SearchHits) error {
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

// func (dc *DescriptionCounter) writeTo(w io.Writer) error {

// jsonOutput, err := json.MarshalIndent(dc, "", " ")
// if err != nil {
// return fmt.Errorf("nable to marshall description to JSON; %w", err)
// }

// _, err = w.Write(jsonOutput)
// if err != nil {
// return fmt.Errorf("unable to write json; %w", err)
// }

// return nil
// }

// func (dc *DescriptionCounter) readFrom(r io.Reader) error {
// d := json.NewDecoder(r)
// return d.Decode(dc)
// }
