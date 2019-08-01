package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterEAD(r chi.Router) {

	// EAD endpoint
	r.Post("/api/ead", eadUpload)
	r.Get("/api/ead/{hubID}", eadManifest)

	// Tree reconstruction endpoint
	r.Get("/api/tree/{spec}", TreeList)
	r.Get("/api/tree/{spec}/{nodeID:.*$}", TreeList)
	r.Get("/api/tree/{spec}/stats", treeStats)
	r.Get("/api/tree/{spec}/desc", TreeDescription)
	r.Get("/api/ead/{spec}/download", EADDownload)
	r.Get("/api/ead/{spec}/desc", TreeDescriptionApi)
	r.Get("/api/ead/desc-test", descTest)
}

func eadUpload(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")

	_, err := ead.ProcessUpload(r, w, spec, BulkProcessor())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func TreeList(w http.ResponseWriter, r *http.Request) {
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
	nodeID := chi.URLParam(r, "nodeID")
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
	searchRequest.AddQueryFilter(fmt.Sprintf("%s:%s", c.Config.ElasticSearch.SpecKey, spec), false)
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
	eadPath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s.pdf", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", spec))
	w.Header().Set("Content-Type", "application/pdf")
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
	eadPath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s.xml", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xml", spec))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	return
}

func TreeDescriptionApi(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	description := filepath.Join(
		c.Config.EAD.CacheDir,
		fmt.Sprintf("%s.json", spec),
	)

	params := r.URL.Query()
	var start int
	var end int
	var query string
	var echo string
	var err error

	for k, _ := range params {
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
		case "query":
			query = params.Get(k)
		case "echo":
			echo = params.Get(k)
		}
	}

	b, err := ioutil.ReadFile(description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply search
	if query != "" {
		dc := ead.NewDescriptionCounter(b)
		text, replaced, hits, err := dc.HighlightQuery(query, b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if echo == "hits" {
			log.Printf("hits: %#v", hits)
			render.JSON(w, r, hits)
			return
		}

		if replaced != 0 {
			b = text
		}
	}

	var desc ead.Description
	json.Unmarshal(b, &desc)

	if start != 0 || end != 0 {
		if end != 0 {
			if end >= desc.NrItems {
				http.Error(w, "end is out of bounds", http.StatusBadRequest)
				return
			}
			desc.Item = desc.Item[start:end]
		} else {
			desc.Item = desc.Item[start:]
		}
	}
	render.JSON(w, r, desc)

	return
}

func eadManifest(w http.ResponseWriter, r *http.Request) {
	hubID := chi.URLParam(r, "hubID")
	parts := strings.Split(hubID, "_")
	if len(parts) != 3 {
		http.Error(w, fmt.Sprintf("badly formatted hubID: %v", hubID), http.StatusBadRequest)
		return
	}
	spec := parts[1]
	ds, err := models.GetDataSet(spec)
	if err != nil {
		http.Error(w, fmt.Sprintf("dataset not found: %v", spec), http.StatusNotFound)
		return
	}

	if ds.Label == "" {
		http.Error(w, fmt.Sprintf("dataset is not an archive: %v", spec), http.StatusBadRequest)
		return
	}

	treeNode, err := fragments.TreeNode(r.Context(), hubID)
	if err != nil || treeNode == nil {
		//if err != nil {
		http.Error(w, fmt.Sprintf("hubID %v not found", hubID), http.StatusNotFound)
		return
	}
	log.Println(treeNode)

	manifest := &ead.Manifest{}
	manifest.InventoryID = ds.Spec
	manifest.ArchiveName = ds.Label
	manifest.UnitID = treeNode.UnitID
	manifest.UnitTitle = treeNode.Label

	render.JSON(w, r, manifest)

	return
}

func TreeDescription(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	//ds, err := models.GetDataSet(spec)
	//if err != nil {
	//render.Status(r, http.StatusNotFound)
	//render.JSON(w, r, APIErrorMessage{
	//HTTPStatus: http.StatusNotFound,
	//Message:    fmt.Sprintln("archive not found"),
	//Error:      nil,
	//})
	//return
	//}

	description := filepath.Join(
		c.Config.EAD.CacheDir,
		fmt.Sprintf("%s.json", spec),
	)

	http.ServeFile(w, r, description)

	return
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

func descTest(w http.ResponseWriter, r *http.Request) {
	archive, err := ead.ReadEAD("hub3/ead/test_data/1.04.02_ead_header.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	desc, err := ead.NewDescription(archive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, desc)
	return
}
