// github.com/delving/hub3/tools/cmd/source_analyzer/models/stats_tree.go
package models

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// )
//
// // TreeStats represents the JSON output structure for the index.json
// type TreeStats struct {
// 	Tag     string                `json:"tag"`
// 	Path    string                `json:"path"`
// 	Count   int                   `json:"count,omitempty"`
// 	Lengths *LengthHistogram      `json:"lengths,omitempty"`
// 	Kids    map[string]*TreeStats `json:"kids,omitempty"`
// }
//
// // NewTreeStats creates a TreeStats from a TreeNode
// func NewTreeStats(node *TreeNode) *TreeStats {
// 	stats := &TreeStats{
// 		Tag:     node.Tag,
// 		Path:    node.Path,
// 		Count:   node.Count,
// 		Lengths: node.Lengths,
// 		Kids:    make(map[string]*TreeStats),
// 	}
//
// 	for tag, kid := range node.Kids {
// 		stats.Kids[tag] = NewTreeStats(kid)
// 	}
//
// 	return stats
// }
//
// // WriteIndexJson writes the tree structure to index.json
// func (n *TreeNode) WriteIndexJson() error {
// 	// Get the path for index.json
// 	indexPath := filepath.Join(n.NodeRepo.DatasetCtx.BaseDir, "index.json")
//
// 	// Create the stats tree
// 	stats := NewTreeStats(n)
//
// 	// Marshal to JSON with indentation
// 	data, err := json.MarshalIndent(stats, "", "  ")
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal index json: %v", err)
// 	}
//
// 	// Write to index.json file
// 	if err := os.WriteFile(indexPath, data, 0o644); err != nil {
// 		return fmt.Errorf("failed to write index.json: %v", err)
// 	}
//
// 	return nil
// }
