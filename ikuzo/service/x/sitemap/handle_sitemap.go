package sitemap

import (
	"errors"
	"net/http"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi/v5"
)

var ErrConfigNotFound = errors.New("sitemap configuration not found")

// handleBaseSitemap is the entry point for all sub-sitemaps.
func (s *Service) handleBaseSitemap(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	sm, err := s.sitemapRoot(r.Context(), cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	_, err = sm.WriteTo(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) sitemapConfig(r *http.Request) (Config, error) {
	var cfg Config

	org, ok := domain.GetOrganization(r)
	if !ok {
		return cfg, domain.ErrOrgNotFound
	}

	configID := chi.URLParam(r, configIDKey)
	for _, sitemap := range org.Config.Sitemaps {
		if strings.EqualFold(sitemap.ID, configID) {
			sitemap.OrgID = org.RawID()
			return sitemap, nil
		}
	}

	return cfg, ErrConfigNotFound
}
