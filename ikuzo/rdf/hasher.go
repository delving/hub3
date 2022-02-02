package rdf

import (
	"fmt"

	"github.com/OneOfOne/xxhash"
)

type hasher string

func hash(input string) string {
	hash := xxhash.Checksum64([]byte(input))
	return fmt.Sprintf("%d", hash)
}

func getHash(input fmt.Stringer) hasher {
	return hasher(hash(input.String()))
}
