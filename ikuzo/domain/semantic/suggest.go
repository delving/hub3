package semantic

import (
	"fmt"
	"sort"
	"strings"
)

// LevenshteinDistance computes the Levenshtein distance between two strings.
// The Levenshtein distance is the minimum number of single-character edits
// (insertions, deletions, or substitutions) required to change one string
// into the other.
func LevenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}

	la := len(a)
	lb := len(b)

	if la == 0 {
		return lb
	}

	if lb == 0 {
		return la
	}

	// Use two rows instead of full matrix for space efficiency.
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i

		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			curr[j] = min(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}

		prev, curr = curr, prev
	}

	return prev[lb]
}

// fieldSuggestion pairs a field name with its distance for sorting.
type fieldSuggestion struct {
	field    string
	distance int
}

// SuggestFields returns field names from knownFields that are within maxDistance
// Levenshtein distance of the unknown field name. Results are sorted by distance
// (closest first), then alphabetically for ties. Returns nil if no matches are
// found within the given distance.
func SuggestFields(unknown string, knownFields []string, maxDistance int) []string {
	var suggestions []fieldSuggestion

	for _, field := range knownFields {
		d := LevenshteinDistance(unknown, field)
		if d > 0 && d <= maxDistance {
			suggestions = append(suggestions, fieldSuggestion{field: field, distance: d})
		}
	}

	if len(suggestions) == 0 {
		return nil
	}

	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].distance != suggestions[j].distance {
			return suggestions[i].distance < suggestions[j].distance
		}

		return suggestions[i].field < suggestions[j].field
	})

	result := make([]string, len(suggestions))
	for i, s := range suggestions {
		result[i] = s.field
	}

	return result
}

// FormatSuggestion formats a human-readable suggestion string from a list of
// suggested field names. Returns an empty string if there are no suggestions,
// a "did you mean" message for a single suggestion, or a "one of" message
// for multiple suggestions.
func FormatSuggestion(unknown string, suggestions []string) string {
	switch len(suggestions) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("; did you mean '%s'?", suggestions[0])
	default:
		quoted := make([]string, len(suggestions))
		for i, s := range suggestions {
			quoted[i] = fmt.Sprintf("'%s'", s)
		}

		return fmt.Sprintf("; did you mean one of: %s?", strings.Join(quoted, ", "))
	}
}

// FilterableFields returns all filter field names from this resource configuration.
func (rc *ResourceConfig) FilterableFields() []string {
	fields := make([]string, len(rc.Filters))
	for i, f := range rc.Filters {
		fields[i] = f.Field
	}

	return fields
}

// FacetableFields returns all facet field names from this resource configuration.
func (rc *ResourceConfig) FacetableFields() []string {
	fields := make([]string, len(rc.Facets))
	for i, f := range rc.Facets {
		fields[i] = f.Field
	}

	return fields
}
