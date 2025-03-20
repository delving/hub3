package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
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

// calculateChecksum generates a SHA-256 checksum of the document
// after normalizing field order.
func calculateChecksum(b []byte) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", fmt.Errorf("unable to unmarshal data: %w", err)
	}

	// Remove existing checksum if present
	delete(data, "_checksum")

	// Sort keys for consistent checksum
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create ordered map
	ordered := make(map[string]interface{})
	for _, k := range keys {
		ordered[k] = data[k]
	}

	bytes, _ := json.Marshal(ordered)
	return checksum(bytes), nil
}

func checksum(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}
