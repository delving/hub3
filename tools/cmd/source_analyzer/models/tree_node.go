// github.com/delving/hub3/tools/cmd/source_analyzer/models/tree_node.go
package models

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxListSize = 10000

// TreeStats represents the JSON output structure for index.json
type TreeStats struct {
	Tag     string                `json:"tag"`
	Path    string                `json:"path"`
	Count   int                   `json:"count"`
	Lengths *LengthHistogram      `json:"lengths"`
	Kids    map[string]*TreeStats `json:"kids"`
}

// TreeNode represents a node in the XML tree structure
type TreeNode struct {
	Tag      string               `json:"tag"`
	Path     string               `json:"path"`
	Count    int                  `json:"count"`
	Lengths  *LengthHistogram     `json:"lengths"`
	Kids     map[string]*TreeNode `json:"kids"`
	Parent   *TreeNode            `json:"-"`
	NodeRepo *NodeRepo            `json:"-"`
	URI      string               `json:"-"`

	// Internal state
	valueBuilder strings.Builder
	valueList    []string
	valueCount   int
	mu           sync.Mutex
}

// NewTreeNode creates a new tree node
func NewTreeNode(repo *NodeRepo, parent *TreeNode, tag, uri string) *TreeNode {
	node := &TreeNode{
		Tag:      tag,
		Parent:   parent,
		NodeRepo: repo,
		URI:      uri,
		Kids:     make(map[string]*TreeNode),
		Lengths:  NewLengthHistogram(),
	}
	node.Path = node.GetPath()
	return node
}

// GetOrCreateKid gets or creates a child node
func (n *TreeNode) GetOrCreateKid(tag, uri string) *TreeNode {
	n.mu.Lock()
	defer n.mu.Unlock()

	kid, exists := n.Kids[tag]
	if !exists {
		kid = NewTreeNode(n.NodeRepo.Child(tag), n, tag, uri)
		n.Kids[tag] = kid
	}
	return kid
}

// Start marks the beginning of a node's content
func (n *TreeNode) Start() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Count++
	n.valueBuilder.Reset()
}

// Value adds content to the node
func (n *TreeNode) Value(text string) {
	if text = strings.TrimSpace(text); text != "" {
		n.mu.Lock()
		defer n.mu.Unlock()

		n.valueBuilder.WriteString(text)
		n.valueCount++
	}
}

// End processes the accumulated value and flushes if needed
func (n *TreeNode) End() error {
	n.mu.Lock()
	value := strings.TrimSpace(n.valueBuilder.String())
	n.valueBuilder.Reset()
	n.mu.Unlock()

	if value != "" {
		n.Lengths.Record(len(value))

		n.mu.Lock()
		n.valueList = append(n.valueList, value)

		// Flush to disk if buffer is full
		if len(n.valueList) >= maxListSize {
			if err := n.flush(); err != nil {
				n.mu.Unlock()
				return err
			}
		}
		n.mu.Unlock()
	}
	return nil
}

// flush writes buffered values to disk
func (n *TreeNode) flush() error {
	if len(n.valueList) == 0 {
		return nil
	}

	writer, closer, err := n.NodeRepo.CreateValuesWriter()
	if err != nil {
		return err
	}
	defer closer.Close()

	for _, value := range n.valueList {
		if _, err := writer.WriteString(value + "\n"); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	// Clear the list after successful write
	n.valueList = n.valueList[:0]
	return nil
}

// Finish completes node processing
func (n *TreeNode) Finish() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Force flush any remaining values
	if len(n.valueList) > 0 {
		slog.Debug("Final flush for node", "node", n.Tag, "valueCount", len(n.valueList))
		if err := n.flush(); err != nil {
			return fmt.Errorf("failed to flush final values: %v", err)
		}
	}

	// Write URI
	if err := n.NodeRepo.WriteURI(n.URI); err != nil {
		return fmt.Errorf("failed to write URI: %v", err)
	}

	// Write node status
	status := map[string]interface{}{
		"count":      n.Count,
		"valueCount": n.valueCount,
	}
	if err := n.NodeRepo.UpdateStatus(status); err != nil {
		return fmt.Errorf("failed to update status: %v", err)
	}

	// Finish children recursively
	for _, kid := range n.Kids {
		if err := kid.Finish(); err != nil {
			return err
		}
	}

	return nil
}

// GetPath returns the full path of the node
func (n *TreeNode) GetPath() string {
	if n.Tag == "" {
		return ""
	}
	if n.Parent == nil {
		return "/" + n.Tag
	}
	return n.Parent.GetPath() + "/" + n.Tag
}

// WriteIndexJson writes the tree structure to index.json
func (n *TreeNode) WriteIndexJson() error {
	stats := &TreeStats{
		Tag:     n.Tag,
		Path:    n.Path,
		Count:   n.Count,
		Lengths: n.Lengths,
		Kids:    make(map[string]*TreeStats),
	}

	for tag, kid := range n.Kids {
		stats.Kids[tag] = NewTreeStats(kid)
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tree stats: %v", err)
	}

	indexPath := filepath.Join(n.NodeRepo.DatasetCtx.BaseDir, "index.json")
	if err := os.WriteFile(indexPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write index.json: %v", err)
	}

	return nil
}

// NewTreeStats creates TreeStats from a TreeNode
func NewTreeStats(node *TreeNode) *TreeStats {
	stats := &TreeStats{
		Tag:     node.Tag,
		Path:    node.Path,
		Count:   node.Count,
		Lengths: node.Lengths,
		Kids:    make(map[string]*TreeStats),
	}

	for tag, kid := range node.Kids {
		stats.Kids[tag] = NewTreeStats(kid)
	}

	return stats
}
