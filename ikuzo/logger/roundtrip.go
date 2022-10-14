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

package logger

import (
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// CustomLogger implements the estransport.Logger interface.
type CustomLogger struct {
	zerolog.Logger
}

// LogRoundTrip prints the information about request and response.
func (l *CustomLogger) LogRoundTrip(
	req *http.Request,
	res *http.Response,
	err error,
	start time.Time,
	dur time.Duration,
) error {
	var (
		e    *zerolog.Event
		nReq int64
		nRes int64
	)

	// Set error level.
	//
	e = l.setErrorLevel(res, err)

	// Count number of bytes in request and response.
	//
	if req != nil && req.Body != nil && req.Body != http.NoBody {
		nReq, _ = io.Copy(io.Discard, req.Body)
	}

	if res != nil && res.Body != nil && res.Body != http.NoBody {
		nRes, _ = io.Copy(io.Discard, res.Body)
	}

	// Log event.
	//
	e.Str("method", req.Method).
		Str("svc", "elasticsearch").
		Int("status", res.StatusCode).
		Dur("duration", dur).
		Int64("req_bytes", nReq).
		Int64("res_bytes", nRes).
		Msg(req.URL.String())

	return nil
}

func (l *CustomLogger) setErrorLevel(res *http.Response, err error) *zerolog.Event {
	var e *zerolog.Event

	switch {
	case err != nil:
		e = l.Error()
	case res != nil && res.StatusCode > 0 && res.StatusCode < 300:
		e = l.Info()
	case res != nil && res.StatusCode > 299 && res.StatusCode < 500:
		e = l.Warn()
	case res != nil && res.StatusCode > 499:
		e = l.Error()
	default:
		e = l.Error()
	}

	return e
}

// RequestBodyEnabled makes the client pass request body to logger
func (l *CustomLogger) RequestBodyEnabled() bool { return true }

// ResponseBodyEnabled makes the client pass response body to logger
func (l *CustomLogger) ResponseBodyEnabled() bool { return true }
