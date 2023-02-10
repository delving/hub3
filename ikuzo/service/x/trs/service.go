package trs

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"gocloud.dev/blob"

	"github.com/delving/hub3/ikuzo/domain"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	store  Store
	bucket *blob.Bucket
	orgs   domain.OrgConfigRetriever
	log    zerolog.Logger
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
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}
