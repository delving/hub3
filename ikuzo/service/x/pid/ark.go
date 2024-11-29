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
