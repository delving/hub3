package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/go-chi/chi/v5"
)

func ApiV3Docs(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, RapiDocPage("/api/v3/swagger.yaml", "API Documentation"))
}

func Render(w http.ResponseWriter, r *http.Request, comp templ.Component) error {
	return comp.Render(r.Context(), w)
}

func APIV3Swagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SwaggerJSON)
}

func APIV3SwaggerYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(SwaggerYAML)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		err := ApiV3Docs(w, r)
		if err != nil {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode:    http.StatusBadRequest,
				Message:       "unable to render docs",
				PreventBubble: false,
			})
		}
	})
	r.Get("/swagger.json", APIV3Swagger)
	r.Get("/swagger.yaml", APIV3SwaggerYAML)

	return r
}
