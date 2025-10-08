package contexts

import (
	"embed"
	"io/fs"
	"sort"
)

//go:embed */*.jsonld */*/*.jsonld
var embeddedContexts embed.FS

// Files returns an fs.FS with the embedded JSON-LD contexts.
func Files() fs.FS {
	return embeddedContexts
}

// List returns the available context filenames in sorted order.
func List() []string {
	names := []string{}

	fs.WalkDir(embeddedContexts, ".", func(path string, d fs.DirEntry, err error) error { //nolint:errcheck // best effort listing
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		names = append(names, path)
		return nil
	})

	sort.Strings(names)

	return names
}
