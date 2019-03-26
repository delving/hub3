package handlers

import (
	"net/http"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
)

// IntrospectionRouter gives access to the configuration at runtime when debug mode is enabled.
func RegisterIntrospection(r chi.Router) {
	r.Get("/introspect/config", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, c.Config)
	})
	r.Get("/introspect/routes", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(docgen.JSONRoutesDoc(r)))
		return
	})
	r.Delete("/introspect/reset", resetAll)
	r.Mount("/debug", mw.Profiler())
}

func resetAll(w http.ResponseWriter, r *http.Request) {
	// reset elasticsearch
	err := index.IndexReset("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// reset Key Value Store
	models.ResetStorm()
	return
}
