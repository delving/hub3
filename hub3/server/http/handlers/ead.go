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
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	elastic "github.com/olivere/elastic/v7"
)

func RegisterEAD(r chi.Router) {
	r.Get("/api/ead/search", eadSearch)
	r.Get("/api/ead/search/{spec}", eadInventorySearch)

	// Tree reconstruction endpoint
	r.Get("/api/tree/{spec}", TreeList)
	r.Get("/api/tree/{spec}/{inventoryID:.*$}", TreeList)
	r.Get("/api/tree/{spec}/stats", treeStats)
	r.Get("/api/ead/{spec}/download", EADDownload)
	r.Get("/api/ead/{spec}/mets/{inventoryID}", METSDownload)
	r.Get("/api/ead/{spec}/desc", TreeDescriptionAPI)
}

func NewOldBulkProcessor() *OldBulkProcessor {
	return &OldBulkProcessor{bi: BulkProcessor()}
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
	// FIXME(kiivihal): add logger for
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
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
	searchRequest, err := fragments.NewSearchRequest(r.URL.Query())
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", spec))
	w.Header().Set("Content-Type", "application/pdf")
	return
}

// MetsDownload is a handler that returns a stored METS XML for an inventory.
func METSDownload(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		http.Error(w, "spec cannot be empty", http.StatusBadRequest)
		return
	}
	inventoryID := chi.URLParam(r, "inventoryID")
	if inventoryID == "" {
		http.Error(w, "inventoryID cannot be empty", http.StatusBadRequest)
		return
	}
	eadPath := path.Join(c.Config.EAD.CacheDir, spec, "mets", fmt.Sprintf("%s.xml", inventoryID))
	http.ServeFile(w, r, eadPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.xml", spec, inventoryID))
	w.Header().Set("Content-Type", "application/xml")
	return
}

// EADDownload is a handler that returns a stored XML for an EAD Archive
func EADDownload(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	eadPath := path.Join(c.Config.EAD.CacheDir, spec, fmt.Sprintf("%s.xml", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xml", spec))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	return
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
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	stats, err := fragments.CreateTreeStats(r.Context(), spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
