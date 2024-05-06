package index

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/rdf"
)

// GraphDocType is the docType for indexing
const GraphDocType = "graph"

func appendUnique(s []string, vals ...string) []string {
	for _, v := range vals {
		if slices.Contains(s, v) {
			continue
		}
		s = append(s, v)
	}

	return s
}

func getPredicate(searchLabel string) (string, error) {
	prefix, label, found := strings.Cut(searchLabel, "_")
	if !found {
		return "", fmt.Errorf("invalid search label must have '_'; %s", searchLabel)
	}

	ns, err := rdf.DefaultNamespaceManager.GetWithPrefix(prefix)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", ns.URI, label), nil
}

// NowInMillis returns time.Now() in miliseconds
func NowInMillis() int64 {
	return time.Now().UTC().UnixMilli()
}
