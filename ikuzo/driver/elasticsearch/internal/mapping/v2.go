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

const (
	defaultShards   = 1
	defaultReplicas = 0
)

func V2ESMapping(shards, replicas int) string {
	shards, replicas = setDefaults(shards, replicas)

	return fmt.Sprintf(
		v2Mapping,
		shards,
		replicas,
	)
}

// v2Mapping is the default mapping for the RDF records enabled by hub3
var v2Mapping = `{
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
						"modified": {"type": "date"},
						"sourceID": {"type": "keyword"},
						"sourcePath": {"type": "keyword"},
						"groupID": {"type": "keyword"}
					}
				},
				"protobuf": {
					"type": "object",
					"properties": {
						"messageType": {"type": "keyword"},
						"data": {
							"type": "keyword",
							"store": true,
							"index": false,
							"doc_values": false
						}
					}
				},
				"tree": {
					"type": "object",
					"properties": {
						"depth": {"type": "integer"},
						"childCount": {"type": "integer"},
						"sortKey": {"type": "integer"},
						"doCount": {"type": "integer"},
						"hubID": {"type": "keyword"},
						"material": {"type": "keyword"},
						"unitID": {"type": "keyword"},
						"type": {"type": "keyword"},
						"cLevel": {"type": "keyword"},
						"physDesc": {"type": "keyword"},
						"agencyCode": {"type": "keyword"},
						"inventoryID": {"type": "keyword"},
						"hasChildren": {"type": "boolean"},
						"label": {
							"type": "text",
							"fields": {
								"keyword": {"type": "keyword", "ignore_above": 512},
								"suggest": { "type": "completion"}
							}
						},
						"title": {"type": "text"},
						"description": {"type": "text"},
						"content": {"type": "text"},
						"periodDesc": { "type": "keyword"},
						"rawContent": {"type": "text", "store": false},
						"access": {
							"type": "text",
							"fields": {
								"keyword": {"type": "keyword", "ignore_above": 512}
							}
						},
						"parent": {"type": "keyword"},
						"leaf": {"type": "keyword"},
						"daoLink": {"type": "keyword"},
						"manifestLink": {"type": "keyword"},
						"mimeType": {"type": "keyword"},
						"periods": {"type": "keyword"},
						"hasDigitalObject": {"type": "boolean"},
						"hasRestriction": {"type": "boolean"},
						"genreform": {"type": "keyword"}
					}
				},
				"recordType": {"type": "short"},
				"full_text": {"type": "text"},
				"resources": {
					"type": "nested",
					"properties": {
						"id": {"type": "keyword"},
						"types": {"type": "keyword"},
						"tags": {"type": "keyword"},
						"context": {
							"type": "nested",
							"properties": {
								"Subject": {"type": "keyword", "ignore_above": 256},
								"SubjectClass": {"type": "keyword", "ignore_above": 256},
								"Predicate": {"type": "keyword", "ignore_above": 256},
								"SearchLabel": {"type": "keyword", "ignore_above": 256},
								"Level": {"type": "integer"},
								"ObjectID": {"type": "keyword", "ignore_above": 256},
								"SortKey": {"type": "integer"},
								"Label": {"type": "keyword"}
							}
						},
						"entries": {
							"type": "nested",
							"properties": {
								"@id": {"type": "keyword"},
								"@value": {
									"type": "text",
									"copy_to": "full_text",
									"fields": {
										"keyword": {"type": "keyword", "ignore_above": 256}
									}
								},
								"searchLabel": {"type": "keyword", "ignore_above": 256},
								"@language": {"type": "keyword", "ignore_above": 256},
								"@type": {"type": "keyword", "ignore_above": 256},
								"entrytype": {"type": "keyword", "ignore_above": 256},
								"predicate": {"type": "keyword", "ignore_above": 256},
								"level": {"type": "integer"},
								"order": {"type": "integer"},
								"integer": {"type": "integer"},
								"tags": {"type": "keyword"},
								"isoDate": {
									"type": "date",
									"format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||dd-MM-yyy||yyyy||epoch_millis"
								},
								"dateRange": {
									"type": "date_range",
									"format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||dd-MM-yyy||yyyy||epoch_millis"
								},
								"intRange": {"type": "integer_range"},
								"float": {"type": "float"},
								"latLong": {"type": "geo_point"}
							}
						}
					}
				}
			}
		}
}`

func V2MappingUpdate() string {
	return v2MappingUpdate
}

// v2MappingUpdate contains updates to the original model that are incremental,
// but will lead to index errors when these fields are not present due to the
// 'strict' on dynamic creating of new fields in the index.
var v2MappingUpdate = `{
  "properties": {
	"meta": {
		"type": "object",
		"properties": {
			"sourceID": {"type": "keyword"},
			"sourcePath": {"type": "keyword"},
			"groupID": {"type": "keyword"}
		}
	},
    "tree": {
      "properties": {
        "physDesc": {"type": "keyword"},
        "periodDesc": { "type": "keyword"},
		"rawContent": {"type": "text", "store": false},
		"genreform": {"type": "keyword"}
      }
    },
	"protobuf": {
		"type": "object",
		"properties": {
			"messageType": {"type": "keyword"},
			"data": {
				"type": "keyword",
				"store": true,
				"index": false,
				"doc_values": false
			}
		}
	},
	"resources": {
		"type": "nested",
		"properties": {
			"entries": {
				"type": "nested",
				"properties": {
					"intRange": {"type": "integer_range"},
					"float": {"type": "float"},
					"level": {"type": "integer"}
				}
			}
		}
	}
  }
}
`
