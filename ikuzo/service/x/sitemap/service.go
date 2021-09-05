package sitemap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/snabb/sitemap"
)

type Service struct {
	store Store
	orgs  domain.OrgConfigRetriever
	log   zerolog.Logger
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

func (s *Service) sitemapRoot(ctx context.Context, cfg Config) (*sitemap.SitemapIndex, error) {
	datasets, err := s.store.Datasets(ctx, cfg)
	if err != nil {
		return nil, err
	}

	smi := sitemap.NewSitemapIndex()

	for _, d := range datasets {
		smi.Add(&sitemap.URL{
			Loc: fmt.Sprintf(
				"%s/api/sitemap/%s/%s",
				cfg.BaseURL,
				cfg.ID,
				d.ID,
			),
		})
	}

	return smi, nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}
