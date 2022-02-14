package es

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	// stats dashboard
	r.Get("/api/stats/bySearchLabel", searchLabelStats)
	// r.Get("/api/stats/bySearchLabel/{:label}", searchLabelStatsValues)
	r.Get("/api/stats/byPredicate", predicateStats)
	// r.Get("/api/stats/byPredicate/{:label}", searchLabelStatsValues)
}
