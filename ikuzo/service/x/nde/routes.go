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
	router.Get("/doc/datacatalog", defaultRedirect)
	router.Get("/doc/dataset/{spec}", defaultRedirect)
	router.Get("/doc/datacatalog/{cfgName}", s.HandleCatalog)
	router.Get("/doc/dataset/{cfgName}/{spec}", s.HandleDataset)
}

func (s *Service) lodRedirect(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/id/") {
		http.Error(w, "unknow lod prefix", http.StatusBadRequest)
		return
	}

	newPath := strings.Replace(r.URL.Path, "/id/", "/doc/", 1)
	http.Redirect(w, r, newPath, http.StatusFound)
}

func defaultRedirect(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		newPath := fmt.Sprintf("%s/default", r.URL.Path)
		http.Redirect(w, r, newPath, http.StatusFound)

		return
	}

	parts := strings.Split(r.URL.Path, "/")
	newpath := []string{
		parts[0],
		parts[1],
		"default",
		parts[2],
	}

	http.Redirect(w, r, path.Join(newpath...), http.StatusFound)
}
