package domain

import (
	"context"
	"net/http"

	"github.com/delving/hub3/ikuzo/logger"
	"github.com/go-chi/chi"
)

// Service defines minimal API service of an ikuzo.Service
type Service interface {
	// Metrics() interface{}
	http.Handler
	Router
	SetServiceBuilder(b *ServiceBuilder)
	Shutdown
}

// Shutdown must be implement by each service that uses background services or connections.
type Shutdown interface {
	Shutdown(ctx context.Context) error
}

// Router implements a callback to register routes to a chi.Router
// If pattern is non-empty this mount point will be used, instead of the
// default specified by the domain.Service implementation
type Router interface {
	Routes(pattern string, router chi.Router)
}

type ServiceBuilder struct {
	Logger *logger.CustomLogger
	Orgs   OrgConfigRetriever
}
