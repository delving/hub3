package ead

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/olivere/elastic"
)

const pathSep string = "~"

func init() {
	path := c.Config.EAD.CacheDir
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatalf("Unable to create cache dir; %s", err)
		}
	}
}

// Manifest holds all the information for an archive to create a IIIF manifest.
type Manifest struct {
	InventoryID string `json:"inventoryID"`
	ArchiveName string `json:"archiveName"`
	UnitID      string `json:"unitID"`
	UnitTitle   string `json:"unitTitle"`
}

// NodeConfig holds all the configuration options fo generating Archive Nodes
type NodeConfig struct {
	Counter  *NodeCounter
	OrgID    string
	Spec     string
	Revision int32
	labels   map[string]string
	Errors   []*DuplicateError
}

type DuplicateError struct {
	Path     string `json:"path"`
	Spec     string `json:"spec"`
	Order    int    `json:"order"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	DupKey   string `json:"dupKey"`
	DupLabel string `json:"dupLabel"`
	CType    string `json:"cType"`
	Depth    int32  `json:"depth"`
}

func (nc *NodeConfig) ErrorToCSV() ([]byte, error) {
	var b bytes.Buffer
	s := func(input string) string {
		re := regexp.MustCompile(`\r?\n`)
		return re.ReplaceAllString(input, " ")
	}

	b.WriteString("nr,spec,order,path,key,label,dupKey,dupLabel,ctype,depth\n")
	for idx, de := range nc.Errors {
		b.WriteString(
			fmt.Sprintf(
				"%d,%s,%d,%s,%s,\"%s\",%s,\"%s\",%s,%d\n",
				idx, strings.TrimSpace(de.Spec), de.Order, de.Path, de.Key, s(de.Label),
				de.DupKey, s(de.DupLabel), de.CType, de.Depth,
			),
		)
	}

	return b.Bytes(), nil
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


// NewNodeList converts the Archival Description Level to a Nodelist
// Nodelist is an optimized lossless Protocol Buffer container.
func (dsc *Cdsc) NewNodeList(cfg *NodeConfig) (*NodeList, uint64, error) {
	nl := &NodeList{}
	nl.Type = dsc.Attrtype
	for _, label := range dsc.Chead {
		nl.Label = append(nl.Label, label.Head)
	}

	for _, nn := range dsc.Nested {
		node, err := NewNode(nn, []string{}, cfg)
		if err != nil {
			return nil, 0, err
		}
		nl.Nodes = append(nl.Nodes, node)
	}

	for _, nn := range dsc.Cc {
		node, err := NewNode(nn, []string{}, cfg)
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

	// recursion on itself for nested nodes on deeper levels
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

// GetPeriods return a list of human readable periods from the EAD unitDate
func (h *Header) GetPeriods() []string {
	periods := []string{}
	for _, date := range h.GetDate() {
		periods = append(periods, date.GetLabel())
	}
	return periods
}

// GetTreeLabel returns the label that needs to be shown with the tree
func (h *Header) GetTreeLabel() string {
	if len(h.Label) == 0 {
		return ""
	}
	return html.UnescapeString(fmt.Sprintf("%s", h.Label[0]))
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

	if cdid.Cdao != nil {
		header.HasDigitalObject = true
		header.DaoLink = cdid.Cdao.Attrhref
	}

	for _, label := range cdid.Cunittitle {
		// todo interpolation of date and title is not correct at the moment.
		dates := []string{}
		if len(label.Cunitdate) != 0 {
			header.DateAsLabel = true
			for _, date := range label.Cunitdate {
				nodeDate, err := date.NewNodeDate()
				if err != nil {
					return nil, err
				}
				header.Date = append(header.Date, nodeDate)
				dates = append(dates, nodeDate.GetLabel())
			}
		}

		header.Label = append(header.Label, label.Title())
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

func (n *Node) getPathID() string {
	eadID := n.GetHeader().GetInventoryNumber()
	if eadID == "" {
		eadID = strconv.FormatUint(n.Order, 10)
	}
	return fmt.Sprintf("%s", eadID)
}

func (n *Node) setPath(parentIDs []string) ([]string, error) {
	if len(parentIDs) > 0 {
		n.BranchID = parentIDs[len(parentIDs)-1]
		n.Path = fmt.Sprintf("%s%s%s", n.BranchID, pathSep, n.getPathID())
	} else {
		n.Path = n.getPathID()
	}
	ids := append(parentIDs, n.Path)
	return ids, nil
}

// NewNode converts EAD c01 to a Archival Node
func NewNode(c CLevel, parentIDs []string, cfg *NodeConfig) (*Node, error) {
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
	prevLabel, ok := cfg.labels[node.Path]
	if ok {
		//data, err := json.MarshalIndent(node, " ", " ")
		//if err != nil {
		//return nil, errors.Wrap(err, "Unable to marshal node during uniqueness check")
		//}
		de := &DuplicateError{
			Path:     node.GetPath(),
			Order:    int(node.GetOrder()),
			Spec:     cfg.Spec,
			Key:      header.GetInventoryNumber(),
			Label:    prevLabel,
			DupKey:   header.GetInventoryNumber(),
			DupLabel: header.GetTreeLabel(),
			CType:    node.GetType(),
			Depth:    node.GetDepth(),
		}
		cfg.Errors = append(cfg.Errors, de)

		//return nil, fmt.Errorf("Found duplicate unique key for %s with previous label %s: \n %s", header.GetInventoryNumber(), prevLabel, data)
		node.Path = fmt.Sprintf("%s%d", node.Path, node.Order)
	}

	cfg.labels[node.Path] = header.GetTreeLabel()

	nested := c.GetNested()
	if len(nested) != 0 {
		for _, nn := range nested {
			n, err := NewNode(nn, parentIDs, cfg)
			if err != nil {
				return nil, err
			}
			node.Nodes = append(node.Nodes, n)
		}
	}
	return node, nil
}
