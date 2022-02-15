// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ikuzo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// nolint:gosec // imported for metrics server. Not exposes on default routes because we don't use the default mux

	_ "net/http/pprof"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/middleware"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	defaultServerPort      = 3000
	defaultShutdownTimeout = 10
)

type Service interface {
	Metrics() interface{}
	http.Handler
	Shutdown
}

// Server provides a net/http compliant WebServer.
type Server interface {
	ListenAndServe() error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Shutdown must be implement by each service that uses background services or connections.
type Shutdown interface {
	Shutdown(ctx context.Context) error
}

type server struct {
	// router is compatible with http.Mux
	router chi.Router
	// port is where the server will listen to TCP requests
	port int
	// metricsPort is the port where expvar is hosted
	metricsPort int
	// TLS certificate
	certFile string
	// TLS keyFile
	keyFile string
	// cancelFunc is called for graceful shutdown of resources and background workers.
	cancelFunc context.CancelFunc
	// workers is a pool that manages all the background WorkerServices
	workers *workerPool
	// gracefulTimeout maximum duration of graceful shutdown of server. (default: 10 seconds)
	gracefulTimeout time.Duration
	// disableRequestLogger stops logging of request information to the global logger
	disableRequestLogger bool
	// logger is the custom zerolog logger
	logger *logger.CustomLogger
	// middleware is an array of middleware options that will be applied.
	// When none are given the default middleware is applied.
	middleware []func(http.Handler) http.Handler
	// routerFuncs are the custom routes.
	// When none are given the default routes are applied.
	routerFuncs []RouterFunc
	// service to access the organization store
	organizations *organization.Service
	// services list registered services
	services []domain.Service
	// shutdownHooks are called on server shutdown
	shutdownHooks map[string]domain.Shutdown
	// service context
	ctx context.Context
	// oto is the OTO generated RCP service
	oto *otohttp.Server
	// introspect enables routes for introspection
	introspect bool
	// sentry shows if sentry is enabled
	sentry bool
}

// NewServer returns the default server.
// The configuration can be modified using Option functions.
// All services are lazy-loaded.
func NewServer(options ...Option) (Server, error) {
	return newServer(options...)
}

// newServer returns the default server.
func newServer(options ...Option) (*server, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	s := &server{
		port:            defaultServerPort,
		cancelFunc:      cancelFunc,
		workers:         newWorkerPool(ctx),
		gracefulTimeout: defaultShutdownTimeout * time.Second,
		shutdownHooks:   make(map[string]domain.Shutdown),
		ctx:             ctx,
	}

	s.setRouterdefaults()

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	// set global logger
	if s.logger != nil {
		log.Logger = s.logger.Logger
	}

	if s.logger == nil {
		l := logger.Nop()
		s.logger = &l
	}

	// append default middleware
	s.middleware = append(s.middleware, DefaultMiddleware()...)

	s.router.Use(s.middleware...)

	// recover is not optional
	s.router.Use(s.recoverer)

	if s.sentry {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		})
		s.router.Use(sentryMiddleware.Handle)
	}

	// cors is not optional
	s.router.Use(
		cors.Handler(cors.Options{
			// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	)

	// setting up request logging middleware
	if !s.disableRequestLogger {
		s.router.Use(middleware.RequestLogger(&log.Logger))
	}

	// setup oto server
	if s.oto != nil {
		log.Info().Msg("starting with oto service")
		s.router.HandleFunc("/oto/*", s.oto.ServeHTTP)
	}

	// setting default services
	s.setDefaultServices()

	// apply default routes
	s.routes()

	// register services with ikuzo server and router
	for _, svc := range s.services {
		if err := s.registerService(svc); err != nil {
			return nil, err
		}
	}

	// apply custom routes
	for _, f := range s.routerFuncs {
		f(s.router)
	}

	if s.introspect {
		s.router.Get("/introspect/routes", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(docgen.JSONRoutesDoc(s.router)))
		})
	}

	// s.logger.Debug().Msg(docgen.JSONRoutesDoc(s.router))
	// TODO: maybe add server validation function

	return s, nil
}

// registerService registers a Service interface to the ikuzo server
func (s *server) registerService(svc domain.Service) error {
	// register routes
	s.registerRouter("", svc)

	builder := &domain.ServiceBuilder{
		Orgs:   s.organizations,
		Logger: s.logger,
	}

	// set organization service
	svc.SetServiceBuilder(builder)

	return nil
}

// registerRouter mounts supplied domain.Router in the server router
//
// This should only be called for routes that are not part of a domain.Service
func (s *server) registerRouter(pattern string, router domain.Router) {
	router.Routes(pattern, s.router)
}

func (s *server) setDefaultServices() {
	// can be used to set default service configurations
}

func (s *server) setRouterdefaults() {
	router := chi.NewRouter()
	router.NotFound(s.handle404)
	router.MethodNotAllowed(s.handleMethodNotAllowed)

	s.router = router
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// ListenAndServe starts a HTTP-server with graceful shutdown.
func (s *server) ListenAndServe() error {
	return s.listenAndServe()
}

func (s *server) listenAndServe(testSignals ...interface{}) error {
	log.Info().
		Int("port", s.port).
		Msg("starting server")

	// gather errors
	allowedErrors := 10
	errChan := make(chan error, allowedErrors)

	if s.metricsPort != 0 {
		log.Info().
			Int("port", s.metricsPort).
			Msg("starting metrics server")

		go func() {
			errChan <- http.ListenAndServe(fmt.Sprintf(":%d", s.metricsPort), nil)
		}()
	}

	// start web-server
	server := http.Server{Addr: fmt.Sprintf(":%d", s.port), Handler: s}

	go func() {
		if s.certFile != "" && s.keyFile != "" {
			errChan <- server.ListenAndServeTLS(s.certFile, s.keyFile)
		} else {
			errChan <- server.ListenAndServe()
		}
	}()

	// watch for quit signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// inject signals for testing
	for _, sign := range testSignals {
		switch v := sign.(type) {
		case os.Signal:
			signalChan <- v
		case error:
			errChan <- v
		}
	}

	// block until a select case is satisfied
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case sig := <-signalChan:
			log.Warn().
				Str("signal", sig.String()).
				Msg("caught shutdown signal, starting graceful shutdown")

			return s.shutdown(&server)
		case <-s.workers.ctx.Done():
			return s.workers.ctx.Err()
		}
	}
}

func (s *server) shutdown(server *http.Server) error {
	// if sentry is running flush the messages
	if s.sentry {
		sentry.Flush(time.Second)
	}
	log.Info().Msg("sending stop signal to background processes")

	// cancel context to shutdown background processes and connections
	s.cancelFunc()

	// set maximum duration for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
	defer cancel()

	log.Info().Msg("stopping web-server")
	server.SetKeepAlivesEnabled(false)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return server.Shutdown(ctx) })

	for _, svc := range s.services {
		svc := svc

		g.Go(func() error { return svc.Shutdown(ctx) })
	}

	for _, h := range s.shutdownHooks {
		h := h

		g.Go(func() error { return h.Shutdown(ctx) })
	}

	// wait until all background workers are finished
	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to shutdown all workers; %w", err)
	}

	log.Info().Msg("finished shutting down background processes")

	return nil
}

// decode decodes the body of the http.Request into the provided interface.
func (s *server) decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// handle404 returns a custom response when a page is not found.
func (s *server) handle404(w http.ResponseWriter, r *http.Request) {
	s.respondWithError(w, r, errors.New("page not found"), http.StatusNotFound)
}

// handleIndex returns default information about the deployment
func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/version", http.StatusFound)
	}
}

// handleMethodNotAllowed returns a custom response when a method is not allowed.
func (s *server) handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.respondWithError(w, r, fmt.Errorf("method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
}

// Get the logger from the request's context. You can safely assume it
// will be always there: if the handler is removed, hlog.FromRequest
// will return a no-op logger.
func (s *server) requestLogger(r *http.Request) *zerolog.Logger {
	return hlog.FromRequest(r)
}

// respond is helper to encode responses from the Server.
func (s *server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	render.Status(r, status)

	if data != nil {
		render.JSON(w, r, data)
	}
}

// respondWithError returns a standardized error message that is encoded by the *server.Respond function.
func (s *server) respondWithError(w http.ResponseWriter, r *http.Request, err error, status int) {
	render.Error(w, r, err, &render.ErrorConfig{
		Log:        &s.logger.Logger,
		StatusCode: status,
	})
}

// recoverer is a middleware that recovers from panics, logs the panic (and a
// stacktrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ID if one is provided.
func (s *server) recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				errText := http.StatusText(http.StatusInternalServerError)

				requestID, ok := hlog.IDFromRequest(r)
				if !ok {
					requestID = xid.New()
				}

				s.logger.WithLevel(zerolog.PanicLevel).
					Stack().
					Str("req_id", requestID.String()).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Int("status", http.StatusInternalServerError).
					Dict("params", middleware.LogParamsAsDict(r.URL.Query())).
					Msg(fmt.Sprintf("Recover from Panic: %s;", rvr))

				err := fmt.Errorf("%s; error logged with request_id: %s", errText, requestID)
				s.respondWithError(w, r, err, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (s *server) addShutdown(name string, hook domain.Shutdown) {
	if _, ok := s.shutdownHooks[name]; !ok {
		s.shutdownHooks[name] = hook
	}
}
