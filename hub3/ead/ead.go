// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate go run number_gen.go

package ead

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	r "github.com/kiivihal/rdf2go"
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/service/x/index"
)

const pathSep string = "~"

// Manifest holds all the information for an archive to create a IIIF manifest.
type Manifest struct {
	InventoryID string `json:"inventoryID"`
	ArchiveName string `json:"archiveName"`
	UnitID      string `json:"unitID"`
	UnitTitle   string `json:"unitTitle"`
}

type NodeEntry struct {
	HubID string
	Path  string
	Order uint64
	Title string
}

// NodeConfig holds all the configuration options fo generating Archive Nodes
type NodeConfig struct {
	ctx                   context.Context
	Counter               *NodeCounter
	MetsCounter           *MetsCounter
	RecordsCreatedCounter uint64
	RecordsUpdated        uint64
	RecordsDeleted        uint64
	OrgID                 string
	Spec                  string
	Title                 []string
	TitleShort            string
	Revision              int32
	PeriodDesc            []string
	labels                map[string]string
	MimeTypes             map[string][]string
	HubIDs                chan *NodeEntry
	Errors                []*DuplicateError
	// TODO(kiivihal): remove later
	IndexService            *index.Service
	CreateTree              func(cfg *NodeConfig, n *Node, hubID string, id string) *fragments.Tree
	DaoFn                   func(cfg *DaoConfig) error
	ContentIdentical        bool
	Nodes                   chan *Node
	ProcessDigital          bool
	ProcessDigitalIfMissing bool
	RetrieveDao             bool
	ProcessAccessTime       time.Time
	m                       sync.Mutex
	Tags                    []string
}

func (cfg *NodeConfig) Labels() map[string]string {
	return cfg.labels
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
		ctx:     ctx,
		Counter: &NodeCounter{},
		MetsCounter: &MetsCounter{
			uniqueCounter: map[string]int{},
			inError:       map[string]string{},
		},
		labels: make(map[string]string),
		HubIDs: make(chan *NodeEntry, 100),
	}
}

// MetsCounter is a concurrency safe counter for number of Mets-files processed
type MetsCounter struct {
	counter        uint64
	digitalObjects uint64
	errors         uint64
	inError        map[string]string
	uniqueCounter  map[string]int
	m              sync.Mutex
}

// Increment increments the count by one
func (mc *MetsCounter) Increment(daoLink string) {
	atomic.AddUint64(&mc.counter, 1)
	mc.m.Lock()
	mc.uniqueCounter[daoLink]++
	mc.m.Unlock()
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

func (mc *MetsCounter) AppendError(unitID string, errMsg string) {
	mc.IncrementError()
	mc.m.Lock()
	defer mc.m.Unlock()
	mc.inError[unitID] = errMsg
}

func (mc *MetsCounter) GetErrors() map[string]string {
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
	defer func() {
		if cfg.Nodes != nil {
			close(cfg.Nodes)
		}
	}()

	nl := &NodeList{}

	if dsc == nil {
		return nl, 0, nil
	}

	nl.Type = dsc.Attrtype
	for _, label := range dsc.Chead {
		nl.Label = append(nl.Label, label.Head)
	}

	for _, p := range dsc.Cp {
		cc, err := p.NewClevel()
		if err != nil {
			return nil, 0, err
		}
		node, err := NewNode(CLevel(cc), []string{}, cfg)
		if err != nil {
			return nil, 0, err
		}
		if cfg.Nodes != nil {
			cfg.Nodes <- node
			continue
		}

		// legacy add should not happen if there is a node channel
		nl.Nodes = append(nl.Nodes, node)
	}

	for _, cc := range dsc.Numbered {
		node, err := NewNode(cc, []string{}, cfg)
		if err != nil {
			return nil, 0, err
		}
		if cfg.Nodes != nil {
			cfg.Nodes <- node
			continue
		}

		// legacy add should not happen if there is a node channel
		nl.Nodes = append(nl.Nodes, node)
	}

	for _, nn := range dsc.Cc {
		node, err := NewNode(CLevel(nn), []string{}, cfg)
		if err != nil {
			return nil, 0, err
		}
		if cfg.Nodes != nil {
			cfg.Nodes <- node
			continue
		}

		// legacy add should not happen if there is a node channel
		nl.Nodes = append(nl.Nodes, node)
	}

	return nl, cfg.Counter.GetCount(), nil
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
	return html.UnescapeString(strings.Join(h.Label, "; "))
}

// NewNodeID converts a unitid field from the EAD did to a NodeID
func (ui *Cunitid) NewNodeID() (*NodeID, error) {
	id := &NodeID{
		ID:       ui.Unitid,
		TypeID:   ui.Attridentifier,
		Type:     ui.Attrtype,
		Audience: ui.Attraudience,
	}
	return id, nil
}

// NewNodeIDs extract Unit Identifiers from the EAD did
func (cdid *Cdid) NewNodeIDs() ([]*NodeID, string, error) {
	var (
		ids         []*NodeID
		inventoryID string
	)

	for _, unitid := range cdid.Cunitid {
		id, err := unitid.NewNodeID()
		if err != nil {
			return nil, "", err
		}

		switch id.Type {
		case "ABS", "series_code", "blank", "analoog", "BD", "brocade", "":
			inventoryID = id.ID
		}

		ids = append(ids, id)
	}

	return ids, inventoryID, nil
}

// NewNodeDate extract date information frme the EAD unitdate
func (date *Cunitdate) NewNodeDate() (*NodeDate, error) {
	nDate := &NodeDate{
		Calendar: date.Attrcalendar,
		Era:      date.Attrera,
		Normal:   date.Attrnormal,
		Label:    date.Unitdate,
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
		Genreform: config.Config.EAD.GenreFormDefault,
	}

	if len(cdid.Cphysdesc) != 0 {
		header.Physdesc = sanitizeXMLAsString(cdid.Cphysdesc[0].Raw)
	}

	if len(cdid.Cdao) != 0 {
		header.HasDigitalObject = true
		header.DaoLink = cdid.Cdao[0].Attrhref
	}

	for _, label := range cdid.Cunittitle {
		// todo interpolation of date and title is not correct at the moment.
		if len(label.Cunitdate) != 0 {
			header.DateAsLabel = true

			for _, date := range label.Cunitdate {
				nodeDate, err := date.NewNodeDate()
				if err != nil {
					return nil, err
				}

				header.Date = append(header.Date, nodeDate)
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
		if strings.EqualFold(unitID.Attrtype, "bd") {
			header.AltRender = "Born Digital"
		}

		// TODO(kiivihal): add series_code identifier
		if unitID.Attridentifier != "" {
			header.Attridentifier = unitID.Attridentifier
		}

		if header.Attridentifier == "" && unitID.Attrtype == "series_code" {
			header.Attridentifier = unitID.Unitid
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

	if len(cdid.Cphysloc) != 0 {
		header.Physloc = string(cdid.Cphysloc[0].Raw)
	}

	return header, nil
}

func (n *Node) getPathID() string {
	eadID := n.Header.InventoryNumber
	if eadID == "" || strings.HasPrefix(eadID, "---") {
		eadID = strconv.FormatUint(n.Order, 10)
	}
	return eadID
}

func (cfg *NodeConfig) UpdatePath(node *Node, parentIDs []string) ([]string, error) {
	cfg.m.Lock()
	defer cfg.m.Unlock()

	if len(parentIDs) > 0 {
		node.BranchID = parentIDs[len(parentIDs)-1]
		node.Path = fmt.Sprintf("%s%s%s", node.BranchID, pathSep, node.getPathID())
	} else {
		node.Path = node.getPathID()
	}

	_, ok := cfg.labels[node.Path]
	if ok {
		newPath := fmt.Sprintf("%s-%d", node.Path, node.Order)
		log.Warn().Str("oldPath", node.Path).Str("newPath", newPath).
			Str("datasetID", cfg.Spec).Msg("renaming duplicate node path entry")
		node.Path = newPath
	}

	cfg.labels[node.Path] = node.Header.GetTreeLabel()

	ids := append(parentIDs, node.Path)

	return ids, nil
}

// NewNode converts EAD c01 to a Archival Node
func NewNode(cl CLevel, parentIDs []string, cfg *NodeConfig) (*Node, error) {
	cfg.Counter.Increment()

	c := cl.GetCc()

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
		if len(c.GetCaccessrestrict()) != 0 {
			arFirst := c.GetCaccessrestrict()[0]
			node.AccessRestrict = strings.TrimSpace(sanitizer.Sanitize(string(arFirst.Raw)))

			for _, p := range arFirst.Cp {
				if len(p.Cref) != 0 && len(p.Cref[0].Cdate) != 0 {
					node.AccessRestrictYear = p.Cref[0].Cdate[0].Attrnormal
				}
			}
		}
	}

	if c.GetMaterial() != "" {
		node.Material = c.GetMaterial()
	}

	for _, p := range c.GetPhystech() {
		node.Phystech = append(node.Phystech, sanitizeXMLAsString(p.Raw))
		node.PhystechType = p.Attrtype
	}

	// check valid date
	for _, d := range node.Header.Date {
		if validErr := d.ValidDateNormal(); validErr != nil {
			de := &DuplicateError{
				Path:     node.Path,
				Order:    int(node.Order),
				Spec:     cfg.Spec,
				Key:      header.InventoryNumber,
				Label:    header.GetTreeLabel(),
				DupLabel: d.Normal,
				CType:    node.Type,
				Depth:    node.Depth,
				Error:    validErr.Error(),
			}
			cfg.Errors = append(cfg.Errors, de)
		}
	}

	parentIDs, err = cfg.UpdatePath(node, parentIDs)
	if err != nil {
		return nil, err
	}

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
	nested := cl.GetNested()
	node.Children = len(nested)

	if node.Children != 0 {
		for _, nn := range nested {
			n, err := NewNode(nn, parentIDs, cfg)
			if err != nil {
				return nil, err
			}

			if cfg.Nodes != nil {
				cfg.Nodes <- n
				continue
			}

			node.Nodes = append(node.Nodes, n)
		}
	}
	return node, nil
}

// ValidateSpec checks for path traversal characters that should not be in the spec identifier.
func ValidateSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("spec cannot be empty")
	}
	if strings.Contains(spec, "..") {
		return fmt.Errorf("spec cannot have two or more dots in sequence")
	}
	if strings.Contains(spec, "/") || strings.Contains(spec, `\`) {
		return fmt.Errorf("spec cannot contains forward or backward slashes")
	}

	return nil
}
