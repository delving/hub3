package elasticsearch

import (
	"strconv"

	"github.com/delving/hub3/ikuzo/service/x/search"
	elastic "github.com/olivere/elastic/v7"
)

type QueryField struct {
	Field string
	Boost float64
}

type QueryBuilder struct {
	defaultFields []QueryField
}

func NewQueryBuilder(defaultFields ...QueryField) *QueryBuilder {
	return &QueryBuilder{
		defaultFields: defaultFields,
	}
}

func (qb *QueryBuilder) NewElasticQuery(q *search.QueryTerm) elastic.Query {
	if !q.IsBoolQuery() && q.Value == "" {
		return elastic.NewMatchAllQuery()
	}

	if !q.IsBoolQuery() {
		switch q.Type() {
		case search.PhraseQuery:
			return buildFieldQueries(q, qb.defaultFields, buildMatchPhraseQuery)
		default:
			return buildFieldQueries(q, qb.defaultFields, buildMatchQuery)
		}
	}

	bq := elastic.NewBoolQuery()

	for _, should := range q.Should() {
		bq = bq.Should(qb.NewElasticQuery(should))
	}

	for _, must := range q.Must() {
		bq = bq.Must(qb.NewElasticQuery(must))
	}

	for _, mustNot := range q.MustNot() {
		bq = bq.MustNot(qb.NewElasticQuery(mustNot))
	}

	return bq
}

type fieldQuery func(q *search.QueryTerm, field QueryField) elastic.Query

func buildFieldQueries(q *search.QueryTerm, fields []QueryField, fn fieldQuery) elastic.Query {
	if len(fields) == 1 {
		return fn(q, fields[0])
	}

	esq := elastic.NewDisMaxQuery()

	queries := []elastic.Query{}

	for _, field := range fields {
		queries = append(queries, fn(q, field))
	}

	return esq.Query(queries...)
}

func buildMatchQuery(q *search.QueryTerm, field QueryField) elastic.Query {
	esq := elastic.NewMatchQuery(field.Field, q.Value)

	if q.Fuzzy != 0 {
		esq = esq.Fuzziness(strconv.Itoa(q.Fuzzy))
	}

	if q.Boost != 0 {
		esq = esq.Boost(q.Boost)
	} else if field.Boost != 0 {
		esq = esq.Boost(field.Boost)
	}

	return esq
}

func buildMatchPhraseQuery(q *search.QueryTerm, field QueryField) elastic.Query {
	esq := elastic.NewMatchPhraseQuery(field.Field, q.Value)

	if q.Fuzzy != 0 {
		esq = esq.Slop(q.Fuzzy)
	}

	if q.Boost != 0 {
		esq = esq.Boost(q.Boost)
	} else if field.Boost != 0 {
		esq = esq.Boost(field.Boost)
	}

	return esq
}
