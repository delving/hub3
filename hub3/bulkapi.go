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

package hub3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"

	c "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/fragments"
	"bitbucket.org/delving/rapid/hub3/models"
	elastic "github.com/olivere/elastic"
	"github.com/parnurzeal/gorequest"
)

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

// BulkActionResponse is the datastructure where we keep the BulkAction statistics
type BulkActionResponse struct {
	Spec               string         `json:"spec"`
	SpecRevision       int            `json:"specRevision"`       // version of the records stored
	TotalReceived      int            `json:"totalReceived"`      // originally json was total_received
	ContentHashMatches int            `json:"contentHashMatches"` // originally json was content_hash_matches
	RecordsStored      int            `json:"recordsStored"`      // originally json was records_stored
	JSONErrors         int            `json:"jsonErrors"`
	TriplesStored      int            `json:"triplesStored"`
	SparqlUpdates      []SparqlUpdate `json:"sparqlUpdates"` // store all the triples here for bulk insert
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(ctx context.Context, r io.Reader, p *elastic.BulkProcessor) (BulkActionResponse, error) {

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
			response.JSONErrors++
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
			continue
		}
		action.p = p
		err = action.Execute(ctx, &response)
		if err != nil {
			return response, err
		}
		response.TotalReceived++

	}
	if c.Config.RDF.RDFStoreEnabled {
		// insert the RDF triples
		errs := response.RDFBulkInsert()
		if errs != nil {
			return response, errs[0]
		}
	}
	log.Printf("%#v", response)
	return response, nil

}

//Execute performs the various BulkActions
func (action BulkAction) Execute(ctx context.Context, response *BulkActionResponse) error {
	if response.Spec == "" {
		response.Spec = action.Spec
	}
	ds, created, err := models.GetOrCreateDataSet(action.Spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", action.Spec)
		return err
	}
	if created {
		err = fragments.SaveDataSet(action.Spec, action.p)
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", action.Spec)
			return err
		}
	}
	response.SpecRevision = ds.Revision
	switch action.Action {
	case "increment_revision":
		err = ds.IncrementRevision()
		if err != nil {
			log.Printf("Unable to increment DataSet for %s\n", action.Spec)
			return err
		}
		response.SpecRevision = ds.Revision + 1
		log.Printf("Incremented dataset %s ", action.Spec)
	case "clear_orphans":
		// clear triples
		ok, err := ds.DropOrphans(ctx)
		if !ok || err != nil {
			log.Printf("Unable to drop orphans for %s: %#v\n", action.Spec, err)
			return err
		}
		log.Printf("Mark orphans and delete them for %s", action.Spec)
	case "disable_index":
		ok, err := ds.DropRecords(ctx)
		if !ok || err != nil {
			log.Printf("Unable to drop records for %s\n", action.Spec)
			return err
		}
		log.Printf("remove dataset %s from the storage", action.Spec)
	case "drop_dataset":
		ok, err := ds.DropAll(ctx)
		if !ok || err != nil {
			log.Printf("Unable to drop dataset %s", action.Spec)
			return err
		}
		log.Printf("remove the dataset %s completely", action.Spec)
	case "index":
		if response.SpecRevision == 0 {
			response.SpecRevision = ds.Revision
		}
		if c.Config.ElasticSearch.Enabled {
			err := action.ESSave(response, c.Config.ElasticSearch.IndexV1)
			if err != nil {
				log.Printf("Unable to save BulkAction for %s because of %s", action.HubID, err)
				return err
			}
		}
		if c.Config.RDF.RDFStoreEnabled {
			action.CreateRDFBulkRequest(response)
		}
	default:
		log.Printf("Unknown action %s", action.Action)
	}
	return nil
}

// RDFBulkInsert inserts all triples from the bulkRequest in one SPARQL update statement
func (r *BulkActionResponse) RDFBulkInsert() []error {
	nrGraphs := len(r.SparqlUpdates)
	if nrGraphs == 0 {
		log.Println("No graphs to store")
		return nil
	}
	strs := make([]string, nrGraphs)
	graphs := make([]string, nrGraphs)
	triplesStored := 0
	for i, v := range r.SparqlUpdates {
		strs[i] = v.String()
		graphs[i] = fmt.Sprintf("DROP GRAPH <%s>;", v.NamedGraphURI)
		count, err := v.TripleCount()
		if err != nil {
			log.Printf("Unable to count triples: %s", err)
			return []error{fmt.Errorf("Unable to count triples for %s because :%s", strs[i], err)}
		}
		triplesStored += count
	}
	sparqlInsert := fmt.Sprintf("%s INSERT DATA {%s}", strings.Join(graphs, "\n"), strings.Join(strs, "\n"))
	errs := models.UpdateViaSparql(sparqlInsert)
	if errs != nil {
		return errs
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
		log.Printf("Scan error: %s", err)
		return "", nil
	}
	return strings.Join(errorContext, "\n"), nil
}

//ESSave the RDF Record to ElasticSearch
func (action BulkAction) ESSave(response *BulkActionResponse, v1StylingIndexing bool) error {
	if action.Graph == "" {
		return fmt.Errorf("hubID %s has an empty graph. This is not allowed", action.HubID)
	}
	fb := action.createFragmentBuilder(response.SpecRevision)
	err := fb.CreateFragments(action.p, true)
	if err != nil {
		log.Printf("Unable to save fragments: %v", err)
		return err
	}
	var r *elastic.BulkIndexRequest
	if v1StylingIndexing {
		indexDoc, err := fragments.CreateV1IndexDoc(fb)
		if err != nil {
			log.Printf("Unable to create index doc: %s", err)
			return err
		}
		r, err = fragments.CreateESAction(indexDoc, action.HubID)
		if err != nil {
			log.Printf("Unable to create v1 es action: %s", err)
			return err
		}
	} else {
		r = elastic.NewBulkIndexRequest().
			Index(c.Config.ElasticSearch.IndexName).
			Type(fragments.DocType).
			Id(action.HubID).
			Doc(fb.Doc())
	}
	if r == nil {
		panic("can't create index doc")
		return fmt.Errorf("Unable create BulkIndexRequest")
	}
	action.p.Add(r)
	return nil
}

func (action BulkAction) createFragmentBuilder(revision int) *fragments.FragmentBuilder {
	fg := fragments.NewFragmentGraph()
	fg.OrgID = c.Config.OrgID
	fg.HubID = action.HubID
	fg.Spec = action.Spec
	fg.Revision = int32(revision)
	fg.NamedGraphURI = action.NamedGraphURI
	fg.Tags = []string{"narthex", "mdr"}
	fb := fragments.NewFragmentBuilder(fg)
	fb.ParseGraph(strings.NewReader(action.Graph), "text/turtle")
	return fb
}

type fusekiStoreResponse struct {
	Count       int `json:"count"`
	TripleCount int `json:"tripleCount"`
	QuadCount   int `json:"quadCount"`
}

//CreateRDFBulkRequest gathers all the triples from an BulkAction to be inserted in bulk.
func (action BulkAction) CreateRDFBulkRequest(response *BulkActionResponse) {
	su := SparqlUpdate{
		Triples:       action.Graph,
		NamedGraphURI: action.NamedGraphURI,
		Spec:          action.Spec,
		SpecRevision:  response.SpecRevision,
	}
	response.SparqlUpdates = append(response.SparqlUpdates, su)
}

//RDFSave save the RDFrecord to the TripleStore.
//This saves each action individually. You should use RDFBulkInsert instead.
func (action BulkAction) RDFSave(response *BulkActionResponse) []error {
	request := gorequest.New()
	postURL := c.Config.GetGraphStoreEndpoint("")
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
