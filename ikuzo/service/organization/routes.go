package organization

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	if pattern == "" {
		pattern = "/organizations"
	}

	r.Route(pattern, func(router chi.Router) {
		router.Get("/", s.handleFilter)
		router.Get("/{id}", s.handleGet)
		router.Put("/", s.handlePut)
	})
}
