package trs

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Service) Routes(pattern string, router chi.Router) {
	router.Get("/api/trs/info", s.handleInfo)
	// r.Get("/api/trs/{id}", s.handleGet)
	// r.Get("/api/trs/{id}/versions", s.handleGet)
	// r.Get("/api/trs/{id}/versions/{first}...{second}", s.handleGet)
	// r.Get("/api/trs/{id}", s.handleGet)
}

func (s *Service) handleInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello thomas")
}
