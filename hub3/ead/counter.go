package ead

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const (
	regex          = `(?i)(^|\W)(%s)(\W|$)`
	replace        = `$1<em class="dchl">$2</em>$3`
	partialRegex   = `(?i)(%s)`
	partialReplace = `<em class="dchl">$1</em>`
)

// DescriptionCounter holds a type-frequency list for the EAD description.
type DescriptionCounter struct {
	Counter     map[string]int             `json:"counter"`
	DataItemIdx map[string]map[uint64]bool `json:"dataItemIdx"`
}

// NewDescriptionCounter creates a type-frequency list for the description.
// The input consist of the EAD description stripped of all XML tags.
func NewDescriptionCounter() *DescriptionCounter {
	dc := &DescriptionCounter{
		Counter:     make(map[string]int),
		DataItemIdx: make(map[string]map[uint64]bool),
	}
	return dc
}

func (dc *DescriptionCounter) writeTo(w io.Writer) error {

	jsonOutput, err := json.MarshalIndent(dc, "", " ")
	if err != nil {
		return errors.Wrapf(err, "Unable to marshall description to JSON")
	}

	_, err = w.Write(jsonOutput)
	if err != nil {
		return errors.Wrapf(err, "Unable to write json")
	}

	return nil
}

func (dc *DescriptionCounter) readFrom(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(dc)
}

// AppendBytes extract words from bytes and updates the type-frequency counter.
func (dc *DescriptionCounter) AppendBytes(b []byte) error {
	words := bytes.Fields(b)
	for _, word := range words {
		err := dc.countWord(string(word), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// AppendString extract words from a string and updates the type-frequency counter.
func (dc *DescriptionCounter) AppendString(s string) error {
	words := strings.Fields(s)
	for _, word := range words {
		err := dc.countWord(string(word), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *DescriptionCounter) countWord(word string, order uint64) error {
	cleanWord := strings.Trim(strings.ToLower(word), ".,;:[]()?")

	dc.Counter[cleanWord]++
	dc.addDataItemIdx(cleanWord, order)

	if strings.Contains(cleanWord, "-") {
		for _, p := range strings.Split(cleanWord, "-") {
			dc.Counter[p]++
			dc.addDataItemIdx(p, order)
		}
	}
	return nil
}

func (dc *DescriptionCounter) addDataItemIdx(word string, order uint64) {
	if order < uint64(1) {
		return
	}

	key, ok := dc.DataItemIdx[word]
	if ok {
		key[order] = true
		return
	}
	dc.DataItemIdx[word] = map[uint64]bool{order: true}
}

func (dc *DescriptionCounter) add(item *DataItem) error {
	words := strings.Fields(item.Text)
	for _, word := range words {
		err := dc.countWord(string(word), item.Order)
		if err != nil {
			return err
		}
	}
	return nil
}

// CountForQuery takes a query string and returns a count and a result map.
// Internally, the type-frequency list is used to quickly count the number of hits,
// apply boolean search parameters, and expand wild-card queries.
//
//  This function should be used to get a quick hit count for a search query.
func (dc DescriptionCounter) CountForQuery(query string) (int, map[string]int) {
	words := strings.Fields(query)
	seen := 0
	hits := map[string]int{}
	for _, word := range words {
		switch word {
		case "AND", "OR", "NOT":
			continue
		}
		word = strings.Trim(strings.ToLower(word), `"()`)
		if strings.HasPrefix(word, "-") {
			continue
		}
		count, ok := dc.Counter[word]
		if ok {
			seen += count
			hits[word] += count

		}

		if strings.HasSuffix(word, "*") {
			prefix := strings.TrimSuffix(word, "*")
			for k, count := range dc.Counter {
				if strings.HasPrefix(k, prefix) {
					seen += count
					hits[k] += count
				}
			}
		}
	}

	return seen, hits
}

// GetDataItemIdx returns an sorted but deduplicated list of all the DataItem.Order
// keys for the search result.
func (dc *DescriptionCounter) GetDataItemIdx(keys map[string]int) []uint64 {
	if len(dc.DataItemIdx) == 0 {
		return []uint64{}
	}

	ids := make(map[uint64]bool)
	itemIdx := []uint64{}

	for k := range keys {
		indices, ok := dc.DataItemIdx[k]
		if !ok {
			continue
		}

		for idx := range indices {
			_, ok := ids[idx]
			if !ok {
				itemIdx = append(itemIdx, idx)
			}
		}
	}

	sort.Slice(itemIdx, func(i, j int) bool { return itemIdx[i] < itemIdx[j] })
	return itemIdx

}

// HighlightQuery surrounds all matches for the query in the text with emphasis tags.
func (dc DescriptionCounter) HighlightQuery(query string, text []byte) ([]byte, int, map[string]int, error) {
	seen, hits := dc.CountForQuery(query)
	if seen == 0 {
		return text, 0, hits, nil
	}
	for k := range hits {
		text = bytes.ReplaceAll(
			text,
			[]byte(fmt.Sprintf("\"%s", k)),
			[]byte(fmt.Sprintf(`"<em class=\"dchl\">%s</em>`, k)),
		)
		text = bytes.ReplaceAll(
			text,
			[]byte(fmt.Sprintf(" %s", k)),
			[]byte(fmt.Sprintf(` <em class=\"dchl\">%s</em>`, k)),
		)
	}
	return text, seen, hits, nil
}

type queryItem struct {
	text     string
	wildcard bool
}

// DescriptionQuery can be used to query and highlight matches in the ead.Description
type DescriptionQuery struct {
	items   []*queryItem
	Seen    int
	Hits    map[string]int
	Partial bool
	Filter  bool
	regex   map[string]*regexp.Regexp
}

func newQueryItem(word string) (*queryItem, bool) {
	switch word {
	case "AND", "OR", "NOT":
		return nil, false
	}
	word = strings.Trim(strings.ToLower(word), `"()`)
	if strings.HasPrefix(word, "-") {
		return nil, false
	}
	var hasSuffix bool
	if strings.HasSuffix(word, "*") {
		word = strings.TrimSuffix(word, "*")
		hasSuffix = true
	}
	queryItem := &queryItem{
		text:     word,
		wildcard: hasSuffix,
	}
	return queryItem, true
}

// NewDescriptionQuery returns a DescriptionQuery that can be used to filter
// and hightlight DataItems.
func NewDescriptionQuery(query string) *DescriptionQuery {
	words := strings.Fields(query)
	dq := &DescriptionQuery{
		Hits:   make(map[string]int),
		regex:  make(map[string]*regexp.Regexp),
		Filter: true,
	}
	for _, word := range words {
		queryItem, ok := newQueryItem(word)
		if ok {
			dq.items = append(dq.items, queryItem)
		}
	}

	return dq
}

// FilterMatches filters a []DataItem for query matches and highlights them.
func (dq *DescriptionQuery) FilterMatches(items []*DataItem) []*DataItem {
	matches := []*DataItem{}
	for _, item := range items {
		text, ok := dq.highlightQuery(item.Text)
		if !ok && dq.Filter {
			continue
		}
		item.Text = text
		matches = append(matches, item)
	}
	return matches
}

// HightlightSummary applied query highlights to the ead.Summary.
func (dq *DescriptionQuery) HightlightSummary(s Summary) Summary {
	if s.Profile != nil {
		s.Profile.Creation, _ = dq.highlightQuery(s.Profile.Creation)
		s.Profile.Language, _ = dq.highlightQuery(s.Profile.Language)
	}

	if s.File != nil {
		s.File.Author, _ = dq.highlightQuery(s.File.Author)
		s.File.Copyright, _ = dq.highlightQuery(s.File.Copyright)
		s.File.PublicationDate, _ = dq.highlightQuery(s.File.PublicationDate)
		s.File.Publisher, _ = dq.highlightQuery(s.File.Publisher)
		s.File.Title, _ = dq.highlightQuery(s.File.Title)
		var editions []string
		for _, e := range s.File.Edition {
			edition, _ := dq.highlightQuery(e)
			editions = append(editions, edition)
		}
		if len(editions) != 0 {
			s.File.Edition = editions
		}
	}

	if s.FindingAid != nil {
		s.FindingAid.AgencyCode, _ = dq.highlightQuery(s.FindingAid.AgencyCode)
		s.FindingAid.Country, _ = dq.highlightQuery(s.FindingAid.Country)
		s.FindingAid.ID, _ = dq.highlightQuery(s.FindingAid.ID)
		s.FindingAid.ShortTitle, _ = dq.highlightQuery(s.FindingAid.ShortTitle)
		var titles []string
		for _, t := range s.FindingAid.Title {
			title, _ := dq.highlightQuery(t)
			titles = append(titles, title)
		}
		if len(titles) != 0 {
			s.FindingAid.Title = titles
		}
	}

	if s.FindingAid != nil && s.FindingAid.UnitInfo != nil {
		unit := s.FindingAid.UnitInfo

		unit.ID, _ = dq.highlightQuery(unit.ID)
		unit.Language, _ = dq.highlightQuery(unit.Language)
		unit.DateBulk, _ = dq.highlightQuery(unit.DateBulk)
		unit.Files, _ = dq.highlightQuery(unit.Files)
		unit.Length, _ = dq.highlightQuery(unit.Length)
		unit.Material, _ = dq.highlightQuery(unit.Material)
		unit.Origin, _ = dq.highlightQuery(unit.Origin)
		unit.Physical, _ = dq.highlightQuery(unit.Physical)
		unit.PhysicalLocation, _ = dq.highlightQuery(unit.PhysicalLocation)
		unit.Repository, _ = dq.highlightQuery(unit.Repository)

		var dates []string
		for _, d := range unit.Date {
			date, _ := dq.highlightQuery(d)
			dates = append(dates, date)
		}
		if len(dates) != 0 {
			unit.Date = dates
		}

		var abstracts []string
		for _, a := range unit.Abstract {
			abstract, _ := dq.highlightQuery(a)
			abstracts = append(abstracts, abstract)
		}
		if len(abstracts) != 0 {
			unit.Abstract = abstracts
		}
		s.FindingAid.UnitInfo = unit
	}

	return s
}

func (dq *DescriptionQuery) regexInput() string {
	if dq.Partial {
		return partialRegex
	}
	return regex
}

func (dq *DescriptionQuery) regexOutput() string {
	if dq.Partial {
		return partialReplace
	}
	return replace
}

func (dq *DescriptionQuery) highlightQuery(text string) (string, bool) {
	found := map[string]bool{}
	for _, word := range strings.Fields(text) {
		partialMatch, ok := dq.match(word)
		if ok {
			found[partialMatch] = true
		}
	}

	for word := range found {
		r, ok := dq.regex[word]
		if !ok {
			r = regexp.MustCompile(fmt.Sprintf(dq.regexInput(), word))
			dq.regex[word] = r
		}
		text = r.ReplaceAllString(
			text,
			dq.regexOutput(),
		)
	}

	return text, len(found) != 0
}

func (dq *DescriptionQuery) match(word string) (string, bool) {
	word = strings.Trim(strings.ToLower(word), `"().,;`)
	var isMatch bool
	var matchWord string
	for _, q := range dq.items {
		switch q.wildcard {
		case true:
			isMatch = strings.HasPrefix(word, q.text)
			matchWord = word
		case false:
			isMatch = word == q.text
			if dq.Partial {
				isMatch = strings.Contains(word, q.text)
			}
			matchWord = q.text
		}
		if isMatch {
			break
		}
	}

	if isMatch {
		dq.Seen++
		dq.Hits[matchWord]++
	}

	return matchWord, isMatch
}
