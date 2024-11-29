package swagger

import "github.com/go-chi/chi/v5"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/api-docs/swagger.json", s.HandleJSON())
	r.Get("/api-docs", s.HandleHTML())
}
