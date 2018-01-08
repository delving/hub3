package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "bitbucket.org/delving/rapid/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

type SparqlResource struct{}

func (rs SparqlResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", sparqlProxy)

	return r
}

func sparqlProxy(w http.ResponseWriter, r *http.Request) {
	if !Config.RDF.SparqlEnabled {
		// todo replace with json later
		render.PlainText(w, r, `{"status": "not enabled"}`)
		return
	}
	query := r.URL.Query().Get("query")
	if query == "" {
		render.Status(r, http.StatusNotFound)
		return
	}
	if !strings.Contains(strings.ToLower(query), " limit ") {
		query = fmt.Sprintf("%s LIMIT 25", query)
	}
	log.Info(query)
	resp, statusCode, contentType, err := runSparqlQuery(query)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(resp)
	render.Status(r, statusCode)
	return
}

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func runSparqlQuery(query string) (body []byte, statusCode int, contentType string, err error) {
	log.Debugf("Sparql Query: %s", query)
	req, err := http.NewRequest("Get", Config.GetSparqlEndpoint(""), nil)
	if err != nil {
		log.Errorf("Unable to create sparql request %s", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Do(req)
	if err != nil {
		log.Errorf("Error in sparql query: %s", err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Unable to read the response body with error: %s", err)
	}
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
