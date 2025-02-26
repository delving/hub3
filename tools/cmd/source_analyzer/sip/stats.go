package sip

import (
	"fmt"
	"sync"
)

// Stats represents statistical information about an XML document
type Stats struct {
	Name                 string                     `json:"name"`
	MaxUniqueValueLength int                        `json:"maxUniqueValueLength"`
	Namespaces           map[string]string          `json:"namespaces"`
	PathStats            map[string]*PathValueStats `json:"pathStats"`
	mu                   sync.Mutex
}

// PathValueStats holds statistics for a specific path
type PathValueStats struct {
	Path         string           `json:"path"`
	ValueCount   int64            `json:"valueCount"`
	UniqueValues map[string]int64 `json:"uniqueValues"`
}

// NewStats creates a new Stats instance
func NewStats(name string, maxUniqueValueLength int) *Stats {
	return &Stats{
		Name:                 name,
		MaxUniqueValueLength: maxUniqueValueLength,
		Namespaces:           make(map[string]string),
		PathStats:            make(map[string]*PathValueStats),
	}
}

// RecordNamespace records a namespace and its prefix
func (s *Stats) RecordNamespace(prefix, uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Namespaces[prefix] = uri
}

// RecordValue records a value for a given path
func (s *Stats) RecordValue(path, value string) {
	if value == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Truncate value if it exceeds max length
	if len(value) > s.MaxUniqueValueLength {
		value = value[:s.MaxUniqueValueLength]
	}

	pathStats, exists := s.PathStats[path]
	if !exists {
		pathStats = &PathValueStats{
			Path:         path,
			UniqueValues: make(map[string]int64),
		}
		s.PathStats[path] = pathStats
	}

	pathStats.ValueCount++
	pathStats.UniqueValues[value]++
}

// GetPathCount returns the number of unique paths
func (s *Stats) GetPathCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.PathStats)
}

// GetPathStats returns statistics for a specific path
func (s *Stats) GetPathStats(path string) *PathValueStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.PathStats[path]
}

// String provides a string representation of the stats
func (s *Stats) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	str := fmt.Sprintf("Statistics for %s\n", s.Name)
	str += fmt.Sprintf("Namespaces: %d\n", len(s.Namespaces))
	str += fmt.Sprintf("Unique Paths: %d\n", len(s.PathStats))

	for path, stats := range s.PathStats {
		str += fmt.Sprintf("\nPath: %s\n", path)
		str += fmt.Sprintf("  Total Values: %d\n", stats.ValueCount)
		str += fmt.Sprintf("  Unique Values: %d\n", len(stats.UniqueValues))
	}

	return str
}
