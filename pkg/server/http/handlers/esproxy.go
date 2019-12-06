package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	c "github.com/delving/hub3/config"

	"github.com/delving/hub3/hub3/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterElasticSearchProxy(router chi.Router) {
	r := chi.NewRouter()

	r.Get("/stats", BulkStats) // GET
	r.Get("/indexes", func(w http.ResponseWriter, r *http.Request) {
		indexes, err := index.ListIndexes()
		if err != nil {
			log.Print(err)
		}
		render.PlainText(w, r, fmt.Sprint("indexes:", indexes))
		return
	})

	if c.Config.ElasticSearch.Proxy {
		r.HandleFunc("/*", esProxy)
	}

	router.Mount("/api/es", r)
}

func esProxy(w http.ResponseWriter, r *http.Request) {
	// parse the url
	url, _ := url.Parse(c.Config.ElasticSearch.Urls[0])

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// strip prefix from path
	r.URL.Path = strings.TrimPrefix(r.URL.EscapedPath(), "/api/es")

	switch {
	case r.Method != "GET":
		http.Error(w, fmt.Sprintf("method %s is not allowed on esProxy", r.Method), http.StatusBadRequest)
		return
	case r.URL.Path == "/":
		// root is allowed to provide version
	case !strings.HasPrefix(r.URL.EscapedPath(), "/_cat"):
		http.Error(w, fmt.Sprintf("path %s is not allowed on esProxy", r.URL.EscapedPath()), http.StatusBadRequest)
		return
	}

	// Update the headers to allow for SSL redirection
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)

}

// Get returns JSON formatted statistics for the BulkProcessor
func BulkStats(w http.ResponseWriter, r *http.Request) {
	stats := index.BulkIndexStatistics(BulkProcessor())
	render.PlainText(w, r, fmt.Sprintf("stats: %v", stats))
	return
}
