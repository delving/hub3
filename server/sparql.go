package server

import (
	"io/ioutil"
	"net/http"
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
	query := r.URL.Query().Get("query")
	log.Info(query)
	if query == "" {
		render.Status(r, http.StatusNotFound)
		return
	}
	render.PlainText(w, r, `{"status": "not enabled"}`)
	return
}

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func runSparqlQuery(query string) (body []byte, statusCode int, err error) {
	log.Debugf("Sparql Query: %s", query)
	req, err := http.NewRequest("Get", Config.GetSparqlEndpoint(""), nil)
	if err != nil {
		log.Errorf("Unable to create sparql request %s", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	rquery := req.URL.Query()
	rquery.Set("query", query)
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
	return
}
