// +build dev

package assets

import (
	"net/http"
	"os"
	"path/filepath"
)

// FileSystem contains the embedded static resources from the "third_parties" directory.
var FileSystem http.FileSystem = getAbsolutePathToFileDir("third_party")

func getAbsolutePathToFileDir(relativePath string) http.Dir {
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, relativePath)
	return http.Dir(filesDir)
}
