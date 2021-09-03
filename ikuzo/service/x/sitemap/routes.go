package sitemap

import "github.com/go-chi/chi"

const configIDKey = "configID"

func (s *Service) Routes(router chi.Router) {
	router.Get("/api/sitemap", s.handleBaseSitemap)
	// TODO(kiivihal): implement routes
	router.Get("/api/sitemap/{configID}", s.handleBaseSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}", s.handleBaseSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}/{page}", s.handleBaseSitemap)
}
