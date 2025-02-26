// github.com/delving/hub3/tools/cmd/source_analyzer/models/node_repo.go
package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	compression "github.com/delving/hub3/tools/cmd/source_analyzer/io"
)

// NodeRepo handles file storage and retrieval for a node
type NodeRepo struct {
	BaseDir    string
	DatasetCtx *DatasetContext
}

// NewNodeRepo creates a new NodeRepo and ensures the directory exists
func NewNodeRepo(ctx *DatasetContext, dir string) *NodeRepo {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create directory %s: %v", dir, err))
	}

	return &NodeRepo{
		BaseDir:    dir,
		DatasetCtx: ctx,
	}
}

// Child creates a new NodeRepo for a child node
func (n *NodeRepo) Child(tag string) *NodeRepo {
	childDir := filepath.Join(n.BaseDir, SanitizePath(tag))
	return NewNodeRepo(n.DatasetCtx, childDir)
}

func (n *NodeRepo) getFilePath(baseName string) string {
	path := filepath.Join(n.BaseDir, baseName)
	if n.DatasetCtx.CompressOutput && !strings.HasSuffix(baseName, ".json") {
		path += ".zst"
	}
	return path
}

// GetValuesPath returns the path for values file
func (n *NodeRepo) GetValuesPath() string {
	return n.getFilePath("values.txt")
}

// GetSortedPath returns the path for sorted values file
func (n *NodeRepo) GetSortedPath() string {
	return n.getFilePath("sorted.txt")
}

// GetCountedPath returns the path for counted values file
func (n *NodeRepo) GetCountedPath() string {
	return n.getFilePath("counted.txt")
}

// CreateValuesWriter creates a buffered writer for values
func (n *NodeRepo) CreateValuesWriter() (*bufio.Writer, io.Closer, error) {
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	writer, err := compression.CreateCompressedWriter(n.GetValuesPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create values writer: %v", err)
	}

	return bufio.NewWriter(writer), writer, nil
}

// CreateSortedWriter creates a writer for sorted values
func (n *NodeRepo) CreateSortedWriter() (*bufio.Writer, io.Closer, error) {
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	writer, err := compression.CreateCompressedWriter(n.GetSortedPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sorted writer: %v", err)
	}

	return bufio.NewWriter(writer), writer, nil
}

// CreateCountedWriter creates a writer for counted values
func (n *NodeRepo) CreateCountedWriter() (*bufio.Writer, io.Closer, error) {
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	writer, err := compression.CreateCompressedWriter(n.GetCountedPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create counted writer: %v", err)
	}

	return bufio.NewWriter(writer), writer, nil
}

// CreateTempFile creates a temporary file in the node's directory
func (n *NodeRepo) CreateTempFile(prefix string) (*os.File, error) {
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Add appropriate extension
	if n.DatasetCtx.CompressOutput {
		prefix = prefix + "*.zst"
	} else {
		prefix = prefix + "*"
	}

	return os.CreateTemp(n.BaseDir, prefix)
}

// CreateValuesReader creates a reader for values
func (n *NodeRepo) CreateValuesReader() (*bufio.Reader, io.Closer, error) {
	valuesPath := n.getFilePath("values.txt")
	return compression.CreateBufferedReader(valuesPath)
}

// CreateSortedReader creates a reader for sorted values
func (n *NodeRepo) CreateSortedReader() (*bufio.Reader, io.Closer, error) {
	sortedPath := n.getFilePath("sorted.txt")
	return compression.CreateBufferedReader(sortedPath)
}

// WriteURI writes the URI to file
func (n *NodeRepo) WriteURI(uri string) error {
	uriPath := filepath.Join(n.BaseDir, "uri.txt")
	return os.WriteFile(uriPath, []byte(uri), 0o644)
}

// UpdateStatus updates the node's status file
func (n *NodeRepo) UpdateStatus(status map[string]interface{}) error {
	statusPath := filepath.Join(n.BaseDir, "status.json")

	// Read existing status if it exists
	currentStatus := make(map[string]interface{})
	if data, err := os.ReadFile(statusPath); err == nil {
		if err := json.Unmarshal(data, &currentStatus); err != nil {
			return fmt.Errorf("failed to parse existing status: %v", err)
		}
	}

	// Merge new status
	for k, v := range status {
		currentStatus[k] = v
	}

	// Write updated status
	data, err := json.MarshalIndent(currentStatus, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %v", err)
	}

	return os.WriteFile(statusPath, data, 0o644)
}

// GetSampleSizes returns the available sample sizes
func (n *NodeRepo) GetSampleSizes() []int {
	return []int{100, 500, 2500}
}

// GetHistogramSizes returns the available histogram sizes
func (n *NodeRepo) GetHistogramSizes() []int {
	return []int{100, 500, 2500, 12500}
}

// ReadURI reads the URI from file
func (n *NodeRepo) ReadURI() (string, error) {
	uriPath := filepath.Join(n.BaseDir, "uri.txt")
	data, err := os.ReadFile(uriPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteHistogram writes a histogram of the specified size
func (n *NodeRepo) WriteHistogram(size int, entries []HistogramEntry) error {
	// Ensure directory exists
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	histogramPath := filepath.Join(n.BaseDir, fmt.Sprintf("histogram-%d.json", size))

	uri, err := n.ReadURI()
	if err != nil {
		uri = "" // Default to empty if not found
	}

	data := map[string]interface{}{
		"tag":         filepath.Base(n.BaseDir),
		"uri":         uri,
		"uniqueCount": len(entries),
		"entries":     entries,
		"complete":    true,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal histogram: %v", err)
	}

	if err := os.WriteFile(histogramPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write histogram: %v", err)
	}

	slog.Debug("wrote histogram file",
		"path", histogramPath,
		"entries", len(entries))

	return nil
}

// SanitizePath sanitizes a path component for file system use
func SanitizePath(path string) string {
	// Replace problematic characters with underscores
	sanitized := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' ||
			r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, path)

	// Ensure the path is not empty and doesn't start with a dot
	if sanitized == "" || strings.HasPrefix(sanitized, ".") {
		sanitized = "_" + sanitized
	}

	return sanitized
}

// CreateCountedReader creates a reader for counted values
func (n *NodeRepo) CreateCountedReader() (*bufio.Reader, io.Closer, error) {
	countedPath := n.getFilePath("counted.txt")
	return compression.CreateBufferedReader(countedPath)
}

// WriteSample writes a sample of values to a JSON file
func (n *NodeRepo) WriteSample(size int, values []string) error {
	samplePath := filepath.Join(n.BaseDir, fmt.Sprintf("sample-%d.json", size))

	data := map[string]interface{}{
		"sample": values,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sample data: %v", err)
	}

	return os.WriteFile(samplePath, jsonData, 0o644)
}

// GetHistogramPath returns the path for the full histogram file
func (n *NodeRepo) GetHistogramPath() string {
	histPath := filepath.Join(n.BaseDir, "histogram.txt")
	if n.DatasetCtx.CompressOutput {
		histPath += ".zst"
	}
	return histPath
}

// CreateHistogramWriter creates a writer for the full histogram
func (n *NodeRepo) CreateHistogramWriter() (*bufio.Writer, io.Closer, error) {
	if err := os.MkdirAll(n.BaseDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	histPath := n.GetHistogramPath()
	writer, err := compression.CreateCompressedWriter(histPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create histogram writer: %v", err)
	}

	return bufio.NewWriter(writer), writer, nil
}
