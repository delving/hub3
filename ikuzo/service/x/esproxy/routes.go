package esproxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (s *Service) Routes(pattern string, r chi.Router) {
	r.Get("/", s.handleRoot)
	r.Handle("/{index}/_search", s.esproxy)
	r.Handle("/{index}/{documentType}/_search", s.esproxy)

	if s.introspect {
		r.HandleFunc("/api/es/*", s.esproxy.SafeHTTP)
		r.Get("/api/es/indexes", func(w http.ResponseWriter, r *http.Request) {
			indices := s.es.Indices()
			indexes, err := indices.List()
			if err != nil {
				log.Print(err)
			}
			render.PlainText(w, r, fmt.Sprint("indexes:", indexes))
		})
	}
}

// esInfo is what an Elasticsearch 7.17 node answers on GET /.
var esInfo = map[string]any{
	"name":         "hub3",
	"cluster_name": "hub3",
	"version": map[string]any{
		"number":       "7.17.0",
		"build_flavor": "default",
	},
	"tagline": "You Know, for Search",
}

// handleRoot lets Elasticsearch clients use the proxy. The official clients
// (elasticsearch-py since 7.14, for one) GET / before their first request and
// refuse a server that does not identify as Elasticsearch, whatever answers
// behind the proxy. They ask for JSON; browsers keep ikuzo's redirect.
func (s *Service) handleRoot(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept"), "json") {
		http.Redirect(w, r, "/version", http.StatusFound)
		return
	}

	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	render.JSON(w, r, esInfo)
}
