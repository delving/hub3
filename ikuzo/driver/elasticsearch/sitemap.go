package elasticsearch

import (
	"context"

	"github.com/delving/hub3/ikuzo/service/x/sitemap"
)

var _ sitemap.Store = (*SitemapStore)(nil)

type SitemapStore struct {
	client *Client
}

func NewSitemapStore() *SitemapStore {
	// TODO(kiivihal): Add client later via method
	return &SitemapStore{}
}

func (s *SitemapStore) Datasets(ctx context.Context, cfg sitemap.Config) ([]sitemap.Location, error) {
	// query := elastic.NewBoolQuery()
	// query = query.Must(elastic.NewTermQuery(sr.OrgIDKey, cfg.OrgID))

	return nil, nil
}

func (s *SitemapStore) LocationCount(ctx context.Context, cfg sitemap.Config) (int, error) {
	return 0, nil
}

func (s *SitemapStore) Locations(ctx context.Context, cfg sitemap.Config, start, end int) []sitemap.Location {
	return nil
}
