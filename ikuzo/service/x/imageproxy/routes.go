package imageproxy

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Service) Routes(pattern string, router chi.Router) {
	if pattern == "" {
		pattern = s.proxyPrefix
	}

	proxyPrefix := fmt.Sprintf("/%s/{options}", s.proxyPrefix)
	router.Get(proxyPrefix+"/*", s.proxyImage)
	router.Get(fmt.Sprintf("/%s/{cacheKey}", s.proxyPrefix), func(w http.ResponseWriter, r *http.Request) {
		cacheKey := chi.URLParam(r, "cacheKey")

		sourceUrl, err := decodeURL(cacheKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		fmt.Fprintf(w, sourceUrl)
	})
}
