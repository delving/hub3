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
	"strings"

	"bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/index"
	"bitbucket.org/delving/rapid/hub3/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
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
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	})
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

func getSearchRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := index.ESClient().Get().
		Index(config.Config.ElasticSearch.IndexName).
		Type("rdfrecord").
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

	record, err := decodeRFDRecord(res.Source)
	if err != nil {
		fmt.Printf("Unable to decode RDFRecord: %#v", res.Source)
		return
	}
	render.JSON(w, r, record)
}

func getSearchResult(w http.ResponseWriter, r *http.Request) {
	s := index.ESClient().Search().
		Index(config.Config.ElasticSearch.IndexName).
		Type("rdfrecord").
		Size(20)
	rawQuery := r.FormValue("q")
	fmt.Println("query: ", rawQuery)
	if rawQuery != "" {
		rawQuery = strings.Replace(rawQuery, "delving_spec:", "spec:", 1)
		s = s.Query(elastic.NewQueryStringQuery(rawQuery))
	} else {
		s = s.Query(elastic.NewMatchAllQuery())
	}
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

	fmt.Printf("%#v\n", res)
	fmt.Printf("%#v\n", res.TotalHits())
	records, err := decodeRDFRecords(res)
	if err != nil {
		log.Printf("Unable to decode records")
		return
	}
	render.JSON(w, r, records)
	return
}

func decodeRFDRecord(hit *json.RawMessage) (*models.RDFRecord, error) {
	r := new(models.RDFRecord)
	if err := json.Unmarshal(*hit, r); err != nil {
		return nil, err
	}
	return r, nil
}

// decodeRDFRecords takes a search result and deserializes the records
func decodeRDFRecords(res *elastic.SearchResult) ([]*models.RDFRecord, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil
	}

	var records []*models.RDFRecord
	for _, hit := range res.Hits.Hits {
		r, err := decodeRFDRecord(hit.Source)
		if err != nil {
			return nil, err
		}
		// TODO Add Score here, e.g.:
		// film.Score = *hit.Score
		records = append(records, r)
	}
	return records, nil
}
