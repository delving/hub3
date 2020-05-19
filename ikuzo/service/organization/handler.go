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
	// b, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// http.Error(w, err.Error(), http.StatusInternalServerError)
	// return
	// }

	// render.JSON(w, r, string(b))
	var org domain.Organization

	err := json.NewDecoder(r.Body).Decode(&org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, org)
	// err = s.Put(r.Context(), org)
	// if err != nil {
	// http.Error(w, err.Error(), http.StatusInternalServerError)
	// return
	// }

	// w.WriteHeader(http.StatusNoContent)
}
