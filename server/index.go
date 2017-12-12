package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/delving/rapid/hub3"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

type IndexResource struct{}

func (rs IndexResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/stats", rs.Get) // GET
	r.Get("/indexes", func(w http.ResponseWriter, r *http.Request) {
		indexes, err := hub3.ListIndexes()
		if err != nil {
			log.Print(err)
		}
		render.PlainText(w, r, fmt.Sprint("indexes:", indexes))
		return
	})

	return r
}

// Get returns JSON formatted statistics for the BulkProcessor
func (rs IndexResource) Get(w http.ResponseWriter, r *http.Request) {
	stats := hub3.IndexStatistics(bp)
	render.PlainText(w, r, fmt.Sprintf("stats: ", stats))
	return
}
