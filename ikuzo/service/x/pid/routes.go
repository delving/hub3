package pid

import (
	"log/slog"

	"github.com/delving/hub3/ikuzo/rdf/formats/html"
	"github.com/go-chi/chi/v5"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/ark:*", s.handleArk())

	htmlHandler := html.NewRDFHandler(slog.Default())
	htmlHandler.RegisterRoutes(r)
}
