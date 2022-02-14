package sitemap

import "github.com/go-chi/chi"

const configIDKey = "configID"

func (s *Service) Routes(pattern string, router chi.Router) {
	router.Get("/api/sitemap", s.handleListSitemapKeys)
	router.Get("/api/sitemap/{configID}", s.handleBaseSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}", s.handleBaseSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}/{page}", s.handleBaseSitemap)
}
