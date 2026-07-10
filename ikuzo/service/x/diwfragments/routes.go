package diwfragments

import "github.com/go-chi/chi/v5"

// Routes mounts the frozen v1 contract. The GET shapes must never change
// incompatibly; new needs mean new fields or /api/ui/v2.
func (s *Service) Routes(pattern string, router chi.Router) {
	router.Get("/api/ui/v1/{collection}/listing", s.handleListing)
	router.Get("/api/ui/v1/{collection}/item/{id}", s.handleItem)
	router.Post("/api/ui/v1/{collection}/fragments", s.handleBulkPut)
}
