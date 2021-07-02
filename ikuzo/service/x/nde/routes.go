package nde

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

func (s *Service) Routes(router chi.Router) {
	router.Get("/id/datacatalog", s.lodRedirect)
	router.Get("/id/dataset/{spec}", s.lodRedirect)
	router.Get("/doc/datacatalog", s.HandleCatalog)
	router.Get("/doc/dataset/{spec}", s.HandleDataset)
}

func (s *Service) lodRedirect(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/id/") {
		http.Error(w, "unknow lod prefix", http.StatusBadRequest)
		return
	}

	path := strings.Replace(r.URL.Path, "/id/", "/doc/", 1)
	http.Redirect(w, r, path, http.StatusFound)
}
