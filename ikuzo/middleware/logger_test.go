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
