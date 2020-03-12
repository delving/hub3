package ikuzo

import (
	"net/http"

	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/go-chi/chi"
)

// RouterFunc is a callback that registers routes to the ikuzo.Server.
type RouterFunc func(router chi.Router)

// Option is a closure to configure the Server.
// It is used in NewServer.
type Option func(*server) error

// SetPort sets the TCP port for the Server.
//
// The Server listens on :3000 by default.
func SetPort(port int) Option {
	return func(s *server) error {
		s.port = port
		return nil
	}
}

// SetLoggerConfig configures the global logger for the server.
func SetLoggerConfig(cfg logger.Config) Option {
	return func(s *server) error {
		s.loggerConfig = cfg
		return nil
	}
}

// SetDisableRequestLogger disables logging of HTTP request
func SetDisableRequestLogger() Option {
	return func(s *server) error {
		s.disableRequestLogger = true
		return nil
	}
}

// SetMiddleware configures the global middleware for the HTTP router.
func SetMiddleware(middleware ...func(next http.Handler) http.Handler) Option {
	return func(s *server) error {
		s.middleware = append(s.middleware, middleware...)
		return nil
	}
}

// SetRouters adds all HTTP routes for the server.
func SetRouters(rb ...RouterFunc) Option {
	return func(s *server) error {
		s.routerFuncs = append(s.routerFuncs, rb...)
		return nil
	}
}

// SetOrganisationService configures the organization service.
// When no service is set a default transient memory-based service is used.
func SetOrganisationService(service *organization.Service) Option {
	return func(s *server) error {
		s.organizations = service
		return nil
	}
}
