package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/delving/hub3/ikuzo/rdf/index"
)

// RDFTagConfig represents the TOML configuration for RDF tags
type RDFTagConfig struct {
	// RDFTag section with different tag types
	RDFTag struct {
		Title         []string `toml:"title"`
		Label         []string `toml:"label"`
		Owner         []string `toml:"owner"`
		Thumbnail     []string `toml:"thumbnail"`
		LandingPage   []string `toml:"landingPage"`
		Description   []string `toml:"description"`
		Subject       []string `toml:"subject"`
		Date          []string `toml:"date"`
		Collection    []string `toml:"collection"`
		SubCollection []string `toml:"subCollection"`
		ObjectType    []string `toml:"objectType"`
		ObjectID      []string `toml:"objectID"`
		Creator       []string `toml:"creator"`
		LatLong       []string `toml:"latLong"`
		DateRange     []string `toml:"dateRange"`
		IsoDate       []string `toml:"isoDate"`
		Integer       []string `toml:"integer"`
		IntegerRange  []string `toml:"integerRange"`
	} `toml:"rdftag"`
}

// LoadRDFTagMapFromTOML loads a TOML file and returns a TagMap for RDF indexing
func LoadRDFTagMapFromTOML(filePath string) (*index.TagMap, error) {
	// Ensure the config file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tag map file not found: %s", filePath)
	}

	var config RDFTagConfig
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode TOML tag map: %w", err)
	}

	// Create a predicate to tags map
	predicateToTags := make(map[string][]string)
	
	// Process each tag type and add to the map
	for _, uri := range config.RDFTag.Title {
		addTag(predicateToTags, uri, "title")
	}
	for _, uri := range config.RDFTag.Label {
		addTag(predicateToTags, uri, "label")
	}
	for _, uri := range config.RDFTag.Owner {
		addTag(predicateToTags, uri, "owner")
	}
	for _, uri := range config.RDFTag.Thumbnail {
		addTag(predicateToTags, uri, "thumbnail")
	}
	for _, uri := range config.RDFTag.LandingPage {
		addTag(predicateToTags, uri, "landingPage")
	}
	for _, uri := range config.RDFTag.Description {
		addTag(predicateToTags, uri, "description")
	}
	for _, uri := range config.RDFTag.Subject {
		addTag(predicateToTags, uri, "subject")
	}
	for _, uri := range config.RDFTag.Date {
		addTag(predicateToTags, uri, "date")
	}
	for _, uri := range config.RDFTag.Collection {
		addTag(predicateToTags, uri, "collection")
	}
	for _, uri := range config.RDFTag.SubCollection {
		addTag(predicateToTags, uri, "subCollection")
	}
	for _, uri := range config.RDFTag.ObjectType {
		addTag(predicateToTags, uri, "objectType")
	}
	for _, uri := range config.RDFTag.ObjectID {
		addTag(predicateToTags, uri, "objectID")
	}
	for _, uri := range config.RDFTag.Creator {
		addTag(predicateToTags, uri, "creator")
	}
	for _, uri := range config.RDFTag.LatLong {
		addTag(predicateToTags, uri, "latLong")
	}
	for _, uri := range config.RDFTag.DateRange {
		addTag(predicateToTags, uri, "dateRange")
	}
	for _, uri := range config.RDFTag.IsoDate {
		addTag(predicateToTags, uri, "isoDate")
	}
	for _, uri := range config.RDFTag.Integer {
		addTag(predicateToTags, uri, "integer")
	}
	for _, uri := range config.RDFTag.IntegerRange {
		addTag(predicateToTags, uri, "integerRange")
	}

	// Create the TagMap from the loaded configuration
	tagMap := index.NewTagMapFromMap(predicateToTags)

	slog.Info("loaded RDF tag map", 
		"filePath", filePath, 
		"predicates", tagMap.Len())

	return tagMap, nil
}

// addTag adds a tag to the predicate map, handling the case where a predicate
// should have multiple tags
func addTag(predicateToTags map[string][]string, uri string, tag string) {
	if uri == "" {
		return
	}
	
	if tags, exists := predicateToTags[uri]; exists {
		predicateToTags[uri] = append(tags, tag)
	} else {
		predicateToTags[uri] = []string{tag}
	}
}