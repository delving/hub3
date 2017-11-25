package server

import (
	"fmt"
	"net/http"

	. "bitbucket.org/delving/rapid/config"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/acme/autocert"
)

// Start starts a graceful webserver process.
func Start() {
	// Setup
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Pre(middleware.RemoveTrailingSlash())

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "You are rocking rapid!")
	})

	// Admin group
	a := e.Group("/admin")
	a.GET("/routes", func(c echo.Context) error {
		return c.JSON(http.StatusOK, e.Routes())
	})
	a.GET("/config", func(c echo.Context) error {
		return c.JSON(
			http.StatusOK,
			Config,
		)
	})

	// WebResource & imageproxy configuration
	//if Config.

	// API configuration
	if Config.OAIPMH.Enabled {
		e.GET("/api/oai-pmh", oaiPmhEndpoint)
	}
	e.POST("/api/index/bulk", bulkAPI)

	// datasets
	e.GET("/api/datasets", listDataSets)
	e.POST("/api/datasets", createDataSet)
	e.GET("/api/datasets/:spec", getDataSet)
	//e.POST("/api/datasets/:spec", updateDataSet)
	//e.DELETE("/api/datasets/:spec", deleteDataset)

	// Start the server
	log.Infof("Using port: %d", Config.Port)
	e.Server.Addr = fmt.Sprintf(":%d", Config.Port)

	// Serve it like a boss
	e.Logger.Fatal(gracehttp.Serve(e.Server))
}

// StartTLS starts a webserver that uses Let's Encrypt to provide SSL cectificates
func StartTLS() {
	e := echo.New()
	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("rapid.delving.org")
	// Cache certificates
	e.AutoTLSManager.Cache = autocert.DirCache(".cert_cache")
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<h1>Welcome to Rapid!</h1>
			<h3>TLS certificates automatically installed from Let's Encrypt :)</h3>
		`)
	})
	e.Logger.Fatal(e.StartAutoTLS(":4443"))
}
