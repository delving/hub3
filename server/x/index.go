package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	c "github.com/delving/hub3/config"

	"github.com/delving/hub3/hub3/index"
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
	//stats := index.BulkIndexStatistics(bp)
	//render.PlainText(w, r, fmt.Sprintf("stats: %v", stats))
	return
}

// NewSingleFinalPathHostReverseProxy proxies QueryString of the request url to the target url
func NewSingleFinalPathHostReverseProxy(target *url.URL, relPath string) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + relPath
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		log.Printf("proxy request: %#v", req)
		log.Printf("proxy request: %#v", req.URL.String())
		log.Printf("proxy request: %#v", req.Body)
	}
	return &httputil.ReverseProxy{Director: director}
}
