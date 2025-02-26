// github.com/delving/hub3/tools/cmd/source_analyzer/sip/analyzer.go
package sip

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	sio "github.com/delving/hub3/tools/cmd/source_analyzer/io"
)

// Analyzer processes XML documents and generates statistics
type Analyzer struct {
	config           *Config
	stats            *Stats
	listener         AnalysisListener
	progressListener ProgressListener
	startTime        time.Time
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(config *Config, listener AnalysisListener) *Analyzer {
	return &Analyzer{
		config:    config,
		stats:     NewStats(config.DatasetName, config.MaxUniqueValueLength),
		listener:  listener,
		startTime: time.Now(),
	}
}

// SetProgressListener sets the progress listener
func (a *Analyzer) SetProgressListener(listener ProgressListener) {
	a.progressListener = listener
	if listener != nil {
		listener.SetProgressMessage("Analyzing Data")
	}
}

// Process analyzes an XML file
func (a *Analyzer) Process() error {
	// Open file with compression support
	reader, closer, err := sio.CreateBufferedReader(a.config.InputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer closer.Close()

	decoder := xml.NewDecoder(reader)
	path := make([]string, 0)
	text := strings.Builder{}
	count := 0

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			a.listener.Failure("XML parsing error", err)
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Progress reporting
			count++
			if count%a.config.ElementStep == 0 && a.progressListener != nil {
				a.progressListener.SetProgress(count)
			}

			// Handle namespaces
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" {
					a.stats.RecordNamespace(attr.Name.Local, attr.Value)
				}
			}

			// Handle text content from previous element
			chunk := strings.TrimSpace(text.String())
			if chunk != "" {
				a.stats.RecordValue(a.buildPath(path), chunk)
			}
			text.Reset()

			// Update path
			path = append(path, t.Name.Local)
			currentPath := a.buildPath(path)

			// Handle attributes
			for _, attr := range t.Attr {
				if attr.Name.Space != "xmlns" {
					attrPath := fmt.Sprintf("%s/@%s", currentPath, attr.Name.Local)
					a.stats.RecordValue(attrPath, attr.Value)
				}
			}

		case xml.CharData:
			text.Write(t)

		case xml.EndElement:
			// Record text content
			chunk := strings.TrimSpace(text.String())
			if chunk != "" {
				a.stats.RecordValue(a.buildPath(path), chunk)
			}
			text.Reset()

			// Update path
			path = path[:len(path)-1]
		}
	}

	// Write stats to output directory
	if err := a.config.WriteStats(a.stats); err != nil {
		a.listener.Failure("Failed to write statistics", err)
		return err
	}

	a.listener.Success(a.stats)
	return nil
}

// buildPath creates a path string from path components
func (a *Analyzer) buildPath(components []string) string {
	return "/" + strings.Join(components, "/")
}
