package imageproxy

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	_ "embed"

	"github.com/go-chi/chi"
)

//go:embed explore.html
var explore string

func (s *Service) handleExplore() http.HandlerFunc {
	ts, err := template.New("explore").Parse(explore)
	if err != nil {
		log.Fatalf("unable to build explore template: %s", err)
	}

	type templateData struct {
		Req *Request
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := chi.URLParam(r, "*")

		rawQuery := r.URL.RawQuery

		req, err := NewRequest(
			targetURL,
			SetRawQueryString(rawQuery),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		td := templateData{Req: req}

		buf := new(bytes.Buffer)
		err = ts.ExecuteTemplate(buf, "explore", td)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := buf.WriteTo(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
