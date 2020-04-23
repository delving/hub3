//nolint:gocritic,scopelint
package ikuzo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func Test_server_fileServer(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
	}{
		{
			"valid path",
			"/docs",
			http.StatusOK,
		},
		{
			"invalid path with URL parameters",
			"/{docs}",
			http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			svr, err := newServer(
				SetDisableRequestLogger(),
			)
			is.NoErr(err)

			workDir, _ := os.Getwd()

			filesDir := filepath.Join(workDir, "../docs")
			svr.fileServer(tt.path, http.Dir(filesDir))

			req, err := http.NewRequest("GET", "/docs/ikuzo/raml/api.raml", nil)
			is.NoErr(err)

			w := httptest.NewRecorder()
			svr.ServeHTTP(w, req)
			is.Equal(w.Code, tt.statusCode)
		})
	}
}
