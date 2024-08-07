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
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	"github.com/go-chi/chi"
	elastic "github.com/olivere/elastic/v7"
)

var (
	contentDispositionKey = "Content-Disposition"
	contentTypeKey        = "Content-Type"
)

func RegisterEAD(r chi.Router) {
	r.Get("/api/ead/search", eadSearch)
	r.Get("/api/ead/search/{spec}", eadInventorySearch)

	// Tree reconstruction endpoint
	r.Get("/api/tree/{spec}", TreeList)
	r.Get("/api/tree/{spec}/{inventoryID:.*$}", TreeList)
	r.Get("/api/tree/{spec}/stats", treeStats)
	r.Get("/api/ead/{spec}/download", EADDownload)
	r.Get("/api/ead/{spec}/desc", TreeDescriptionAPI)
	r.Get("/api/ead/{spec}/desc/index", TreeDescriptionSearch)
	r.Get("/api/ead/{spec}/meta", EADMeta)
}

func NewOldBulkProcessor() *OldBulkProcessor {
	return &OldBulkProcessor{bi: nil}
}

type OldBulkProcessor struct {
	bi *elastic.BulkProcessor
}

func (bp OldBulkProcessor) Publish(ctx context.Context, msg ...*domainpb.IndexMessage) error {
	for _, m := range msg {
		r := elastic.NewBulkIndexRequest().
			Index(m.GetIndexName()).
			RetryOnConflict(3).
			Id(m.GetRecordID()).
			Doc(fmt.Sprintf("%s", m.GetSource()))

		bp.bi.Add(r)
	}

	return nil
}

func TreeList(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Error(w, r, fmt.Errorf(emptySpecMsg()), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	page := r.URL.Query().Get("page")
	if page != "" {
		q := r.URL.Query()
		q.Del("page")

		q.Add("treePage", page)

		r.URL.RawQuery = q.Encode()
	}

	nodeID := chi.URLParam(r, "inventoryID")
	if nodeID != "" {
		id, err := url.QueryUnescape(nodeID)
		if err != nil {
			log.Println("Unable to unescape QueryParameters.")
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, err.Error())
			return
		}

		q := r.URL.Query()
		isPaging := q.Get("paging") == "true"
		if isPaging {
			q.Add("byUnitID", id)
		} else {
			q.Add("byLeaf", id)
		}

		r.URL.RawQuery = q.Encode()
	}

	searchRequest, err := fragments.NewSearchRequest(orgID.String(), r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}
	searchRequest.ItemFormat = fragments.ItemFormatType_TREE
	err = searchRequest.AddQueryFilter(fmt.Sprintf("%s:%s", c.Config.ElasticSearch.SpecKey, spec), false)
	if err != nil {
		log.Println("Unable to add QueryFilter")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch searchRequest.Tree {
	case nil:
		searchRequest.Tree = &fragments.TreeQuery{
			Depth: []string{"1", "2"},
			Spec:  spec,
		}
	default:
		searchRequest.Tree.Spec = spec
	}
	ProcessSearchRequest(w, r, searchRequest)
	return
}

// PDFDownload is a handler that returns a stored PDF for an EAD Archive
func PDFDownload(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		http.Error(w, "spec cannot be empty", http.StatusBadRequest)
		return
	}
	eadPath := path.Join(c.Config.EAD.CacheDir, spec, fmt.Sprintf("%s.pdf", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set(contentDispositionKey, fmt.Sprintf("attachment; filename=%s.pdf", spec))
	w.Header().Set(contentTypeKey, "application/pdf")
	return
}

// EADDownload is a handler that returns a stored XML for an EAD Archive
func EADDownload(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Error(w, r, fmt.Errorf(emptySpecMsg()), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})
		return
	}
	eadPath := path.Join(c.Config.EAD.CacheDir, spec, fmt.Sprintf("%s.xml", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set(contentDispositionKey, fmt.Sprintf("attachment; filename=%s.xml", spec))
	w.Header().Set(contentTypeKey, r.Header.Get(contentTypeKey))
	return
}

func EADMeta(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	err := ead.ValidateSpec(spec)
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}
	meta, err := ead.GetMeta(spec)
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
		})
		return
	}

	render.JSON(w, r, meta)
}

func TreeDescriptionSearch(w http.ResponseWriter, r *http.Request) {
	var hits int
	rawQuery := r.URL.Query().Get("q")
	if rawQuery == "" {
		render.JSON(w, r, map[string]int{"total": hits})
		return
	}

	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Error(w, r, fmt.Errorf(emptySpecMsg()), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	descriptionIndex, getErr := ead.GetDescriptionIndex(spec)
	if getErr != nil && !errors.Is(getErr, ead.ErrNoDescriptionIndex) {
		render.Error(w, r, getErr, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
			Message:    "error with retrieving description index",
		})
		return
	}

	if descriptionIndex != nil {
		searhHits, searchErr := descriptionIndex.SearchWithString(rawQuery)
		if searchErr != nil && !errors.Is(searchErr, memory.ErrSearchNoMatch) {
			c.Config.Logger.Error().Err(searchErr).
				Str("subquery", "description").
				Msg("unable to search description")

			http.Error(w, searchErr.Error(), http.StatusNotFound)
			return
		}

		hits = searhHits.Total()
	}

	render.JSON(w, r, map[string]int{"total": hits})
}

func TreeDescriptionAPI(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")

	params := r.URL.Query()

	var (
		start  int
		end    int
		query  string
		echo   string
		err    error
		filter bool
	)

	for k := range params {
		switch k {
		case "start":
			start, err = strconv.Atoi(params.Get(k))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "end":
			end, err = strconv.Atoi(params.Get(k))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "query", "q":
			query = params.Get(k)
		case "echo":
			echo = params.Get(k)
		case "filter":
			filter = strings.EqualFold(params.Get(k), "true")
		}
	}

	if end != 0 && start > end {
		http.Error(w, "Start cannot be greater than end", http.StatusBadRequest)
		return
	}

	var searchHits int

	desc, err := ead.GetDescription(spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if query != "" {
		descIndex, getErr := ead.GetDescriptionIndex(spec)
		if getErr != nil {
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}

		hits, searchErr := descIndex.SearchWithString(query)
		if searchErr != nil && !errors.Is(searchErr, memory.ErrSearchNoMatch) {
			http.Error(w, searchErr.Error(), http.StatusInternalServerError)
			return
		}

		desc.Item = descIndex.HighlightMatches(hits, desc.Item, filter)

		if echo == "hits" {
			render.JSON(w, r, hits.TermFrequency())
			return
		}

		// TODO(kiivihal): should we implement search and highlighting for summary
		// desc.Summary = dq.HightlightSummary(desc.Summary)
		searchHits = hits.Total()
		desc.NrItems = len(desc.Item)

		if filter {
			desc.NrSections = 0
			desc.Section = []*ead.SectionInfo{}
		}
	}

	desc.NrHits = searchHits

	if start != 0 || end != 0 {
		if end != 0 {
			if end >= desc.NrItems {
				end = desc.NrItems
			}
			desc.Item = desc.Item[start:end]
		} else {
			desc.Item = desc.Item[start:]
		}
	}
	render.JSON(w, r, desc)
}

func treeStats(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Error(w, r, fmt.Errorf(emptySpecMsg()), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}
	stats, err := fragments.CreateTreeStats(r.Context(), string(orgID), spec)
	if err != nil {
		render.Error(w, r, fmt.Errorf(emptySpecMsg()), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})
		return
	}
	// todo return 404 if stats.Leafs == 0
	if stats.Leafs == 0 {
		render.Status(r, http.StatusNotFound)
		return
	}
	render.JSON(w, r, stats)
	return
}

func eadSearch(w http.ResponseWriter, r *http.Request) {
	resp, err := ead.PerformClusteredSearch(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, resp)
}

func eadInventorySearch(w http.ResponseWriter, r *http.Request) {
	resp, err := ead.PerformDetailSearch(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, resp)
}

func emptySpecMsg() string {
	return fmt.Sprintln("spec can't be empty")
}
