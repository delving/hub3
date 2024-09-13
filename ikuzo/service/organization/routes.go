package organization

import (
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	if pattern == "" {
		pattern = "/organizations"
	}

	if strings.EqualFold("HUB3_ORG_DEBUG", "true") {
		r.Route(pattern, func(router chi.Router) {
			router.Get("/", s.handleFilter)
			router.Get("/{id}", s.handleGet)
			router.Put("/", s.handlePut)
		})
	}
}
