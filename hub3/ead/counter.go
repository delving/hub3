package ead

import (
	"bytes"
	"fmt"
	"strings"
)

type DescriptionCounter struct {
	counter map[string]int
}

func NewDescriptionCounter(text []byte) *DescriptionCounter {
	words := bytes.Fields(text)

	dc := &DescriptionCounter{
		counter: make(map[string]int, len(words)),
	}

	for _, word := range words {
		cleanWord := string(bytes.Trim(bytes.ToLower(word), ".,;:[]()?"))

		dc.counter[cleanWord]++
		if strings.Contains(cleanWord, "-") {
			for _, p := range strings.Split(cleanWord, "-") {
				dc.counter[p]++
			}
		}
	}

	return dc
}

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
		count, ok := dc.counter[word]
		if ok {
			seen += count
			hits[word] += count

		}

		if strings.HasSuffix(word, "*") {
			prefix := strings.TrimSuffix(word, "*")
			for k, count := range dc.counter {
				if strings.HasPrefix(k, prefix) {
					seen += count
					hits[k] += count
				}
			}
		}
	}

	return seen, hits
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
			//[]byte(fmt.Sprintf("+++%s+++", k)),
		)
		text = bytes.ReplaceAll(
			text,
			[]byte(fmt.Sprintf(" %s", k)),
			[]byte(fmt.Sprintf(` <em class=\"dchl\">%s</em>`, k)),
			//[]byte(fmt.Sprintf("+++%s+++", k)),
		)
	}
	return text, seen, hits, nil
}
