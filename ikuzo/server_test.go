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

// nolint:gocritic,scopelint,gomnd
package ikuzo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/delving/hub3/ikuzo/logger"
	"github.com/matryer/is"
)

func Test_server_handleIndex(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusFound)
}

func Test_server_handleHeartbeat(t *testing.T) {
	is := is.New(t)
	svr, err := newServer()
	is.NoErr(err)

	req, err := http.NewRequest("GET", "/ping", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), ".")
	is.Equal(w.Header().Get("Content-Type"), "text/plain")
}

func Test_server_handleStripSlashes(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	svr.router.Get("/ping2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ping2")
	})

	req, err := http.NewRequest("GET", "/ping2/", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), "ping2")
}

func Test_server_handle404(t *testing.T) {
	is := is.New(t)
	svr, err := NewServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	req, err := http.NewRequest("GET", "/404", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusNotFound)
	is.Equal(w.Body.String(), `{"status":"Not Found","code":404,"message":"page not found"}`+"\n")
}

func Test_server_handleMethodNotAllowed(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	req, err := http.NewRequest("HEAD", "/", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusMethodNotAllowed)
	is.Equal(w.Body.String(), `{"status":"Method Not Allowed","code":405,"message":"method HEAD is not allowed"}`+"\n")
}

func Test_server_respondReturnsEncodingError(t *testing.T) {
	is := is.New(t)
	svr, err := newServer(
		SetDisableRequestLogger(),
	)
	is.NoErr(err)

	w := httptest.NewRecorder()
	c := make(chan int)
	svr.respond(w, nil, c, http.StatusInternalServerError)
	is.Equal(w.Code, http.StatusInternalServerError)
	is.Equal(
		w.Body.String(),
		`{"status":"Internal Server Error","code":500,"message":"json: unsupported type: chan int"}`+"\n",
	)
}

func Test_server_decode(t *testing.T) {
	is := is.New(t)

	type response struct {
		Message string `json:"message"`
	}

	type args struct {
		body    []byte
		message string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"empty body returns error",
			args{
				body: []byte(""),
			},
			true,
		},
		{
			"simple message",
			args{
				body:    []byte(`{"message": "ikuzo"}`),
				message: "ikuzo",
			},
			false,
		},
		{
			"malformed JSON",
			args{
				body: []byte(`{"message": "ikuzo"`),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{}
			var v response
			req, err := http.NewRequest("get", "/echo", bytes.NewReader(tt.args.body))
			is.NoErr(err)
			if err := s.decode(req, &v); (err != nil) != tt.wantErr {
				t.Errorf("server.decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			is.Equal(v.Message, tt.args.message)
		})
	}
}

func Test_server_Shutdown(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer

	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	// simple shutdown
	server := &http.Server{Handler: svr}
	err = svr.shutdown(server)
	is.NoErr(err)

	// wait for workerpool to be done
	server = &http.Server{Handler: svr}
	svr.gracefulTimeout = 1 * time.Nanosecond
	svr.workers.wg.Add(1)

	errChan := make(chan error, 1)

	go func() {
		errChan <- svr.shutdown(server)
	}()

	svr.workers.wg.Done()

	for err := range errChan {
		is.NoErr(err)
		return
	}
}

func Test_server_ListenAndServe(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer

	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		errChan <- svr.ListenAndServe()
	}()

	// stop the server
	svr.cancelFunc()
	<-doneChan

	for err := range errChan {
		is.True(err != nil)
		is.Equal(err, context.Canceled)

		return
	}
}

func Test_server_listenAndServeWithError(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer

	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)

	is.NoErr(err)

	testError := errors.New("test error")

	err = svr.listenAndServe(testError)
	is.True(err != nil)
	is.Equal(err, testError)
}

func Test_server_listenAndServeWithSignal(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer

	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	err = svr.listenAndServe(syscall.SIGTERM)
	is.NoErr(err)
}

func Test_server_requestLogger(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	svr.router.Get("/ping2", func(w http.ResponseWriter, r *http.Request) {
		svr.requestLogger(r).Debug().Msg("extra message")
		fmt.Fprint(w, "ping2")
	})

	req, err := http.NewRequest("GET", "/ping2", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), "ping2")

	is.True(strings.Contains(buf.String(), "extra message"))
}

func Test_server_recoverer(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	l := logger.NewLogger(
		logger.Config{Output: &buf},
	)

	svr, err := newServer(
		SetLogger(&l),
	)
	is.NoErr(err)

	svr.router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		svr.requestLogger(r).Debug().Msg("extra message")
		panic("panicing here ")
	})

	req, err := http.NewRequest("GET", "/panic", nil)
	is.NoErr(err)

	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusInternalServerError)

	is.True(strings.Contains(w.Body.String(), "error logged with request_id:"))

	is.True(strings.Contains(buf.String(), "Recover from Panic"))
	is.True(strings.Contains(buf.String(), `"level":"panic"`))

	// svr should still be running
	req, err = http.NewRequest("GET", "/", nil)
	is.NoErr(err)

	w = httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusFound)
}
