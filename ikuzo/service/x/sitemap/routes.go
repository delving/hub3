package sitemap

import "github.com/go-chi/chi/v5"

const configIDKey = "configID"

func (s *Service) Routes(pattern string, router chi.Router) {
	router.Get("/api/sitemap", s.handleListSitemapKeys)
	router.Get("/api/sitemap/{configID}", s.handleBaseSitemap)
	router.Post("/api/sitemap/{configID}", s.handleGenerateAll)
	router.Get("/api/sitemap/{configID}/sitemap.xml", s.handleBaseSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}", s.handleSitemap)
	router.Get("/api/sitemap/{configID}/{datasetID}/{page}", s.handleSitemap)
	router.Post("/api/sitemap/{configID}/{datasetID}", s.handleGenerateSitemap)
	router.Delete("/api/sitemap/{configID}/{datasetID}", s.handleDeleteSitemap)
}
