// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mapping

import "fmt"

func FragmentESMapping(shards, replicas int) string {
	shards, replicas = setDefaults(shards, replicas)

	return fmt.Sprintf(
		fragmentMapping,
		shards,
		replicas,
	)
}

// fragmentMapping is the default mapping for the RDF fragments in hub3
var fragmentMapping = `{
	"settings": {
		"index": {
			"mapping.total_fields.limit": 1000,
			"mapping.depth.limit": 20,
			"mapping.nested_fields.limit": 50,
			"number_of_shards": %d,
			"number_of_replicas": %d
		},
		"analysis": {
			"analyzer": {
				"path_hierarchy": {
					"tokenizer": "path_hierarchy"
				}
			}
		}
	},
	"mappings":{
			"dynamic": "strict",
			"date_detection" : false,
			"properties": {
				"meta": {
					"type": "object",
					"properties": {
						"spec": {"type": "keyword"},
						"orgID": {"type": "keyword"},
						"hubID": {"type": "keyword"},
						"revision": {"type": "long"},
						"tags": {"type": "keyword"},
						"docType": {"type": "keyword"},
						"namedGraphURI": {"type": "keyword"},
						"entryURI": {"type": "keyword"},
						"modified": {"type": "date"}
					}
				},
				"subject": {"type": "keyword"},
				"predicate": {"type": "keyword"},
				"searchLabel": {"type": "keyword", "ignore_above": 256},
				"object": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"language": {"type": "keyword"},
				"dataType": {"type": "keyword"},
				"triple": {"type": "keyword", "index": false, "store": true},
				"lodKey": {"type": "keyword"},
				"objectType": {"type": "keyword"},
				"recordType": {"type": "short"},
				"order": {"type": "integer"},
				"level": {"type": "integer"},
				"resourceType": {"type": "keyword"},
				"nestedPath": {
					"type": "text",
					"analyzer": "path_hierarchy",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"path": {
					"type": "text",
					"analyzer": "path_hierarchy",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
}}}`
