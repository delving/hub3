package mapping

import (
	"fmt"

	"github.com/OneOfOne/xxhash"
)

// TODO(kiivihal): add compile time safety for each mapping
// add init function that creates a hash for each mapping and compares it to a constant
// this should prevent changes to the mapping that are not reflected in the update.
// this is needed for all mappings that have strict fields.
const (
	v2MappingSha       = "332a55b29c0b868d"
	v2UpdateMappingSha = "9b20cec1c69f8b7b"
	fragmentMappingSha = "7607ca7737d17e4a"
)

var keys = map[string]string{
	v2MappingSha:       v2Mapping,
	v2UpdateMappingSha: v2MappingUpdate,
	fragmentMappingSha: fragmentMapping,
}

func ValidateMappings() (old, current string, ok bool) {
	return validate(keys)
}

func validate(keys map[string]string) (old, current string, ok bool) {
	for old, mapping := range keys {
		if current := hash(mapping); current != old {
			return old, current, false
		}
	}

	return "", "", true
}

func hash(input string) string {
	hash := xxhash.Checksum64([]byte(input))

	return fmt.Sprintf("%016x", hash)
}
