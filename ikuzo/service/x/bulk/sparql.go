package bulk

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
)

func diffAsSparqlUpdate(previous, current, graphURI, spec string) (updateQuery string, err error) {
	if graphURI == "" {
		return "", fmt.Errorf("graphURI cannot be empty: %q", graphURI)
	}
	added, removed := diffTriples(previous, current)

	updateQuery = generateSPARQLUpdateQuery(added, removed, graphURI, spec)

	return updateQuery, nil
}

func generateSPARQLUpdateQuery(added, removed []string, graphURI, spec string) string {
	var sb strings.Builder

	if len(added) == 0 && len(removed) == 0 {
		return ""
	}

	if len(removed) != 0 {
		// Construct the DELETE statement
		sb.WriteString("DELETE DATA {\n")
		sb.WriteString("GRAPH <")
		sb.WriteString(graphURI)
		sb.WriteString("> { ")
		for _, triple := range removed {
			sb.WriteString(triple + "\n")
		}
		sb.WriteString(" }\n")
		sb.WriteString("}\n")

	}

	if len(added) != 0 {
		// Construct the INSERT statement
		sb.WriteString("INSERT DATA {\n")
		sb.WriteString("GRAPH <")
		sb.WriteString(graphURI)
		sb.WriteString("> { ")
		if spec != "" {
			sb.WriteString(fmt.Sprintf("<%s> <http://schemas.delving.eu/nave/terms/datasetSpec> \"%s\" .\n", graphURI, spec))
		}
		for _, triple := range added {
			sb.WriteString(triple + "\n")
		}
		sb.WriteString(" }\n")
		sb.WriteString("}\n")
	}

	return sb.String()
}

func diffTriples(a, b string) (insertedLines, deletedLines []string) {
	return diffStrings(getSortedLines(a), getSortedLines(b))
}

func diffStrings(a, b []string) (added, removed []string) {
	mapA := make(map[string]bool)
	for _, val := range a {
		mapA[val] = true
	}
	mapB := make(map[string]bool)
	for _, val := range b {
		mapB[val] = true
	}

	for val := range mapB {
		if !mapA[val] {
			added = append(added, val)
		}
	}

	for val := range mapA {
		if !mapB[val] {
			removed = append(removed, val)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)

	return added, removed
}

func getSortedLines(s string) []string {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	return lines
}
