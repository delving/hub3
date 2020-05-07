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
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/gammazero/workerpool"
	r "github.com/kiivihal/rdf2go"

	"github.com/parnurzeal/gorequest"
)

type BulkIndex interface {
	Publish(ctx context.Context, message ...*domainpb.IndexMessage) error
}

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
	bi            BulkIndex
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
	sparqlUpdates      []fragments.SparqlUpdate // store all the triples here for bulk insert
	ds                 *models.DataSet
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(ctx context.Context, r io.Reader, bi BulkIndex, wp *workerpool.WorkerPool) (BulkActionResponse, error) {
	//log.Println("Start reading actions.")
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 5*1024*1024)

	response := BulkActionResponse{
		sparqlUpdates: []fragments.SparqlUpdate{},
		TotalReceived: 0,
	}
	for scanner.Scan() {
		var action BulkAction
		//log.Printf("bulkAction: \n %s\n", line)
		err := json.Unmarshal(scanner.Bytes(), &action)
		if err != nil {
			response.JSONErrors++
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
			log.Printf("%s", scanner.Bytes())
			continue
		}
		action.bi = bi
		action.wp = wp

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
	if response.ds == nil {
		response.Spec = action.Spec

		ds, created, err := models.GetOrCreateDataSet(action.Spec)
		if err != nil {
			log.Printf("Unable to get DataSet for %s\n", action.Spec)
			return err
		}
		if created {
			err = fragments.SaveDataSet(action.Spec, nil)
			if err != nil {
				log.Printf("Unable to Save DataSet Fragment for %s\n", action.Spec)
				return err
			}
		}
		response.SpecRevision = ds.Revision
		response.ds = ds
	}

	switch action.Action {
	case "increment_revision":
		ds, err := response.ds.IncrementRevision()
		if err != nil {
			log.Printf("Unable to increment DataSet for %s\n", action.Spec)
			return err
		}
		response.SpecRevision = ds.Revision + 1
		log.Printf("Incremented dataset %s ", action.Spec)
	case "clear_orphans":
		// clear triples
		ok, err := response.ds.DropOrphans(context.Background(), nil, action.wp)
		if !ok || err != nil {
			log.Printf("Unable to drop orphans for %s: %#v\n", action.Spec, err)
			return err
		}
		log.Printf("Mark orphans and delete them for %s", action.Spec)
	case "disable_index":
		ok, err := response.ds.DropRecords(ctx, action.wp)
		if !ok || err != nil {
			log.Printf("Unable to drop records for %s\n", action.Spec)
			return err
		}
		log.Printf("remove dataset %s from the storage", action.Spec)
	case "drop_dataset":
		ok, err := response.ds.DropAll(ctx, action.wp)
		if !ok || err != nil {
			log.Printf("Unable to drop dataset %s", action.Spec)
			return err
		}
		log.Printf("remove the dataset %s completely", action.Spec)
	case "index":
		if c.Config.ElasticSearch.Enabled {
			err := action.ESSave(response, c.Config.ElasticSearch.IndexTypes...)
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
func (resp *BulkActionResponse) RDFBulkInsert() []error {
	// remove sparqlUpdates because they are no longer needed
	triplesStored, errs := fragments.RDFBulkInsert(resp.sparqlUpdates)
	resp.sparqlUpdates = []fragments.SparqlUpdate{}
	resp.TriplesStored = triplesStored
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

func (action *BulkAction) processV1(fb *fragments.FragmentBuilder) error {
	fb.GetSortedWebResources()

	// TODO(kiivihal): update v1 support
	indexDoc, err := fragments.CreateV1IndexDoc(fb)
	if err != nil {
		log.Printf("Unable to create index doc: %s", err)
		return err
	}

	b, err := json.Marshal(indexDoc)
	if err != nil {
		return err
	}

	m := &domainpb.IndexMessage{
		OrganisationID: c.Config.OrgID,
		DatasetID:      action.Spec,
		RecordID:       action.HubID,
		IndexName:      c.Config.ElasticSearch.GetV1IndexName(),
		Deleted:        false,
		Source:         b,
	}

	if err := action.bi.Publish(context.Background(), m); err != nil {
		return err
	}

	// add to posthook worker from v1
	subject := strings.TrimSuffix(action.NamedGraphURI, "/graph")
	g := fb.SortedGraph
	ph := models.NewPostHookJob(g, action.Spec, false, subject, action.HubID)
	if ph.Valid() && action.wp != nil {
		// non async posthook
		models.ApplyPostHookJob(ph)

		// async
		// action.wp.Submit(func() { models.ApplyPostHookJob(ph) })
	}

	return nil
}

func (action *BulkAction) processFragments(fb *fragments.FragmentBuilder) error {
	return fb.IndexFragments(action.bi)
}

func (action *BulkAction) processV2(fb *fragments.FragmentBuilder) error {
	// index FragmentGraph
	m, err := fb.Doc().IndexMessage()
	if err != nil {
		return err
	}

	if err := action.bi.Publish(context.Background(), m); err != nil {
		return err
	}

	return nil
}

//ESSave the RDF Record to ElasticSearch
func (action *BulkAction) ESSave(response *BulkActionResponse, indexTypes ...string) error {
	if action.Graph == "" {
		return fmt.Errorf("hubID %s has an empty graph. This is not allowed", action.HubID)
	}
	fb, err := action.createFragmentBuilder(response.SpecRevision)
	if err != nil {
		log.Printf("Unable to build fragmentBuilder: %v", err)
		return err
	}

	_, err = fb.ResourceMap()
	if err != nil {
		log.Printf("Unable to create resource map: %v", err)
		return err
	}

	for _, indexType := range indexTypes {
		switch indexType {
		case "v1":
			if err := action.processV1(fb); err != nil {
				return err
			}
		case "v2":
			if err := action.processV2(fb); err != nil {
				return err
			}
		case "fragments":
			if err := action.processFragments(fb); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown indexType: %s", indexType)
		}
	}

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
	response.sparqlUpdates = append(response.sparqlUpdates, su)
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
