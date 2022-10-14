// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//nolint:gocritic,scopelint
package ikuzo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/delving/hub3/ikuzo/logger"
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

	l := logger.NewLogger(logger.Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			svr, err := newServer(
				SetDisableRequestLogger(),
				SetLogger(&l),
			)
			is.NoErr(err)

			workDir, _ := os.Getwd()

			filesDir := filepath.Join(workDir, "./webapp/testdata/docs")
			svr.fileServer(tt.path, http.Dir(filesDir))

			req, err := http.NewRequest("GET", "/docs/doc.md", nil)
			is.NoErr(err)

			w := httptest.NewRecorder()
			svr.ServeHTTP(w, req)
			is.Equal(w.Code, tt.statusCode)
		})
	}
}
