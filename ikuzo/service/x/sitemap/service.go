package sitemap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/snabb/sitemap"

	"github.com/delving/hub3/ikuzo/domain"
)

var _ domain.Service = (*Service)(nil)

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

func getMaxPages(count int64) (pages []int) {
	maxPages := (count / maxSitemapURLs) + 1
	for i := 0; i < int(maxPages); i++ {
		pages = append(pages, i+1)
	}

	return pages
}

func (s *Service) sitemapRoot(ctx context.Context, cfg domain.SitemapConfig) (*sitemap.SitemapIndex, error) {
	datasets, err := s.store.Datasets(ctx, cfg)
	if err != nil {
		return nil, err
	}

	smi := sitemap.NewSitemapIndex()

	for _, d := range datasets {
		for _, page := range getMaxPages(d.RecordCount) {
			smi.Add(&sitemap.URL{
				Loc: fmt.Sprintf(
					"%s/api/sitemap/%s/%s/%d",
					cfg.BaseURL,
					cfg.ID,
					d.ID,
					page,
				),
			})
		}
	}

	return smi, nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}
