// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
