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
	"fmt"

	"github.com/renevanderark/goharvest/oai"
)

// Dump a snippet of the Record metadata
func dump(record *oai.Record) {
	fmt.Printf("%s\n\n", record.Metadata.Body[0:500])
}

// harvestToFile demonstrates harvesting using the ListRecords verb with HarvestRecords
func harvestToFile(baseUrl string, set string, prefix string, from string) {
	req := &oai.Request{
		BaseURL:        baseUrl,
		Set:            set,
		MetadataPrefix: prefix,
		Verb:           "ListRecords",
		From:           from,
	}
	// HarvestRecords passes each individual metadata record to the dump
	// function as a Record object
	req.HarvestRecords(dump)
}
