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

//nolint:gocritic,scopelint,gochecknoglobals
package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestRequestLogger(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	logger := log.Output(&buf)

	r := chi.NewRouter()

	requestLogger := RequestLogger(&logger)
	r.Use(requestLogger)
	r.Get("/test-ping", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "test-ping")
		is.NoErr(err)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test-ping", nil)
	is.NoErr(err)

	r.ServeHTTP(w, req)
	is.True(strings.Contains(buf.String(), `"url":"/test-ping",`))
}

func Test_LogParamsAsDict(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		name        string
		queryParams string
		want        string
	}{
		{
			name:        "single param",
			queryParams: "q=query",
			want:        `{"params":{"q":["query"]}}`,
		},
		{
			name:        "double param",
			queryParams: "filter=f1&filter=f2&v=",
			want:        `{"params":{"filter":["f1","f2"]}}`,
		},
		{
			name:        "empty params",
			queryParams: "",
			want:        `{"params":{}}`,
		},
		{
			name:        "param with whitespace",
			queryParams: "q=",
			want:        `{"params":{}}`,
		},
		{
			name:        "params with qf as object",
			queryParams: "qf.dateRange=ead-rdf_periodDesc:1189~1863",
			want:        `{"params":{"qf.dateRange":["ead-rdf_periodDesc:1189~1863"]}}`,
		},
		{
			name:        "params with qf as string needs to be stored as object",
			queryParams: "qf=meta.tags:suriname",
			want:        `{"params":{"qf.meta.tags":["suriname"]}}`,
		},
		{
			name:        "multi qf params",
			queryParams: "qf=meta.tags:suriname&qf=meta.tags:ead",
			want:        `{"params":{"qf.meta.tags":["suriname","ead"]}}`,
		},
		{
			name:        "simple string. store as qf.value object",
			queryParams: "qf=thisCanHappen",
			want:        `{"params":{"qf.value":["thisCanHappen"]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/echo?%s", tt.queryParams),
				nil,
			)
			is.NoErr(err)

			var buf bytes.Buffer
			logger := zerolog.New(&buf).With().Logger()

			logger.Log().Dict("params", LogParamsAsDict(req.URL.Query())).Send()

			if got := strings.TrimSpace(buf.String()); !cmp.Equal(got, tt.want) {
				t.Errorf("paramsToDict() = %v, want %v", got, tt.want)
			}
		})
	}
}

// package level variable to eliminate compiler optimisations
var event *zerolog.Event

func BenchmarkLogParamsAsDictUnsorted(b *testing.B) {
	// run the Fib function b.N times
	var e *zerolog.Event

	req, _ := http.NewRequest("GET", "/echo?q=123&limit=10", nil)

	for n := 0; n < b.N; n++ {
		e = LogParamsAsDict(req.URL.Query())
	}

	event = e
}

func Test_newLineChecker(t *testing.T) {
	is := is.New(t)

	ignoredPath := "/version1"

	t.Run("no paths", func(t *testing.T) {
		lc := newLineChecker()
		is.True(lc.allowLine(http.StatusNotFound, ignoredPath))
	})

	t.Run("disable all", func(t *testing.T) {
		lc := newLineChecker("*")
		is.True(lc.disableAll404)
		is.True(!lc.allowLine(http.StatusNotFound, ignoredPath))
		is.True(lc.enabled)
	})

	t.Run("disable single path", func(t *testing.T) {
		lc := newLineChecker(ignoredPath)
		is.True(!lc.disableAll404)
		is.True(lc.enabled)

		_, ok := lc.lookUps[ignoredPath]
		is.True(ok)
		is.True(lc.allowLine(http.StatusNotFound, "/version10"))
		is.True(!lc.allowLine(http.StatusNotFound, ignoredPath))
	})

	t.Run("paths with wildcards", func(t *testing.T) {
		lc := newLineChecker("/version*")
		is.True(!lc.disableAll404)
		is.True(lc.enabled)

		is.True(!lc.allowLine(http.StatusNotFound, "/version10"))
		is.True(!lc.allowLine(http.StatusNotFound, ignoredPath))
	})
}
