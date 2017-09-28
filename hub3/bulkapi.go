package hub3

import (
	"bufio"
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

type BulkAction struct {
	hubId       string `json:"hubId"`
	dataset     string `json:"dataset"`
	graphUri    string `json:"graphUri"`
	actionType  string `json:"type"`
	action      string `json:"action"`
	contentHash string `json:"contentHash"`
	graph       string `json:"graph"`
}

// readActions reads BulkActions from an io.Reader line by line.
func readActions(reader io.Reader) error {

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return err
	}
	return nil

}
