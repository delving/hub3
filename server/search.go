package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type SearchResource struct{}

func (rs SearchResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/v2", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	})

	return r
}
