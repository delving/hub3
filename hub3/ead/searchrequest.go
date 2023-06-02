package ead

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OneOfOne/xxhash"
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"

	cfg "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
)

// SearchRequest holds all information for EAD search
type SearchRequest struct {
	Page             int
	Rows             int
	Query            *elastic.BoolQuery
	RawQuery         string
	Service          *elastic.SearchService
	FacetFields      []string
	ContextIndex     []string
	Filters          []*fragments.QueryFilter
	FacetSize        int
	FacetAndBoolType bool
	SortBy           string
	NestedSortField  string
	SortAsc          bool
	NoCache          bool
	CacheRefresh     bool
	CacheReset       bool
	InventoryID      string
	Explain          bool
	EchoService      bool
	rlog             *zerolog.Logger
	fub              *fragments.FacetURIBuilder
	postFilter       elastic.Query
}

func newSearchRequest(params url.Values) (*SearchRequest, error) {
	sr := &SearchRequest{
		Page:            1,
		Rows:            10,
		NestedSortField: "@value.keyword",
		FacetSize:       50,
		Filters:         []*fragments.QueryFilter{},
	}

	rlog := cfg.Config.Logger.With().
		Str("application", "hub3").
		Str("search.type", "request builder").
		Logger()

	sr.rlog = &rlog

	logConvErr := func(p string, v []string, err error) {
		sr.rlog.Error().Err(err).
			Str("param", p).
			Msgf("unable to convert %v to int", v)
	}

	for p, v := range params {
		switch p {
		case "rows":
			size, err := strconv.Atoi(params.Get(p))
			if err != nil {
				logConvErr(p, v, err)

				return nil, err
			}

			if size > 100 {
				size = 100
			}

			sr.Rows = size
		case "facet.size":
			size, err := strconv.Atoi(params.Get(p))
			if err != nil {
				logConvErr(p, v, err)

				return nil, err
			}

			if size > 100 {
				size = 100
			}

			sr.FacetSize = size
		case "FacetBoolType":
			fbt := params.Get(p)
			if fbt != "" {
				sr.FacetAndBoolType = strings.EqualFold(fbt, "and")
			}
		case "page":
			rawPage, err := strconv.Atoi(params.Get(p))
			if err != nil {
				logConvErr(p, v, err)

				return nil, err
			}

			if rawPage == 0 {
				err := fmt.Errorf("0 pages is not allowed. Paging starts at 1")
				sr.rlog.Error().Err(err).
					Str("param", p).
					Msg("")

				return nil, err
			}

			sr.Page = rawPage
		case "sortBy":
			sortKey := params.Get(p)
			if strings.HasPrefix(sortKey, "^") {
				sr.SortAsc = true
				sortKey = strings.TrimPrefix(sortKey, "^")
			}

			if strings.HasPrefix(sortKey, "int.") {
				sr.NestedSortField = "integer"
				sortKey = strings.TrimPrefix(sortKey, "int.")
			}

			sr.SortBy = sortKey
		case "q", "query":
			sr.RawQuery = params.Get(p)
		case "facet.field":
			sr.FacetFields = v
		case "qf", "qf[]":
			for _, filter := range v {
				qf, err := fragments.NewQueryFilter(filter)
				if err != nil {
					sr.rlog.Error().Err(err).
						Str("param", p).
						Msg("error in filter gerenation")

					return nil, err
				}

				sr.Filters = append(sr.Filters, qf)
			}
		case "qf.dateRange", "qf.dateRange[]":
			for _, filter := range v {
				qf, err := fragments.NewDateRangeFilter(filter)
				if err != nil {
					sr.rlog.Error().Err(err).
						Str("param", p).
						Msg("error in daterange filter gerenation")

					return sr, err
				}

				sr.Filters = append(sr.Filters, qf)
			}
		case "noCache":
			sr.NoCache = strings.EqualFold(params.Get(p), "true")
		case "cacheRefresh":
			sr.CacheRefresh = strings.EqualFold(params.Get(p), "true")
		case "cacheReset":
			sr.CacheReset = strings.EqualFold(params.Get(p), "true")
		}
	}

	return sr, nil
}

func (sr *SearchRequest) SetFragmentURIBuilder() error {
	fub, err := fragments.NewFacetURIBuilder(sr.RawQuery, sr.Filters)
	if err != nil {
		return err
	}

	sr.fub = fub
	return nil
}

func (sr *SearchRequest) SetPostFilter() error {
	if sr.fub == nil {
		if err := sr.SetFragmentURIBuilder(); err != nil {
			return err
		}
	}

	postFilter, err := sr.fub.CreateFacetFilterQuery("", sr.FacetAndBoolType)
	if err != nil {
		sr.rlog.Error().Err(err).Msg("unable to create search postfilter")
		return err
	}

	sr.Service = sr.Service.PostFilter(postFilter)
	sr.postFilter = postFilter

	return nil
}

func (sr *SearchRequest) requestKey() string {
	jsonBytes, err := json.Marshal(sr)
	if err != nil {
		cfg.Config.Logger.Error().Err(err).
			Msg("unable to marshal request key")

		return ""
	}

	hash := xxhash.Checksum64(jsonBytes)

	return fmt.Sprintf("%016x", hash)
}

func (sr *SearchRequest) enableDescriptionSearch() bool {
	for _, f := range sr.Filters {
		if f.SearchLabel != "ead-rdf_periodDesc" {
			return false
		}
	}

	return true
}

func (sr *SearchRequest) Do(ctx context.Context, label string) (*elastic.SearchResult, error) {
	resp, err := sr.Service.Do(ctx)

	queryStart := time.Now()
	sr.rlog.Info().
		Int("status", resp.Status).
		Int64("esTimeInMillis", resp.TookInMillis).
		Dur("duration", time.Since(queryStart)).
		Str("searchType", label).
		Msg("elastic ead cluster search request")

	if err != nil {
		sr.rlog.Error().Err(err).Msg("error in elasticsearch response")
		return nil, err
	}

	return resp, nil
}

func (sr *SearchRequest) createSearchResponse(collapseResponse, aggResponse *elastic.SearchResult) (*SearchResponse, error) {
	return createSearchResponse(sr, collapseResponse, aggResponse)
}
