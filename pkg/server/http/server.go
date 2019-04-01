package http

import (
	"fmt"
	"log"
	"net/http"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/pkg/server/http/assets"
	"github.com/delving/hub3/pkg/server/http/handlers"
	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
)

type server struct {
	n         *negroni.Negroni
	r         chi.Router
	buildInfo *c.BuildVersionInfo
}

type Server interface {
	Flush() error
	ListenAndServe() error
}

func NewServer() (Server, error) {
	s := &server{}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s.n = negroniWithDefaults()
	s.r = chiWithDefaults()

	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "You are rocking hub3!")
	})

	s.r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", s.buildInfo)
		render.JSON(w, r, s.buildInfo)
		return
	})

	// introspection
	if c.Config.DevMode {
		handlers.RegisterIntrospection(s.r)
	}

	s.n.UseHandler(s.r)

	return s, nil
}

func chiWithDefaults() chi.Router {
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

	// Setup Router
	r := chi.NewRouter()
	r.Use(cors.Handler)
	r.Use(mw.StripSlashes)
	r.Use(mw.Heartbeat("/ping"))

	return r

}

func negroniWithDefaults() *negroni.Negroni {

	n := negroni.New()

	// recovery
	recovery := negroni.NewRecovery()
	recovery.Formatter = &negroni.HTMLPanicFormatter{}
	n.Use(recovery)

	// logger
	l := negroni.NewLogger()
	l.SetFormat("{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}} {{.Path}} {{.Request.URL.RawQuery}}\n")
	n.Use(l)

	// compress the responses
	n.Use(gzip.Gzip(gzip.DefaultCompression))

	// setup fileserver for third_party directory
	n.Use(negroni.NewStatic(assets.FileSystem))

	return n
}

func (s server) ListenAndServe() error {
	log.Printf("Using port: %d", c.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Config.Port), s.n)
	// TODO catch ctrl-c for graceful shutdown
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s server) Flush() error {
	return nil
}