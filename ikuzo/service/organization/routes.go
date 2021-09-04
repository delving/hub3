package organization

import "github.com/go-chi/chi"

func (s *Service) Routes() chi.Router {
	router := chi.NewRouter()

	router.Get("/", s.handleFilter)
	router.Get("/{id}", s.handleGet)
	router.Put("/", s.handlePut)

	return router
}
