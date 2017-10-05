package hub3

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
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
}

type BulkActionResponse struct {
	Total        int      `json:"totalItemCount"`
	IndexCount   int      `json:"indexedItemCount"`
	DeletedCount int      `json:"deletedItemCount"`
	InvalidCount int      `json:"invalidItemCount"`
	InvalidItems []string `json:"invalidItems"`
}

// ReadActions reads BulkActions from an io.Reader line by line.
func ReadActions(reader io.Reader) (BulkActionResponse, error) {

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var action BulkAction
		err := json.Unmarshal(scanner.Bytes(), &action)
		if err != nil {
			log.Println("Unable to unmarshal JSON.")
			log.Print(err)
		}
		//log.Println(action.HubID)
		log.Println(action.Action)
	}

	response := BulkActionResponse{
		InvalidItems: []string{},
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return response, err
	}
	return response, nil

}
