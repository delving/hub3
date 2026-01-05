package main

import (
	"log/slog"

	"github.com/delving/hub3/ikuzo/rdf/index"
)

// processSingleDirectoryWithTagMap extends processSingleDirectory with TagMap support
func processSingleDirectoryWithTagMap(dirPath, targetIndex, customPattern string, tagMap *index.TagMap, maxRecordSize int) (*NarthexStats, error) {
	rdfXML, err := findProcessedFileWithPattern(dirPath, customPattern)
	if err != nil {
		return nil, err
	}

	slog.Info("processing directory", "dir", dirPath, "rdfFile", rdfXML, "tagMapPresent", tagMap != nil, "maxRecordSizeMB", maxRecordSize)

	r, err := openXMLReader(rdfXML)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	cfg := Config{
		IndexName:     targetIndex,
		TagMap:        tagMap,
		MaxRecordSize: maxRecordSize,
	}
	cfg.OutputPath, cfg.DatePrefix = processPath(rdfXML)
	if cfg.DatePrefix == "" {
		return nil, err
	}

	stats, err := ParseNarthex(r, cfg)
	if err != nil {
		return nil, err
	}

	return stats, nil
}