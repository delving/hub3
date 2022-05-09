package elasticsearch

import (
	"context"
	"fmt"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/service/x/sitemap"
	"github.com/olivere/elastic/v7"
)

var _ sitemap.Store = (*SitemapStore)(nil)

type SitemapStore struct {
	client *Client
}

func (c *Client) NewSitemapStore() *SitemapStore {
	return &SitemapStore{
		client: c,
	}
}

func (s *SitemapStore) Datasets(ctx context.Context, cfg sitemap.Config) ([]sitemap.Location, error) {
	var locations []sitemap.Location
	agg := elastic.NewCompositeAggregation().
		Sources(
			elastic.NewCompositeAggregationTermsValuesSource("datasets").Field("meta.spec"),
		).Size(200)

	tagQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery(PathTags, "nt"),
			elastic.NewTermQuery(PathTags, "mdr"),
		)

	query := elastic.NewBoolQuery()
	query = query.Must(
		tagQuery,
		elastic.NewTermQuery(PathOrgID, cfg.OrgID),
	)

	search := s.client.search.Search().
		Index(config.Config.ElasticSearch.GetIndexName(cfg.OrgID)).
		Aggregation("datasets", agg).
		Size(0).
		Query(query)

	resp, err := search.
		Do(ctx)
	if err != nil {
		return locations, err
	}

	if resp.Error != nil {
		return locations, fmt.Errorf("%s", resp.Error.Reason)
	}

	comp, ok := resp.Aggregations.Composite("datasets")
	if ok {
		for _, nt := range comp.Buckets {
			spec := nt.Key["datasets"].(string)
			if cfg.IsExcludedSpec(spec) {
				continue
			}

			locations = append(locations, sitemap.Location{
				ID:          spec,
				RecordCount: nt.DocCount,
			})
		}
	}

	return locations, nil
}

func (s *SitemapStore) LocationCount(ctx context.Context, cfg sitemap.Config) (int, error) {
	return 0, nil
}

func (s *SitemapStore) Locations(ctx context.Context, cfg sitemap.Config, start, end int) []sitemap.Location {
	return nil
}
