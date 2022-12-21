package nde

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi"
)

func (s *Service) Routes(router chi.Router) {
	router.Get("/id/datacatalog/*", s.lodRedirect)
	router.Get("/id/dataset/*", s.lodRedirect)
	router.Get("/id/datacatalog/", s.lodRedirect)
	router.Get("/doc/datacatalog", s.defaultRedirect)
	router.Get("/doc/dataset/{spec}", s.defaultRedirect)
	router.Get("/doc/datacatalog/{cfgName}", s.HandleCatalog)
	router.Get("/doc/dataset/{cfgName}/{spec}", s.HandleDataset)
}

func (s *Service) lodRedirect(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/id/") {
		http.Error(w, "unknown lod prefix", http.StatusBadRequest)
		return
	}

	newPath := strings.Replace(r.URL.Path, "/id/", "/doc/", 1)
	http.Redirect(w, r, newPath, http.StatusFound)
}

func (s *Service) defaultRedirect(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		newPath := fmt.Sprintf("%s/%s", r.URL.Path, s.defaultCfg.URLPrefix)
		http.Redirect(w, r, newPath, http.StatusFound)

		return
	}

	parts := strings.Split(r.URL.Path, "/")
	newpath := []string{
		parts[0],
		parts[1],
		s.defaultCfg.URLPrefix,
		parts[2],
	}

	http.Redirect(w, r, path.Join(newpath...), http.StatusFound)
}
