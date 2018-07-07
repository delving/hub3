// +build dev

package assets

import (
	"net/http"
	"os"
	"path/filepath"
)

// Assets contains project assets.
var Assets http.FileSystem = getAbsolutePathToFileDir("public")

func getAbsolutePathToFileDir(relativePath string) http.Dir {
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, relativePath)
	return http.Dir(filesDir)
}
