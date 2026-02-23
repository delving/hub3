package pid

import (
	"log/slog"

	"github.com/delving/hub3/ikuzo/rdf/formats/html"
	"github.com/go-chi/chi/v5"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/ark:*", s.handleArk())
	r.Get("/ark:/*", s.handleArk()) // Support legacy ark:/ format

	// Catch bare NAAN paths where arks.org strips the ark: prefix.
	// Redirects e.g. /54386/02a4d96c... → /ark:54386/02a4d96c...
	r.Get("/{naan:[0-9]{5}}/*", handleBareNAAN())

	htmlHandler := html.NewRDFHandler(slog.Default())
	htmlHandler.RegisterRoutes(r)
}
