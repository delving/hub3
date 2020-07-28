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

package organization

import (
	"encoding/json"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func (s *Service) Routes() chi.Router {
	router := chi.NewRouter()

	router.Get("/", s.handleFilter)
	router.Get("/{id}", s.handleGet)
	router.Put("/", s.handlePut)

	return router
}

func (s *Service) handleFilter(w http.ResponseWriter, r *http.Request) {
	orgs, err := s.Filter(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, orgs)
}

func (s *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	org, err := s.Get(r.Context(), domain.OrganizationID(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, org)
}

func (s *Service) handlePut(w http.ResponseWriter, r *http.Request) {
	var org domain.Organization

	err := json.NewDecoder(r.Body).Decode(&org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, org)
}
