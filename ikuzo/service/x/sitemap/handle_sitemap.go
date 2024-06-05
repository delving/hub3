package sitemap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/snabb/sitemap"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

var ErrConfigNotFound = errors.New("sitemap configuration not found")

func (s *Service) handleGenerateAll(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		s.renderError(w, r, err, http.StatusNotFound)
		return
	}

	datasets, err := s.store.Datasets(r.Context(), cfg)
	if err != nil {
		s.renderError(w, r, err, http.StatusInternalServerError)
		return
	}

	go func() {
		for _, ds := range datasets {
			_, err := s.generateSitemaps(cfg, ds.ID)
			if err != nil {
				slog.Error("unable to generate sitemap", "spec", ds.ID, "orgID", cfg.OrgID, "error", err)
			}
		}

		slog.Info("finished generating all sitemaps", "size", len(datasets))
	}()

	render.PlainText(w, r, fmt.Sprintf("started generating %d sitemaps", len(datasets)))
}

// handleBaseSitemap is the entry point for all sub-sitemaps.
func (s *Service) handleBaseSitemap(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		s.renderError(w, r, err, http.StatusNotFound)
		return
	}

	sm, err := s.sitemapRoot(r.Context(), cfg)
	if err != nil {
		s.renderError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	_, err = sm.WriteTo(w)
	if err != nil {
		s.renderError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Service) handleSitemap(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		s.renderError(w, r, err, http.StatusNotFound)
		return
	}

	spec := chi.URLParam(r, "datasetID")
	page := chi.URLParam(r, "page")
	if page == "" {
		page = "1"
	}

	if cfg.IsExcludedSpec(spec) {
		s.renderError(w, r, fmt.Errorf("no sitemap available for this dataset: %s", spec), http.StatusNotFound)
		return
	}

	pageNr, err := strconv.Atoi(page)
	if err != nil {
		slog.Warn("unable to parse page returning 1", "input", page)
		pageNr = 1
	}

	sitemapPath := cfg.Path(spec, pageNr)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("sitemap-%s-%d.xml", spec, pageNr))
	http.ServeFile(w, r, sitemapPath)
}

func (s *Service) generateSitemaps(cfg domain.SitemapConfig, spec string) (int, error) {
	var (
		rowsSeen    int
		sitemapPage int
		totalSeen   int
	)

	sm := sitemap.New()
	sitemapPage = 1

	cb := func(loc Location) error {
		rowsSeen++
		totalSeen++

		if rowsSeen >= maxSitemapURLs {

			path := cfg.Path(spec, sitemapPage)
			slog.Info("writing sitemap path", "path", path)
			sitemapErr := s.writeSitemap(sm, path)
			if sitemapErr != nil {
				return sitemapErr
			}

			sm = sitemap.New()
			rowsSeen = 1
			sitemapPage++
		}

		sm.Add(&sitemap.URL{
			Loc:     cfg.URL(loc.ID),
			LastMod: loc.LastMod,
		})
		return nil
	}

	err := s.store.Locations(context.Background(), spec, cfg, cb)
	if err != nil {
		return 0, err
	}

	path := cfg.Path(spec, sitemapPage)
	sitemapErr := s.writeSitemap(sm, path)
	slog.Info("writing sitemap path", "path", path)
	if sitemapErr != nil {
		return 0, sitemapErr
	}

	return totalSeen, nil
}

func (s *Service) handleGenerateSitemap(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		s.renderError(w, r, err, http.StatusNotFound)
		return
	}
	spec := chi.URLParam(r, "datasetID")

	if removeErr := s.removeSitemaps(cfg, spec); removeErr != nil {
		s.renderError(w, r, removeErr, http.StatusInternalServerError)
		return
	}

	processed, err := s.generateSitemaps(cfg, spec)
	if err != nil {
		s.renderError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, "generated sitemaps for %s with %d records", spec, processed)
}

func (s *Service) handleDeleteSitemap(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.sitemapConfig(r)
	if err != nil {
		s.renderError(w, r, err, http.StatusNotFound)
		return
	}

	spec := chi.URLParam(r, "datasetID")
	if err := s.removeSitemaps(cfg, spec); err != nil {
		s.renderError(w, r, err, http.StatusInternalServerError)
		return
	}

	slog.Info("removed cached sitemaps", "datasetID", spec)
	fmt.Fprintf(w, "removed sitemaps for dataset %s", spec)
}

func (s *Service) sitemapConfig(r *http.Request) (cfg domain.SitemapConfig, err error) {
	org, ok := domain.GetOrganization(r)
	if !ok {
		return cfg, domain.ErrOrgNotFound
	}

	configID := chi.URLParam(r, configIDKey)
	for idx, sitemap := range org.Config.Sitemaps {
		if strings.EqualFold(sitemap.ID, configID) {
			sitemap.OrgID = org.RawID()
			org.Config.Sitemaps[idx] = sitemap
			return sitemap, nil
		}
	}

	return cfg, ErrConfigNotFound
}

func (s *Service) removeSitemaps(cfg domain.SitemapConfig, spec string) error {
	path := filepath.Join(cfg.DataPath, cfg.OrgID, spec)
	return os.RemoveAll(path)
}

func (s *Service) writeSitemap(sm *sitemap.Sitemap, fname string) error {
	path := filepath.Dir(fname)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	_, err = sm.WriteTo(f)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) renderError(w http.ResponseWriter, r *http.Request, err error, code int) {
	slog.Error("error handling request", "error", err, "url", r.URL.String(), "status", code)
	http.Error(w, err.Error(), code)
}

func sitemapName(spec, page string) string {
	return fmt.Sprintf("%s_%s.xml", spec, page)
}
