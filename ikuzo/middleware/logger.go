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

package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func (lc *lineChecker) generateLookupPaths(paths ...string) {
	if len(paths) == 0 {
		return
	}

	lc.enabled = true

	for _, path := range paths {
		if path == "*" {
			lc.disableAll404 = true
			continue
		}

		if strings.HasSuffix(path, "*") {
			path = strings.TrimSuffix(path, "*")
			lc.lookUps[path] = true

			continue
		}

		lc.lookUps[path] = false
	}

	lc.enabled = true
}

type lineChecker struct {
	lookUps       map[string]bool
	disableAll404 bool
	enabled       bool
}

func newLineChecker(paths ...string) lineChecker {
	lc := lineChecker{
		lookUps: map[string]bool{},
	}

	lc.generateLookupPaths(paths...)

	return lc
}

func (lc lineChecker) allowLine(status int, requestPath string) bool {
	if !lc.enabled {
		return true
	}

	if status != http.StatusNotFound {
		return true
	}

	if lc.disableAll404 {
		return false
	}

	for path, wildcard := range lc.lookUps {
		matcher := strings.EqualFold
		if wildcard {
			matcher = strings.HasPrefix
		}

		if matcher(requestPath, path) {
			return false
		}
	}

	return true
}

// RequestLogger creates a middleware chain for request logging
func RequestLogger(log *zerolog.Logger, disable404Paths ...string) func(next http.Handler) http.Handler {
	c := alice.New()

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(*log))

	lc := newLineChecker(disable404Paths...)

	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		if !lc.allowLine(status, r.URL.String()) {
			return
		}

		l := hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Dict("params", LogParamsAsDict(r.URL.Query()))

		setChiURLParams(l, r, "spec", "dataset_id")
		setChiURLParams(l, r, "datasetID", "dataset_id")
		setChiURLParams(l, r, "inventoryID", "inventory_id")

		addHeader(l, r, "cache_url", "Cache-Url")
		addHeader(l, r, "cache_type", "Cache-Type")

		l.Msg("")
	}))

	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	c = c.Append(orgIDHandler("org_id"))

	// Here is your final handler
	return c.Then
}

func setChiURLParams(l *zerolog.Event, r *http.Request, paramKey, fieldKey string) {
	if val := chi.URLParamFromCtx(r.Context(), paramKey); val != "" {
		l.Str(fieldKey, val)
	}
}

func addHeader(l *zerolog.Event, r *http.Request, key, header string) {
	if r.Header.Get(header) != "" {
		l.Str(key, r.Header.Get(header))
	}
}

// orgIDHandler adds the request's domain.OrganizationID as a field to the context's logger
// using fieldKey as field key.
func orgIDHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if orgID := domain.GetOrganizationID(r); orgID != "" {
				l := zerolog.Ctx(r.Context())
				l.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str(fieldKey, string(orgID))
				})
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LogParamsAsDict logs the request params as a zerolog.Dict.
func LogParamsAsDict(params url.Values) *zerolog.Event {
	dict := zerolog.Dict()

	for key, values := range params {
		arr := zerolog.Arr()

		var nonEmpty bool

		alteredKey := ""
		for _, v := range values {
			if v != "" {
				if key == "qf" {
					parts := strings.Split(v, ":")
					if len(parts) == 2 {
						alteredKey = fmt.Sprintf("%s.%s", key, parts[0])
						v = parts[1]
					}
					if len(parts) == 1 {
						alteredKey = "qf.value"
					}
				}
				arr = arr.Str(v)

				if !nonEmpty {
					nonEmpty = true
				}
			}
		}

		if nonEmpty {
			if alteredKey != "" {
				key = alteredKey
			}
			dict = dict.Array(key, arr)
		}
	}

	return dict
}
