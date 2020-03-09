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

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/gammazero/workerpool"
	r "github.com/kiivihal/rdf2go"
	"github.com/olivere/elastic/v7"

	"github.com/parnurzeal/gorequest"
)

// BulkAction is used to unmarshal the information from the BulkAPI
type BulkAction struct {
	HubID         string `json:"hubId"`
	OrgID         string `json:"orgID"`
	Spec          string `json:"dataset"`
	LocalID       string `json:"localID"`
	NamedGraphURI string `json:"graphUri"`
	RecordType    string `json:"type"`
	Action        string `json:"action"`
	ContentHash   string `json:"contentHash"`
	Graph         string `json:"graph"`
	RDF           string `json:"rdf"`
	GraphMimeType string `json:"graphMimeType"`
	SubjectType   string `json:"subjectType"`
	p             *elastic.BulkProcessor
	wp            *workerpool.WorkerPool
}

// BulkActionResponse is the datastructure where we keep the BulkAction statistics
type BulkActionResponse struct {
	Spec               string                   `json:"spec"`
	SpecRevision       int                      `json:"specRevision"`       // version of the records stored
	TotalReceived      int                      `json:"totalReceived"`      // originally json was total_received
	ContentHashMatches int                      `json:"contentHashMatches"` // originally json was content_hash_matches
	RecordsStored      int                      `json:"recordsStored"`      // originally json was records_stored
	JSONErrors         int                      `json:"jsonErrors"`
	TriplesStored      int                      `json:"triplesStored"`
	SparqlUpdates      []fragments.SparqlUpdate `json:"sparqlUpdates"` // store all the triples here for bulk insert
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(ctx context.Context, r io.Reader, p *elastic.BulkProcessor, wp *workerpool.WorkerPool) (BulkActionResponse, error) {
	//log.Println("Start reading actions.")
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	response := BulkActionResponse{
		SparqlUpdates: []fragments.SparqlUpdate{},
		TotalReceived: 0,
	}
	var line []byte
	for scanner.Scan() {
		line = scanner.Bytes()
		var action BulkAction
		//log.Printf("bulkAction: \n %s\n", line)
		err := json.Unmarshal(line, &action)
		if err != nil {
			response.JSONErrors++
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
			log.Printf("%s", line)
			continue
		}
		action.p = p
		action.wp = wp

		//err = ioutil.WriteFile(fmt.Sprintf("/tmp/es_actions/%s.json", action.HubID), []byte(action.Graph), 0644)
		//err = ioutil.WriteFile(fmt.Sprintf("/tmp/raw_graph/%s.json", action.HubID), []byte(action.Graph), 0644)
		//if err != nil {
		//log.Printf("Processing error: %#v", err)
		//return response, err
		//}
		err = action.Execute(ctx, &response)
		if err != nil {
			log.Printf("Processing error: %v", err)
			return response, err
		}
		response.TotalReceived++

	}
	if scanner.Err() != nil {
		log.Printf("Error scanning bulkActions: %s", scanner.Err())
		return response, scanner.Err()
	}
	if c.Config.RDF.RDFStoreEnabled {
		// insert the RDF triples
		errs := response.RDFBulkInsert()
		if errs != nil {
			return response, errs[0]
		}
	}
	//log.Printf("%#v", response)
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
		ds, err = ds.IncrementRevision()
		if err != nil {
			log.Printf("Unable to increment DataSet for %s\n", action.Spec)
			return err
		}
		response.SpecRevision = ds.Revision + 1
		log.Printf("Incremented dataset %s ", action.Spec)
	case "clear_orphans":
		// clear triples
		ok, err := ds.DropOrphans(ctx, action.p, action.wp)
		if !ok || err != nil {
			log.Printf("Unable to drop orphans for %s: %#v\n", action.Spec, err)
			return err
		}
		log.Printf("Mark orphans and delete them for %s", action.Spec)
	case "disable_index":
		ok, err := ds.DropRecords(ctx, action.wp)
		if !ok || err != nil {
			log.Printf("Unable to drop records for %s\n", action.Spec)
			return err
		}
		log.Printf("remove dataset %s from the storage", action.Spec)
	case "drop_dataset":
		ok, err := ds.DropAll(ctx, action.wp)
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
	default:
		log.Printf("Unknown action %s", action.Action)
	}
	return nil
}

// RDFBulkInsert inserts all triples from the bulkRequest in one SPARQL update statement
func (r *BulkActionResponse) RDFBulkInsert() []error {
	// remove sparqlUpdates because they are no longer needed
	triplesStored, errs := fragments.RDFBulkInsert(r.SparqlUpdates)
	r.SparqlUpdates = []fragments.SparqlUpdate{}
	r.TriplesStored = triplesStored
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
func (action *BulkAction) ESSave(response *BulkActionResponse, v1StylingIndexing bool) error {
	if action.Graph == "" {
		return fmt.Errorf("hubID %s has an empty graph. This is not allowed", action.HubID)
	}
	fb, err := action.createFragmentBuilder(response.SpecRevision)
	if err != nil {
		log.Printf("Unable to build fragmentBuilder: %v", err)
		return err
	}

	var r *elastic.BulkIndexRequest
	if v1StylingIndexing {
		// cleanup the graph and sort rdf webresources
		fb.GetSortedWebResources()

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
		// add to posthook worker from v1
		subject := strings.TrimSuffix(action.NamedGraphURI, "/graph")
		g := fb.SortedGraph
		ph := models.NewPostHookJob(g, action.Spec, false, subject, action.HubID)
		if ph.Valid() {
			action.wp.Submit(func() { models.ApplyPostHookJob(ph) })
			//action.wp.Submit(func() { log.Println(ph.Subject) })
		}
	} else {
		// index the LoD Fragments
		if c.Config.ElasticSearch.Fragments {
			err = fb.IndexFragments(action.p)
			if err != nil {
				return err
			}
		}

		// index FragmentGraph
		r = elastic.NewBulkIndexRequest().
			Index(c.Config.ElasticSearch.GetIndexName()).
			RetryOnConflict(3).
			Id(action.HubID).
			Doc(fb.Doc())
	}
	if r == nil {
		// todo add code back to create index doc
		//panic("can't create index doc")
		return fmt.Errorf("Unable create BulkIndexRequest")
	}

	// submit the bulkIndexRequest for indexing
	action.p.Add(r)

	if c.Config.RDF.RDFStoreEnabled {
		if c.Config.RDF.HasStoreTag(fb.FragmentGraph().Meta.Tags) {
			action.CreateRDFBulkRequest(response, fb.Graph)
		}
	}

	return nil
}

func (action BulkAction) createFragmentBuilder(revision int) (*fragments.FragmentBuilder, error) {
	fg := fragments.NewFragmentGraph()
	fg.Meta.OrgID = c.Config.OrgID
	fg.Meta.HubID = action.HubID
	fg.Meta.Spec = action.Spec
	fg.Meta.Revision = int32(revision)
	fg.Meta.NamedGraphURI = action.NamedGraphURI
	fg.Meta.EntryURI = fg.GetAboutURI()
	fg.Meta.Modified = fragments.NowInMillis()
	//fg.RecordType = fragments.RecordType_NARTHEX
	fg.Meta.Tags = []string{"narthex", "mdr"}
	fb := fragments.NewFragmentBuilder(fg)
	if action.GraphMimeType == "" {
		action.GraphMimeType = c.Config.RDF.DefaultFormat
	}
	err := fb.ParseGraph(strings.NewReader(action.Graph), action.GraphMimeType)
	if err != nil {
		log.Printf("Unable to parse the graph: %s", err)
		return fb, fmt.Errorf("Source RDF is not in format: %s", action.GraphMimeType)
	}
	return fb, nil
}

type fusekiStoreResponse struct {
	Count       int `json:"count"`
	TripleCount int `json:"tripleCount"`
	QuadCount   int `json:"quadCount"`
}

//CreateRDFBulkRequest gathers all the triples from an BulkAction to be inserted in bulk.
func (action BulkAction) CreateRDFBulkRequest(response *BulkActionResponse, g *r.Graph) {
	var b bytes.Buffer
	g.Serialize(&b, "text/turtle")

	su := fragments.SparqlUpdate{
		Triples:       b.String(),
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
	defer resp.Body.Close()
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
