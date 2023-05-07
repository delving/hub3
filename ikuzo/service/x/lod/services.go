package lod

import (
	"context"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs         domain.OrgConfigRetriever
	log          zerolog.Logger
	stores       map[string]Resolver
	defaultStore string
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		stores: make(map[string]Resolver),
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	// set default store if none is given
	if len(s.stores) == 1 && s.defaultStore == "" {
		for key := range s.stores {
			s.defaultStore = key
			break
		}
	}

	return s, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "lod").Logger()
	s.orgs = b.Orgs
}
