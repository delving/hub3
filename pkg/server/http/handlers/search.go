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

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	log "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	elastic "github.com/olivere/elastic"
)

type contextKey string

const retryKey contextKey = "retry"

func RegisterSearch(router chi.Router) {
	r := chi.NewRouter()

	// throttle queries on elasticsearch
	r.Use(middleware.Throttle(100))

	r.Get("/v2", GetScrollResult)
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

	router.Mount("/api/search", r)

}

func GetScrollResult(w http.ResponseWriter, r *http.Request) {
	searchRequest, err := fragments.NewSearchRequest(r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}
	ProcessSearchRequest(w, r, searchRequest)
	return
}

//
func ProcessSearchRequest(w http.ResponseWriter, r *http.Request, searchRequest *fragments.SearchRequest) {

	s, fub, err := searchRequest.ElasticSearchService(index.ESClient())
	if err != nil {
		log.Printf("Unable to create Search Service: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// suggestion
	//s.Suggester(elastic.NewSuggestField)

	res, err := s.Do(r.Context())
	echoRequest := r.URL.Query().Get("echo")
	if err != nil {
		if echoRequest != "" {
			echo, err := searchRequest.Echo(echoRequest, res.TotalHits())
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
		log.Println("Unable to get search result.")
		log.Println(err)
		return
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return
	}

	if searchRequest.Peek != "" {
		aggs, err := searchRequest.DecodeFacets(res, nil)
		if err != nil {
			log.Printf("Unable to decode facets: %#v", err)
			return
		}
		peek := make(map[string]int64)
		for _, facet := range aggs {
			for _, link := range facet.Links {
				peek[link.Value] = link.Count
			}
		}

		result := &fragments.ScrollResultV4{}
		result.Peek = peek
		render.JSON(w, r, result)
		return
	}

	if searchRequest.CollapseOn != "" {
		records, err := decodeCollapsed(res, searchRequest)
		if err != nil {
			log.Printf("Unable to render collapse")
			return
		}
		result := &fragments.ScrollResultV4{}
		result.Collapsed = records
		result.Pager = &fragments.ScrollPager{Total: res.TotalHits()}
		render.JSON(w, r, result)
		return
	}

	records, searchAfter, err := decodeFragmentGraphs(res)
	searchAfterBin, err := searchRequest.CreateBinKey(searchAfter)
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
	w.Header().Add("P_SCROLL_ID", pager.ScrollID)
	w.Header().Add("P_CURSOR", strconv.Itoa(int(pager.Cursor)))
	w.Header().Add("P_TOTAL", strconv.Itoa(int(pager.Total)))
	w.Header().Add("P_ROWS", strconv.Itoa(int(pager.Rows)))

	// workaround warmer issue ES
	if len(res.Hits.Hits) == 0 && pager.Cursor < int32(pager.Total) {
		log.Printf("bad response from ES retrying the request")
		time.Sleep(1 * time.Second)
		retryCount := r.Context().Value(retryKey)
		if retryCount == nil {
			retryCount = interface{}(0)
		}
		if retryCount.(int) > 3 {
			msg := "empty response from elasticsearch. failed after 3 tries"
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), retryKey, retryCount.(int)+1)
		http.Redirect(w, r.WithContext(ctx), r.URL.RequestURI(), http.StatusSeeOther)
		return
	}

	// Echo requests when requested
	if echoRequest != "" {
		echo, err := searchRequest.Echo(echoRequest, res.TotalHits())
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

	switch echoRequest {
	case "nextScrollID", "searchAfter":
		srNext, err := fragments.SearchRequestFromHex(pager.ScrollID)
		if err != nil {
			http.Error(w, "unable to decode nextScrollID", http.StatusInternalServerError)
			return
		}
		if echoRequest != "searchAfter" {
			render.JSON(w, r, srNext)
			return

		}
		sa, err := srNext.DecodeSearchAfter()
		if err != nil {
			log.Printf("unable to decode searchAfter: %#v", err)
			http.Error(w, "unable to decode next SearchAfter", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, sa)
		return
	case "searchResponse":
		render.JSON(w, r, res)
		return
	case "searchService":
		ss := reflect.ValueOf(s).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)
		srcMap, err := src.Source()
		if err != nil {
			log.Printf("Unable to decode SearchSource: got %s", err)
			http.Error(w, "unable to decode next SearchSource", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, srcMap)
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

	// meta formats that don't use search result
	switch searchRequest.GetResponseFormatType() {
	case fragments.ResponseFormatType_LDJSON:
		entries := []map[string]interface{}{}
		for _, rec := range records {

			for _, json := range rec.NewJSONLD() {
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
				GraphMimeType: "application/ld+json",
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

	switch searchRequest.ItemFormat {
	case fragments.ItemFormatType_FLAT:
		for _, rec := range records {
			rec.NewFields()
			rec.Resources = nil
		}
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
	case fragments.ItemFormatType_TREE:
		leafs := []*fragments.Tree{}

		searching := &fragments.TreeSearching{
			IsSearch: searchRequest.Tree.IsSearch,
			ByLabel:  searchRequest.Tree.Label,
			ByUnitID: searchRequest.Tree.UnitID,
		}
		paging := &fragments.TreePaging{
			PageSize:       searchRequest.Tree.GetPageSize(),
			HitsTotalCount: int32(res.TotalHits()),
		}

		// traditional collapsed tree view
		if searchRequest.Tree.IsExpanded() && len(records) > 0 {
			searching.HitsTotal = int32(res.TotalHits())

			if searchRequest.Tree.IsPaging && searchRequest.Tree.IsSearch {

				leaf := records[0].Tree
				//log.Printf("sortKey: %d (%s)", leaf.SortKey, leaf.CLevel)
				pages, err := searchRequest.Tree.SearchPages(int32(leaf.SortKey))
				if err != nil {
					log.Printf("Unable to get searchPages: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//log.Printf("searchPages: %#v", pages)
				var pageParam string
				for _, page := range pages {
					pageParam = fmt.Sprintf("%s&page=%d", pageParam, page)
				}
				//log.Printf("searchParams: %#v", pageParam)
				qs := fmt.Sprintf("paging=true%s", pageParam)
				m, _ := url.ParseQuery(qs)
				sr, _ := fragments.NewSearchRequest(m)
				err = sr.AddQueryFilter(
					fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
					false,
				)
				if err != nil {
					log.Printf("Unable to add QueryFilter: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				s, _, err := sr.ElasticSearchService(index.ESClient())
				if err != nil {
					log.Printf("Unable to create Search Service: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err := s.Do(r.Context())
				if err != nil {
					return
				}
				if res == nil {
					log.Printf("expected response != nil; got: %v", res)
					return
				}
				paging.HitsTotalCount = int32(res.TotalHits())
				searching.SetPreviousNext(searchRequest.Start)
				records, _, err = decodeFragmentGraphs(res)
				if err != nil {
					return
				}

				// todo set page
				searchRequest.Tree.Page = pages
				searchRequest.Tree.Leaf = leaf.CLevel
			}
			if !searchRequest.Tree.IsPaging {
				leaf := records[0].Tree.CLevel
				qs := fmt.Sprintf("byLeaf=%s&fillTree=true", leaf)
				m, _ := url.ParseQuery(qs)
				sr, _ := fragments.NewSearchRequest(m)
				err := sr.AddQueryFilter(
					fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
					false,
				)
				if err != nil {
					log.Printf("Unable to add QueryFilter: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				s, _, err := sr.ElasticSearchService(index.ESClient())
				if err != nil {
					log.Printf("Unable to create Search Service: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err := s.Do(r.Context())
				if err != nil {
					return
				}
				if res == nil {
					log.Printf("expected response != nil; got: %v", res)
					return
				}
				paging.HitsTotalCount = int32(res.TotalHits())
				searching.SetPreviousNext(searchRequest.Start)
				records, _, err = decodeFragmentGraphs(res)
				if err != nil {
					return
				}
				searchRequest.Tree.FillTree = true
				searchRequest.Tree.Leaf = leaf
			}
		}

		if len(records) == 0 {
			break
		}
		// get paging header for first node returned on the page
		firstNode := records[0]
		lastNode := records[len(records)-1]
		if firstNode.Tree.SortKey != 1 && searchRequest.Tree.IsPaging {
			qs := fmt.Sprintf("byUnitID=%s&allParents=true", firstNode.Tree.Leaf)
			m, _ := url.ParseQuery(qs)
			sr, _ := fragments.NewSearchRequest(m)
			err := sr.AddQueryFilter(
				fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
				false,
			)
			if err != nil {
				log.Printf("Unable to add QueryFilter: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			s, _, err := sr.ElasticSearchService(index.ESClient())
			if err != nil {
				log.Printf("Unable to create Search Service: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			res, err := s.Do(r.Context())
			if err != nil {
				return
			}
			if res == nil {
				log.Printf("expected response != nil; got: %v", res)
				return
			}
			parents, _, err := decodeFragmentGraphs(res)
			if err != nil {
				return
			}
			for _, parent := range parents {
				parent.Tree.HasChildren = parent.Tree.ChildCount != 0
				leafs = append(leafs, parent.Tree)
			}
		}

		for _, rec := range records {
			rec.Tree.HasChildren = rec.Tree.ChildCount != 0
			leafs = append(leafs, rec.Tree)
		}
		// add cursor hint
		if searchRequest.Tree.CursorHint != 0 {
			result.Pager.Cursor = searchRequest.Tree.CursorHint
		}
		records = nil
		if searchRequest.Tree.GetFillTree() || searchRequest.Tree.IsPaging {
			var nodeMap map[string]*fragments.Tree
			result.Tree, nodeMap, err = fragments.InlineTree(leafs, searchRequest.Tree, res.TotalHits())
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to render grouped TreeNodes: %s\n", err.Error())
				render.PlainText(w, r, err.Error())
				return
			}

			tq := searchRequest.Tree
			result.TreeHeader = &fragments.TreeHeader{}

			if tq.IsPaging {
				paging.PageCurrent = tq.GetPage()
				paging.IsSearch = tq.IsSearch
				paging.CalculatePaging()
				result.TreeHeader.Paging = paging
			}

			// leaf based searching
			if tq.GetLeaf() != "" {
				activeNode, ok := nodeMap[tq.GetLeaf()]
				if !ok {
					render.Status(r, http.StatusInternalServerError)
					errMsg := fmt.Sprintf("Unable to find node %s in map", tq.GetLeaf())
					log.Println(errMsg)
					render.PlainText(w, r, errMsg)
					return
				}
				result.TreeHeader.ExpandedIDs = fragments.ExpandedIDs(activeNode)
				result.TreeHeader.ActiveID = tq.GetLeaf()
				paging.ResultActive = activeNode.PageEntry()
			}

			// update paging
			if result.TreeHeader.Paging != nil {
				paging := result.TreeHeader.Paging
				paging.ResultFirst = firstNode.Tree.PageEntry()
				paging.ResultLast = lastNode.Tree.PageEntry()
				paging.SameLeaf = paging.ResultFirst.SameLeaf(paging.ResultLast)
			}
			if searchRequest.Tree.IsSearch {
				result.TreeHeader.Searching = searching
			}

			if searchRequest.Tree.IsNavigatedQuery() {
				result.TreeHeader.PreviousScrollIDs, err = searchRequest.Tree.GetPreviousScrollIDs(
					result.TreeHeader.ActiveID,
					searchRequest,
					pager,
				)
			}
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to render previousScrollIDs: %s\n", err.Error())
				render.PlainText(w, r, err.Error())
				return
			}

			switch searchRequest.Tree.GetPageMode() {
			case "append":
				log.Println("appending")
				page := paging.ResultFirst.CreateTreePage(nodeMap, result.Tree, true, 0)
				result.TreePage = page
				result.Tree = nil
			case "prepend":
				log.Println("prepending")
				page := paging.ResultLast.CreateTreePage(nodeMap, result.Tree, false, paging.ResultFirst.SortKey)
				result.TreePage = page
				result.Tree = nil
			}

			break
		}
		result.Tree = leafs
	case fragments.ItemFormatType_GROUPED:
		for _, rec := range records {
			_, err = rec.NewGrouped()

			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to render grouped resources: %s\n", err.Error())
				render.PlainText(w, r, err.Error())
				return
			}
		}

	}
	result.Items = records

	if !searchRequest.Paging {
		q, _, err := searchRequest.NewUserQuery()
		if err != nil {
			log.Printf("Unable to create User Query")
			return
		}
		q.Numfound = int32(res.TotalHits())
		result.Query = q

		// decode Aggregations
		aggs, err := searchRequest.DecodeFacets(res, fub)
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
		Do(r.Context())
	if err != nil {
		log.Println("Unable to get search result.")
		log.Println(err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, []string{})
		return
	}
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, []string{})
		return
	}
	if !res.Found {
		log.Printf("%s was not found", id)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, []string{})
		return
	}

	record, err := decodeFragmentGraph(res.Source)
	if err != nil {
		fmt.Printf("Unable to decode RDFRecord: %#v", res.Source)
		render.JSON(w, r, []string{})
		render.Status(r, 404)
		return
	}

	switch r.URL.Query().Get("itemFormat") {
	case "flat":
		record.NewFields()
		record.Resources = nil
	case "jsonld":
		record.NewJSONLD()
		record.Resources = nil
	case "summary":
		record.NewResultSummary()
		record.Resources = nil
	case "grouped":
		_, err := record.NewGrouped()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			log.Printf("Unable to render grouped resources: %s\n", err.Error())
			render.PlainText(w, r, err.Error())
			return
		}

	}

	switch r.URL.Query().Get("format") {
	case "jsonld":
		entries := []map[string]interface{}{}
		for _, json := range record.NewJSONLD() {
			entries = append(entries, json)
		}
		record.Resources = nil
		render.JSON(w, r, entries)
		w.Header().Set("Content-Type", "application/json-ld; charset=utf-8")
		return

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

func decodeResourceEntry(hit *json.RawMessage) (*fragments.ResourceEntry, error) {
	re := new(fragments.ResourceEntry)
	if err := json.Unmarshal(*hit, re); err != nil {
		return nil, err
	}
	return re, nil
}

func decodeCollapsed(res *elastic.SearchResult, sr *fragments.SearchRequest) ([]*fragments.Collapsed, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil
	}

	var collapsed []*fragments.Collapsed

	for _, hit := range res.Hits.Hits {
		coll := &fragments.Collapsed{}
		fields, ok := hit.Fields[sr.CollapseOn]
		if ok {
			coll.Field = fields.([]interface{})[0].(string)
		}

		collapseInner := hit.InnerHits["collapse"]
		coll.HitCount = collapseInner.Hits.TotalHits
		for _, inner := range collapseInner.Hits.Hits {
			r, err := decodeFragmentGraph(inner.Source)

			if err != nil {
				return nil, err
			}
			err = decodeHighlights(r, inner)
			if err != nil {
				return nil, err
			}
			coll.Items = append(coll.Items, r)
		}

		collapsed = append(collapsed, coll)
	}

	return collapsed, nil
}

func decodeHighlights(r *fragments.FragmentGraph, hit *elastic.SearchHit) error {

	hl, ok := hit.InnerHits["highlight"]
	if ok {
		for _, hlHit := range hl.Hits.Hits {
			re, err := decodeResourceEntry(hlHit.Source)
			if err != nil {
				return err
			}
			hlEntry, ok := hlHit.Highlight["resources.entries.@value"]
			if ok {
				// add highlighting
				r.Highlights = append(
					r.Highlights,
					&fragments.ResourceEntryHighlight{
						SearchLabel: re.SearchLabel,
						MarkDown:    hlEntry,
					},
				)

			}
		}
	}
	return nil
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

		err = decodeHighlights(r, hit)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, r)
	}
	return records, searchAfter, nil
}
