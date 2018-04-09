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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3"
	"github.com/delving/rapid-saas/hub3/fragments"
	"github.com/delving/rapid-saas/hub3/harvesting"
	"github.com/delving/rapid-saas/hub3/index"
	"github.com/delving/rapid-saas/hub3/models"
	"github.com/gammazero/workerpool"
	elastic "github.com/olivere/elastic"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/kiivihal/goharvest/oai"
)

var bp *elastic.BulkProcessor
var wp *workerpool.WorkerPool
var ctx context.Context

func init() {
	var err error
	ctx = context.Background()
	bps := index.CreateBulkProcessorService()
	bp, err = bps.Do(ctx)
	if err != nil {
		log.Fatalf("Unable to start BulkProcessor: ", err)
	}
	wp = workerpool.New(10)
}

// APIErrorMessage contains the default API error messages
type APIErrorMessage struct {
	HTTPStatus int    `json:"code"`
	Message    string `json:"type"`
	Error      error  `json:error`
}

// NewSingleFinalPathHostReverseProxy proxies QueryString of the request url to the target url
func NewSingleFinalPathHostReverseProxy(target *url.URL, relPath string) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + relPath
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func bulkAPI(w http.ResponseWriter, r *http.Request) {
	response, err := hub3.ReadActions(ctx, r.Body, bp, wp)
	if err != nil {
		log.Println("Unable to read actions")
		errR := ErrRender(err)
		// todo fix errr renderer for better narthex consumption.
		_ = errR.Render(w, r)
		//render.Render(w, r, rrRender(err))
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
	return
}

// bindPMHRequest the query parameters to the OAI-Request
func bindPMHRequest(r *http.Request) oai.Request {
	baseURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	q := r.URL.Query()
	req := oai.Request{
		Verb:            q.Get("verb"),
		MetadataPrefix:  q.Get("metadataPrefix"),
		Set:             q.Get("set"),
		From:            q.Get("from"),
		Until:           q.Get("until"),
		Identifier:      q.Get("identifier"),
		ResumptionToken: q.Get("resumptionToken"),
		BaseURL:         baseURL,
	}
	return req
}

// oaiPmhEndpoint processed OAI-PMH request and returns the results
func oaiPmhEndpoint(w http.ResponseWriter, r *http.Request) {
	req := bindPMHRequest(r)
	log.Println(req)
	resp := harvesting.ProcessVerb(&req)
	render.XML(w, r, resp)
}

// listFragments returns a list of matching fragments
// See for more info: http://linkeddatafragments.org/
func listFragments(w http.ResponseWriter, r *http.Request) {
	fr := fragments.NewFragmentRequest()
	spec := chi.URLParam(r, "spec")
	if spec != "" {
		fr.Spec = spec
	}
	err := fr.ParseQueryString(r.URL.Query())
	if err != nil {
		log.Printf("Unable to list fragments because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprint("Unable to list fragments was not found"),
			Error:      err,
		})
		return
	}
	frags, err := fr.Find(ctx, index.ESClient())
	if err != nil || frags.Len() == 0 {
		log.Printf("Unable to list fragments because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusNotFound,
			Message:    fmt.Sprint("No fragmenst for query were found."),
			Error:      err,
		})
		return
	}
	w.Header().Set("Content-Type", "text/turtle")
	err = frags.Serialize(w, "text/turtle")
	if err != nil {
		log.Printf("Unable to list serialize fragments because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusNotFound,
			Message:    fmt.Sprintf("Unable to serialize fragments: %s", err),
			Error:      err,
		})
		return

	}
	return
}

// listDataSets returns a list of all public datasets
func listDataSets(w http.ResponseWriter, r *http.Request) {
	sets, err := models.ListDataSets()
	if err != nil {
		log.Printf("Unable to list datasets because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusInternalServerError,
			Message:    fmt.Sprint("Unable to list datasets was not found"),
			Error:      err,
		})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, sets)
	return
}

// getDataSetStats returns a dataset when found or a 404
func getDataSetStats(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	log.Printf("Get stats for spec %s", spec)
	stats, err := models.CreateDataSetStats(ctx, spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", chi.URLParam(r, "spec")),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Println("Unable to create dataset stats: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return
	}
	render.JSON(w, r, stats)
	return

}

// getDataSet returns a dataset when found or a 404
func getDataSet(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	ds, err := models.GetDataSet(spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", spec),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Println("Unable to get dataset: %s", spec)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return

	}
	render.JSON(w, r, ds)
	return
}

func deleteDataset(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		render.Status(r, http.StatusNotFound)
		log.Printf("Dataset is not found: %s", spec)
		return
	}
	ok, err := ds.DropAll(ctx, wp)
	if !ok || err != nil {
		render.Status(r, http.StatusBadRequest)
		log.Printf("Unable to delete request because: %s", err)
		return
	}
	log.Printf("Dataset is deleted: %s", spec)
	render.Status(r, http.StatusAccepted)
	return
}

// createDataSet creates a new dataset.
func createDataSet(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")
	if spec == "" {
		spec = chi.URLParam(r, "spec")
	}
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		var created bool
		ds, created, err = models.CreateDataSet(spec)
		if created {
			err = fragments.SaveDataSet(spec, bp)
		}
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprintf("Unable to create dataset for %s", spec),
				Error:      nil,
			})
			log.Printf("Unable to create dataset for %s.\n", spec)
			return
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, ds)
		return
	}
	render.Status(r, http.StatusNotModified)
	render.JSON(w, r, ds)
	return
}

// listNameSpaces list all currently defined NameSpace object
func listNameSpaces(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, c.Config.NameSpaceMap.ByPrefix())
	//render.JSON(w, r, c.Config.NameSpaces)
	return
}
