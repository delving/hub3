package imageproxy

import (
	"fmt"
	"html"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) Routes(pattern string, router chi.Router) {
	if pattern == "" {
		pattern = s.proxyPrefix
	}

	router.Get(fmt.Sprintf("/%s/cachemetrics", pattern), s.rebuildCacheMetrics)
	router.Get(fmt.Sprintf("/%s/stats", pattern), s.handleCacheStats())
	router.Get(fmt.Sprintf("/%s/explore/*", pattern), s.handleExplore())

	proxyPrefix := fmt.Sprintf("/%s/{options}", pattern)
	router.Get(proxyPrefix+"/*", s.handleProxyRequest)
	router.Get(fmt.Sprintf("/%s/{cacheKey}", pattern), func(w http.ResponseWriter, r *http.Request) {
		cacheKey := html.EscapeString(chi.URLParam(r, "cacheKey"))

		sourceURL, err := decodeURL(cacheKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, html.EscapeString(sourceURL))
	})
}
