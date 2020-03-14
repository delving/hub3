package ead

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
	"github.com/olivere/elastic/v7"
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
	Counter          *NodeCounter
	MetsCounter      *MetsCounter
	OrgID            string
	Spec             string
	Title            []string
	TitleShort       string
	Revision         int32
	PeriodDesc       []string
	labels           map[string]string
	MimeTypes        map[string][]string
	Errors           []*DuplicateError
	Client           *http.Client
	BulkProcessor    BulkProcessor
	CreateTree       func(cfg *NodeConfig, n *Node, hubID string, id string) *fragments.Tree
	ContentIdentical bool
}

// BulkProcessor is an interface for oliver/elastice BulkProcessor.
type BulkProcessor interface {
	Add(request elastic.BulkableRequest)
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
	Error    string
}

func (nc *NodeConfig) ErrorToCSV() ([]byte, error) {
	var b bytes.Buffer
	s := func(input string) string {
		re := regexp.MustCompile(`\r?\n`)
		return re.ReplaceAllString(input, " ")
	}

	b.WriteString("nr,spec,order,path,key,label,dupKey,dupLabel,ctype,depth,error\n")
	for idx, de := range nc.Errors {
		b.WriteString(
			fmt.Sprintf(
				"%d,%s,%d,%s,%s,\"%s\",%s,\"%s\",%s,%d,%#v\n",
				idx, strings.TrimSpace(de.Spec), de.Order, de.Path, de.Key, s(de.Label),
				de.DupKey, s(de.DupLabel), de.CType, de.Depth, de.Error,
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
		MetsCounter: &MetsCounter{
			uniqueCounter: map[string]int{},
		},
		Client: &http.Client{Timeout: 10 * time.Second},
		labels: make(map[string]string),
	}
}

// MetsCounter is a concurrency safe counter for number of Mets-files processed
type MetsCounter struct {
	counter        uint64
	digitalObjects uint64
	errors         uint64
	inError        []string
	uniqueCounter  map[string]int
}

// Increment increments the count by one
func (mc *MetsCounter) Increment(daoLink string) {
	atomic.AddUint64(&mc.counter, 1)
	mc.uniqueCounter[daoLink]++
}

// GetUniqueCounter returns the map of unique METS links.
func (mc *MetsCounter) GetUniqueCounter() map[string]int {
	return mc.uniqueCounter
}

// IncrementDigitalObject increments the count by one
func (mc *MetsCounter) IncrementDigitalObject(delta uint64) {
	atomic.AddUint64(&mc.digitalObjects, delta)
}

// GetCount returns the snapshot of the current count
func (mc *MetsCounter) GetCount() uint64 {
	return atomic.LoadUint64(&mc.counter)
}

// GetDigitalObjectCount returns the snapshot of the current count
func (mc *MetsCounter) GetDigitalObjectCount() uint64 {
	return atomic.LoadUint64(&mc.digitalObjects)
}

// IncrementError increments the error count by one
func (mc *MetsCounter) IncrementError() {
	atomic.AddUint64(&mc.errors, 1)
}

// GetErrorCount returns the snapshot of the current error count
func (mc *MetsCounter) GetErrorCount() uint64 {
	return atomic.LoadUint64(&mc.errors)
}

func (mc *MetsCounter) AppendError(err string) {
	mc.IncrementError()
	mc.inError = append(mc.inError, err)
}

func (mc *MetsCounter) GetErrors() []string {
	return mc.inError
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
	if dsc == nil {
		return nl, 0, nil
	}
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

// ESSave saves the list of Archive Nodes to ElasticSearch
func (nl *NodeList) ESSave(cfg *NodeConfig, p *elastic.BulkProcessor) error {
	for _, n := range nl.Nodes {
		err := n.ESSave(cfg, p)
		if err != nil {
			return err
		}
	}
	// todo store cfg.Counter.GetCount() in dataset
	return nil
}

// ESSave stores a Fragments and a FragmentGraph in ElasticSearch
func (n *Node) ESSave(cfg *NodeConfig, p *elastic.BulkProcessor) error {
	fg, rm, err := n.FragmentGraph(cfg)
	if err != nil {
		return err
	}
	r := elastic.NewBulkIndexRequest().
		Index(c.Config.ElasticSearch.GetIndexName()).
		RetryOnConflict(3).
		Id(fg.Meta.HubID).
		Doc(fg)
	p.Add(r)

	if c.Config.ElasticSearch.Fragments {
		err := fragments.IndexFragments(rm, fg, p)
		if err != nil {
			return err
		}
	}

	// recursion on itself for nested nodes on deeper levels
	for _, n := range n.Nodes {
		err := n.ESSave(cfg, p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sparse creates a sparse version of Header
func (h *Header) Sparse() {
	if h.DateAsLabel {
		h.DateAsLabel = false
		for _, date := range h.Date {
			h.Label = append(h.Label, date.Label)
		}
	}
	h.Date = nil
	h.ID = nil
	h.Physdesc = ""
}

// GetPeriods return a list of human readable periods from the EAD unitDate
func (h *Header) GetPeriods() []string {
	periods := []string{}
	for _, date := range h.Date {
		periods = append(periods, date.Label)
	}
	return periods
}

// GetTreeLabel returns the label that needs to be shown with the tree
func (h *Header) GetTreeLabel() string {
	if len(h.Label) == 0 {
		return ""
	}
	return html.UnescapeString(h.Label[0])
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
	var ids []*NodeID
	var inventoryID string
	for _, unitid := range cdid.Cunitid {
		id, err := unitid.NewNodeID()
		if err != nil {
			return nil, "", err
		}
		switch id.Type {
		case "ABS", "series_code", "blank", "analoog", "BD", "":
			inventoryID = id.ID
		}
		ids = append(ids, id)
	}
	return ids, inventoryID, nil
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

// ValidDateNormal returns if the range in Normal is valid.
func (nd *NodeDate) ValidDateNormal() error {
	if nd.Normal == "" {
		return nil
	}

	if strings.Contains(nd.Normal, "/") {
		nd.Normal = strings.TrimPrefix(strings.TrimSuffix(nd.Normal, "/"), "/")
		parts := strings.Split(nd.Normal, "/")

		if len(parts) == 2 && parts[0] > parts[1] {
			return fmt.Errorf("first date %s is later than second date %s", parts[0], parts[1])
		}
	}

	return nil
}

// NewHeader creates an Archival Header
func (cdid *Cdid) NewHeader() (*Header, error) {
	header := &Header{
		Genreform: c.Config.EAD.GenreFormDefault,
	}

	if cdid.Cphysdesc != nil {
		header.Physdesc = sanitizeXMLAsString(cdid.Cphysdesc.Raw)
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
				dates = append(dates, nodeDate.Label)
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

	for _, unitID := range cdid.Cunitid {
		// Mark the header as Born Digital when we find a BD type unitid.
		if strings.ToLower(unitID.Attrtype) == "bd" {
			header.AltRender = "Born Digital"
		}
	}

	nodeIDs, inventoryID, err := cdid.NewNodeIDs()
	if err != nil {
		return nil, err
	}
	if inventoryID != "" {
		header.InventoryNumber = inventoryID
	}
	header.ID = append(header.ID, nodeIDs...)

	if cdid.Cphysloc != nil {
		header.Physloc = string(cdid.Cphysloc.Raw)
	}

	return header, nil
}

func (n *Node) getPathID() string {
	eadID := n.Header.InventoryNumber
	if eadID == "" || strings.HasPrefix(eadID, "---") {
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
	if header.DaoLink != "" {
		cfg.MetsCounter.Increment(header.DaoLink)
	}

	if c.GetAttraltrender() != "" {
		node.Header.AltRender = c.GetAttraltrender()
	}

	if c.GetGenreform() != "" {
		node.Header.Genreform = c.GetGenreform()
	}

	// add accessrestrict
	if ar := c.GetCaccessrestrict(); ar != nil {
		node.AccessRestrict = strings.TrimSpace(sanitizer.Sanitize(string(c.GetCaccessrestrict().Raw)))
		for _, p := range ar.Cp {
			if p.Cref != nil && p.Cref.Cdate != nil {
				node.AccessRestrictYear = p.Cref.Cdate.Attrnormal
			}
		}
	}

	if c.GetMaterial() != "" {
		node.Material = c.GetMaterial()
	}

	for _, p := range c.GetPhystech() {
		node.Phystech = append(node.Phystech, sanitizeXMLAsString(p.Raw))
	}

	parentIDs, err = node.setPath(parentIDs)
	if err != nil {
		return nil, err
	}

	// check valid date
	for _, d := range node.Header.Date {
		if err := d.ValidDateNormal(); err != nil {
			de := &DuplicateError{
				Path:     node.Path,
				Order:    int(node.Order),
				Spec:     cfg.Spec,
				Key:      header.InventoryNumber,
				Label:    header.GetTreeLabel(),
				DupLabel: d.Normal,
				CType:    node.Type,
				Depth:    node.Depth,
				Error:    err.Error(),
			}
			cfg.Errors = append(cfg.Errors, de)
		}
	}

	_, ok := cfg.labels[node.Path]
	if ok {
		node.Path = fmt.Sprintf("%s%d", node.Path, node.Order)
	}

	cfg.labels[node.Path] = header.GetTreeLabel()

	subject := r.NewResource(node.GetSubject(cfg))

	didTriples, err := c.GetCdid().Triples(subject)
	if err != nil {
		return nil, err
	}

	node.triples = append(node.triples, didTriples...)

	cLevelTriples, err := c.Triples(subject)
	if err != nil {
		return nil, err
	}

	node.triples = append(node.triples, cLevelTriples...)

	// add nested
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
