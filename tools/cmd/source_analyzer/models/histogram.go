// github.com/delving/hub3/tools/cmd/source_analyzer/models/histogram.go
package models

// HistogramEntry represents an entry in the histogram
type HistogramEntry struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
