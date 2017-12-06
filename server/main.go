package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "bitbucket.org/delving/rapid/config"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// Start starts a graceful webserver process.
func Start() {

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
	n.Use(negroni.NewStatic(http.Dir(Config.HTTP.StaticDir)))

	// Setup Router
	r := chi.NewRouter()
	r.Use(mw.StripSlashes)
	n.UseHandler(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You are rocking rapid!"))
	})

	// example fileserver
	//FileServer(r, "/docs", getAbsolutePathToFileDir("public"))

	log.Printf("Using port: %d", Config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), n)

	//// Admin group
	//a := e.Group("/admin")
	//a.GET("/routes", func(c echo.Context) error {
	//return c.JSON(http.StatusOK, e.Routes())
	//})
	//a.GET("/config", func(c echo.Context) error {
	//return c.JSON(
	//http.StatusOK,
	//Config,
	//)
	//})

	// WebResource & imageproxy configuration
	//if Config.

	// API configuration
	//if Config.OAIPMH.Enabled {
	//e.GET("/api/oai-pmh", oaiPmhEndpoint)
	//}
	//e.POST("/api/index/bulk", bulkAPI)

	//// datasets
	//e.GET("/api/datasets", listDataSets)
	//e.POST("/api/datasets", createDataSet)
	//e.GET("/api/datasets/:spec", getDataSet)
	////e.POST("/api/datasets/:spec", updateDataSet)
	////e.DELETE("/api/datasets/:spec", deleteDataset)

	//// Start the server
	//log.Infof("Using port: %d", Config.Port)
	//e.Server.Addr = fmt.Sprintf(":%d", Config.Port)

	//// Serve it like a boss
	//e.Logger.Fatal(gracehttp.Serve(e.Server))
}

func ConfigRouter() http.Handler {
	r := chi.NewRouter()
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
