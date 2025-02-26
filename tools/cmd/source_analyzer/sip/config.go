// github.com/delving/hub3/tools/cmd/source_analyzer/sip/config.go
package sip

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the analyzer configuration
type Config struct {
	// Input parameters
	InputFile            string
	OutputDir            string
	DatasetName          string
	MaxUniqueValueLength int

	// Optional parameters
	CompressOutput bool
	ElementStep    int // Progress reporting interval
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig(inputFile, outputDir string) *Config {
	return &Config{
		InputFile:            inputFile,
		OutputDir:            outputDir,
		DatasetName:          filepath.Base(inputFile),
		MaxUniqueValueLength: 1000,
		ElementStep:          10000,
	}
}

// WriteStats writes the statistics to the output directory
func (c *Config) WriteStats(stats *Stats) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(c.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Write full statistics
	statsPath := filepath.Join(c.OutputDir, "stats.json")
	statsData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %v", err)
	}
	if err := os.WriteFile(statsPath, statsData, 0o644); err != nil {
		return fmt.Errorf("failed to write stats: %v", err)
	}

	// Write summary
	summaryPath := filepath.Join(c.OutputDir, "summary.txt")
	if err := os.WriteFile(summaryPath, []byte(stats.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write summary: %v", err)
	}

	// Write path statistics separately for easier processing
	pathStatsDir := filepath.Join(c.OutputDir, "paths")
	if err := os.MkdirAll(pathStatsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create paths directory: %v", err)
	}

	for path, pathStats := range stats.PathStats {
		// Create a safe filename from the path
		safeFileName := makePathSafe(path) + ".json"
		pathFile := filepath.Join(pathStatsDir, safeFileName)

		pathData, err := json.MarshalIndent(pathStats, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal path stats: %v", err)
		}
		if err := os.WriteFile(pathFile, pathData, 0o644); err != nil {
			return fmt.Errorf("failed to write path stats: %v", err)
		}
	}

	return nil
}

// makePathSafe converts a path to a safe filename
func makePathSafe(path string) string {
	return filepath.Base(path)
}
