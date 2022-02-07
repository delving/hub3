// Package elasticsearch provides a Driver for the ElasticSerach
// search engine.
//
// The elasticsearch.Client wraps both the official ElasticSearch client
// and the olivere/elastic search DSL. Ikuzo package should not directly use
// these libraries but use this wrapper client instead.
package elasticsearch
