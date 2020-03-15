package elasticsearch

import (
	"github.com/delving/hub3/ikuzo/service/x/search"
	elastic "github.com/olivere/elastic/v7"
)

func NewBoolQuery(q *search.QueryTerm) elastic.Query {

	if !q.IsBoolQuery() {
		match := elastic.NewMatchQuery("", q.Value)

		return match
	}

	boolQuery := elastic.NewBoolQuery()

	return boolQuery
}
