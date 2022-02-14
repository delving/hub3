package elasticsearch

import (
	"fmt"
	"net/http"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/olivere/elastic/v7"
)

func (c *Client) GetResourceEntryStats(field string, r *http.Request) (*elastic.SearchResult, error) {
	fieldPath := fmt.Sprintf("resources.entries.%s", field)

	labelAgg := elastic.NewTermsAggregation().Field(fieldPath).Size(100)

	order := r.URL.Query().Get("order")
	switch order {
	case "term":
		labelAgg = labelAgg.OrderByKeyAsc()
	default:
		labelAgg = labelAgg.OrderByCountDesc()
	}

	searchLabelAgg := elastic.NewNestedAggregation().Path("resources.entries")
	searchLabelAgg = searchLabelAgg.SubAggregation(field, labelAgg)

	org, ok := domain.GetOrganization(r)
	if !ok {
		return nil, domain.ErrOrgNotFound
	}

	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewTermQuery("meta.docType", fragments.FragmentGraphDocType),
		elastic.NewTermQuery(PathOrgID, org.RawID()),
	)

	spec := r.URL.Query().Get("spec")
	if spec != "" {
		q = q.Must(elastic.NewTermQuery(PathDatasetID, spec))
	}

	res, err := c.search.Search().
		Index(org.Config.GetIndexName()).
		TrackTotalHits(org.Config.ElasticSearch.TrackTotalHits).
		Query(q).
		Size(0).
		Aggregation(field, searchLabelAgg).
		Do(r.Context())

	return res, err
}
