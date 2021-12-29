package imageproxy

import (
	"net/http"

	"github.com/go-chi/render"
)

func (s *Service) rebuildCacheMetrics(w http.ResponseWriter, r *http.Request) {
	if err := s.buildCacheMetrics(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, s.cm)
}
