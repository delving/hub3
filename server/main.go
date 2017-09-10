package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/delving/rapid/config"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/acme/autocert"
)

// Start starts a graceful webserver process.
func Start(config config.RawConfig) {
	// Setup
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "You are rocking rapid!")
	})
	log.Infof("Using port: %d", config.Port)
	e.Server.Addr = fmt.Sprintf(":%d", config.Port)

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
