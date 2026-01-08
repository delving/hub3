package bulk

import "github.com/go-chi/chi/v5"

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Post("/api/index/bulk", s.Handle)
	r.Post("/api/index/rdf", s.HandleRDF)
	r.Get("/api/bulk/debug/sparql", s.handleBulkDebug)
	r.Get("/api/bulk/recdefs", s.handleRecDefs)
	// Admin endpoints to trigger orphan cleanup
	// Global cleanup: scans all datasets and cleans up those with orphans
	r.Post("/api/admin/cleanup-orphans", s.handleGlobalCleanupOrphans)
	// Single dataset cleanup: cleans up orphans for a specific dataset
	r.Post("/api/admin/cleanup-orphans/{orgID}/{datasetID}", s.handleCleanupOrphans)
}
