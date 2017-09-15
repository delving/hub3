package hub3

import (
	"fmt"

	"github.com/OneOfOne/xxhash"
)

// CreateHash creates an xxhash-based hash of a string
func CreateHash(input string) string {
	hash := xxhash.Checksum64([]byte(input))
	return fmt.Sprintf("%016x", hash)
}
