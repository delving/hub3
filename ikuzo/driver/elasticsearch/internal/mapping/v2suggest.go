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

func V2SuggestMapping(shards, replicas int) string {
	shards, replicas = setDefaults(shards, replicas)

	return fmt.Sprintf(
		v2SuggestMapping,
		shards,
		replicas,
	)
}

// v2Mapping is the default mapping for the RDF records enabled by hub3
var v2SuggestMapping = `{
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
				"default": {
					"tokenizer": "standard",
					"char_filter":  ["html_strip"],
					"filter" : ["lowercase","asciifolding"]
				}
			}
		}
	},
	"mappings":{
		"dynamic": true,
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
					"modified": {"type": "date"},
					"sourceID": {"type": "keyword"},
					"sourcePath": {"type": "keyword"},
					"groupID": {"type": "keyword"}
				}
			},
			"id": {"type": "keyword"},
			"brocadeID": {"type": "keyword"},
			"orgID": {"type": "keyword"},
			"suggestType": {"type": "keyword"},
			"json": {
				"type": "keyword",
				"store": true,
				"index": false,
				"doc_values": false
			},
			"text": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512},
					"suggest": { "type": "completion"}
				}
			},
			"parent": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512}
				}
			},
			"name": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512},
					"suggest": { "type": "completion"}
				}
			},
			"capacity": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512},
					"suggest": { "type": "completion"}
				}
			},
			"capacityID": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512}
				}
			},
			"hasCapacity": {"type": "boolean"},
			"isCapacity": {"type": "boolean"},
			"nameWithContext": {
				"type": "text",
				"fields": {
					"keyword": {"type": "keyword", "ignore_above": 512},
					"suggest": { "type": "completion"}
				}
			}
		}
	}
}`

func V2SuggestMappingUpdate() string {
	return v2SuggestMappingUpdate
}

// v2MappingUpdate contains updates to the original model that are incremental,
// but will lead to index errors when these fields are not present due to the
// 'strict' on dynamic creating of new fields in the index.
var v2SuggestMappingUpdate = `{
  "properties": {
	"meta": {
		"type": "object",
		"properties": {
			"sourceID": {"type": "keyword"},
			"sourcePath": {"type": "keyword"},
			"groupID": {"type": "keyword"}
		}
	},
	"text": {
		"type": "text",
		"fields": {
			"keyword": {"type": "keyword", "ignore_above": 512},
			"suggest": { "type": "completion"}
		}
	}
  }
}
`
