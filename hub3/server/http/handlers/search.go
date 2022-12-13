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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/delving/hub3/config"
	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/delving/hub3/ikuzo/search"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	elastic "github.com/olivere/elastic/v7"
)

var (
	noSearchServiceMsg        = "Unable to create Search Service: %v"
	unexpectedResponseMsg     = "expected response != nil; got: %v"
	unableToAddQueryFilterMsg = "Unable to add QueryFilter: %v"
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
		render.Error(w, r, fmt.Errorf("v1 not enabled"), &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
		})
		return
	})
	r.Get("/v1/{id}", func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, r, fmt.Errorf("v1 not enabled"), &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
		})
		return
	})

	router.Mount("/api/search", r)

	v2 := chi.NewRouter()
	v2.Use(middleware.Throttle(100))
	v2.Get("/search", GetScrollResult)
	v2.Get("/search/{id}", func(w http.ResponseWriter, r *http.Request) {
		getSearchRecord(w, r)
		return
	})

	router.Mount("/v2", v2)
}

func GetScrollResult(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	searchRequest, err := fragments.NewSearchRequest(orgID.String(), r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}
	ProcessSearchRequest(w, r, searchRequest)
	return
}

func ProcessSearchRequest(w http.ResponseWriter, r *http.Request, searchRequest *fragments.SearchRequest) {
	orgID := domain.GetOrganizationID(r)

	s, fub, err := searchRequest.ElasticSearchService(index.ESClient())
	if err != nil {
		log.Printf(noSearchServiceMsg, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if searchRequest.Tree != nil {
		tree := searchRequest.Tree
		// empty spec queries cause issues with paging
		if tree.Spec == "" {
			searchRequest.Tree = nil
		}
	}

	res, err := s.Do(r.Context())
	echoRequest := NewEchoSearchRequest(r, searchRequest, s, res)
	if err != nil {
		if echoRequest.HasEcho() {
			if err := echoRequest.RenderEcho(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		if echoErr := echoRequest.RenderEcho(w); echoErr != nil {
			http.Error(w, fmt.Sprintf("unable to render echo: %s", echoErr), http.StatusInternalServerError)
			return
		}

		status := http.StatusInternalServerError
		if res != nil {
			status = res.Status
		}

		http.Error(w, fmt.Sprintf("unable to get search results: %s", err), status)

		log.Printf("Unable to get search result; %s", err)
		return
	}
	if res == nil {
		log.Printf(unexpectedResponseMsg, res)
		return
	}

	if searchRequest.Peek != "" {
		if echoRequest.HasEcho() {
			if err := echoRequest.RenderEcho(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

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
		if echoRequest.HasEcho() {
			if err := echoRequest.RenderEcho(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		records, err := decodeCollapsed(res, searchRequest)
		if err != nil {
			log.Printf("Unable to render collapse")
			return
		}

		if searchRequest.CollapseFormat == "flat" {
			for _, rec := range records {
				for _, item := range rec.Items {
					item.NewFields(nil)
					item.Resources = nil
					item.Summary = nil
					item.ProtoBuf = nil
				}
			}
		}

		result := &fragments.ScrollResultV4{}
		result.Collapsed = records

		var collapsedIds int

		filteredAgg, ok := res.Aggregations.Filter("counts")
		if ok {
			collapseCount, ok := filteredAgg.Aggregations.Cardinality("collapseCount")
			if ok {
				collapsedIds = int(*collapseCount.Value)
			}
		}
		searchRequest.ResponseSize = searchRequest.CollapseSize

		result.Pager, err = searchRequest.ScrollPagers(int64(collapsedIds))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// decode Aggregations
		aggs, err := searchRequest.DecodeFacets(res, fub)
		if err != nil {
			log.Printf("Unable to decode facets: %#v", err)
			return
		}
		result.Facets = aggs

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

	var paginator *search.Paginator
	if searchRequest.V1Mode {
		paginator, err = search.NewPaginator(
			int(res.TotalHits()),
			int(searchRequest.GetResponseSize()),
			int(searchRequest.GetPage()),
			int(searchRequest.GetStart()),
		)
		if err != nil {
			log.Printf("Unable to create Paginator: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		searchRequest.Start = int32(paginator.Start) - 1

		if err := paginator.AddPageLinks(); err != nil {
			log.Println("Unable to create PageLinks")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	pager, err := searchRequest.ScrollPagers(res.TotalHits())
	if err != nil {
		log.Println("Unable to create Scroll Pager. ")
		return
	}

	if paginator != nil {
		if pager.Cursor == int32(0) && paginator.Start != 0 {
			pager.Cursor = int32(paginator.Start)
		}
	}

	// Add scrollID pager information to the header
	w.Header().Add("P_PREVIOUS_SCROLL_ID", pager.PreviousScrollID)
	w.Header().Add("P_NEXT_SCROLL_ID", pager.NextScrollID)
	w.Header().Add("P_CURSOR", strconv.Itoa(int(pager.Cursor)))
	w.Header().Add("P_TOTAL", strconv.Itoa(int(pager.Total)))
	w.Header().Add("P_ROWS", strconv.Itoa(int(pager.Rows)))

	// workaround warmer issue ES
	if res.Hits == nil && int64(pager.Cursor) < pager.Total {
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

	echoRequest.ScrollPager = pager

	if echoRequest.HasEcho() {
		if echoErr := echoRequest.RenderEcho(w); echoErr != nil {
			http.Error(w, echoErr.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// meta formats that don't use search result
	switch searchRequest.GetResponseFormatType() {
	case fragments.ResponseFormatType_LDJSON:
		entries := []map[string]interface{}{}

		for _, rec := range records {
			entries = append(entries, rec.NewJSONLD()...)
			rec.Resources = nil
		}

		render.JSON(w, r, entries)
		w.Header().Set("Content-Type", "application/json-ld; charset=utf-8")

		return
	case fragments.ResponseFormatType_BULKACTION:
		actions := []string{}

		for _, rec := range records {
			rec.NewJSONLD()
			graph, marshalErr := json.Marshal(rec.JSONLD)
			if marshalErr != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to marshal json-ld to string : %s\n", marshalErr.Error())
				render.PlainText(w, r, marshalErr.Error())
				return
			}

			action := &bulk.Request{
				HubID:         rec.Meta.HubID,
				DatasetID:     rec.Meta.Spec,
				NamedGraphURI: rec.Meta.NamedGraphURI,
				Action:        "index",
				Graph:         string(graph),
				GraphMimeType: "application/ld+json",
				RecordType:    "mdr",
			}

			bytes, marshalErr := json.Marshal(action)
			if marshalErr != nil {
				render.Status(r, http.StatusInternalServerError)
				log.Printf("Unable to create Bulkactions: %s\n", marshalErr.Error())
				render.PlainText(w, r, marshalErr.Error())
				return
			}

			actions = append(actions, string(bytes))
		}
		render.PlainText(w, r, strings.Join(actions, "\n"))
		// w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	result := &fragments.ScrollResultV4{}
	result.Pager = pager
	// TODO(kiivihal): how to enable or disable this

	if paginator != nil {
		result.Pagination = paginator
	}

	var textQuery *memory.TextQuery

	if searchRequest.Query != "" {
		var textQueryErr error

		textQuery, textQueryErr = memory.NewTextQueryFromString(searchRequest.Query)
		if textQueryErr != nil {
			log.Printf("unable to build text query: %s\n", err.Error())
			http.Error(w, textQueryErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch searchRequest.ItemFormat {
	case fragments.ItemFormatType_FLAT:
		for _, rec := range records {
			rec.NewFields(textQuery)
			rec.Resources = nil
			rec.Summary = nil
			rec.ProtoBuf = nil
		}
	case fragments.ItemFormatType_SUMMARY:
		for _, rec := range records {
			rec.NewResultSummary()
			rec.Resources = nil
			rec.ProtoBuf = nil
		}
	case fragments.ItemFormatType_JSONLD:
		for _, rec := range records {
			rec.NewJSONLD()
			rec.Resources = nil
			rec.ProtoBuf = nil
		}
	case fragments.ItemFormatType_TREE:
		result.Pagination = nil
		leafs := []*fragments.Tree{}

		searching := &fragments.TreeSearching{
			IsSearch: searchRequest.Tree.IsSearch,
			ByLabel:  searchRequest.Tree.Label,
			ByUnitID: searchRequest.Tree.UnitID,
			ByQuery:  searchRequest.Tree.Query,
		}
		paging := &fragments.TreePaging{
			PageSize:       searchRequest.Tree.GetPageSize(),
			HitsTotalCount: int32(res.TotalHits()),
		}

		// with zero results load the default first page
		if paging.HitsTotalCount == 0 && searchRequest.Tree.IsSearch && searchRequest.Tree.IsPaging {
			// if there is no query in the params then this is already a redirect
			if !r.URL.Query().Has("q") && !r.URL.Query().Has("query") && !r.URL.Query().Has("byQuery") {

				http.Error(w, fmt.Sprintf("inventoryID '%s' not found", searchRequest.Tree.UnitID), http.StatusNotFound)
				return
			}
			newPath := fmt.Sprintf("%s?paging=true&page=1", r.URL.Path)
			http.Redirect(w, r, newPath, http.StatusSeeOther)
			return
		}

		if searchRequest.Tree.Query != "" {
			var textQueryErr error

			textQuery, textQueryErr = memory.NewTextQueryFromString(searchRequest.Tree.Query)
			if textQueryErr != nil {
				log.Printf("unable to build text query: %s\n", err.Error())
				http.Error(w, textQueryErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Traditional expanded tree view.
		if searchRequest.Tree.IsExpanded() && len(records) > 0 {
			searching.HitsTotal = int32(res.TotalHits())

			// Paging call with a searchHit.
			if searchRequest.Tree.IsPaging && searchRequest.Tree.IsSearch {
				leaf := records[0].Tree
				pages, err := searchRequest.Tree.SearchPages(int32(leaf.SortKey))
				if err != nil {
					log.Printf("Unable to get searchPages: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				var pageParam string

				for _, page := range pages {
					pageParam = fmt.Sprintf("%s&treePage=%d", pageParam, page)
				}
				qs := fmt.Sprintf("paging=true%s", pageParam)
				m, _ := url.ParseQuery(qs)
				sr, _ := fragments.NewSearchRequest(orgID.String(), m)
				sr.Tree.WithFields = searchRequest.Tree.WithFields

				err = sr.AddQueryFilter(
					fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
					false,
				)
				if err != nil {
					log.Printf(unableToAddQueryFilterMsg, domain.LogUserInput(err.Error()))
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				s, _, err := sr.ElasticSearchService(index.ESClient())
				if err != nil {
					log.Printf(noSearchServiceMsg, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err := s.Do(r.Context())
				if err != nil {
					return
				}
				if res == nil {
					log.Printf(unexpectedResponseMsg, res)
					return
				}
				paging.HitsTotalCount = int32(res.TotalHits())

				searching.SetPreviousNext(searchRequest.Start)

				records, _, err = decodeFragmentGraphs(res)
				if err != nil {
					return
				}

				// TODO(kiivihal): set page
				searchRequest.Tree.Page = pages
				searchRequest.Tree.Leaf = leaf.CLevel
			}

			if !searchRequest.Tree.IsPaging {
				leaf := records[0].Tree.CLevel
				qs := fmt.Sprintf("byLeaf=%s&fillTree=true", leaf)
				m, _ := url.ParseQuery(qs)
				sr, _ := fragments.NewSearchRequest(orgID.String(), m)
				sr.Tree.WithFields = searchRequest.Tree.WithFields

				err := sr.AddQueryFilter(
					fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
					false,
				)
				if err != nil {
					log.Printf(unableToAddQueryFilterMsg, domain.LogUserInput(err.Error()))
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				s, _, err := sr.ElasticSearchService(index.ESClient())
				if err != nil {
					log.Printf(noSearchServiceMsg, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err := s.Do(r.Context())
				if err != nil {
					return
				}
				if res == nil {
					log.Printf(unexpectedResponseMsg, res)
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
			sr, _ := fragments.NewSearchRequest(orgID.String(), m)
			sr.Tree.WithFields = searchRequest.Tree.WithFields

			err := sr.AddQueryFilter(
				fmt.Sprintf("%s:%s", config.Config.ElasticSearch.SpecKey, searchRequest.Tree.GetSpec()),
				false,
			)
			if err != nil {
				log.Printf(unableToAddQueryFilterMsg, domain.LogUserInput(err.Error()))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			s, _, err := sr.ElasticSearchService(index.ESClient())
			if err != nil {
				log.Printf(noSearchServiceMsg, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			res, err := s.Do(r.Context())
			if err != nil {
				return
			}
			if res == nil {
				log.Printf(unexpectedResponseMsg, res)
				return
			}
			parents, _, err := decodeFragmentGraphs(res)
			if err != nil {
				return
			}

			for _, parent := range parents {
				parent.Tree.HasChildren = parent.Tree.ChildCount != 0
				if sr.Tree.WithFields {
					parent.Tree.Fields = parent.NewFields(textQuery, c.Config.EAD.TreeFields...)
					parent.Tree.Content = []string{}
				}

				leafs = append(leafs, parent.Tree)
			}
		}

		for _, rec := range records {
			rec.Tree.HasChildren = rec.Tree.ChildCount != 0
			if searchRequest.Tree.WithFields {
				rec.Tree.Fields = rec.NewFields(textQuery, c.Config.EAD.TreeFields...)
				rec.Tree.Content = []string{}
			}

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
				// It is possible the treeQuery leaf is not in the nodeMap because we are in the treePage
				// append/prepend call with nodes only for the requested page.
				if ok {
					result.TreeHeader.ExpandedIDs = fragments.ExpandedIDs(activeNode)
					result.TreeHeader.ActiveID = tq.GetLeaf()
					paging.ResultActive = activeNode.PageEntry()
				}
				if !ok {
					errMsg := fmt.Sprintf("Unable to find node %s in map", tq.GetLeaf())
					log.Println(domain.LogUserInput(errMsg))
				}
			}

			// update paging
			if result.TreeHeader.Paging != nil {
				paging = result.TreeHeader.Paging
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
				page := paging.ResultFirst.CreateTreePage(nodeMap, result.Tree, true, 0)
				result.TreePage = page
				result.Tree = nil
			case "prepend":
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
			rec.ProtoBuf = nil
			rec.Summary = nil

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
		for _, agg := range aggs {
			if agg.Field == "meta.tags" {
				peek := make(map[string]int64)
				for _, facet := range aggs {
					for _, link := range facet.Links {
						peek[link.Value] = link.Count
					}
				}

				result.Peek = peek
			}
		}
		result.Facets = aggs
	}

	// currently only JSON is supported. Add switch when protobuf must be returned
	render.JSON(w, r, result)
	return
}

func GetSearchRecord(ctx context.Context, id string) (*fragments.FragmentGraph, error) {
	orgID := strings.Split(id, "_")[0]

	res, err := index.ESClient().Get().
		Index(config.Config.ElasticSearch.GetIndexName(orgID)).
		Id(id).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf(unexpectedResponseMsg, res)
	}

	if !res.Found {
		return nil, fmt.Errorf("%s was not found", domain.LogUserInput(id))
	}

	return decodeFragmentGraph(res.Source)
}

func getSearchRecord(w http.ResponseWriter, r *http.Request) {
	// TODO(kiivihal): add more like this support to the query
	id := chi.URLParam(r, "id")

	record, err := GetSearchRecord(r.Context(), id)
	if err != nil {
		fmt.Printf("Unable to decode RDFRecord: %#v", err)
		render.JSON(w, r, []string{})
		render.Status(r, 404)
		return
	}

	switch r.URL.Query().Get("itemFormat") {
	case "flat":
		record.NewFields(nil)
		record.Resources = nil
	case "jsonld":
		record.NewJSONLD()
		record.Resources = nil
	case "summary":
		record.NewResultSummary()
		record.Resources = nil
	case "grouped":
		_, err := record.NewGrouped()
		render.Status(r, http.StatusInternalServerError)
		log.Printf("Unable to render grouped resources: %s\n", err.Error())
		render.PlainText(w, r, err.Error())
		return
	}

	switch r.URL.Query().Get("format") {
	case "jsonld":
		entries := []map[string]interface{}{}
		entries = append(entries, record.NewJSONLD()...)

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

func decodeFragmentGraph(hit json.RawMessage) (*fragments.FragmentGraph, error) {
	r := new(fragments.FragmentGraph)
	if err := json.Unmarshal(hit, r); err != nil {
		return nil, err
	}
	return r, nil
}

func decodeResourceEntry(hit json.RawMessage) (*fragments.ResourceEntry, error) {
	re := new(fragments.ResourceEntry)
	if err := json.Unmarshal(hit, re); err != nil {
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
		coll.HitCount = collapseInner.Hits.TotalHits.Value
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
