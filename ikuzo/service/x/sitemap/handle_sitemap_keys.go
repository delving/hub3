package sitemap

import (
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/render"
)

// handleBaseSitemap is the entry point for all sub-sitemaps.
func (s *Service) handleListSitemapKeys(w http.ResponseWriter, r *http.Request) {
	org, ok := domain.GetOrganization(r)
	if !ok {
		http.Error(w, domain.ErrOrgNotFound.Error(), http.StatusNotFound)
		return
	}

	render.JSON(w, r, org.Config.Sitemaps)
}
