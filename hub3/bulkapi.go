package hub3

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

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
	HubID       string `json:"hubId"`
	Spec        string `json:"dataset"`
	GraphURI    string `json:"graphUri"`
	RecordType  string `json:"type"`
	Action      string `json:"action"`
	ContentHash string `json:"contentHash"`
	Graph       string `json:"graph"`
	p           *elastic.BulkProcessor
}

type BulkActionResponse struct {
	Spec               string `json:"spec"`
	TotalReceived      int    `json:"totalReceived"`      // originally json was total_received
	ContentHashMatches int    `json:"contentHashMatches"` // originally json was content_hash_matches
	RecordsStored      int    `json:"recordsStored"`      // originally json was records_stored
	JsonErrors         int    `json:"jsonErrors"`
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(r io.Reader, p *elastic.BulkProcessor) (BulkActionResponse, error) {

	reader := bufio.NewReader(r)
	response := BulkActionResponse{}
	var line string
	var err error
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Printf(" > Failed!: %v\n", err)
				continue
			}
		}
		response.TotalReceived += 1

		var action BulkAction
		err := json.Unmarshal([]byte(line), &action)
		if err != nil {
			response.JsonErrors += 1
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
			continue
		}
		action.p = p
		action.Excute(&response)

	}
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
		err := action.ESSave(response)
		//err := action.Save(response)
		if err != nil {
			log.Printf("Unable to save BulkAction for %s because of %s", action.HubID, err)
			return err
		}
		//errs := action.RDFSave(response)
		//if errs != nil {
		//log.Printf("Unable to save BulkAction for %s because of %s", action.HubID, errs)
		//return errs[0]
		//}
		//fmt.Println("Store and index")
	default:
		log.Printf("Unknown action %s", action.Action)
	}
	return nil
}

// ESSaves the RDFRecord to ElasticSearch
func (action BulkAction) ESSave(response *BulkActionResponse) error {
	record := models.NewRDFRecord(action.HubID, action.Spec)
	record.ContentHash = action.ContentHash
	record.Graph = action.Graph
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

// RDFSave save the RDFrecord to the TripleStore
func (action BulkAction) RDFSave(response *BulkActionResponse) []error {
	request := gorequest.New()
	postURL := fmt.Sprintf("%s?graph=%s", "http://localhost:3030/rapid/data", action.GraphURI)
	resp, body, errs := request.Post(postURL).
		Set("Content-Type", "application/n-triples; charset=utf-8").
		Type("multipart").
		SendFile(bytes.NewBufferString(action.Graph)).
		//Send([]byte(action.Graph)).
		End()
	fmt.Printf("%#v", []byte(action.Graph))
	fmt.Println(action.Graph)
	fmt.Println(body)
	fmt.Println(resp)
	fmt.Println(errs)
	if errs != nil {
		log.Fatal(errs)
	}
	return errs
}

// Save converts the bulkAction request to an RDFRecord and saves it in Boltdb
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
