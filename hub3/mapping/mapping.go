package mapping

// ESMapping is the default mapping for the RDF records enabled by hub3
var ESMapping = `{
	"settings": {
		"index": {
			"mapping.total_fields.limit": 1000,
			"mapping.depth.limit": 20,
			"mapping.nested_fields.limit": 50,
			"number_of_shards": 1,
			"number_of_replicas": 1
		}
	},
	"mappings":{
		"doc": {
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
				"tree": {
					"type": "object",
					"properties": {
						"depth": {"type": "integer"},
						"childCount": {"type": "integer"},
						"sortKey": {"type": "integer"},
						"doCount": {"type": "integer"},
						"hubID": {"type": "keyword"},
						"unitID": {"type": "keyword"},
						"type": {"type": "keyword"},
						"cLevel": {"type": "keyword"},
						"agencyCode": {"type": "keyword"},
						"inventoryID": {"type": "keyword"},
						"hasChildren": {"type": "boolean"},
						"label": {"type": "text"},
						"title": {"type": "text"},
						"description": {"type": "text"},
						"content": {"type": "text"},
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
						"hasRestriction": {"type": "boolean"}
					}
				},
				"subject": {"type": "keyword"},
				"predicate": {"type": "keyword"},
				"object": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						},
						"suggest": {
							"type": "completion",
							"contexts": [
								{ "name": "spec", "type": "category", "path": "meta.spec" },
								{ "name": "specType", "type": "category", "path": "_specType" },
								{ "name": "rdfType", "type": "category", "path": "_rdfType" },
								{ "name": "objectType", "type": "category", "path": "objectType" },
								{ "name": "language", "type": "category", "path": "language" },
								{ "name": "searchLabel", "type": "category", "path": "searchLabel" },
								{ "name": "orgID", "type": "category", "path": "meta.orgID"}
							]
						}
					}
				},
				"_specType": {"type": "keyword", "index": false, "store": false},
				"_rdfType": {"type": "keyword", "index": false, "store": false},
				"searchLabel": {"type": "keyword", "ignore_above": 256},
				"language": {"type": "keyword"},
				"dataType": {"type": "keyword"},
				"triple": {"type": "keyword", "index": "false", "store": "true"},
				"lodKey": {"type": "keyword"},
				"objectType": {"type": "keyword"},
				"recordType": {"type": "short"},
				"order": {"type": "integer"},
				"path": {"type": "keyword"},
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
								"latLong": {"type": "geo_point"}
							}
						}
					}
				}
			}
		}
}}`

// V1ESMapping has the legacy mapping for V1 indexes. It should only be used when indexV1 is enabled in the
// configuration.
var V1ESMapping = `
{
    "settings": {
        "analysis": {
            "filter": {
                "dutch_stop": {
                    "type":       "stop",
                    "stopwords":  "_dutch_"
                },
                "dutch_stemmer": {
                    "type":       "stemmer",
                    "language":   "dutch"
                },
                "dutch_override": {
                    "type":       "stemmer_override",
                    "rules": [
                        "fiets=>fiets",
                        "bromfiets=>bromfiets",
                        "ei=>eier",
                        "kind=>kinder"
                    ]
                }
            },
            "analyzer": {
                "dutch": {
                    "tokenizer":  "standard",
                    "filter": [
                        "lowercase",
                        "dutch_stop",
                        "dutch_override",
                        "dutch_stemmer"
                    ]
                }
            }
        }
    },
    "mappings": {
        "_default_":
            {
                "_all": {
                    "enabled": "true"
                },
                "date_detection": "false",
                "properties": {
					"full_text": {"type": "text"},
                    "id": {"type": "integer"},
                    "absolute_url": {"type": "keyword"},
                    "point": { "type": "geo_point" },
                    "delving_geohash": { "type": "geo_point" },
                    "delving_geoHash": { "type": "geo_point" },
                    "system": {
                        "properties": {
							"about_uri": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"caption": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"preview": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "created_at": {"format": "dateOptionalTime", "type": "date"},
							"graph_name": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "modified_at": {"format": "dateOptionalTime", "type": "date"},
							"slug": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "geohash": { "type": "geo_point" },
                            "source_graph": { "index": "false", "type": "text", "doc_values": "false" },
							"source_uri": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"spec": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"thumbnail": {"fields": {"raw": { "type": "keyword"}}, "type": "text"}
                    }
                }},
                "dynamic_templates": [
                    {"legacy": { "path_match": "legacy.*",
                        "mapping": { "type": "keyword",
                            "fields": { "raw": { "type": "keyword"}, "value": { "type": "text" } }
                        }
                    }},
                    {"dates": { "match": "*_at", "mapping": { "type": "date" } }},
                    {"rdf": {
                        "path_match": "rdf.*",
                        "mapping": {
                            "type": "text",
                            "fields": {
                                "raw": {
                                    "type": "keyword"
                                },
                                "value": {
                                    "type": "text"
                                }
                            }
                        }
                    }},
                    {"uri": { "match": "id", "mapping": { "type": "keyword" } }},
                    {"point": { "match": "point", "mapping": { "type": "geo_point" }}},
                    {"geo_hash": { "match": "delving_geohash", "mapping": { "type": "geo_point" } }},
                    {"value": {
						"match": "value",
						"mapping": { "type": "text" }
					}},
                    {"raw": {
						"match": "raw",
						"mapping": {"type": "keyword", "ignore_above": 1024}
					}},
                    {"id": { "match": "id", "mapping": { "type": "keyword" } }},
                    {"graphs": { "match": "*_graph", "mapping": { "type": "text", "index": "false" } }},
                    {"inline": { "match": "inline", "mapping": { "type": "object", "include_in_parent": "true" } }},
                    {"strings": {
                        "match_mapping_type": "string",
                        "mapping": {"type": "text", "fields": {"raw": {"type": "keyword", "ignore_above": 1024 }}}
                    }}
                ]
            }
    }}
`
