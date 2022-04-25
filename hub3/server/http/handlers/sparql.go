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
}

var limitExp = regexp.MustCompile(`(?im) limit (\d*)`)

func ensureSparqlLimit(query string) (string, error) {
	matches := limitExp.FindAllStringSubmatch(query, -1)
	if len(matches) == 0 || matches == nil {
		return fmt.Sprintf("%s LIMIT 25", query), nil
	}

	last := matches[len(matches)-1]
	number, err := strconv.ParseInt(last[1], 10, 0)
	if err != nil {
		return "", fmt.Errorf("limit attribute %q is not a valid number", last[1])
	}

	if number > 1000 {
		return "", fmt.Errorf("sparql limit is not allowed to be greater than 1000")
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

	// log.Println(query)
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

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func runSparqlQuery(orgID, query string) (body []byte, statusCode int, contentType string, err error) {
	log.Printf("Sparql Query: %s", query)
	req, err := http.NewRequest("Get", c.Config.GetSparqlEndpoint(orgID, ""), http.NoBody)
	if err != nil {
		log.Printf("Unable to create sparql request %s", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

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
