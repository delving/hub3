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

package ead

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/go-chi/chi/v5"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/hlog"

	cfg "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/search"
)

const (
	metaTags       = "meta.tags"
	specField      = "meta.spec"
	trueParamValue = "true"
)

var once sync.Once

func buildSearchRequest(r *http.Request, includeDescription bool) (*SearchRequest, error) {
	requestID, _ := hlog.IDFromRequest(r)
	rlog := cfg.Config.Logger.With().
		Str("req_id", requestID.String()).
		Str("searchType", "ead cluster search").
		Logger()

	client := index.ESClient()
	orgID := domain.GetOrganizationID(r)

	s := client.Search(cfg.Config.ElasticSearch.GetIndexName(orgID.String())).
		TrackTotalHits(cfg.Config.ElasticSearch.TrackTotalHits)

	sr, err := newSearchRequest(r.URL.Query())
	if err != nil {
		rlog.Error().Err(err).
			Msg("unable to create ead.SearchRequest")

		return nil, err
	}

	slog.Info("ead search request", "sr", sr, "filters", sr.Filters)

	tagQuery := elastic.NewBoolQuery().Should(elastic.NewTermQuery(metaTags, "ead"))
	if includeDescription && sr.enableDescriptionSearch() {
		tagQuery = tagQuery.Should(elastic.NewTermQuery(metaTags, "eadDesc"))
	}

	query := elastic.NewBoolQuery()
	query = query.Must(tagQuery)

	if sr.RawQuery != "" {
		// TODO(kiivihal): replace querystring below with search.QueryTerm
		q, err := fragments.QueryFromSearchFields(sr.RawQuery, cfg.Config.EAD.SearchFields...)
		if err != nil {
			return sr, err
		}

		query = query.Must(q)
	}

	if r.URL.Query().Get("explain") == trueParamValue {
		s = s.Explain(true)
		sr.Explain = true
	}

	if r.URL.Query().Get("service") == trueParamValue {
		sr.EchoService = true
	}

	sr.Query = query

	slog.Info("elasticsearch query", "query", query)

	sr.Service = s

	return sr, nil
}

func buildCollapseRequest(r *http.Request) (*SearchRequest, error) {
	req, requestErr := buildSearchRequest(r, true)
	if requestErr != nil {
		if req.rlog != nil {
			req.rlog.Error().Err(requestErr).
				Msg("performClusteredSearch error")
		}

		return nil, requestErr
	}

	s := req.Service

	b := elastic.NewCollapseBuilder(specField).
		InnerHit(
			elastic.NewInnerHit().
				Name("collapse").
				Size(1).
				Sort("tree.inventoryID", true),
		).
		MaxConcurrentGroupRequests(4)
	s = s.Collapse(b)
	s = s.FetchSource(false)

	if req.SortBy != "" {
		switch key := req.SortBy; {
		case req.SortBy == "_score":
			s = s.Sort(req.SortBy, req.SortAsc)
		case strings.Contains(key, "_"):
			path := fmt.Sprintf("resources.entries.%s", req.NestedSortField)
			fieldSort := elastic.NewFieldSort(path).
				Order(req.SortAsc).
				Nested(
					elastic.NewNestedSort("resources.entries").
						Filter(
							elastic.NewTermQuery("resources.entries.searchLabel", key),
						),
				)
			s = s.SortBy(fieldSort)
		default:
			s = s.Sort(req.SortBy, req.SortAsc)
		}
	}

	s = s.Size(req.Rows)
	if req.Page > 1 {
		s = s.From(getCursor(req.Rows, req.Page))
	}

	s = s.Query(req.Query)

	req.Service = s

	return req, nil
}

func buildFacetRequest(r *http.Request) (*SearchRequest, error) {
	req, requestErr := buildSearchRequest(r, true)
	if requestErr != nil {
		if req.rlog != nil {
			req.rlog.Error().Err(requestErr).
				Msg("performClusteredSearch error")
		}

		return nil, requestErr
	}

	req.FacetFields = append(
		req.FacetFields,
		[]string{
			"tree.hasDigitalObject",
			"tree.mimeType",
			"ead-rdf_genreform",
		}...,
	)

	if err := req.SetFragmentURIBuilder(); err != nil {
		return nil, err
	}

	for _, facetField := range req.FacetFields {
		if facetField == "" {
			continue
		}

		ff, facetErr := fragments.NewFacetField(facetField)
		if facetErr != nil {
			return nil, facetErr
		}

		agg, facetErr := fragments.CreateAggregationBySearchLabel("resources.entries", ff, req.FacetAndBoolType, req.fub)
		if facetErr != nil {
			return nil, facetErr
		}

		req.Service = req.Service.Aggregation(facetField, agg)
	}

	if err := req.SetPostFilter(); err != nil {
		return nil, err
	}
	// spec count aggregation
	specCountAgg := elastic.NewCardinalityAggregation().
		Field(specField)

	eadTypeCountAgg := elastic.NewTermsAggregation().
		Field(metaTags)

	countFilterAgg := elastic.NewFilterAggregation().
		Filter(req.postFilter).
		SubAggregation("specCount", specCountAgg).
		SubAggregation("typeCount", eadTypeCountAgg)

	req.Service = req.Service.
		Query(req.Query).
		Aggregation("counts", countFilterAgg).
		Aggregation("noFiltTypeCount", eadTypeCountAgg).
		Size(0)

	return req, nil
}

func PerformClusteredSearch(r *http.Request) (*SearchResponse, error) {
	once.Do(newBigCache)

	searchRequest, err := buildCollapseRequest(r)
	if err != nil {
		searchRequest.rlog.Error().Err(err).Msg("unable to build collapse request")
		return nil, err
	}

	if searchRequest.CacheReset {
		newBigCache()
		// already cache this request
		searchRequest.CacheReset = false
	}

	requestKey := searchRequest.requestKey()
	searchRequest.rlog.Debug().
		Str("cache_key", requestKey).
		Msg("generating cache request key")

	if httpCache != nil && requestKey != "" && !searchRequest.NoCache && !searchRequest.CacheRefresh {
		response := getCachedRequest(requestKey, searchRequest.rlog)
		if response != nil {
			return response, nil
		}
	}

	if searchRequest.CacheRefresh {
		searchRequest.CacheRefresh = false
		requestKey = searchRequest.requestKey()
	}

	aggRequest, err := buildFacetRequest(r)
	if err != nil {
		aggRequest.rlog.Error().Err(err).Msg("unable to build aggregation request")
		return nil, err
	}

	aggResponse, err := aggRequest.Do(r.Context(), "agg")
	if err != nil {
		return nil, err
	}

	searchResponse, err := searchRequest.Do(r.Context(), "search")
	if err != nil {
		return nil, err
	}

	eadResponse, err := aggRequest.createSearchResponse(searchResponse, aggResponse)
	if err != nil {
		aggRequest.rlog.Error().Err(err).
			Msg("error in building the response")

		return nil, err
	}

	eadResponse.Archives, err = decodeHits(searchRequest, searchResponse.Hits)
	if err != nil {
		return nil, fmt.Errorf("unable to decode elasticsearch hits: %w", err)
	}

	if httpCache != nil && requestKey != "" && !searchRequest.NoCache {
		// don't cache no results
		if eadResponse.TotalHits == 0 {
			return eadResponse, nil
		}

		storeResponseInCache(requestKey, eadResponse, searchRequest.rlog)
	}

	eadResponse.IsSearch = searchRequest.IsSearch()

	return eadResponse, nil
}

func getPageNumber(cursor, totalPages, rowSize int) int {
	if cursor == 0 || totalPages == 0 {
		return 1
	}

	return (cursor / rowSize) + 1
}

func PerformDetailSearch(r *http.Request) (*SearchResponse, error) {
	once.Do(newBigCache)

	inventoryID := chi.URLParam(r, "spec")

	requestID, _ := hlog.IDFromRequest(r)
	rlog := cfg.Config.Logger.With().
		Str("req_id", requestID.String()).
		Str("searchType", "ead detail search").
		Str("inventoryID", inventoryID).
		Logger()

	req, err := buildSearchRequest(r, false)
	if err != nil {
		rlog.Error().Err(err).
			Msg("EAD detail error")

		return nil, err
	}

	req.InventoryID = inventoryID

	if req.CacheReset {
		newBigCache()
		// already cache this request
		req.CacheReset = false
	}

	requestKey := req.requestKey()
	rlog.Debug().
		Str("cache_key", requestKey).
		Msg("generating cache request key")

	if httpCache != nil && requestKey != "" && !req.NoCache && !req.CacheRefresh {
		response := getCachedRequest(requestKey, &rlog)
		if response != nil {
			return response, nil
		}
	}

	if req.CacheRefresh {
		req.CacheRefresh = false
		requestKey = req.requestKey()
	}

	s := req.Service
	s = s.Size(req.Rows)
	if req.Page > 1 {
		s = s.From(getCursor(req.Rows, req.Page))
	}

	query := req.Query

	query = query.Must(elastic.NewTermQuery(specField, inventoryID))

	// only return the tree part of the search response
	fsc := elastic.NewFetchSourceContext(true)
	fsc.Include("tree")
	s = s.FetchSourceContext(fsc)

	postFilter := elastic.NewBoolQuery()

	for _, qf := range req.Filters {
		switch {
		case strings.HasPrefix(qf.SearchLabel, "tree."):
			postFilter = postFilter.Must(elastic.NewTermQuery(qf.SearchLabel, qf.Value))
		default:
			f, filterErr := qf.ElasticFilter()
			if filterErr != nil {
				return nil, filterErr
			}

			if qf.Exclude {
				postFilter = postFilter.MustNot(f)
				continue
			}

			postFilter = postFilter.Must(f)
		}
	}

	query = query.Must(postFilter)
	req.Query = query
	slog.Debug("ead detail search query", "req", req, "query", query)

	resp, err := s.
		Query(query).
		Sort("tree.sortKey", true).
		Do(r.Context())
	if err != nil {
		rlog.Error().Err(err).
			Msg("error in elasticsearch response")
		return nil, err
	}

	eadResponse := &SearchResponse{
		Archives:         []Archive{},
		TotalClevelCount: int(resp.TotalHits()),
	}

	if r.URL.Query().Get("explain") == trueParamValue {
		eadResponse.Explain = resp
		ss := reflect.ValueOf(s).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)

		srcMap, err := src.Source()
		if err != nil {
			rlog.Error().Err(err).
				Msg("unable to decode elastich search request")

			return nil, err
		}

		eadResponse.Service = srcMap
	}

	eadResponse.TotalHits = eadResponse.TotalClevelCount + eadResponse.TotalDescriptionCount

	if eadResponse.TotalHits > 0 {
		eadResponse.ArchiveCount = 1
		eadResponse.TotalPages = getPageCount(eadResponse.TotalClevelCount, req.Rows)
	}

	cursor := getCursor(req.Rows, req.Page)
	if cursor > eadResponse.TotalClevelCount {
		err := fmt.Errorf(
			"page start %d requested is greater then records returned: %d",
			cursor,
			eadResponse.ArchiveCount,
		)
		rlog.Error().Err(err).
			Msg("request error")

		return nil, err
	}

	eadResponse.Cursor = cursor

	if resp == nil || resp.TotalHits() == 0 {
		return eadResponse, nil
	}

	var pageErr error
	eadResponse.Pagination, pageErr = search.NewPaginator(eadResponse.ArchiveCount, req.Rows, getPageNumber(cursor, eadResponse.TotalPages, req.Rows), cursor)
	if pageErr != nil {
		return nil, fmt.Errorf("unable to create paginator; %w", pageErr)
	}

	eadResponse.CurrentPage = getPageNumber(cursor, eadResponse.TotalPages, req.Rows)
	eadResponse.HasPrevious = eadResponse.CurrentPage > 1
	eadResponse.HasNext = (eadResponse.CurrentPage + 1) < eadResponse.TotalPages
	eadResponse.SelectedArchive = inventoryID
	if eadResponse.HasPrevious {
		eadResponse.PreviousPage = eadResponse.CurrentPage - 1
	}
	if eadResponse.HasNext {
		eadResponse.NextPage = eadResponse.CurrentPage + 1
	}

	eadResponse.CLevels = []CLevelEntry{}

	for _, hit := range resp.Hits.Hits {
		r := new(fragments.FragmentGraph)
		if err := json.Unmarshal(hit.Source, r); err != nil {
			return nil, err
		}

		tree := r.Tree

		cLevel := CLevelEntry{
			UnitID:      tree.UnitID,
			Label:       tree.Label,
			HubID:       tree.HubID,
			ResultOrder: tree.SortKey,
			Path:        tree.CLevel,
		}

		eadResponse.CLevels = append(eadResponse.CLevels, cLevel)
	}

	if httpCache != nil && requestKey != "" && !req.NoCache {
		// don't cache no results
		if eadResponse.TotalHits == 0 {
			return eadResponse, nil
		}

		storeResponseInCache(requestKey, eadResponse, &rlog)
	}

	return eadResponse, nil
}

func getCursor(rows, page int) int {
	if page == 1 || rows == 0 {
		return 0
	}

	start := ((page - 1) * rows)
	if start < 1 {
		return 0
	}

	return start
}

func getPageCount(archives, rows int) int {
	if rows == 0 || archives == 0 {
		return 0
	}

	if archives < rows {
		return 1
	}

	pages := archives / rows
	if archives%rows != 0 {
		pages++
	}

	return pages
}

// isAdvancedSearch checks if the query contains Lucene QueryString
// advanced search query syntax.
func isAdvancedSearch(query string) bool {
	parts := strings.Fields(query)
	for _, p := range parts {
		switch {
		case p == "AND":
			return true
		case p == "OR":
			return true
		case p == "NOT":
			return true
		case strings.HasPrefix(p, "-"):
			return true
		case strings.HasPrefix(p, "+"):
			return true
		case strings.HasPrefix(p, "\""):
			return true
		case strings.HasSuffix(p, "\""):
			return true
		}
	}

	return false
}
