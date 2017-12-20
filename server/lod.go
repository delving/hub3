package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/delving/rapid/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type LODResource struct{}

func (rs LODResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get(
		fmt.Sprintf("/%s", config.Config.LOD.RDF), func(w http.ResponseWriter, r *http.Request) {
			render.PlainText(w, r, `{"type": "rdf data endpoint"}`)
			return
		})
	r.Get(
		fmt.Sprintf("/%s", config.Config.LOD.HTML), func(w http.ResponseWriter, r *http.Request) {
			render.PlainText(w, r, `{"type": "rdf html endpoint"}`)
			return
		})
	r.Get(
		fmt.Sprintf("/%s", config.Config.LOD.Resource), func(w http.ResponseWriter, r *http.Request) {
			render.PlainText(w, r, `{"type": "rdf routing endpoint"}`)
			return
		})

	return r
}
