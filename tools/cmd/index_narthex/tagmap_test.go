package main

import (
	"os"
	"testing"
)

func TestLoadRDFTagMapFromTOML(t *testing.T) {
	// Create a temporary test TOML file
	content := `[rdftag]
# used for title of a resource
title = [
  "http://purl.org/dc/elements/1.1/title",
  "https://archief.nl/def/ead/idUnittitle",
]

label = [
  "http://purl.org/dc/elements/1.1/title",
  "http://www.w3.org/2004/02/skos/core#prefLabel",
]

latLong = ["http://schemas.delving.eu/nave/terms/latLong"]
dateRange = ["https://archief.nl/def/ead/dateNormal"]
isoDate = ["https://archief.nl/def/ead/dateiso"]
`

	tmpFile, err := os.CreateTemp("", "rdf-tagmap-test-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err = tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Test loading the tag map
	tagMap, err := LoadRDFTagMapFromTOML(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load tag map: %v", err)
	}

	// Verify the map was properly loaded
	if tagMap.Len() != 6 {
		t.Errorf("Expected 6 entries in tag map, got %d", tagMap.Len())
	}

	// Check specific entries
	titleTags := tagMap.GetTags("http://purl.org/dc/elements/1.1/title")
	if len(titleTags) != 2 || titleTags[0] != "title" || titleTags[1] != "label" {
		t.Errorf("Expected title predicate to have tags [title label], got %v", titleTags)
	}

	dateRangeTags := tagMap.GetTags("https://archief.nl/def/ead/dateNormal")
	if len(dateRangeTags) != 1 || dateRangeTags[0] != "dateRange" {
		t.Errorf("Expected dateRange predicate to have tag [dateRange], got %v", dateRangeTags)
	}
}