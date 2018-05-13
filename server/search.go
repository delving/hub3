// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"fmt"
	log "log"
	"net/http"

	"github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/fragments"
	"github.com/delving/rapid-saas/hub3/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	//elastic "github.cocm/olivere/elastic"
	elastic "gopkg.in/olivere/elastic.v5"
)

// SearchResource is a struct for the Search routes
type SearchResource struct{}

// Routes returns the chi.Router
func (rs SearchResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/v2", getSearchResult)
	r.Get("/v2/{id}", func(w http.ResponseWriter, r *http.Request) {
		getSearchRecord(w, r)
		return
	})
	r.Get("/v2/scroll", getScrollResult)
	r.Get("/v1", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	})
	r.Get("/v1/{id}", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	})

	return r
}

func getScrollResult(w http.ResponseWriter, r *http.Request) {

	searchRequest, err := fragments.NewSearchRequest(r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		return
	}

	log.Printf("%#v\n", searchRequest)

	s := index.ESClient().Search().
		Index(config.Config.ElasticSearch.IndexName).
		Size(int(searchRequest.GetResponseSize()))

	query, err := searchRequest.ElasticQuery()
	if err != nil {
		log.Println("Unable to get search result.")
		log.Println(err)
		return
	}

	source, _ := query.Source()
	qs, _ := json.MarshalIndent(source, "", " ")
	log.Printf("%s\n", qs)

	s = s.Query(query)
	res, err := s.Do(ctx)
	if err != nil {
		log.Println("Unable to get search result.")
		log.Println(err)
		return
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return
	}
	//records, err := decodeFragmentGraphs(res)
	//if err != nil {
	//log.Printf("Unable to decode records")
	//return
	//}

	pager, err := searchRequest.NextScrollID(res.TotalHits())
	if err != nil {
		log.Println("Unable to create Scroll Pager. ")
		return
	}

	result := fragments.ScrollResultV3{}
	result.Pager = pager
	//result.Items = records
	render.JSON(w, r, result)
	return
}

func getSearchRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := index.ESClient().Get().
		Index(config.Config.ElasticSearch.IndexName).
		Id(id).
		Do(ctx)
	if err != nil {
		log.Println("Unable to get search result.")
		log.Println(err)
		return
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return
	}
	if !res.Found {
		log.Printf("%s was not found", id)
		return
	}

	record, err := decodeFragmentGraph(res.Source)
	if err != nil {
		fmt.Printf("Unable to decode RDFRecord: %#v", res.Source)
		return
	}
	render.JSON(w, r, record)
}

func getSearchResult(w http.ResponseWriter, r *http.Request) {
	s := index.ESClient().Search().
		Index(config.Config.ElasticSearch.IndexName).
		Size(20)
	rawQuery := r.FormValue("q")
	fmt.Println("query: ", rawQuery)
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("docType", fragments.FragmentGraphDocType))
	// todo enable query later
	//if rawQuery != "" {
	//rawQuery = strings.Replace(rawQuery, "delving_spec:", "spec:", 1)
	//s = s.Query(elastic.NewQueryStringQuery(rawQuery))
	//} else {
	//s = s.Query(elastic.NewMatchAllQuery())
	//}
	s = s.Query(query)
	res, err := s.Do(ctx)
	if err != nil {
		log.Println("Unable to get search result.")
		log.Println(err)
		return
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return
	}
	records, err := decodeFragmentGraphs(res)
	if err != nil {
		log.Printf("Unable to decode records")
		return
	}
	result := fragments.SearchResultV3{}
	result.Items = records
	render.JSON(w, r, result)
	return
}

func decodeFragmentGraph(hit *json.RawMessage) (*fragments.FragmentGraph, error) {
	r := new(fragments.FragmentGraph)
	if err := json.Unmarshal(*hit, r); err != nil {
		return nil, err
	}
	return r, nil
}

// decodeFragmentGraphs takes a search result and deserializes the records
func decodeFragmentGraphs(res *elastic.SearchResult) ([]*fragments.FragmentGraph, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil
	}

	var records []*fragments.FragmentGraph
	for _, hit := range res.Hits.Hits {
		r, err := decodeFragmentGraph(hit.Source)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
