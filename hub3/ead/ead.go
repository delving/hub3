package ead

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/fragments"
	elastic "gopkg.in/olivere/elastic.v5"
)

func init() {
	path := c.Config.EAD.CacheDir
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatalf("Unable to create cache dir; %s", err)
		}
	}

}

// NodeConfig holds all the configuration options fo generating Archive Nodes
type NodeConfig struct {
	Counter  *NodeCounter
	OrgID    string
	Spec     string
	Revision int32
	labels   map[string]string
}

// AddLabel adds a cLevel id and its label to the label map
// This map is used to resolve the label for each clevel for rendering the tree
func (nc *NodeConfig) AddLabel(id, label string) {
	nc.labels[id] = label
}

// NewNodeConfig creates a new NodeConfig
func NewNodeConfig(ctx context.Context) *NodeConfig {
	return &NodeConfig{
		Counter: &NodeCounter{},
		labels:  make(map[string]string),
	}
}

// NodeCounter is a concurrency safe counter for number of Nodes processed
type NodeCounter struct {
	counter uint64
}

// Increment increments the count by one
func (nc *NodeCounter) Increment() {
	atomic.AddUint64(&nc.counter, 1)
}

// GetCount returns the snapshot of the current count
func (nc *NodeCounter) GetCount() uint64 {
	return atomic.LoadUint64(&nc.counter)
}

// NewNodeList converts the Archival Description Level to a Nodelist
// Nodelist is an optimized lossless Protocol Buffer container.
func (dsc *Cdsc) NewNodeList(cfg *NodeConfig) (*NodeList, uint64, error) {
	nl := &NodeList{}
	nl.Type = dsc.Attrtype
	for _, label := range dsc.Chead {
		nl.Label = append(nl.Label, label.Head)
	}

	for idx, nn := range dsc.Nested {
		node, err := NewNode(nn, []string{}, idx, cfg)
		if err != nil {
			return nil, 0, err
		}
		nl.Nodes = append(nl.Nodes, node)
	}
	return nl, cfg.Counter.GetCount(), nil
}

// Sparse creates a sparse version of the list of Archive Nodes
func (nl *NodeList) Sparse() {
	Sparsify(nl.Nodes)
}

// ESSave saves the list of Archive Nodes to ElasticSearch
func (nl *NodeList) ESSave(cfg *NodeConfig, p *elastic.BulkProcessor) error {
	for _, n := range nl.GetNodes() {
		err := n.ESSave(cfg, p)
		if err != nil {
			return err
		}
	}
	log.Printf("Unique labels %d; cLevel counter %d", len(cfg.labels), cfg.Counter.GetCount())
	return nil
}

// ESSave stores a Fragments and a FragmentGraph in ElasticSearch
func (n *Node) ESSave(cfg *NodeConfig, p *elastic.BulkProcessor) error {
	fg, rm, err := n.FragmentGraph(cfg)
	if err != nil {
		return err
	}
	//log.Printf("node: %#v", n)
	//log.Printf("hubID: %s", fg.Meta.HubID)
	r := elastic.NewBulkIndexRequest().
		Index(c.Config.ElasticSearch.IndexName).
		Type(fragments.DocType).
		RetryOnConflict(3).
		Id(fg.Meta.HubID).
		Doc(fg)
	p.Add(r)
	err = fragments.IndexFragments(rm, fg, p)
	if err != nil {
		return err
	}

	for _, n := range n.GetNodes() {
		err := n.ESSave(cfg, p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sparse creates a sparse version of Header
func (h *Header) Sparse() {
	if h.GetDateAsLabel() {
		h.DateAsLabel = false
		for _, date := range h.GetDate() {
			h.Label = append(h.Label, date.GetLabel())
		}
	}
	h.Date = nil
	h.ID = nil
	h.Physdesc = ""
}

// GetTreeLabel returns the label that needs to be shown with the tree
func (h *Header) GetTreeLabel() string {
	if len(h.Label) == 0 {
		return ""
	}
	return fmt.Sprintf("%s %s", h.GetInventoryNumber(), h.Label[0])
}

// Sparsify is a recursive function that creates a Sparse representation
// of a list of Nodes. This is mostly used to efficiently create Tree Views
// of the Archive C-Levels
func Sparsify(nodes []*Node) {
	for _, n := range nodes {
		n.HTML = ""
		n.CTag = ""
		n.Header.Sparse()
		if len(n.Nodes) != 0 {
			Sparsify(n.Nodes)
		}

	}
}

// NewNodeID converts a unitid field from the EAD did to a NodeID
func (ui *Cunitid) NewNodeID() (*NodeID, error) {
	id := &NodeID{
		ID:       ui.ID,
		TypeID:   ui.Attridentifier,
		Type:     ui.Attrtype,
		Audience: ui.Attraudience,
	}
	return id, nil
}

// NewNodeIDs extract Unit Identifiers from the EAD did
func (cdid *Cdid) NewNodeIDs() ([]*NodeID, string, error) {
	ids := []*NodeID{}
	var invertoryNumber string
	for _, unitid := range cdid.Cunitid {
		id, err := unitid.NewNodeID()
		if err != nil {
			return nil, "", err
		}
		switch id.GetType() {
		case "ABS", "series_code", "":
			invertoryNumber = id.GetID()
		}
		ids = append(ids, id)
	}
	return ids, invertoryNumber, nil
}

// NewNodeDate extract date infomation frme the EAD unitdate
func (date *Cunitdate) NewNodeDate() (*NodeDate, error) {
	nDate := &NodeDate{
		Calendar: date.Attrcalendar,
		Era:      date.Attrera,
		Normal:   date.Attrnormal,
		Label:    date.Date,
		Type:     date.Attrtype,
	}
	return nDate, nil
}

// NewHeader creates an Archival Header
func (cdid *Cdid) NewHeader() (*Header, error) {
	header := &Header{}
	if cdid.Cphysdesc != nil {
		header.Physdesc = cdid.Cphysdesc.PhyscDesc
	}

	for _, label := range cdid.Cunittitle {
		if len(label.Cunitdate) != 0 {
			header.DateAsLabel = true
			for _, date := range label.Cunitdate {
				nodeDate, err := date.NewNodeDate()
				if err != nil {
					return nil, err
				}
				header.Date = append(header.Date, nodeDate)
			}
			continue
		}
		header.Label = append(header.Label, label.Title)
	}

	for _, date := range cdid.Cunitdate {
		nodeDate, err := date.NewNodeDate()
		if err != nil {
			return nil, err
		}
		header.Date = append(header.Date, nodeDate)
	}

	nodeIDs, inventoryID, err := cdid.NewNodeIDs()
	if err != nil {
		return nil, err
	}
	if inventoryID != "" {
		header.InventoryNumber = inventoryID
	}
	header.ID = append(header.ID, nodeIDs...)

	return header, nil
}

func (n *Node) setPath(parentIDs []string) ([]string, error) {
	if len(parentIDs) > 0 {
		n.BranchID = parentIDs[len(parentIDs)-1]
		n.Path = fmt.Sprintf("%s-%s", n.BranchID, n.GetHeader().GetInventoryNumber())
	} else {
		n.Path = n.GetHeader().GetInventoryNumber()
	}
	ids := append(parentIDs, n.Path)
	return ids, nil
}

// NewNode converts EAD c01 to a Archival Node
func NewNode(c CLevel, parentIDs []string, order int, cfg *NodeConfig) (*Node, error) {
	cfg.Counter.Increment()
	node := &Node{
		CTag:      c.GetXMLName().Local,
		Depth:     int32(len(parentIDs) + 1),
		Type:      c.GetAttrlevel(),
		SubType:   c.GetAttrotherlevel(),
		ParentIDs: parentIDs,
		Order:     cfg.Counter.GetCount(),
	}

	header, err := c.GetCdid().NewHeader()
	if err != nil {
		return nil, err
	}
	if header.GetInventoryNumber() == "" {
		header.InventoryNumber = fmt.Sprintf("%d", order)
	}
	node.Header = header

	// add scope content
	if c.GetScopeContent() != nil {
		html, err := xml.Marshal(c.GetScopeContent().Cp)
		if err != nil {
			return nil, err
		}
		node.HTML = string(html)
	}

	parentIDs, err = node.setPath(parentIDs)
	if err != nil {
		return nil, err
	}

	// add nested
	_, ok := cfg.labels[node.Path]
	if ok {
		return nil, fmt.Errorf("Found duplicate unique key for %s", header.GetInventoryNumber())

	}
	cfg.labels[node.Path] = header.GetTreeLabel()

	nested := c.GetNested()
	if len(nested) != 0 {
		for idx, nn := range nested {
			n, err := NewNode(nn, parentIDs, idx, cfg)
			if err != nil {
				return nil, err
			}
			node.Nodes = append(node.Nodes, n)
		}
	}
	return node, nil
}
