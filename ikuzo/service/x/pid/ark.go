package pid

import (
	"fmt"
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
