// Copyright 2017 Delving B.V.
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

package handlers

import (
	"fmt"
	"log"
	"net/http"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	elastic "github.com/olivere/elastic/v7"
)

func RegisterContentStats(r chi.Router) {

	// stats dashboard
	r.Get("/api/stats/bySearchLabel", searchLabelStats)
	//r.Get("/api/stats/bySearchLabel/{:label}", searchLabelStatsValues)
	r.Get("/api/stats/byPredicate", predicateStats)
	//r.Get("/api/stats/byPredicate/{:label}", searchLabelStatsValues)

}

func getResourceEntryStats(field string, r *http.Request) (*elastic.SearchResult, error) {

	fieldPath := fmt.Sprintf("resources.entries.%s", field)

	labelAgg := elastic.NewTermsAggregation().Field(fieldPath).Size(100)

	order := r.URL.Query().Get("order")
	switch order {
	case "term":
		labelAgg = labelAgg.OrderByTermAsc()
	default:
		labelAgg = labelAgg.OrderByCountDesc()
	}
	searchLabelAgg := elastic.NewNestedAggregation().Path("resources.entries")
	searchLabelAgg = searchLabelAgg.SubAggregation(field, labelAgg)

	orgID := domain.GetOrganizationID(r)

	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewTermQuery("meta.docType", fragments.FragmentGraphDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, string(orgID)),
	)
	spec := r.URL.Query().Get("spec")
	if spec != "" {
		q = q.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, spec))
	}
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.GetIndexName()).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		Query(q).
		Size(0).
		Aggregation(field, searchLabelAgg).
		Do(r.Context())
	return res, err
}

func searchLabelStats(w http.ResponseWriter, r *http.Request) {

	res, err := getResourceEntryStats("searchLabel", r)
	if err != nil {
		log.Print("Unable to get statistics for searchLabels")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits.Value)
	render.JSON(w, r, res)
	return
}
func predicateStats(w http.ResponseWriter, r *http.Request) {

	res, err := getResourceEntryStats("predicate", r)
	if err != nil {
		log.Print("Unable to get statistics for predicate")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits.Value)
	render.JSON(w, r, res)
	return
}
