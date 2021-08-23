package webapp

import (
	"io/fs"
	"net/http"
	"os"
	"time"
)

// NewStaticHandler receives an fs.FS (like embed.FS) and returns a http.HandlerFunc.
// The main purpose of this handler is to support embed.Fs static files with proper
// cache control. Without this wrapper the static content is never cached.
func NewStaticHandler(files fs.FS) http.HandlerFunc {
	staticFileHandler := http.FileServer(
		&StaticFSWrapper{
			FileSystem:   http.FS(files),
			FixedModTime: time.Now(),
		},
	)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Expires", "")
		w.Header().Set("Cache-Control", "max-age=259200")
		w.Header().Set("Pragma", "")
		staticFileHandler.ServeHTTP(w, r)
	}
}

type StaticFSWrapper struct {
	http.FileSystem
	FixedModTime time.Time
}

func (f *StaticFSWrapper) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)

	return &StaticFileWrapper{File: file, fixedModTime: f.FixedModTime}, err
}

type StaticFileWrapper struct {
	http.File
	fixedModTime time.Time
}

func (f *StaticFileWrapper) Stat() (os.FileInfo, error) {
	fileInfo, err := f.File.Stat()

	return &StaticFileInfoWrapper{FileInfo: fileInfo, fixedModTime: f.fixedModTime}, err
}

type StaticFileInfoWrapper struct {
	os.FileInfo
	fixedModTime time.Time
}

func (f *StaticFileInfoWrapper) ModTime() time.Time {
	return f.fixedModTime
}
