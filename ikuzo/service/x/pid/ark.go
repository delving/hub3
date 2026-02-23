package pid

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func isUUIDWithoutHyphens(s string) bool {
	if len(s) != 32 {
		return false
	}

	// Check if the string contains only hexadecimal characters
	re := regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	return re.MatchString(s)
}

func addHyphensToUUID(s string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		s[0:8],
		s[8:12],
		s[12:16],
		s[16:20],
		s[20:32])
}

func validateARK(ark string) string {
	ark = strings.TrimPrefix(ark, "/")
	parts := strings.Split(ark, "/")
	uuid := parts[len(parts)-1]

	if !isUUIDWithoutHyphens(uuid) {
		return ark
	}
	uuid = addHyphensToUUID(uuid)
	parts[len(parts)-1] = uuid

	return strings.Join(parts, "/")
}

// isBareNAAN checks if a URL path is a bare NAAN path without the ark: prefix.
// This happens when arks.org strips the ark: prefix before redirecting to our resolver.
// Example: /54386/02a4d96cdacc2d686f58f0a66daec47d → true
func isBareNAAN(path string) bool {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return false
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return false
	}

	naan := parts[0]
	if len(naan) != 5 {
		return false
	}

	for _, c := range naan {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// handleBareNAAN returns a handler that redirects bare NAAN paths to the ark: format.
// arks.org strips the ark: prefix when redirecting to our resolver, so
// /54386/02a4d96c... becomes /ark:54386/02a4d96c...
func handleBareNAAN() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ark:"+strings.TrimPrefix(r.URL.Path, "/"), http.StatusFound)
	}
}

// normalizeARK toggles between ark:/ and ark: formats for retry.
// This handles the difference between modern ARKs (ark:123/123) and
// N2T resolver format which adds a slash (ark:/123/123).
//
// Examples:
//   - ark:/123/123 → ark:123/123
//   - ark:123/123 → ark:/123/123
func normalizeARK(ark string) string {
	if strings.HasPrefix(ark, "ark:/") {
		// Remove slash: ark:/123/123 → ark:123/123
		return strings.Replace(ark, "ark:/", "ark:", 1)
	}
	if strings.HasPrefix(ark, "ark:") && !strings.HasPrefix(ark, "ark:/") {
		// Add slash: ark:123/123 → ark:/123/123
		return strings.Replace(ark, "ark:", "ark:/", 1)
	}
	return ark
}
