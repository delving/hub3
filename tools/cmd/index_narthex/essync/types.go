// Package essync provides tools for efficient synchronization of documents
// between a source and Elasticsearch, with support for change detection,
// checksumming, and bulk operations.
package essync

import (
	"time"
)

// Change represents a single document operation (index or delete)
// to be performed in Elasticsearch.
type Change struct {
	Operation string                 `json:"operation"`
	ID        string                 `json:"_id"`
	Index     string                 `json:"_index"`
	Source    map[string]interface{} `json:"_source,omitempty"`
	Checksum  string                 `json:"_checksum,omitempty"`
}

// ErrorStats contains statistics about the sync operation
// and any errors that occurred.
type ErrorStats struct {
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalChanges int32          `json:"total_changes"`
	Successful   int32          `json:"successful"`
	Failed       int32          `json:"failed"`
	ErrorTypes   map[string]int `json:"error_types"`
}

// ErrorRecord represents a single failed operation attempt.
type ErrorRecord struct {
	Timestamp time.Time   `json:"timestamp"`
	Index     string      `json:"index"`
	Error     string      `json:"error"`
	Document  interface{} `json:"document"`
	Attempt   int         `json:"attempt"`
}

// ErrorFile represents the structure of the error log file.
type ErrorFile struct {
	Stats  ErrorStats               `json:"stats"`
	Errors map[string][]ErrorRecord `json:"errors"` // key is document ID
}

// LoadOptions configures how existing documents are loaded from Elasticsearch.
type LoadOptions struct {
	Index       string                 // Index name to load from
	Query       map[string]interface{} // Query to filter documents
	BatchSize   int                    // Number of documents per scroll batch
	FnameFilter string
}

// DefaultLoadOptions returns LoadOptions with sensible defaults for the given index.
func DefaultLoadOptions(index string) LoadOptions {
	return LoadOptions{
		Index:     index,
		Query:     map[string]interface{}{"match_all": map[string]interface{}{}},
		BatchSize: 1000,
	}
}
