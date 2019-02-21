package fragments

import (
	"bufio"
	"bytes"
	fmt "fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/delving/rapid-saas/config"
	"github.com/parnurzeal/gorequest"
)

// SparqlUpdate contains the elements to perform a SPARQL update query
type SparqlUpdate struct {
	Triples       string `json:"triples"`
	NamedGraphURI string `json:"graphUri"`
	Spec          string `json:"datasetSpec"`
	SpecRevision  int    `json:"specRevision"`
}

// TripleCount counts the number of Ntriples in a string
func (su SparqlUpdate) TripleCount() (int, error) {
	r := strings.NewReader(su.Triples)
	return lineCounter(r)
}

func lineCounter(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, nil
}

func executeTemplate(tmplString string, name string, model interface{}) string {
	tmpl, err := template.New(name).Parse(tmplString)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, model)
	if err != nil {
		panic(err)
	}
	return b.String()
}

func (su SparqlUpdate) String() string {
	t := `GRAPH <{{.NamedGraphURI}}> {
		<{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}" .
		<{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/specRevision> "{{.SpecRevision}}"^^<http://www.w3.org/2001/XMLSchema#integer> .
		{{ .Triples }}
	}`
	return executeTemplate(t, "update", su)
}

func RDFBulkInsert(sparqlUpdates []SparqlUpdate) (int, []error) {
	nrGraphs := len(sparqlUpdates)
	if nrGraphs == 0 {
		log.Println("No graphs to store")
		return 0, nil
	}
	strs := make([]string, nrGraphs)
	graphs := make([]string, nrGraphs)
	triplesStored := 0
	for i, v := range sparqlUpdates {
		strs[i] = v.String()
		graphs[i] = fmt.Sprintf("DROP GRAPH <%s>;", v.NamedGraphURI)
		count, err := v.TripleCount()
		if err != nil {
			log.Printf("Unable to count triples: %s", err)
			return 0, []error{fmt.Errorf("Unable to count triples for %s because :%s", strs[i], err)}
		}
		triplesStored += count
	}
	sparqlInsert := fmt.Sprintf("%s INSERT DATA {%s}", strings.Join(graphs, "\n"), strings.Join(strs, "\n"))
	errs := UpdateViaSparql(sparqlInsert)
	if errs != nil {
		return 0, errs
	}
	return triplesStored, nil
}

// UpdateViaSparql is a post to sparql function that tasks a valid SPARQL update query
func UpdateViaSparql(update string) []error {
	request := gorequest.New()
	postURL := config.Config.GetSparqlUpdateEndpoint("")

	parameters := url.Values{}
	parameters.Add("update", update)
	updateQuery := parameters.Encode()

	resp, body, errs := request.Post(postURL).
		Send(updateQuery).
		Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		Retry(3, 4*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		End()
	if errs != nil {
		log.Fatalf("errors for query %s: %#v", postURL, errs)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		//log.Println(body)
		//log.Println(resp)
		log.Printf("unable to store sparqlUpdate: %s", update)
		return []error{fmt.Errorf("store error for SparqlUpdate:%s", body)}
	}
	return errs
}