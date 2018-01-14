package hub3

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	. "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/models"
	"github.com/parnurzeal/gorequest"
	elastic "gopkg.in/olivere/elastic.v5"
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

// BulkAction is used to unmarshal the information from the BulkAPI
type BulkAction struct {
	HubID         string `json:"hubId"`
	Spec          string `json:"dataset"`
	NamedGraphURI string `json:"graphUri"`
	RecordType    string `json:"type"`
	Action        string `json:"action"`
	ContentHash   string `json:"contentHash"`
	Graph         string `json:"graph"`
	p             *elastic.BulkProcessor
}

type SparqlUpdate struct {
	Triples       string `json:"triples"`
	NamedGraphURI string `json:"graphUri"`
	Spec          string `json:"datasetSpec"`
	SpecRevision  int    `json:"specRevision"`
}

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

type BulkActionResponse struct {
	Spec               string         `json:"spec"`
	SpecRevision       int            `json:"specRevision"`       // version of the records stored
	TotalReceived      int            `json:"totalReceived"`      // originally json was total_received
	ContentHashMatches int            `json:"contentHashMatches"` // originally json was content_hash_matches
	RecordsStored      int            `json:"recordsStored"`      // originally json was records_stored
	JsonErrors         int            `json:"jsonErrors"`
	TriplesStored      int            `json:"triplesStored"`
	SparqlUpdates      []SparqlUpdate `json:"sparqlUpdates"` // store all the triples here for bulk insert
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(r io.Reader, p *elastic.BulkProcessor) (BulkActionResponse, error) {

	scanner := bufio.NewScanner(r)
	response := BulkActionResponse{
		SparqlUpdates: []SparqlUpdate{},
		TotalReceived: 0,
	}
	var line []byte
	for scanner.Scan() {
		line = scanner.Bytes()
		var action BulkAction
		err := json.Unmarshal(line, &action)
		if err != nil {
			response.JsonErrors += 1
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
			continue
		}
		action.p = p
		action.Excute(&response)
		response.TotalReceived += 1

	}
	// insert the RDF triples
	errs := response.RDFBulkInsert()
	if errs != nil {
		////log.Fatal(errs)
		return response, errs[0]
	}
	log.Printf("%#v", response)
	return response, nil

}

// Execute performs the various BulkActions
func (action BulkAction) Excute(response *BulkActionResponse) error {
	if response.Spec == "" {
		response.Spec = action.Spec
	}
	switch action.Action {
	case "increment_revision":
		ds, err := models.GetOrCreateDataSet(action.Spec)
		if err != nil {
			log.Printf("Unable to get DataSet for %s\n", action.Spec)
			return err
		}
		err = ds.IncrementRevision()
		if err != nil {
			log.Printf("Unable to increment DataSet for %s\n", action.Spec)
			return err
		}
		log.Printf("Incremented dataset %s to %d", action.Spec, ds.Revision)
	case "clear_orphans":
		fmt.Println("Mark orphans and delete them")
	case "disable_index":
		fmt.Println("remove dataset from the index")
	case "drop_dataset":
		fmt.Println("remove the dataset completely")
	case "index":
		if response.SpecRevision == 0 {
			ds, err := models.GetOrCreateDataSet(action.Spec)
			if err != nil {
				log.Printf("Unable to get DataSet for %s\n", action.Spec)
				return err
			}
			response.SpecRevision = ds.Revision
		}
		err := action.ESSave(response)
		if err != nil {
			log.Printf("Unable to save BulkAction for %s because of %s", action.HubID, err)
			return err
		}
		action.CreateRDFBulkRequest(response)
		//fmt.Println("Store and index")
	default:
		log.Printf("Unknown action %s", action.Action)
	}
	return nil
}

func (r *BulkActionResponse) RDFBulkInsert() []error {
	nrGraphs := len(r.SparqlUpdates)
	if nrGraphs == 0 {
		log.Println("No graphs to store")
		return nil
	}
	request := gorequest.New()
	postURL := Config.GetSparqlUpdateEndpoint("")
	strs := make([]string, nrGraphs)
	triplesStored := 0
	for i, v := range r.SparqlUpdates {
		strs[i] = v.String()
		count, err := v.TripleCount()
		if err != nil {
			log.Printf("Unable to count triples: %s", err)
			return []error{fmt.Errorf("Unable to count triples for %s because :%s", strs[i], err)}
		}
		triplesStored += count
	}
	parameters := url.Values{}
	sparqlInsert := fmt.Sprintf("INSERT DATA {%s}", strings.Join(strs, "\n"))
	parameters.Add("update", sparqlInsert)
	dropInsert := parameters.Encode()

	resp, body, errs := request.Post(postURL).
		Send(dropInsert).
		Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		Retry(3, 4*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		End()
	if errs != nil {
		log.Fatal(errs)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Println(body)
		log.Println(resp)
		log.Printf("unable to store sparqlUpdate: %s", dropInsert)
		return []error{fmt.Errorf("store error for SparqlUpdate:%s", body)}
	}
	r.TriplesStored = triplesStored
	// remove sparqlUpdates because they are no longer needed
	r.SparqlUpdates = []SparqlUpdate{}
	return errs
}

func getContext(input string, lineNumber int) (string, error) {
	lineContext := 10
	start := lineNumber - lineContext
	if start < 2 {
		start = 1
	}
	end := lineNumber + lineContext
	// Splits on newlines by default.
	scanner := bufio.NewScanner(strings.NewReader(input))

	line := 1
	errorContext := []string{}
	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		if line > start && line < end {
			text := fmt.Sprintf("%d:\t%s", line, scanner.Text())
			if line == lineNumber {
				text = fmt.Sprintf("\n%s\n", text)
			}
			errorContext = append(errorContext, text)
		}
		if line > end {
			break
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		log.Println("Scan error: %s", err)
		return "", nil
	}
	return strings.Join(errorContext, "\n"), nil
}

// ESSaves the RDFRecord to ElasticSearch
func (action BulkAction) ESSave(response *BulkActionResponse) error {
	record := models.NewRDFRecord(action.HubID, action.Spec)
	record.ContentHash = action.ContentHash
	record.Graph = action.Graph
	record.Revision = response.SpecRevision
	record.NamedGraphURI = action.NamedGraphURI
	if action.Graph == "" {
		return fmt.Errorf("hubID %s has an empty graph. This is not allowed", action.HubID)
	}
	r := elastic.NewBulkIndexRequest().Index(Config.ElasticSearch.IndexName).Type("rdfrecord").Id(action.HubID).Doc(record)
	if r == nil {
		return fmt.Errorf("Unable create BulkIndexRequest")
	}
	action.p.Add(r)
	return nil
}

type fusekiStoreResponse struct {
	Count       int `json:"count"`
	TripleCount int `json:"tripleCount"`
	QuadCount   int `json:"quadCount"`
}

// RDFBulkInsert gathers all the
func (action BulkAction) CreateRDFBulkRequest(response *BulkActionResponse) {
	su := SparqlUpdate{
		Triples:       action.Graph,
		NamedGraphURI: action.NamedGraphURI,
		Spec:          action.Spec,
		SpecRevision:  response.SpecRevision,
	}
	response.SparqlUpdates = append(response.SparqlUpdates, su)
}

// RDFSave save the RDFrecord to the TripleStore.
// This saves each action individualy. You should use RDFBulkInsert instead.
func (action BulkAction) RDFSave(response *BulkActionResponse) []error {
	request := gorequest.New()
	postURL := Config.GetGraphStoreEndpoint("")
	resp, body, errs := request.Post(postURL).
		Query(fmt.Sprintf("graph=%s", action.NamedGraphURI)).
		Set("Content-Type", "application/n-triples; charset=utf-8").
		Type("text").
		Send(action.Graph).
		End()
	if errs != nil {
		log.Fatal(errs)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Printf("Unable to store GraphURI: %s", action.NamedGraphURI)
		return []error{fmt.Errorf("Store error for %s with message:%s", action.NamedGraphURI, body)}
	}
	fres := new(fusekiStoreResponse)
	err := json.Unmarshal([]byte(body), &fres)
	if err != nil {
		return []error{err}
	}
	log.Printf("Stored %d triples for graph %s", fres.TripleCount, action.NamedGraphURI)
	response.TriplesStored += fres.TripleCount
	return errs
}

// Save converts the bulkAction request to an RDFRecord and saves it in Boltdb
// TODO remove this later. Now superseded by ESSave
func (action BulkAction) Save(response *BulkActionResponse) error {
	record, err := models.GetOrCreateRDFRecord(action.HubID, action.Spec)
	if err != nil {
		log.Printf("Unable to get or create RDFrecord for %s", action.HubID)
		return err
	}
	if record.ContentHash == action.ContentHash && action.ContentHash != "" {
		// do nothing when the contentHash matches.
		response.ContentHashMatches++
		return nil
	}
	record.ContentHash = action.ContentHash
	//record.Graph = action.Graph
	if action.Graph == "" {
		return fmt.Errorf("hubID %s has an empty graph. This is not allowed", action.HubID)
	}
	//record.Revision
	err = record.Save()
	if err != nil {
		log.Printf("Unable to save %s because of %s", action.HubID, err)
	}
	return err
}
