package ead

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/delving/hub3/hub3/fragments"
	elastic "github.com/olivere/elastic/v7"
)

// SearchResponse contains the EAD Search response.
type SearchResponse struct {
	// ArchiveCount returns the number of collapsed Archives that match the search  query
	ArchiveCount int `json:"archiveCount"`

	// Cursor the location of the first result in the ElasticSearch search response
	Cursor int `json:"cursor"`

	// SelectedArchive in the detail search. When empty it is a generic ead.SearchResponse
	SelectedArchive string `json:"selectedArchive"`

	// CurrentPage the page the result is currently on
	CurrentPage int `json:"currentPage"`

	// HasNext if there is a next Page
	HasNext bool `json:"hasNext"`

	// HasPrevious - if there is a previous page
	HasPrevious bool `json:"hasPrevious"`

	// PreviousPage in the search response
	PreviousPage int

	// NextPage in the search response
	NextPage int

	// TotalPages is the total number of pages in the search response
	TotalPages int `json:"totalPages"`

	// TotalClevelCount returns the total number of clevel that mathc the search query
	// this counts is per clevel, so multiple hits inside a clevel are counted as one
	TotalClevelCount int `json:"totalClevelCount"`

	// TotalDescriptionCount returns the total number of hits in the description.
	// This is an cardinatility aggregation so each hit inside the decription counts as a hit.
	TotalDescriptionCount int `json:"totalDescriptionCount"`

	// TotalHits is a combination of TotalClevelCount and TotalDescriptiontCount.
	TotalHits int `json:"totalHits"`

	// Archives contains the list of archives from the response constrained by the search pagination
	Archives []Archive `json:"archives"`

	// CLevels contains a paged result of the cLevels for a specific archive that match the search query
	// It is ordered by the ead orderKey.
	CLevels []CLevelEntry `json:"cLevels,omitempty"`

	// Facets holds the QueryFacets for filtering
	Facets []*fragments.QueryFacet `json:"facets,omitempty"`

	// Explain response from elasticsearch
	Explain *elastic.SearchResult `json:"explain,omitempty"`

	// Service is the elasticsearch query
	Service interface{} `json:"service,omitempty"`
}

// CLevel holds the search results per clevel entry in the an EAD Archive.
type CLevelEntry struct {
	// Path is the unique key to the path of the clevel in the archive tree
	Path string `json:"path"`

	// UnitID is the identifier of the clevel
	UnitID string `json:"unitID"`

	// Label is the title of the clevel
	Label string `json:"label"`

	// HubID is the unique identifier of the clevel as stored in the hub3 index
	HubID string `json:"hubID"`

	// ResultOrder is the place the search result has in the total list of results.
	// This can be used to aid the search pagination on the Archive result page.
	ResultOrder uint64 `json:"sortKey"`
}

// Archive holds all information for the EAD search results that are grouped
// by inventoryID. This is the EadID from the EAD header.
type Archive struct {
	InventoryID      string   `json:"inventoryID"`
	Title            string   `json:"title"`
	Period           []string `json:"period"`
	CLevelCount      int      `json:"cLevelCount"`
	DescriptionCount int      `json:"descriptionCount"`
	Files            string   `json:"files,omitempty"`
	Length           string   `json:"length,omitempty"`
	Abstract         []string `json:"abstract,omitempty"`
	Material         string   `json:"material,omitempty"`
	Language         string   `json:"language,omitempty"`
	Origin           []string `json:"origin,omitempty"`
	MetsFiles        int      `json:"metsFiles,omitempty"`
	ClevelsTotal     int      `json:"clevelsTotal"`
}

func createSearchResponse(sr *SearchRequest, collapseResponse, aggResponse *elastic.SearchResult) (*SearchResponse, error) {
	eadResponse := &SearchResponse{
		Archives: []Archive{},
	}

	if sr.Explain {
		eadResponse.Explain = collapseResponse
	}

	if sr.EchoService {
		ss := reflect.ValueOf(sr.Service).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)

		srcMap, sourceErr := src.Source()
		if sourceErr != nil {
			sr.rlog.Error().Err(sourceErr).
				Msg("unable to decode elastich search request")

			return nil, sourceErr
		}

		eadResponse.Service = srcMap
	}

	// set total description count
	unFilteredEadTypeCount, ok := aggResponse.Aggregations.Terms("noFiltTypeCount")
	if ok {
		for _, b := range unFilteredEadTypeCount.Buckets {
			if b.Key == "eadDesc" {
				eadResponse.TotalDescriptionCount = int(b.DocCount)
			}
		}
	}

	// set counts for number of archives and inventories (clevels) returned
	filteredAgg, ok := aggResponse.Aggregations.Filter("counts")
	if ok {
		specCount, ok := filteredAgg.Aggregations.Cardinality("specCount")
		if ok {
			eadResponse.ArchiveCount = int(*specCount.Value)
			eadResponse.TotalPages = getPageCount(eadResponse.ArchiveCount, sr.Rows)
		}

		eadTypeCount, ok := filteredAgg.Aggregations.Terms("typeCount")
		if ok {
			for _, b := range eadTypeCount.Buckets {
				if b.Key == "ead" {
					eadResponse.TotalClevelCount = int(b.DocCount)
				}
			}
		}
	}

	// total unique search hits in the response
	eadResponse.TotalHits = eadResponse.TotalClevelCount + eadResponse.TotalDescriptionCount

	cursor := getCursor(sr.Rows, sr.Page)
	if cursor > eadResponse.ArchiveCount {
		pageErr := fmt.Errorf(
			"page start %d requested is greater then records returned: %d",
			cursor,
			eadResponse.ArchiveCount,
		)
		sr.rlog.Error().Err(pageErr).
			Msg("request error")

		return nil, pageErr
	}

	eadResponse.Cursor = cursor

	if !sr.enableDescriptionSearch() {
		eadResponse.TotalDescriptionCount = 0
	}

	// build facets
	aggs, err := fragments.DecodeFacets(aggResponse, sr.fub)
	if err != nil {
		sr.rlog.Error().Err(err).
			Msg("facet decode error")

		return nil, err
	}

	eadResponse.Facets = aggs

	if !sr.enableDescriptionSearch() {
		eadResponse.TotalDescriptionCount = 0
	}

	return eadResponse, nil
}

func decodeHits(sr *SearchRequest, hits *elastic.SearchHits) ([]Archive, error) {
	archives := []Archive{}

	var err error

	for _, hit := range hits.Hits {
		fields, ok := hit.Fields[specField]
		if ok {
			spec := fields.([]interface{})[0].(string)

			meta, metaErr := GetMeta(spec)
			if metaErr != nil {
				sr.rlog.Error().Err(metaErr).
					Str("spec", spec).
					Msg("unable to ead meta information")

				return nil, metaErr
			}

			archive := Archive{
				InventoryID:      spec,
				Title:            meta.Label,
				Period:           meta.Period,
				DescriptionCount: 0,
				ClevelsTotal:     meta.Inventories,
			}

			var hitHasDescription bool

			inner, ok := hit.InnerHits["collapse"]
			if ok {
				archive.CLevelCount = int(inner.Hits.TotalHits.Value)

				if len(inner.Hits.Hits) > 0 {
					r := new(fragments.FragmentGraph)

					if unmarshallErr := json.Unmarshal(inner.Hits.Hits[0].Source, r); unmarshallErr != nil {
						sr.rlog.Error().Err(unmarshallErr).
							Msg("unable to unmarshal json for elasticsearch hit")

						return nil, unmarshallErr
					}

					if r.Tree.InventoryID != "" {
						archive.CLevelCount--
						archive.DescriptionCount = 1
						hitHasDescription = true
					}
				}
			}

			if sr.RawQuery != "" && sr.enableDescriptionSearch() && hitHasDescription {
				archive.DescriptionCount, err = GetDescriptionCount(spec, sr.RawQuery)
				if err != nil {
					return nil, err
				}
			}

			archives = append(archives, archive)
		}
	}

	return archives, nil
}
