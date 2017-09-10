package server

import (
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

// Start starts a graceful webserver process.
func Start() {
	// Setup
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "You are rocking rapid!")
	})
	e.Server.Addr = ":3001"

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
