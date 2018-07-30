// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/index"
	"github.com/delving/rapid-saas/hub3/models"
	"github.com/delving/rapid-saas/server/assets"
	"github.com/phyber/negroni-gzip/gzip"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thoas/stats"
	"github.com/urfave/negroni"
	negroniprometheus "github.com/zbindenren/negroni-prometheus"
)

// ErrorMessage is a placeholder for disabled endpoints
type ErrorMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Start starts a graceful webserver process.
func Start(buildInfo *c.BuildVersionInfo) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Negroni middleware manager
	n := negroni.New()

	// recovery
	recovery := negroni.NewRecovery()
	recovery.Formatter = &negroni.HTMLPanicFormatter{}
	n.Use(recovery)

	// logger
	l := negroni.NewLogger()
	n.Use(l)

	// compress the responses
	n.Use(gzip.Gzip(gzip.DefaultCompression))

	// stats middleware
	s := stats.New()
	n.Use(s)

	// stats prometheus
	m := negroniprometheus.NewMiddleware("rapid")
	n.Use(m)

	// configure CORS, see https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
	cors := cors.New(cors.Options{
		//AllowedOrigins: []string{"*"},
		AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	// setup fileserver for public directory
	n.Use(negroni.NewStatic(assets.Assets))

	// Setup Router
	r := chi.NewRouter()
	r.Use(cors.Handler)
	r.Use(mw.StripSlashes)
	r.Use(mw.Heartbeat("/ping"))

	// stats page
	r.Get("/api/stats/http", func(w http.ResponseWriter, r *http.Request) {
		stats := s.Data()
		render.JSON(w, r, stats)
		return
	})

	r.Handle("/metrics", prometheus.Handler())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "You are rocking rapid!")
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", buildInfo)
		render.JSON(w, r, buildInfo)
		return
	})

	// stats dashboard
	r.Get("/api/stats/bySearchLabel", searchLabelStats)
	//r.Get("/api/stats/bySearchLabel/{:label}", searchLabelStatsValues)
	r.Get("/api/stats/byPredicate", predicateStats)
	//r.Get("/api/stats/byPredicate/{:label}", searchLabelStatsValues)

	// stastic serving on vfsgen files
	r.Get("/api/search/v2/_docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/_docs", http.StatusSeeOther)
	})
	r.Get("/api/_docs", func(w http.ResponseWriter, r *http.Request) {
		serveHTML(w, r, "api-docs.html")
		return
	})
	r.Get("/explore/sparql", func(w http.ResponseWriter, r *http.Request) {
		serveHTML(w, r, "explore/sparql.html")
		return
	})
	r.Get("/explore/fragments", func(w http.ResponseWriter, r *http.Request) {
		serveHTML(w, r, "explore/ldf.html")
		return
	})

	// gaf ZVT
	//r.Get("/gaf/search-alt/*", func(w http.ResponseWriter, r *http.Request) {
	//http.ServeFile(w, r, "./public/gaf/index.html")
	//return
	//})
	//r.Get("/gaf/search-alt", func(w http.ResponseWriter, r *http.Request) {
	//http.ServeFile(w, r, "./public/gaf/index.html")
	//return
	//})
	//r.Get("/gaf/search-cache/*", func(w http.ResponseWriter, r *http.Request) {
	//http.ServeFile(w, r, "./public/gaf/index-cache.html")
	//return
	//})
	//r.Get("/gaf/search-cache", func(w http.ResponseWriter, r *http.Request) {
	//http.ServeFile(w, r, "./public/gaf/index-cache.html")
	//return
	//})

	// WebResource & imageproxy configuration
	proxyPrefix := fmt.Sprintf("/%s/*", c.Config.ImageProxy.ProxyPrefix)
	r.With(StripPrefix).Get(proxyPrefix, serveProxyImage)

	if c.Config.WebResource.Enabled {
		r.Mount("/thumbnail", ThumbnailResource{}.Routes())
		r.Mount("/deepzoom", DeepZoomResource{}.Routes())
		r.Mount("/explore", ExploreResource{}.Routes())
		r.Mount("/api/webresource", WebResourceAPIResource{}.Routes())
		// legacy route
		r.Get("/iip/deepzoom/mnt/tib/tiles/{orgId}/{spec}/{localId}.tif.dzi", renderDeepZoom)
		// render cached directories
		FileServer(r, "/webresource", getAbsolutePathToFileDir(c.Config.WebResource.CacheResourceDir))
	}
	//r.Get("/deepzoom", func(w http.ResponseWriter, r *http.Request) {
	//cmd := exec.Command("vips", "dzsave", "/tmp/webresource/dev-org-id/test2/source/123.jpg", "/tmp/123")
	//stdoutStderr, err := cmd.Output()
	//if err != nil {
	//log.Println("Something went wrong")
	//fmt.Printf("%s\n", stdoutStderr)
	//log.Println(err)
	//}
	//w.Write([]byte("zoomed"))
	//})

	// API configuration
	if c.Config.OAIPMH.Enabled {
		r.Get("/api/oai-pmh", oaiPmhEndpoint)
	}

	// Narthex endpoint
	r.Post("/api/rdf/bulk", bulkAPI)
	r.Get("/api/bulk/sync", bulkSyncList)
	r.Post("/api/bulk/sync", bulkSyncStart)
	r.Get("/api/bulk/sync/{id}", bulkSyncProgress)
	r.Delete("/api/bulk/sync/{id}", bulkSyncCancel)
	// TODO remove later
	r.Post("/api/index/bulk", bulkAPI)
	r.Post("/api/index/fuzzed", generateFuzzed)

	// CSV upload endpoint
	r.Post("/api/rdf/csv", csvUpload)
	r.Delete("/api/rdf/csv", csvDelete)

	// SKOS sync endpoint
	r.Get("/api/rdf/skos", skosSync)
	r.Post("/api/rdf/skos", skosUpload)

	// RDF upload endpoint
	r.Post("/api/rdf/source", rdfUpload)

	// EAD endpoint
	r.Post("/api/ead", eadUpload)

	// Tree reconstruction endpoint
	r.Get("/api/tree/{spec}", treeList)
	r.Get("/api/tree/{spec}/{nodeID}", treeList)
	r.Get("/api/tree/{spec}/stats", treeStats)

	// Search endpoint
	r.Mount("/api/search", SearchResource{}.Routes())

	// Sparql endpoint
	r.Mount("/sparql", SparqlResource{}.Routes())

	// RDF indexing endpoint
	r.Mount("/api/es", IndexResource{}.Routes())

	// datasets
	r.Get("/api/datasets", listDataSets)
	r.Get("/api/datasets/histogram", listDataSetHistogram)
	r.Post("/api/datasets", createDataSet)
	r.Get("/api/datasets/{spec}", getDataSet)
	r.Get("/api/datasets/{spec}/stats", getDataSetStats)
	// later change to update dataset
	r.Post("/api/datasets/{spec}", createDataSet)
	r.Delete("/api/datasets/{spec}", deleteDataset)

	// fragments
	r.Get("/api/fragments", listFragments)
	r.Get("/fragments/{spec}", listFragments)
	r.Get("/fragments", listFragments)

	// namespaces
	r.Get("/api/namespaces", listNameSpaces)

	// LoD routingendpoint
	r.Mount("/", LODResource{}.Routes())

	// introspection
	if c.Config.DevMode {
		r.Mount("/introspect", IntrospectionRouter(r))
		r.Mount("/debug", mw.Profiler())
	}

	if c.Config.Cache.Enabled {
		r.Mount("/api/cache", CacheResource{}.Routes())
		r.Handle(fmt.Sprintf("%s/*", c.Config.Cache.APIPrefix), cacheHandler())
	}

	n.UseHandler(r)
	log.Printf("Using port: %d", c.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Config.Port), n)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: graceful shutdown with flushing and closing connections.
	//// Start the server
	//log.Infof("Using port: %d", c.Config.Port)
	//e.Server.Addr = fmt.Sprintf(":%d", c.Config.Port)

	//// Serve it like a boss
	//e.Logger.Fatal(gracehttp.Serve(e.Server))

}

// StripPrefix removes the leading '/' from the HTTP path
func StripPrefix(h http.Handler) http.Handler {
	proxyPrefix := fmt.Sprintf("/%s", c.Config.ImageProxy.ProxyPrefix)
	return http.StripPrefix(proxyPrefix, h)
}

// IntrospectionRouter gives access to the configuration at runtime when debug mode is enabled.
func IntrospectionRouter(chiRouter chi.Router) http.Handler {
	r := chi.NewRouter()
	r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, c.Config)
	})
	r.Get("/routes", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(docgen.JSONRoutesDoc(chiRouter)))
		return
	})
	r.Delete("/reset", resetAll)
	return r
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

func getAbsolutePathToFileDir(relativePath string) http.Dir {
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, relativePath)
	return http.Dir(filesDir)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func serveHTML(w http.ResponseWriter, r *http.Request, filePath string) error {
	file, err := assets.Assets.Open(filePath)
	if err != nil {
		log.Printf("Unable to open file %s: %v", filePath, err)
		render.Status(r, http.StatusNotFound)
		render.PlainText(w, r, "")
		return err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Unable to read file %s: %v", filePath, err)
		render.Status(r, http.StatusNotFound)
		render.PlainText(w, r, "")
		return err
	}
	render.HTML(w, r, string(body))
	return nil
}
