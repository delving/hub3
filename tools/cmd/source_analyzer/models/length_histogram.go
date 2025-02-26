// github.com/delving/hub3/tools/cmd/source_analyzer/models/length_histogram.go
package models

import (
	"encoding/json"
	"fmt"
	"sync"
)

// LengthRange represents a range for length histogram buckets
type LengthRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// String provides a string representation of the range
func (lr *LengthRange) String() string {
	if lr.To < 0 {
		return fmt.Sprintf("%d-*", lr.From)
	}
	if lr.To > 0 {
		return fmt.Sprintf("%d-%d", lr.From, lr.To)
	}
	return fmt.Sprintf("%d", lr.From)
}

// Fits checks if a value length falls within this range
func (lr *LengthRange) Fits(value int) bool {
	if lr.To == 0 {
		return value == lr.From
	}
	if lr.To < 0 {
		return value >= lr.From
	}
	return value >= lr.From && value <= lr.To
}

// Counter represents count for a specific length range
type Counter struct {
	Range LengthRange `json:"range"`
	Count int         `json:"count"`
}

// LengthHistogram maintains counts of string lengths in predefined ranges
type LengthHistogram struct {
	Counters []*Counter `json:"counters"`
	mu       sync.Mutex
}

// NewLengthHistogram creates a new histogram with predefined ranges
func NewLengthHistogram() *LengthHistogram {
	ranges := []LengthRange{
		{From: 0, To: 0},    // Empty strings
		{From: 1, To: 1},    // Single character
		{From: 2, To: 2},    // Two characters
		{From: 3, To: 3},    // Three characters
		{From: 4, To: 4},    // Four characters
		{From: 5, To: 5},    // Five characters
		{From: 6, To: 10},   // Short strings
		{From: 11, To: 15},  // Medium strings
		{From: 16, To: 20},  // Medium-long strings
		{From: 21, To: 30},  // Long strings
		{From: 31, To: 50},  // Very long strings
		{From: 51, To: 100}, // Extra long strings
		{From: 101, To: -1}, // Everything else
	}

	counters := make([]*Counter, len(ranges))
	for i, r := range ranges {
		counters[i] = &Counter{Range: r}
	}

	return &LengthHistogram{
		Counters: counters,
	}
}

// Record adds a string length to the histogram
func (h *LengthHistogram) Record(length int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, counter := range h.Counters {
		if counter.Range.Fits(length) {
			counter.Count++
		}
	}
}

// IsEmpty checks if the histogram has any counts
func (h *LengthHistogram) IsEmpty() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, counter := range h.Counters {
		if counter.Count > 0 {
			return false
		}
	}
	return true
}

// GetTotal returns the total number of values recorded
func (h *LengthHistogram) GetTotal() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	total := 0
	for _, counter := range h.Counters {
		total += counter.Count
	}
	return total
}

// GetRangeCount returns the count for a specific range
func (h *LengthHistogram) GetRangeCount(from, to int) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, counter := range h.Counters {
		if counter.Range.From == from && counter.Range.To == to {
			return counter.Count
		}
	}
	return 0
}

// MarshalJSON implements custom JSON marshaling
func (h *LengthHistogram) MarshalJSON() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Only include counters with non-zero counts
	nonZeroCounters := make([]*Counter, 0)
	for _, counter := range h.Counters {
		if counter.Count > 0 {
			nonZeroCounters = append(nonZeroCounters, counter)
		}
	}

	type Alias LengthHistogram
	return json.Marshal(&struct {
		Counters []*Counter `json:"counters"`
	}{
		Counters: nonZeroCounters,
	})
}
