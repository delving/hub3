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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	c "bitbucket.org/delving/rapid/config"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

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

	// configure CORS, see https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	n.Use(cors)

	// setup fileserver for public directory
	n.Use(negroni.NewStatic(http.Dir(c.Config.HTTP.StaticDir)))

	// Setup Router
	r := chi.NewRouter()
	r.Use(mw.StripSlashes)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("You are rocking rapid!"))
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", buildInfo)
		render.JSON(w, r, buildInfo)
		return
	})

	// static fileserver
	FileServer(r, "/static", getAbsolutePathToFileDir("public"))

	// WebResource & imageproxy configuration
	proxyPrefix := fmt.Sprintf("/%s/*", c.Config.ImageProxy.ProxyPrefix)
	r.With(StripPrefix).Get(proxyPrefix, serveProxyImage)

	if c.Config.WebResource.Enabled {
		r.Mount("/thumbnail", ThumbnailResource{}.Routes())
		r.Mount("/deepzoom", DeepZoomResource{}.Routes())
		r.Mount("/explore", ExploreResource{}.Routes())
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

	// introspection
	if c.Config.DevMode {
		r.Mount("/introspect", IntrospectionRouter(r))
	}

	// API configuration
	if c.Config.OAIPMH.Enabled {
		r.Get("/api/oai-pmh", oaiPmhEndpoint)
	}

	// Narthex endpoint
	r.Post("/api/index/bulk", bulkAPI)

	// Search endpoint
	r.Mount("/api/search", SearchResource{}.Routes())

	// Sparql endpoint
	r.Mount("/sparql", SparqlResource{}.Routes())

	// RDF indexing endpoint
	r.Mount("/api/es", IndexResource{}.Routes())

	// datasets
	r.Get("/api/datasets", listDataSets)
	r.Post("/api/datasets", createDataSet)
	r.Get("/api/datasets/{spec}", getDataSet)
	r.Get("/api/datasets/{spec}/stats", getDataSetStats)
	// later change to update dataset
	r.Post("/api/datasets/{spec}", createDataSet)
	r.Delete("/api/datasets/{spec}", deleteDataset)

	r.Get("/api/fragments", listFragments)

	// namespaces
	r.Get("/api/namespaces", listNameSpaces)

	// LoD routingendpoint
	r.Mount("/", LODResource{}.Routes())

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
	// todo add routes
	//r.Get("routes", func(w http.ResponseWriter, req *http.Request) {
	//render.JSON(w, req, docgen.JSONRoutesDoc(chiRouter.Routes))
	//})
	return r
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
