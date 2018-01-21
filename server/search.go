// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// SearchResource is a struct for the Search routes
type SearchResource struct{}

// Routes returns the chi.Router
func (rs SearchResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/v2", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	})

	return r
}
