package diwfragments

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

var _ domain.Service = (*Service)(nil)

// Store abstracts fragment persistence so the ES implementation stays a
// swap-out detail behind the frozen contract (pletka principle).
type Store interface {
	// Put upserts fragments by their DocID.
	Put(ctx context.Context, fragments []Fragment) error
	// Get returns the fragment or (nil, nil) when absent.
	Get(ctx context.Context, orgID, collection string, kind Kind, recordID, lang string) (*Fragment, error)
}

// Service serves the /api/ui/v1 fragment contract.
type Service struct {
	store Store
}

// Option configures a Service at construction time.
type Option func(*Service) error

// SetStore injects the fragment store.
func SetStore(store Store) Option {
	return func(s *Service) error {
		s.store = store
		return nil
	}
}

// NewService builds a Service from options.
func NewService(options ...Option) (*Service, error) {
	s := &Service{}
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// ServeHTTP satisfies http.Handler by routing through the chi routes.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// SetServiceBuilder satisfies domain.Service; no builder facilities needed.
func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {}

// Shutdown satisfies domain.Shutdown; the service holds no background work.
func (s *Service) Shutdown(ctx context.Context) error { return nil }
