// github.com/delving/hub3/tools/cmd/source_analyzer/processor/xml_processor.go
package processor

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	compress "github.com/delving/hub3/tools/cmd/source_analyzer/io"
	"github.com/delving/hub3/tools/cmd/source_analyzer/models"
)

// XMLProcessor handles the parsing of XML documents and tree construction
type XMLProcessor struct {
	decoder   *xml.Decoder
	closer    io.Closer
	tracker   *ProgressTracker
	startTime time.Time
}

// ProgressTracker tracks processing progress
type ProgressTracker struct {
	processed  int64
	lastReport time.Time
	phase      string
	mu         sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		lastReport: time.Now(),
		phase:      "XML Parsing",
	}
}

// SetPhase changes the current processing phase
func (p *ProgressTracker) SetPhase(phase string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase = phase
	p.processed = 0
	p.lastReport = time.Now()
}

// Increment increases the processed count and returns true if it's time to report
func (p *ProgressTracker) Increment() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.processed++
	if time.Since(p.lastReport) >= time.Second {
		p.lastReport = time.Now()
		fmt.Printf("[%s] Processed %d elements\n", p.phase, p.processed)
		return true
	}
	return false
}

// GetProcessed returns the current processed count
func (p *ProgressTracker) GetProcessed() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.processed
}

func ProcessXML(filename string, ctx *models.DatasetContext) (*models.TreeNode, error) {
	// Create root node and progress tracker
	rootRepo := ctx.CreateRootRepo()
	root := models.NewTreeNode(rootRepo, nil, "", "")
	tracker := NewProgressTracker()

	// Open and process file
	reader, closer, err := compress.CreateBufferedReader(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer closer.Close()

	processor := &XMLProcessor{
		decoder:   xml.NewDecoder(reader),
		closer:    closer,
		tracker:   tracker,
		startTime: time.Now(),
	}
	defer processor.Close()

	slog.Info("Starting XML analysis", "fileName", filename)

	// Track current node and process XML
	currentNode := root
	for {
		token, err := processor.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML parsing error: %v", err)
		}

		processor.tracker.Increment()

		switch t := token.(type) {
		case xml.StartElement:
			// Create new node for element
			uri := processor.getNamespaceURI(t.Name)
			node := currentNode.GetOrCreateKid(t.Name.Local, uri)
			node.Start()

			// Process attributes
			for _, attr := range t.Attr {
				attrURI := processor.getNamespaceURI(attr.Name)
				attrTag := fmt.Sprintf("@%s", attr.Name.Local)
				kid := node.GetOrCreateKid(attrTag, attrURI)
				kid.Start()
				kid.Value(attr.Value)
				if err := kid.End(); err != nil {
					return nil, fmt.Errorf("error processing attribute: %v", err)
				}
			}

			currentNode = node

		case xml.EndElement:
			if err := currentNode.End(); err != nil {
				return nil, fmt.Errorf("error ending element: %v", err)
			}
			currentNode = currentNode.Parent

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				currentNode.Value(text)
			}
		}
	}

	// Finish all nodes before processing statistics
	slog.Info("Finishing all nodes...")
	if err := root.Finish(); err != nil {
		return nil, fmt.Errorf("error finishing nodes: %v", err)
	}

	// Find actual root node
	var actualRoot *models.TreeNode
	for _, node := range root.Kids {
		actualRoot = node
		break
	}

	if actualRoot == nil {
		return nil, fmt.Errorf("no root element found in XML")
	}

	// Process nodes and generate statistics
	slog.Info("Generating statistics...")
	processor.tracker.SetPhase("Statistics")

	if err := processor.processNodes(actualRoot); err != nil {
		return nil, fmt.Errorf("error processing nodes: %v", err)
	}

	// Write index.json
	if err := actualRoot.WriteIndexJson(); err != nil {
		return nil, fmt.Errorf("failed to write index.json: %v", err)
	}

	// Print statistics
	elapsed := time.Since(processor.startTime)
	slog.Info("Processing complete",
		"totalElements", processor.tracker.GetProcessed(),
		"timeElapsed", elapsed.Round(time.Second),
		"outputDir", actualRoot.NodeRepo.DatasetCtx.BaseDir)

	return actualRoot, nil
}

// processNodes handles the sorting and collating of node data
func (p *XMLProcessor) processNodes(node *models.TreeNode) error {
	// Process current node if it has values
	if !node.Lengths.IsEmpty() {
		if err := p.processNodeData(node); err != nil {
			return err
		}
	}

	// Process children
	for _, child := range node.Kids {
		if err := p.processNodes(child); err != nil {
			return err
		}
	}

	return nil
}

// processNodeData handles sorting and collating for a single node
func (p *XMLProcessor) processNodeData(node *models.TreeNode) error {
	nodePath := node.GetPath()

	// Sort values
	slog.Debug("\nProcessing node", "nodePath", nodePath)
	p.tracker.SetPhase(fmt.Sprintf("Sorting %s", nodePath))
	sorter := NewSorter(node.NodeRepo)
	if err := sorter.Sort(ValueSort); err != nil {
		return fmt.Errorf("error sorting values for %s: %v", nodePath, err)
	}

	// Collate and generate statistics
	p.tracker.SetPhase(fmt.Sprintf("Collating %s", nodePath))
	collator := NewCollator(node.NodeRepo)
	result, err := collator.Process()
	if err != nil {
		return fmt.Errorf("error collating values for %s: %v", nodePath, err)
	}

	slog.Debug("Processed Node %s: found %d unique values\n", "node", nodePath, "uniqueValues", result.UniqueCount)

	// Generate histograms
	p.tracker.SetPhase(fmt.Sprintf("Generating histograms for %s", nodePath))
	if err := sorter.Sort(HistogramSort); err != nil {
		return fmt.Errorf("error generating histograms for %s: %v", nodePath, err)
	}

	return node.NodeRepo.UpdateStatus(map[string]interface{}{
		"uniqueCount": result.UniqueCount,
		"sampleSizes": result.SampleSizes,
		"processed":   true,
	})
}

// Close cleanup resources
func (p *XMLProcessor) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}

// getNamespaceURI returns the full namespace URI for an XML name
func (p *XMLProcessor) getNamespaceURI(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	return fmt.Sprintf("{%s}%s", name.Space, name.Local)
}

// handleComment processes XML comments
func (p *XMLProcessor) handleComment(node *models.TreeNode, comment xml.Comment) {
	// Convert comment to string and clean it up
	commentText := string(comment)
	commentText = cleanupComment(commentText)

	// If comment contains meaningful content, add it to the node
	if commentText != "" {
		node.Value(commentText)
	}
}

// cleanupComment removes unnecessary whitespace and normalizes comment content
func cleanupComment(comment string) string {
	// Remove comment markers and trim
	comment = strings.TrimPrefix(comment, "<!--")
	comment = strings.TrimSuffix(comment, "-->")
	comment = strings.TrimSpace(comment)

	// Normalize whitespace
	return strings.Join(strings.Fields(comment), " ")
}
