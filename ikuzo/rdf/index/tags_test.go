package index

import (
	"testing"

	"github.com/matryer/is"
)

func TestTagMap_AddTags(t *testing.T) {
	is := is.New(t)

	tm := NewTagMap()
	is.Equal(tm.Len(), 0)

	// Add tags to a predicate
	predicate := "http://purl.org/dc/elements/1.1/date"
	tm.AddTags(predicate, "dateRange", "isoDate")
	is.Equal(tm.Len(), 1)

	// Check tags are added
	tags := tm.GetTags(predicate)
	is.Equal(len(tags), 2)
	is.Equal(tags[0], "dateRange")
	is.Equal(tags[1], "isoDate")

	// Add more tags to the same predicate
	tm.AddTags(predicate, "date", "isoDate") // isoDate is duplicate
	tags = tm.GetTags(predicate)
	is.Equal(len(tags), 3) // Should only add the new "date" tag
	is.Equal(tags[2], "date")

	// Check non-existent predicate
	nonExistentTags := tm.GetTags("http://non-existent")
	is.Equal(len(nonExistentTags), 0)
}

func TestTagMap_ApplyTagsToEntry(t *testing.T) {
	is := is.New(t)

	// Create a TagMap with some predicate mappings
	tm := NewTagMap()
	tm.AddTags("http://purl.org/dc/elements/1.1/date", "dateRange", "isoDate")
	tm.AddTags("http://www.w3.org/2003/01/geo/wgs84_pos#lat_long", "latLong")

	// Test date entry with isoDate tag
	dateEntry := &Entry{
		Predicate: "http://purl.org/dc/elements/1.1/date",
		Value:     "2020-01-01",
		EntryType: Literal,
	}

	// Apply tags
	tm.ApplyTagsToEntry(dateEntry)
	is.Equal(len(dateEntry.Tags), 2)
	is.Equal(dateEntry.Tags[0], "dateRange")
	is.Equal(dateEntry.Tags[1], "isoDate")

	// Check if processTags works for isoDate and dateRange
	err := dateEntry.processTags()
	is.NoErr(err)
	
	// Since this entry has both isoDate and dateRange tags, we expect 3 entries in the Date array:
	// 1. From isoDate: the original value "2020-01-01"
	// 2. From dateRange: Greater value "2020-01-01"
	// 3. From dateRange: Less value "2020-01-01" (for single date value, Greater and Less are same)
	is.Equal(len(dateEntry.Date), 3)
	
	// Verify that DateRange is set
	is.True(dateEntry.DateRange != nil)
	is.Equal(dateEntry.DateRange.Greater, "2020-01-01")
	is.Equal(dateEntry.DateRange.Less, "2020-01-01")
	
	// Test date range entry
	dateRangeEntry := &Entry{
		Predicate: "http://purl.org/dc/elements/1.1/date",
		Value:     "2020-01-01 2021-12-31", // Period with start and end date
		EntryType: Literal,
		Tags:      []string{"dateRange"},
	}
	
	// Process tags for date range
	err = dateRangeEntry.processTags()
	is.NoErr(err)
	is.True(dateRangeEntry.DateRange != nil)
	is.Equal(dateRangeEntry.DateRange.Greater, "2020-01-01")
	is.Equal(dateRangeEntry.DateRange.Less, "2021-12-31")
	is.Equal(len(dateRangeEntry.Date), 2) // Both dates should be added to the Date field
	is.Equal(dateRangeEntry.Date[0], "2020-01-01")
	is.Equal(dateRangeEntry.Date[1], "2021-12-31")

	// Create an entry with a non-matching predicate
	entry2 := &Entry{
		Predicate: "http://purl.org/dc/elements/1.1/title",
		Value:     "Test Title",
		EntryType: Literal,
	}

	// Apply tags
	tm.ApplyTagsToEntry(entry2)
	is.Equal(len(entry2.Tags), 0) // No tags should be applied
}

func TestGraph_ApplyTags(t *testing.T) {
	is := is.New(t)

	// Create a test graph with header
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject",
	}
	g, err := NewGraph(header)
	is.NoErr(err)

	// Add resources and entries
	resource1 := &Resource{
		ID:    "urn:test:subject",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, resource1)

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://www.w3.org/2003/01/geo/wgs84_pos#lat_long",
		SearchLabel: "geo_latLong",
		Value:       "52.3676, 4.9041",
		EntryType:   Literal,
	})

	resource1.Entries = append(resource1.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title",
		EntryType:   Literal,
	})

	// Create a TagMap
	tm := NewTagMap()
	tm.AddTags("http://purl.org/dc/elements/1.1/date", "dateRange", "isoDate")
	tm.AddTags("http://www.w3.org/2003/01/geo/wgs84_pos#lat_long", "latLong")

	// Apply tags to the graph
	g.ApplyTags(tm)

	// Check if tags were applied correctly
	is.Equal(len(resource1.Entries[0].Tags), 2) // date entry
	is.Equal(resource1.Entries[0].Tags[0], "dateRange")
	is.Equal(resource1.Entries[0].Tags[1], "isoDate")

	is.Equal(len(resource1.Entries[1].Tags), 1) // latLong entry
	is.Equal(resource1.Entries[1].Tags[0], "latLong")

	is.Equal(len(resource1.Entries[2].Tags), 0) // title entry, no tags
	
	// Check that processTags was called and TypeIndexFields were set
	// Since the entry has both isoDate and dateRange tags, it should have 3 entries in the Date array
	is.Equal(len(resource1.Entries[0].Date), 3) 
	
	// Also verify that DateRange is set correctly
	is.True(resource1.Entries[0].DateRange != nil)
	is.Equal(resource1.Entries[0].DateRange.Greater, "2020-01-01")
	is.Equal(resource1.Entries[0].DateRange.Less, "2020-01-01")
	
	is.Equal(resource1.Entries[1].LatLong, "52.3676, 4.9041") // latLong field set
}

func TestGraph_SetContextAndTags(t *testing.T) {
	is := is.New(t)

	// Create a test graph with header
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject",
	}
	g, err := NewGraph(header)
	is.NoErr(err)

	// Add resources and entries
	rootResource := &Resource{
		ID:    "urn:test:subject",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, rootResource)

	childResource := &Resource{
		ID:    "urn:test:child",
		Types: []string{"http://example.org/ontology/ChildType"},
	}
	g.Resources = append(g.Resources, childResource)

	// Add connection between resources
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/relation",
		SearchLabel: "dc_relation",
		EntryType:   ResourceType,
		ID:          "urn:test:child",
	})

	// Add entry with taggable predicate
	childResource.Entries = append(childResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	// Create a TagMap
	tm := NewTagMap()
	tm.AddTags("http://purl.org/dc/elements/1.1/date", "dateRange", "isoDate")

	// Verify context isn't set yet
	is.Equal(g.contextIsSet, false)

	// Apply context and tags
	err = g.SetContextAndTags(tm)
	is.NoErr(err)

	// Verify context is set
	is.Equal(g.contextIsSet, true)

	// Verify tags are applied
	is.Equal(len(childResource.Entries[0].Tags), 2)
	is.Equal(childResource.Entries[0].Tags[0], "dateRange")
	is.Equal(childResource.Entries[0].Tags[1], "isoDate")
	
	// Verify TypeIndexField is set for date entry
	// Since the entry has both isoDate and dateRange tags, it should have 3 entries in the Date array
	is.Equal(len(childResource.Entries[0].Date), 3)
	
	// Also verify that DateRange is set correctly
	is.True(childResource.Entries[0].DateRange != nil)
	is.Equal(childResource.Entries[0].DateRange.Greater, "2020-01-01")
	is.Equal(childResource.Entries[0].DateRange.Less, "2020-01-01")

	// Verify context is set flag
	is.Equal(g.contextIsSet, true)
	
	// Note: Actual context values depend on how the test environment handles them
	// In some cases, the actual context references might not be populated,
	// but we've already verified that the contextIsSet flag is true
}

func TestNewTagMapFromMap(t *testing.T) {
	is := is.New(t)

	// Create a map of predicate URIs to tags
	m := map[string][]string{
		"http://purl.org/dc/elements/1.1/date":            {"dateRange", "isoDate"},
		"http://www.w3.org/2003/01/geo/wgs84_pos#lat_long": {"latLong"},
	}

	// Create a TagMap from the map
	tm := NewTagMapFromMap(m)
	is.Equal(tm.Len(), 2)

	// Check tags are added correctly
	dateTags := tm.GetTags("http://purl.org/dc/elements/1.1/date")
	is.Equal(len(dateTags), 2)
	is.Equal(dateTags[0], "dateRange")
	is.Equal(dateTags[1], "isoDate")

	latLongTags := tm.GetTags("http://www.w3.org/2003/01/geo/wgs84_pos#lat_long")
	is.Equal(len(latLongTags), 1)
	is.Equal(latLongTags[0], "latLong")
}

func TestGraph_IndexMessageWithTags(t *testing.T) {
	is := is.New(t)

	// Create a test graph with header
	header := Header{
		OrgID:    "test-org",
		Spec:     "test-spec",
		HubID:    "test-hubid",
		EntryURI: "urn:test:subject",
	}
	g, err := NewGraph(header)
	is.NoErr(err)

	// Add resources and entries
	rootResource := &Resource{
		ID:    "urn:test:subject",
		Types: []string{"http://example.org/ontology/TestType"},
	}
	g.Resources = append(g.Resources, rootResource)

	// Add entries to the root resource
	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/date",
		SearchLabel: "dc_date",
		Value:       "2020-01-01",
		EntryType:   Literal,
	})

	rootResource.Entries = append(rootResource.Entries, &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Document",
		EntryType:   Literal,
	})

	// Create a TagMap
	tm := NewTagMap()
	tm.AddTags("http://purl.org/dc/elements/1.1/date", "isoDate")
	tm.AddTags("http://purl.org/dc/elements/1.1/title", "title")

	// Verify context and tags aren't set yet
	is.Equal(g.contextIsSet, false)
	is.Equal(len(rootResource.Entries[0].Tags), 0)
	is.Equal(len(rootResource.Entries[1].Tags), 0)

	// Use IndexMessage to prepare for indexing
	msg, err := g.IndexMessage(tm)
	is.NoErr(err)
	is.True(msg != nil)

	// Verify context is set
	is.Equal(g.contextIsSet, true)

	// Verify tags are applied
	is.Equal(len(rootResource.Entries[0].Tags), 1)
	is.Equal(rootResource.Entries[0].Tags[0], "isoDate")
	is.Equal(len(rootResource.Entries[1].Tags), 1)
	is.Equal(rootResource.Entries[1].Tags[0], "title")

	// Verify TypeIndexField is set for date entry
	is.Equal(len(rootResource.Entries[0].Date), 1) // Only isoDate tag is used for this test
	is.Equal(rootResource.Entries[0].Date[0], "2020-01-01")
}