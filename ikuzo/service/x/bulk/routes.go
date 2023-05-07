package bulk

import "github.com/go-chi/chi/v5"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Post("/api/index/bulk", s.Handle)
	r.Post("/api/index/rdf", s.HandleRDF)
}
