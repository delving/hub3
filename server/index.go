package server

import (
	"fmt"
	"net/http"
	"net/url"

	c "github.com/delving/rapid/config"

	"github.com/delving/rapid/hub3/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

type IndexResource struct{}

func (rs IndexResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/stats", rs.Get) // GET
	r.Get("/indexes", func(w http.ResponseWriter, r *http.Request) {
		indexes, err := index.ListIndexes()
		if err != nil {
			log.Print(err)
		}
		render.PlainText(w, r, fmt.Sprint("indexes:", indexes))
		return
	})
	// Anything we don't do in Go, we pass to the old platform
	es, _ := url.Parse(c.Config.ElasticSearch.Urls[0])
	es.Path = fmt.Sprintf("/%s/", c.Config.ElasticSearch.IndexName)
	if c.Config.ElasticSearch.Proxy {
		r.Handle("/_search", NewSingleFinalPathHostReverseProxy(es, "_search"))
		r.Handle("/_mapping", NewSingleFinalPathHostReverseProxy(es, "_mapping"))
	}

	return r
}

// Get returns JSON formatted statistics for the BulkProcessor
func (rs IndexResource) Get(w http.ResponseWriter, r *http.Request) {
	stats := index.BulkIndexStatistics(bp)
	render.PlainText(w, r, fmt.Sprintf("stats: ", stats))
	return
}
