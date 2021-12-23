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

	router.Get(fmt.Sprintf("/%s/explore/*", pattern), s.handleExplore())

	proxyPrefix := fmt.Sprintf("/%s/{options}", pattern)
	router.Get(proxyPrefix+"/*", s.handleProxyRequest)
	router.Get(fmt.Sprintf("/%s/{cacheKey}", pattern), func(w http.ResponseWriter, r *http.Request) {
		cacheKey := chi.URLParam(r, "cacheKey")

		sourceURL, err := decodeURL(cacheKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, sourceURL)
	})
}
