package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

// SparqlResource is a struct for the Search routes
type SparqlResource struct{}

// Routes returns the chi.Router
func (rs SparqlResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", sparqlProxy)
	r.Post("/", sparqlProxy)

	return r
}

func sparqlProxy(w http.ResponseWriter, r *http.Request) {
	if !c.Config.RDF.SparqlEnabled {
		log.Printf("sparql is disabled\n")
		render.JSON(w, r, &ErrorMessage{"not enabled", ""})
		return
	}
	var query string
	log.Print(r.Method)
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
	if !strings.Contains(strings.ToLower(query), "limit ") {
		query = fmt.Sprintf("%s LIMIT 25", query)
	}
	log.Info(query)
	resp, statusCode, contentType, err := runSparqlQuery(query)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, string(resp))
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

// runSparqlQuery sends a SPARQL query to the SPARQL-endpoint specified in the configuration
func runSparqlQuery(query string) (body []byte, statusCode int, contentType string, err error) {
	log.Debugf("Sparql Query: %s", query)
	req, err := http.NewRequest("Get", c.Config.GetSparqlEndpoint(""), nil)
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
		if strings.Contains(err.Error(), "connection refused") {
			body = []byte("triple store unavailable")
		}
		statusCode = http.StatusBadRequest
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Unable to read the response body with error: %s", err)
		return
	}
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
