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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	log "log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3"
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

	r.Get("/v2", getScrollResult)
	r.Get("/v2/{id}", func(w http.ResponseWriter, r *http.Request) {
		getSearchRecord(w, r)
		return
	})

	r.Get("/v1", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	})
	r.Get("/v1/{id}", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	})

	return r
}

func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getScrollResult(w http.ResponseWriter, r *http.Request) {

	searchRequest, err := fragments.NewSearchRequest(r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}

	// Echo requests when requested
	echoRequest := r.URL.Query().Get("echo")
	if echoRequest != "" {
		echo, err := searchRequest.Echo(echoRequest, 0)
		if err != nil {
			log.Println("Unable to echo request")
			log.Println(err)
			return
		}
		if echo != nil {
			render.JSON(w, r, echo)
			return
		}
	}

	s, err := searchRequest.ElasticSearchService(index.ESClient())
	if err != nil {
		log.Println("Unable to create Search Service")
		return
	}

	switch echoRequest {
	case "searchService":
		render.JSON(w, r, searchRequest)
		return
	case "request":
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			msg := fmt.Sprintf("Unable to dump request: %s", err)
			log.Print(msg)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprint(msg),
				Error:      err,
			})
			return
		}

		render.PlainText(w, r, string(dump))
		return
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

	if echoRequest == "searchResponse" {
		render.JSON(w, r, res)
		return
	}

	records, searchAfter, err := decodeFragmentGraphs(res)
	searchAfterBin, err := getBytes(searchAfter)
	if err != nil {
		log.Printf("Unable to encode searchAfter")
		return
	}

	searchRequest.SearchAfter = searchAfterBin
	if err != nil {
		log.Printf("Unable to decode records")
		return
	}

	pager, err := searchRequest.NextScrollID(res.TotalHits())
	if err != nil {
		log.Println("Unable to create Scroll Pager. ")
		return
	}

	// Add scrollID pager information to the header
	w.Header().Add("P_SCROLL_ID", pager.GetScrollID())
	w.Header().Add("P_CURSOR", strconv.Itoa(int(pager.GetCursor())))
	w.Header().Add("P_TOTAL", strconv.Itoa(int(pager.GetTotal())))
	w.Header().Add("P_ROWS", strconv.Itoa(int(pager.GetRows())))

	// meta formats that don't use search result
	switch searchRequest.GetResponseFormatType() {
	case fragments.ResponseFormatType_LDJSON:
		entries := []map[string]interface{}{}
		for _, rec := range records {

			for _, json := range rec.NewJSONLD() {
				log.Println(json)
				entries = append(entries, json)
			}
			rec.Resources = nil
		}
		render.JSON(w, r, entries)
		w.Header().Set("Content-Type", "application/json-ld; charset=utf-8")
		return
	case fragments.ResponseFormatType_BULKACTION:
		actions := []string{}
		for _, rec := range records {
			rec.NewJSONLD()
			graph, err := json.Marshal(rec.JSONLD)
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to marshal json-ld to string : %s\n", err.Error())
				render.PlainText(w, r, err.Error())
				return
			}
			action := &hub3.BulkAction{
				HubID:         rec.Meta.HubID,
				Spec:          rec.Meta.Spec,
				NamedGraphURI: rec.Meta.NamedGraphURI,
				Action:        "index",
				Graph:         string(graph),
				RecordType:    "mdr",
			}
			bytes, err := json.Marshal(action)
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to create Bulkactions: %s\n", err.Error())
				render.PlainText(w, r, err.Error())
				return
			}
			actions = append(actions, string(bytes))
		}
		render.PlainText(w, r, strings.Join(actions, "\n"))
		//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	result := &fragments.ScrollResultV4{}
	result.Pager = pager

	// FragmentGraph is the default decoded issue
	switch searchRequest.ItemFormat {
	case fragments.ItemFormatType_SUMMARY:
		for _, rec := range records {
			rec.NewResultSummary()
			rec.Resources = nil
		}
	case fragments.ItemFormatType_JSONLD:
		for _, rec := range records {
			rec.NewJSONLD()
			rec.Resources = nil
		}
	}
	result.Items = records

	if !searchRequest.Paging {
		// decode Aggregations
		aggs, err := searchRequest.DecodeFacets(res)
		if err != nil {
			log.Printf("Unable to decode facets: %#v", err)
			return
		}
		result.Facets = aggs
	}

	switch searchRequest.GetResponseFormatType() {
	// TODO enable later again
	//case fragments.ResponseFormatType_PROTOBUF:
	//output, err := proto.Marshal(result)
	//if err != nil {
	//log.Println("Unable to marshal result to protobuf format.")
	//return
	//}
	//render.Data(w, r, output)
	default:
		render.JSON(w, r, result)
	}
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
	switch r.URL.Query().Get("format") {
	// TODO enable protobuf later again
	//case "protobuf":
	//output, err := proto.Marshal(record)
	//if err != nil {
	//log.Println("Unable to marshal result to protobuf format.")
	//return
	//}
	//render.Data(w, r, output)
	default:
		render.JSON(w, r, record)
	}
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
func decodeFragmentGraphs(res *elastic.SearchResult) ([]*fragments.FragmentGraph, []interface{}, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil, nil
	}

	var records []*fragments.FragmentGraph
	var searchAfter []interface{}
	for _, hit := range res.Hits.Hits {
		searchAfter = hit.Sort
		r, err := decodeFragmentGraph(hit.Source)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, r)
	}
	return records, searchAfter, nil
}
