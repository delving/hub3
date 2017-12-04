package server

import (
	"fmt"
	"net/http"

	. "bitbucket.org/delving/rapid/config"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/labstack/gommon/log"
)

// Start starts a graceful webserver process.
func Start() {
	// Setup
	r := chi.NewRouter()
	r.Use(mw.Logger)
	r.Use(mw.Recoverer)
	r.Use(mw.StripSlashes)

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
	r.Use(cors.Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You are rocking rapid!"))
	})

	log.Infof("Using port: %d", Config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), r)

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
