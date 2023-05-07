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

package ikuzo

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	mw "github.com/go-chi/chi/v5/middleware"
)

// DefaultMiddleware are the default functions applied to the global router.
func DefaultMiddleware() []func(http.Handler) http.Handler {
	handlers := []func(http.Handler) http.Handler{
		mw.Heartbeat("/ping"),
		mw.StripSlashes,
	}

	return handlers
}

// routes are the default routes for ikuzo.
// These can be overwritten using SetRouters.
//
// All handlers should have lazy initialization, so when they are not called
// no connections should be initialized.
func (s *server) routes() {
	s.router.Get("/", s.handleIndex())
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func (s *server) fileServer(path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		log.Error().
			Str("path", path).
			Msg("FileServer does not permit URL parameters. (fileserver: disabled)")

		return
	}

	fsServer := http.FileServer(root)

	fs := http.StripPrefix(path, fsServer)

	path += "*"

	s.router.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
