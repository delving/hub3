package harvest

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/oai/!open_oai.OAIHandler", s.ServeHTTP)
	r.Post("/oai/harvest-now", s.HarvestNow)
}
