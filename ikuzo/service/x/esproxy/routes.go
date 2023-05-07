package esproxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Handle("/{index}/_search", s.esproxy)
	r.Handle("/{index}/{documentType}/_search", s.esproxy)

	if s.introspect {
		r.HandleFunc("/api/es/*", s.esproxy.SafeHTTP)
		r.Get("/api/es/indexes", func(w http.ResponseWriter, r *http.Request) {
			indices := s.es.Indices()
			indexes, err := indices.List()
			if err != nil {
				log.Print(err)
			}
			render.PlainText(w, r, fmt.Sprint("indexes:", indexes))
		})
	}
}
