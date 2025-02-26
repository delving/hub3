// github.com/delving/hub3/tools/cmd/source_analyzer/sip/example_test.go
package sip

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matryer/is"
)

// testProgressListener implements ProgressListener for tests
type testProgressListener struct {
	t      *testing.T
	counts []int
}

func (l *testProgressListener) SetProgress(count int) {
	l.counts = append(l.counts, count)
	l.t.Logf("Processed %d elements", count)
}

func (l *testProgressListener) SetProgressMessage(message string) {
	l.t.Logf("Status: %s", message)
}

// testAnalysisListener implements AnalysisListener for tests
type testAnalysisListener struct {
	t          *testing.T
	startTime  time.Time
	stats      *Stats
	errMessage string
	err        error
}

func (l *testAnalysisListener) Success(stats *Stats) {
	l.stats = stats
	elapsed := time.Since(l.startTime)
	l.t.Logf("Analysis complete in %v", elapsed)
	l.t.Logf("Stats: %s", stats.String())
}

func (l *testAnalysisListener) Failure(message string, err error) {
	l.errMessage = message
	l.err = err
	l.t.Errorf("Analysis failed: %s - %v", message, err)
}

func TestAnalyzer(t *testing.T) {
	is := is.New(t)

	// Create temp directory for output
	outputDir, err := os.MkdirTemp("", "sip-test-*")
	is.NoErr(err)
	defer os.RemoveAll(outputDir)

	// Setup test file
	testFile := filepath.Join("testdata", "test.xml")

	// Create configuration
	config := &Config{
		InputFile:            testFile,
		OutputDir:            outputDir,
		DatasetName:          "test_dataset",
		MaxUniqueValueLength: 1000,
		ElementStep:          100, // Smaller step for testing
	}

	// Create listeners
	progress := &testProgressListener{t: t}
	analysis := &testAnalysisListener{
		t:         t,
		startTime: time.Now(),
	}

	// Create and run analyzer
	analyzer := NewAnalyzer(config, analysis)
	analyzer.SetProgressListener(progress)

	err = analyzer.Process()
	is.NoErr(err)

	// Verify analysis completed successfully
	is.True(analysis.stats != nil)    // Should have stats
	is.True(analysis.err == nil)      // Should have no error
	is.True(len(progress.counts) > 0) // Should have progress updates

	// Check output files exist
	files := []string{
		filepath.Join(outputDir, "stats.json"),
		filepath.Join(outputDir, "summary.txt"),
		filepath.Join(outputDir, "paths"),
	}
	for _, file := range files {
		_, err := os.Stat(file)
		is.NoErr(err) // Output files should exist
	}
}

func ExampleAnalyzer() {
	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "sip-example-*")
	if err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}
	defer os.RemoveAll(outputDir)

	// Create basic configuration
	config := NewDefaultConfig("testdata/test.xml", outputDir)

	// Create simple listeners
	progress := &DefaultProgressListener{}
	analysis := &DefaultAnalysisListener{}

	// Create and run analyzer
	analyzer := NewAnalyzer(config, analysis)
	analyzer.SetProgressListener(progress)

	if err := analyzer.Process(); err != nil {
		fmt.Printf("Analysis failed: %v\n", err)
		return
	}
}

func TestAnalyzerWithCompression(t *testing.T) {
	is := is.New(t)

	// Create temp directory for output
	outputDir, err := os.MkdirTemp("", "sip-test-compressed-*")
	is.NoErr(err)
	defer os.RemoveAll(outputDir)

	// Setup test file
	testFile := filepath.Join("testdata", "test.xml")

	// Create configuration with compression
	config := &Config{
		InputFile:            testFile,
		OutputDir:            outputDir,
		DatasetName:          "test_dataset",
		MaxUniqueValueLength: 1000,
		ElementStep:          100,
		CompressOutput:       true,
	}

	// Create listeners
	progress := &testProgressListener{t: t}
	analysis := &testAnalysisListener{
		t:         t,
		startTime: time.Now(),
	}

	// Create and run analyzer
	analyzer := NewAnalyzer(config, analysis)
	analyzer.SetProgressListener(progress)

	err = analyzer.Process()
	is.NoErr(err)

	// Verify analysis completed
	is.True(analysis.stats != nil)
	is.True(analysis.err == nil)

	// Check for compressed output files
	pathsDir := filepath.Join(outputDir, "paths")
	entries, err := os.ReadDir(pathsDir)
	is.NoErr(err)
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if filepath.Ext(name) != ".zst" {
				t.Errorf("Expected compressed file with .zst extension, got: %s", name)
			}
		}
	}
}
