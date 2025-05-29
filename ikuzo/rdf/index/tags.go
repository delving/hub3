package index

import (
	"log/slog"
	"sync"
)

// TagMap contains a map of predicate URIs to tag names
type TagMap struct {
	sync.RWMutex
	predicateToTags map[string][]string
}

// NewTagMap creates a new TagMap
func NewTagMap() *TagMap {
	return &TagMap{
		predicateToTags: make(map[string][]string),
	}
}

// FromMap creates a TagMap from a map of predicate URIs to tag slices
func NewTagMapFromMap(m map[string][]string) *TagMap {
	tm := NewTagMap()
	
	if m == nil {
		return tm
	}
	
	for uri, tags := range m {
		tm.AddTags(uri, tags...)
	}
	
	return tm
}

// AddTags adds tags to a predicate URI
func (tm *TagMap) AddTags(predicateURI string, tags ...string) {
	tm.Lock()
	defer tm.Unlock()
	
	current, exists := tm.predicateToTags[predicateURI]
	if !exists {
		tm.predicateToTags[predicateURI] = tags
		return
	}
	
	// Add only unique tags
	for _, tag := range tags {
		if !containsTag(current, tag) {
			current = append(current, tag)
		}
	}
	tm.predicateToTags[predicateURI] = current
}

// GetTags returns tags for a predicate URI
func (tm *TagMap) GetTags(predicateURI string) []string {
	tm.RLock()
	defer tm.RUnlock()
	
	tags, exists := tm.predicateToTags[predicateURI]
	if !exists {
		return nil
	}
	
	return tags
}

// Len returns the number of predicate URIs in the TagMap
func (tm *TagMap) Len() int {
	tm.RLock()
	defer tm.RUnlock()
	
	return len(tm.predicateToTags)
}

// ApplyTagsToEntry adds tags to an Entry based on its predicate
func (tm *TagMap) ApplyTagsToEntry(entry *Entry) {
	if entry == nil {
		return
	}
	
	tags := tm.GetTags(entry.Predicate)
	if len(tags) == 0 {
		return
	}
	
	entry.AddTags(tags...)
}

// containsTag checks if a string is in a slice
func containsTag(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ApplyTags adds tags to all entries in a Graph and processes them
func (g *Graph) ApplyTags(tagMap *TagMap) {
	if tagMap == nil || tagMap.Len() == 0 {
		return
	}
	
	count := 0
	processedCount := 0
	for _, resource := range g.Resources {
		for _, entry := range resource.Entries {
			tagMap.ApplyTagsToEntry(entry)
			if len(entry.Tags) > 0 {
				count++
				// Process the tags immediately to ensure the TypeIndexField values are set
				if err := entry.processTags(); err != nil {
					slog.Warn("Failed to process tags", "error", err, "entry", entry)
				} else {
					processedCount++
				}
			}
		}
	}
	
	slog.Debug("Applied tags to entries", 
		"count", count, 
		"processed", processedCount, 
		"total_resources", len(g.Resources))
}

// SetContextAndTags sets both context levels and applies tags to the Graph
func (g *Graph) SetContextAndTags(tagMap *TagMap) error {
	// First set context levels
	if !g.contextIsSet {
		if err := g.addContextLevels(); err != nil {
			return err
		}
	}
	
	// Then apply tags
	g.ApplyTags(tagMap)
	
	return nil
}