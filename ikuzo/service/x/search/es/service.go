package es

import (
	"context"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs   domain.OrgConfigRetriever
	log    zerolog.Logger
	client *elasticsearch.Client
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
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
	s.log = b.Logger.With().Str("svc", "{}").Logger()
	s.orgs = b.Orgs
}
