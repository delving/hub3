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
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterSparql(r chi.Router) {
	r.Get("/sparql", sparqlProxy)
	r.Post("/sparql", sparqlProxy)
	r.Put("/api/rdf/graph-store", graphStoreUpdate)
	r.Delete("/api/rdf/graph-store", graphStoreDelete)
	r.Post("/api/rdf/graph-store/delete", graphStoreDelete)
}

var limitExp = regexp.MustCompile(`(?im)\slimit\s*(\d*)`)

func ensureSparqlLimit(query string) (string, error) {
	matches := limitExp.FindAllStringSubmatch(query, -1)
	if len(matches) == 0 || matches == nil {
		return fmt.Sprintf("%s LIMIT 25", query), nil
	}

	for _, m := range matches {
		number, err := strconv.ParseInt(m[1], 10, 0)
		if err != nil {
			return "", fmt.Errorf("limit attribute %q is not a valid number", m[1])
		}

		if number > 1000 {
			return "", fmt.Errorf("sparql limit is not allowed to be greater than 1000")
		}
	}

	return query, nil
}

func sparqlProxy(w http.ResponseWriter, r *http.Request) {
	if !c.Config.RDF.SparqlEnabled {
		log.Printf("sparql is disabled\n")
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	}

	var query string

	switch r.Method {
	case http.MethodGet:
		query = r.URL.Query().Get("query")
	case http.MethodPost:
		query = r.FormValue("query")
	}

	if query == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &ErrorMessage{"Bad Request", "a value in the query param is required."})
		return
	}

	query, err := ensureSparqlLimit(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID := domain.GetOrganizationID(r)
	resp, statusCode, contentType, err := runSparqlQuery(orgID.String(), query)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, string(resp))
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

func makeSparqlRequest(req *http.Request) (body []byte, statusCode int, contentType string, err error) {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Do(req)
	if err != nil {
		log.Printf("Error in sparql query: %s", err)
		if strings.Contains(err.Error(), "connection refused") {
			body = []byte("triple store unavailable")
		}
		statusCode = http.StatusBadRequest
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Unable to read the response body with error: %s", err)
		return
	}
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")

	return body, statusCode, contentType, err
}

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func runSparqlQuery(orgID, query string) (body []byte, statusCode int, contentType string, err error) {
	log.Printf("Sparql Query: %s", query)
	req, err := http.NewRequest("Get", c.Config.GetSparqlEndpoint(orgID, ""), http.NoBody)
	if err != nil {
		log.Printf("Unable to create sparql request %s", err)
		return
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	return makeSparqlRequest(req)
}

func graphStoreDelete(w http.ResponseWriter, r *http.Request) {
	if !c.Config.RDF.SparqlEnabled {
		log.Printf("sparql is disabled\n")
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	}

	var graphName string

	switch r.Method {
	case http.MethodPost:
		graphName = r.FormValue("graph")
	case http.MethodDelete:
		graphName = r.URL.Query().Get("graph")
	}

	if graphName == "" {
		render.JSON(w, r, &ErrorMessage{"invalid graph form value", ""})
		return
	}

	orgID := domain.GetOrganizationID(r)
	resp, statusCode, contentType, err := runGraphStoreDelete(orgID.String(), graphName)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, string(resp))
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

func graphStoreUpdate(w http.ResponseWriter, r *http.Request) {
	if !c.Config.RDF.SparqlEnabled {
		log.Printf("sparql is disabled\n")
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	}

	in, _, err := r.FormFile("rdf")
	if err != nil {
		http.Error(w, "cannot find ead form file", http.StatusBadRequest)
		return
	}

	defer in.Close()

	// cleanup upload
	defer func() {
		err = r.MultipartForm.RemoveAll()
	}()

	graphName := r.FormValue("graph")
	if graphName == "" {
		render.JSON(w, r, &ErrorMessage{"invalid graph form value", ""})
		return
	}

	fileContentType := r.FormValue("content-type")

	orgID := domain.GetOrganizationID(r)
	resp, statusCode, contentType, err := runGraphStoreQuery(orgID.String(), graphName, fileContentType, in)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, string(resp))
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

// runGraphStoreQuery sends a GraphStore request to the SPARQL-endpoint specified in the configuration
func runGraphStoreQuery(orgID, graphName, inContentType string, in io.Reader) (body []byte, statusCode int, contentType string, err error) {
	req, err := http.NewRequest(http.MethodPut, c.Config.GetGraphStoreEndpoint(orgID, ""), in)
	if err != nil {
		log.Printf("Unable to create sparql request %s", err)
		return
	}
	req.Header.Add("Content-Type", inContentType)
	params := req.URL.Query()
	params.Set("graph", graphName)
	req.URL.RawQuery = params.Encode()

	return makeSparqlRequest(req)
}

// runGraphStoreQuery sends a GraphStore request to the SPARQL-endpoint specified in the configuration
func runGraphStoreDelete(orgID, graphName string) (body []byte, statusCode int, contentType string, err error) {
	req, err := http.NewRequest(http.MethodDelete, c.Config.GetGraphStoreEndpoint(orgID, ""), http.NoBody)
	if err != nil {
		log.Printf("Unable to create sparql request %s", err)
		return
	}
	params := req.URL.Query()
	params.Set("graph", graphName)
	req.URL.RawQuery = params.Encode()

	return makeSparqlRequest(req)
}
