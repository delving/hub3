package esproxy

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Handle("/{index}/_search", s.esproxy)
	r.Handle("/{index}/{documentType}/_search", s.esproxy)
}
