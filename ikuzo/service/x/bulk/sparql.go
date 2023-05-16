package bulk

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"github.com/delving/hub3/hub3/fragments"
)

var (
	datasetSpecFmt  string = "http://schemas.delving.eu/nave/terms/datasetSpec"
	contentHashFmt  string = "http://schemas.delving.eu/nave/terms/contentHash"
	specRevisionFmt string = "http://schemas.delving.eu/nave/terms/specRevision"
)

type DiffConfig struct {
	su              *fragments.SparqlUpdate
	previousTriples string
	previousHash    string
}

func diffAsSparqlUpdate(cfg *DiffConfig) (updateQuery string, err error) {
	if cfg.su.NamedGraphURI == "" {
		return "", fmt.Errorf("graphURI cannot be empty: %q", cfg.su.NamedGraphURI)
	}
	added, removed := diffTriples(cfg.previousTriples, cfg.su.Triples)

	updateQuery = generateSPARQLUpdateQuery(added, removed, cfg)

	return updateQuery, nil
}

func generateSPARQLUpdateQuery(added, removed []string, cfg *DiffConfig) string {
	var sb strings.Builder

	if len(added) == 0 && len(removed) == 0 {
		return ""
	}
	// <{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}" .
	// <{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/contentHash> "{{.RDFHash}}" .
	// <{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/specRevision> "{{.SpecRevision}}"^^<http://www.w3.org/2001/XMLSchema#integer> .

	if len(removed) != 0 {
		// Construct the DELETE statement
		sb.WriteString("DELETE DATA {\n")
		sb.WriteString("GRAPH <")
		sb.WriteString(cfg.su.NamedGraphURI)
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
		sb.WriteString(cfg.su.NamedGraphURI)
		sb.WriteString("> { ")
		// TODO: generate from DiffConfig
		// if su.Spec != "" {
		// 	sb.WriteString(fmt.Sprintf("<%s> <http://schemas.delving.eu/nave/terms/datasetSpec> \"%s\" .\n", su.NamedGraphURI, su.Spec))
		// 	sb.WriteString(fmt.Sprintf("<%s> <http://schemas.delving.eu/nave/terms/contentHash> \"%s\" .\n", su.NamedGraphURI, su.RDFHash))
		// }
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
