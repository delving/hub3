package oaipmh

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/api/oaipmh", s.handleVerb())
	r.Get("/api/oai-pmh", s.handleVerb())
}
