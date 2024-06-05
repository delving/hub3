package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/olivere/elastic/v7"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/sitemap"
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

func sitemapIndices(cfg domain.SitemapConfig) []string {
	indices := []string{config.Config.ElasticSearch.GetIndexName(cfg.OrgID)}
	if cfg.ContextIndex != "" {
		indices = append(indices, cfg.ContextIndex)
	}

	return indices
}

func (s *SitemapStore) Datasets(ctx context.Context, cfg domain.SitemapConfig) ([]sitemap.Location, error) {
	var locations []sitemap.Location
	agg := elastic.NewCompositeAggregation().
		Sources(
			elastic.NewCompositeAggregationTermsValuesSource("datasets").Field("meta.spec"),
		).Size(500)

	query := buildQuery(cfg, "")

	hasSearchAfter := true
	var searchAfter map[string]any

	for hasSearchAfter {
		if len(searchAfter) > 0 {
			slog.Info("sitemap aggregration searchAfter", "searchAfter", searchAfter)
			agg = agg.AggregateAfter(searchAfter)
		}

		search := s.client.search.Search().
			Index(sitemapIndices(cfg)...).
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
			searchAfter = comp.AfterKey
			if len(searchAfter) == 0 {
				hasSearchAfter = false
			}
		}
	}

	return locations, nil
}

func (s *SitemapStore) Locations(ctx context.Context, spec string, cfg domain.SitemapConfig, cb func(loc sitemap.Location) error) error {
	var err error

	// Create a Point In Time
	openResp, err := s.client.search.OpenPointInTime(sitemapIndices(cfg)...).
		KeepAlive("1m").
		Pretty(true).
		Do(context.Background())
	if err != nil {
		slog.Error("error in elasticsearch scrolling point in time", "error", err)
		return err
	}

	defer func() {
		_, closeErr := s.client.search.ClosePointInTime(openResp.Id).Pretty(true).Do(context.Background())
		if closeErr != nil {
			slog.Error("unable to close point in time", "error", closeErr)
		}
	}()

	var (
		headers     []*fragments.Header
		searchAfter []interface{}
		esPage      int
	)

	pitID := openResp.Id

	for {
		esPage++

		headers, searchAfter, err = s.next(cfg, spec, pitID, searchAfter)
		if err != nil {
			break
		}

		for _, h := range headers {
			lastMod := h.LastModified()
			loc := sitemap.Location{
				ID:      h.EntryURI,
				LastMod: &lastMod,
			}
			if cbErr := cb(loc); cbErr != nil {
				return cbErr
			}
		}
	}

	if !errors.Is(err, ErrEndOfScroll) {
		return err
	}

	return nil
}

func (s *SitemapStore) next(cfg domain.SitemapConfig, spec, pitID string, searchAfter []interface{}) ([]*fragments.Header, []interface{}, error) {
	headers := []*fragments.Header{}

	query := buildQuery(cfg, spec)

	fsc := elastic.NewFetchSourceContext(true)
	fsc.Include("meta")

	pageSize := 1000

	search := s.client.search.Search().
		Size(pageSize).
		Query(query).
		PointInTime(elastic.NewPointInTimeWithKeepAlive(pitID, "1m")).
		FetchSourceContext(fsc).
		Sort("meta.hubID", true)

	if len(searchAfter) > 0 {
		search = search.SearchAfter(searchAfter...)
	}

	resp, err := search.Do(context.Background())
	if err != nil {
		return headers, nil, err
	}

	if resp.Error != nil {
		return headers, nil, fmt.Errorf("%s", resp.Error.Reason)
	}

	if len(resp.Hits.Hits) == 0 {
		return headers, nil, ErrEndOfScroll
	}

	var newSearchAfter []interface{}
	for _, hit := range resp.Hits.Hits {
		newSearchAfter = hit.Sort

		r, err := decodeFragmentGraph(hit.Source)
		if err != nil {
			return nil, nil, err
		}

		headers = append(headers, r.Meta)
	}

	return headers, newSearchAfter, nil
}

func buildQuery(cfg domain.SitemapConfig, spec string) elastic.Query {
	filters := elastic.NewBoolQuery()
	for _, filter := range cfg.Filters {
		field, value, ok := strings.Cut(filter, ":")
		if !ok {
			continue
		}

		if strings.HasPrefix(field, "-") {
			filters = filters.MustNot(elastic.NewTermQuery(strings.TrimPrefix(field, "-"), value))
			continue
		}
		filters = filters.Should(elastic.NewTermsQuery(field, value))
	}

	orgFilter := []elastic.Query{elastic.NewTermQuery(PathOrgID, cfg.OrgID)}
	if cfg.ContextIndex != "" {
		contextOrg := strings.TrimSuffix(cfg.ContextIndex, "v2")
		orgFilter = append(orgFilter, elastic.NewTermQuery(PathOrgID, contextOrg))
	}

	query := elastic.NewBoolQuery()
	query = query.Should(
		orgFilter...,
	)

	if spec != "" {
		query = query.Must(
			elastic.NewTermQuery("meta.spec", spec),
		)
	}

	if len(cfg.Filters) > 0 {
		query = query.Must(filters)
	}

	return query
}
