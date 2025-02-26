// github.com/delving/hub3/tools/cmd/source_analyzer/models/dataset_context.go
package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DatasetContext holds configuration and state for dataset processing
type DatasetContext struct {
	BaseDir        string // Base directory for all output
	TreeRoot       string // Root directory for the tree structure
	CompressOutput bool   // Whether to compress output files
}

// DropTree removes all existing tree data
func (ctx *DatasetContext) DropTree() error {
	if err := os.RemoveAll(ctx.TreeRoot); err != nil {
		return err
	}
	return os.MkdirAll(ctx.TreeRoot, 0o755)
}

// CreateNodeRepo creates a new NodeRepo for the root
func (ctx *DatasetContext) CreateRootRepo() *NodeRepo {
	return NewNodeRepo(ctx, ctx.TreeRoot)
}

// GetStatus reads the status file for a node
func (ctx *DatasetContext) GetStatus(path string) (map[string]interface{}, error) {
	statusFile := filepath.Join(ctx.TreeRoot, path, "status.json")

	data, err := os.ReadFile(statusFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	var status map[string]interface{}
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return status, nil
}

// UpdateStatus updates the status file for a node
func (ctx *DatasetContext) UpdateStatus(path string, status map[string]interface{}) error {
	statusFile := filepath.Join(ctx.TreeRoot, path, "status.json")

	// Read existing status
	currentStatus, err := ctx.GetStatus(path)
	if err != nil {
		return err
	}

	// Merge new status
	for k, v := range status {
		currentStatus[k] = v
	}

	// Write updated status
	data, err := json.MarshalIndent(currentStatus, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statusFile, data, 0o644)
}

// ListNodes returns a list of all node paths in the tree
func (ctx *DatasetContext) ListNodes() ([]string, error) {
	var nodes []string
	err := filepath.Walk(ctx.TreeRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != ctx.TreeRoot {
			relPath, err := filepath.Rel(ctx.TreeRoot, path)
			if err != nil {
				return err
			}
			nodes = append(nodes, relPath)
		}
		return nil
	})
	return nodes, err
}
