package server

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/delving/rapid/hub3"
	"bitbucket.org/delving/rapid/hub3/models"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/kiivihal/goharvest/oai"
)

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

// APIErrorMessage contains the default API error messages
type APIErrorMessage struct {
	HttpStatus int    `json:"code"`
	Message    string `json:"type"`
	Error      error  `json:error`
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func bulkAPI(w http.ResponseWriter, r *http.Request) {
	response, err := hub3.ReadActions(r.Body)
	if err != nil {
		log.Println("Unable to read actions")
		render.Render(w, r, ErrRender(err))
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
	resp := hub3.ProcessVerb(&req)
	render.XML(w, r, resp)
}

// listDataSets returns a list of all public datasets
func listDataSets(w http.ResponseWriter, r *http.Request) {
	sets, err := models.ListDataSets()
	if err != nil {
		log.Printf("Unable to list datasets because of: %s", err)
		render.Render(w, r, ErrRender(err))
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, sets)
	return
}

// getDataSet returns a dataset when found or a 404
func getDataSet(w http.ResponseWriter, r *http.Request) {
	ds, err := models.GetDataSet(chi.URLParam(r, "spec"))
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HttpStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", chi.URLParam(r, "spec")),
				Error:      err,
			})
			return
		}
		render.Render(w, r, ErrRender(err))
		return

	}
	render.JSON(w, r, ds)
	return
}

// createDataSet creates a new dataset.
func createDataSet(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		ds, err = models.CreateDataSet(spec)
		if err != nil {
			log.Printf("Unable to create dataset for %s", spec)
			render.Render(w, r, ErrRender(err))
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
