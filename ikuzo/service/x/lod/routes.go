package lod

import "github.com/go-chi/chi/v5"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/resolve", s.handleResolve)
}
