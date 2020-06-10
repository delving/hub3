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

// nolint:gocritic
package ikuzo

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"

	"github.com/delving/hub3/ikuzo/logger"
	"github.com/matryer/is"
	"github.com/rs/zerolog/log"
)

func TestOptionSetPort(t *testing.T) {
	is := is.New(t)

	// default
	svr, err := newServer()
	is.NoErr(err)
	is.Equal(svr.port, 3000)

	// custom port
	customPort := 3001
	svr, err = newServer(
		SetPort(customPort),
	)
	is.NoErr(err)
	is.Equal(svr.port, customPort)

	// option with error
	errFunc := func(s *server) error { return errors.New("bad option") }
	svr, err = newServer(
		errFunc,
	)
	// err should not be nil
	is.True(err != nil)
	is.True(svr == nil)
}

func TestOptionSetLoggerConfig(t *testing.T) {
	is := is.New(t)

	l := logger.NewLogger(
		logger.Config{
			LogLevel:            logger.DebugLevel,
			EnableConsoleLogger: true,
		},
	)
	_, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	var buf bytes.Buffer
	log.Logger = log.Output(&buf)
	log.Debug().Int("answer", 42).Msg("debug message")
	is.True(strings.Contains(buf.String(), `"answer":42,`))
}

func TestSetDisableRequestLogger(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)
	is.True(svr.disableRequestLogger)
}

func TestSetMiddleware(t *testing.T) {
	is := is.New(t)
	svr, err := newServer()
	is.NoErr(err)

	// when no middleware is supplied the defaults are set
	is.True(len(svr.middleware) != 0)

	svr, err = newServer(
		SetMiddleware(mw.Heartbeat("/ping-test")),
	)
	is.NoErr(err)

	// only the set middleware is applied
	expectedNrMiddleware := 1
	is.True(len(svr.middleware) == expectedNrMiddleware)

	req, err := http.NewRequest("GET", "/ping-test", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), ".")
}

func TestSetRouters(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetRouters(
			func(r chi.Router) {
				r.Get("/router-test", func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, "router-test")
				})
			},
		),
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	req, err := http.NewRequest("GET", "/router-test", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), "router-test")
}
