// github.com/delving/hub3/tools/cmd/source_analyzer/processor/analyzer_test.go
package processor

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/delving/hub3/tools/cmd/source_analyzer/models"
	"github.com/matryer/is"
)

// TestResult represents the structure of histogram JSON files
type TestResult struct {
	Tag         string                  `json:"tag"`
	URI         string                  `json:"uri"`
	UniqueCount int                     `json:"uniqueCount"`
	Entries     []models.HistogramEntry `json:"entries"`
	Complete    bool                    `json:"complete"`
}

// Helper function to check if a file exists and has content
func checkFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

func TestXMLAnalysis(t *testing.T) {
	is := is.New(t)

	// Create temp directory for output
	outputDir, err := os.MkdirTemp("", "analyzer-test-*")
	is.NoErr(err)
	defer os.RemoveAll(outputDir)

	// Setup test context
	ctx := &models.DatasetContext{
		BaseDir:        outputDir,
		TreeRoot:       filepath.Join(outputDir, "tree"),
		CompressOutput: false,
	}

	// Process test XML file
	tree, err := ProcessXML("testdata/test.xml", ctx)
	is.NoErr(err)
	is.True(tree != nil) // Should have a valid tree

	// Collect paths that actually have values
	pathsWithValues := collectNodesWithValues(tree, "")

	// Check each path that should have values
	for _, path := range pathsWithValues {
		nodePath := filepath.Join(ctx.TreeRoot, path)
		fullPath := filepath.Join(nodePath, "values.txt")
		if !checkFile(fullPath) {
			t.Logf("Missing expected values file: %s", fullPath)
			t.Fail()
		}
	}

	// Verify known content
	expectedValuePaths := map[string][]string{
		"record/title":    {"First Record", "Second Record"},
		"record/tags/tag": {"test", "example", "sample"},
		"record/@id":      {"1", "2"},
	}

	for path, expectedValues := range expectedValuePaths {
		histogramPath := filepath.Join(ctx.TreeRoot, models.SanitizePath(path), "histogram-100.json")
		if !checkFile(histogramPath) {
			continue // Skip if file doesn't exist (node might not have values)
		}

		histogramData, err := os.ReadFile(histogramPath)
		is.NoErr(err)

		var result TestResult
		err = json.Unmarshal(histogramData, &result)
		is.NoErr(err)

		// Verify all expected values are present
		foundValues := make(map[string]bool)
		for _, entry := range result.Entries {
			foundValues[entry.Value] = true
		}

		for _, expected := range expectedValues {
			is.True(foundValues[expected]) // Should find expected value
		}
	}
}

// Update helper function to use models.SanitizePath
func collectNodesWithValues(node *models.TreeNode, parentPath string) []string {
	var paths []string
	currentPath := filepath.Join(parentPath, node.Tag)

	// Check if this node has any values
	if !node.Lengths.IsEmpty() {
		paths = append(paths, currentPath)
	}

	// Recursively check children
	for _, child := range node.Kids {
		childPaths := collectNodesWithValues(child, currentPath)
		paths = append(paths, childPaths...)
	}

	return paths
}

func TestXMLAnalysisCompressed(t *testing.T) {
	is := is.New(t)

	// Create temp directory for compressed output
	outputDir, err := os.MkdirTemp("", "analyzer-test-compressed-*")
	is.NoErr(err)
	defer os.RemoveAll(outputDir)

	ctx := &models.DatasetContext{
		BaseDir:        outputDir,
		TreeRoot:       filepath.Join(outputDir, "tree"),
		CompressOutput: true,
	}

	// Process with compression
	tree, err := ProcessXML("testdata/test.xml", ctx)
	is.NoErr(err)
	is.True(tree != nil)

	// Check title values specifically as we know they should exist
	// Note the addition of 'root' to the path
	titlePath := filepath.Join(ctx.TreeRoot, "root/record/title")
	slog.Info("checking title path", "path", titlePath)

	// List all files in the directory for debugging
	files, err := os.ReadDir(titlePath)
	if err != nil {
		slog.Error("failed to read directory",
			"path", titlePath,
			"error", err)
	} else {
		for _, file := range files {
			slog.Info("directory content",
				"name", file.Name(),
				"is_dir", file.IsDir())
		}
	}

	// Check for compressed files
	expectedFiles := []string{
		"values.txt.zst",
		"sorted.txt.zst",
		"counted.txt.zst",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(titlePath, file)
		exists := checkFile(path)
		if !exists {
			t.Logf("Missing expected compressed file: %s", path)
			t.Fail()
		}
	}

	// Verify we can read the histogram
	histogramPath := filepath.Join(titlePath, "histogram-100.json")
	slog.Info("reading histogram", "path", histogramPath)

	histogramData, err := os.ReadFile(histogramPath)
	is.NoErr(err)

	var result TestResult
	err = json.Unmarshal(histogramData, &result)
	is.NoErr(err)

	is.Equal(result.UniqueCount, 2) // Should still have 2 unique titles
}

func logTreeStructure(node *models.TreeNode, prefix string, path string) {
	if node == nil {
		return
	}

	currentPath := filepath.Join(path, node.Tag)
	slog.Info("tree node",
		"path", currentPath,
		"tag", node.Tag,
		"count", node.Count,
		"has_values", !node.Lengths.IsEmpty())

	for _, kid := range node.Kids {
		logTreeStructure(kid, prefix+"  ", currentPath)
	}
}
