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

package ead

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	rdf "github.com/kiivihal/rdf2go"
)

// Description is simplified version of the 'eadheader', 'archdesc/did' and
// 'archdesc/descgroups'.
//
// The goal of the simplification is reduce the complexity of the Archival Description
// for searching and rendering without loosing semantic meaning.
type Description struct {
	Summary    Summary        `json:"summary,omitempty"`
	Section    []*SectionInfo `json:"sections,omitempty"`
	NrSections int            `json:"nrSections,omitempty"`
	NrItems    int            `json:"nrItems,omitempty"`
	NrHits     int            `json:"nrHits"`
	Item       []*DataItem    `json:"item,omitempty"`
}

// SectionInfo holds meta information about each section so that it could
// render the DataItems per section.
type SectionInfo struct {
	Text  string `json:"text,omitempty"`
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
	Order int    `json:"order,omitempty"`
}

// Summary holds the essential metadata information to describe an Archive.
type Summary struct {
	FindingAid *FindingAid `json:"findingAid,omitempty"`
	File       *File       `json:"file,omitempty"`
	Profile    *Profile    `json:"profile,omitempty"`
}

// DataItem holds every data entry in the Archival Description.
type DataItem struct {
	Type  DataType `json:"type"`
	Text  string   `json:"text,omitempty"`
	Label string   `json:"label,omitempty"`
	Note  string   `json:"note,omitempty"`

	// language type
	LangCode   string `json:"langCode,omitempty"`
	ScriptCode string `json:"scriptCode,omitempty"`

	// Unit
	Number string `json:"number,omitempty"`
	Units  string `json:"units,omitempty"`

	// datatype tag
	Tag     string `json:"tag,omitempty"`
	TagType string `json:"tagType,omitempty"`

	// list tag
	ListNumber string `json:"listNumber,omitempty"`
	ListType   string `json:"listType,omitempty"`

	// Repository type
	Link     string `json:"link,omitempty"`
	LinkType string `json:"linkType,omitempty"`
	Activate string `json:"activate,omitempty"`
	ShowLink string `json:"showLink,omitempty"`

	// nested blocks
	//Inner []DataItem `json:"inner,omitempty"`
	Depth     int    `json:"depth,omitempty"`
	ParentIDS string `json:"parentIDS"`

	//FlowType between data items
	FlowType FlowType `json:"flowType"`
	Order    uint64   `json:"order,omitempty"`
	Closed   bool     `json:"closed,omitempty"`

	// date type
	Normal   string `json:"normal,omitempty"`
	Era      string `json:"era,omitempty"`
	Calendar string `json:"calendar,omitempty"`
}

// FindingAid holds the core information about the Archival Record
type FindingAid struct {
	ID         string    `json:"id,omitempty"`
	Country    string    `json:"country,omitempty"`
	AgencyCode string    `json:"agencyCode,omitempty"`
	Title      []string  `json:"title,omitempty"`
	ShortTitle string    `json:"shortTitle,omitempty"`
	UnitInfo   *UnitInfo `json:"unit,omitempty"`
}

// UnitInfo holds the meta information of the Archival Record
type UnitInfo struct {
	Date             []string `json:"date,omitempty"`
	DateBulk         string   `json:"dateBulk"`
	ID               string   `json:"id,omitempty"`
	Physical         string   `json:"physical,omitempty"`
	Files            string   `json:"files,omitempty"`
	Length           string   `json:"length,omitempty"`
	Language         string   `json:"language,omitempty"`
	Material         string   `json:"material,omitempty"`
	Repository       string   `json:"repository,omitempty"`
	PhysicalLocation string   `json:"physicalLocation,omitempty"`
	Origin           []string `json:"origin,omitempty"`
	Abstract         []string `json:"abstract,omitempty"`
}

// File holds the meta-information about the EAD file
type File struct {
	Title           string   `json:"title,omitempty"`
	Author          string   `json:"author,omitempty"`
	Edition         []string `json:"edition,omitempty"`
	Publisher       string   `json:"publisher,omitempty"`
	PublicationDate string   `json:"publicationDate,omitempty"`
	Copyright       string   `json:"copyright,omitempty"`
	CopyrightURI    string   `json:"copyrightURI"`
}

// Profile details information about the creation of the Archival Record
type Profile struct {
	Creation string `json:"creation,omitempty"`
	Language string `json:"language,omitempty"`
}

// itemCounter is a concurrency safe counter for number of Nodes processed
type itemCounter struct {
	counter uint64
}

// Increment increments the count by one
func (dc *itemCounter) Increment() {
	atomic.AddUint64(&dc.counter, 1)
}

// GetCount returns the snapshot of the current count
func (dc *itemCounter) GetCount() uint64 {
	return atomic.LoadUint64(&dc.counter)
}

type itemBuilder struct {
	counter itemCounter
	items   []*DataItem
	q       *Deque
}

// ParentIDs returns a ~ separated list of order identifiers
func (ib *itemBuilder) ParentIDs() string {
	var parentIDs []string

	for _, i := range ib.q.List() {
		if i == nil {
			continue
		}
		item := i.(*DataItem)
		parentIDs = append(
			parentIDs,
			fmt.Sprintf("%d", item.Order),
		)
	}
	return strings.Join(parentIDs, "~")
}
func (ib *itemBuilder) append(item *DataItem) {
	ib.counter.Increment()

	item.Order = ib.counter.GetCount()
	item.ParentIDS = ib.ParentIDs()
	ib.items = append(ib.items, item)
	ib.q.PushBack(item)
	item.Depth = ib.q.Len()
}

// push updates an DataItem or adds a new one.
func (ib *itemBuilder) push(se xml.StartElement) error {
	id := &DataItem{Tag: se.Name.Local}

	switch se.Name.Local {
	case "head":
		// do nothing
		return nil
	case "emph":
		ib.addTextPrevious(" <em>")
		return nil
	case "bioghist", "custodhist", "acqinfo", "scopecontent", "phystech", "otherfindaid", "odd":
		id.Type = SubSection
	case "prefercite", "altformavail", "relatedmaterial":
		id.Type = SubSection
	case "lb":
		if previous := ib.previous(); previous != nil {
			switch previous.Tag {
			case "bibref":
				id.FlowType = Inline
			}
		}
	case "bibref":
		id.FlowType = Inline
	case "title":
		id.FlowType = Inline
		err := ib.close()

		if err != nil {
			return err
		}
	case "p":
		if previous := ib.previous(); previous != nil {
			switch previous.Type {
			case Note:
				id.Type = Note
				id.FlowType = Inline
			}
		} else {
			id.Type = Paragraph
		}
	case "note":
		err := ib.addFlowType(Inline)
		if err != nil {
			return err
		}

		err = ib.close()
		if err != nil {
			return err
		}

		id.Type = Note
		id.FlowType = Inline
	case "list":
		err := ib.close()
		if err != nil {
			return err
		}
		id.Type = List

		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "numeration":
				id.ListNumber = attr.Value
			case "type":
				id.ListType = attr.Value
			}
		}
	case "item":
		id.Type = ListItem
	case "defitem":
		id.Type = DefItem
	case "table":
		id.Type = Table
	case "row":
		previous := ib.previous()
		switch previous.Tag {
		case "thead":
			id.Type = TableHead
		default:
			id.Type = TableRow
		}
	case "entry":
		id.Type = TableCel
	case "label":
		id.Type = ListLabel
	case "chronlist":
		id.Type = ChronList
	case "chronitem":
		id.Type = ChronItem
	case "date":
		id.Type = Date

		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "calendar":
				id.Calendar = attr.Value
			case "era":
				id.Era = attr.Value
			case "normal":
				id.Normal = attr.Value
			}
		}
	case "event":
		id.Type = Event
	case "extref", "extptr":
		err := ib.close()
		if err != nil {
			return err
		}

		id.Type = Link
		id.FlowType = Inline

		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "actuate":
				id.Activate = attr.Value
			case "href":
				id.Link = attr.Value
			case "linktype":
				id.LinkType = attr.Value
			case "show":
				id.ShowLink = attr.Value
			}
		}
	default:
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "label":
				id.Label = attr.Value
			case "type":
				id.TagType = attr.Value
			default:
			}
		}
	}

	ib.append(id)
	return nil
}

// addFlowType marks the last dataItem on the queue as with the specified FlowType.
func (ib *itemBuilder) addFlowType(ft FlowType) error {
	last, ok := ib.q.PopBack()
	if !ok {
		return fmt.Errorf("unable to pop from queue")
	}

	if last != nil {
		elem := last.(*DataItem)
		elem.FlowType = ft
		//log.Printf("Adding flowType to: %#v", elem)
		ib.q.PushBack(elem)
	}

	return nil
}

// addTextPrevious concatenates text with the previous DataItem on the queue.
func (ib *itemBuilder) addTextPrevious(text string) error {
	last, ok := ib.q.PopBack()
	if !ok {
		return fmt.Errorf("unable to pop from queue")
	}
	if last != nil {
		elem := last.(*DataItem)
		elem.Text = fmt.Sprintf("%s%s", elem.Text, text)
		ib.q.PushBack(elem)
	}

	return nil
}

// previous returns the previous element on the queue.
func (ib *itemBuilder) previous() *DataItem {
	last, _ := ib.q.Back()
	if last != nil {
		return last.(*DataItem)
	}
	return nil
}

// close closes the last dataItem on the queue, so that the dataflow is not
// interupted by nested elements.
func (ib *itemBuilder) close() error {
	last, ok := ib.q.PopBack()
	if !ok {
		return fmt.Errorf("unable to pop from queue")
	}
	if last != nil {
		elem := last.(*DataItem)
		elem.Closed = true
		ib.q.PushBack(elem)
	}

	return nil
}

func (ib *itemBuilder) pop(ee xml.EndElement) error {
	switch ee.Name.Local {
	case "head":
	case "emph":
		ib.addTextPrevious("</em> ")
	case "lb":
		last, _ := ib.q.Back()
		if last != nil {
			elem := last.(*DataItem)
			if elem.Text == "-" {
				elem.FlowType = Inline
			}
		}

	default:
		last, ok := ib.q.PopBack()
		if !ok {
			return fmt.Errorf("Unable to pop element from the queue")
		}
		if last != nil {
			elem := last.(*DataItem)
			if elem.Type == Paragraph {
				switch elem.Text {
				case "":
					elem.FlowType = Inline
				case ".":
					elem.FlowType = Next
				case "-":
					elem.FlowType = Inline
				}
			}
		}
	}
	return nil
}

// addText adds text to the latest DataItem on the queue.
// If the DataItem is already closed, then a new one is cloned with a new count.
func (ib *itemBuilder) addText(text []byte) error {
	last, ok := ib.q.Back()
	if !ok {
		// TODO determine what to do
		// when nothing is on the queue. It should never happen.
	}
	if last != nil {
		elem := last.(*DataItem)
		if elem.Closed {
			//log.Printf("creating new item")
			di := elem.clone()
			di.Text = string(text)
			ib.append(di)
			ee := xml.EndElement{Name: xml.Name{Local: "p"}}
			ib.pop(ee)
			return nil
		}
		if elem.Text != "" {
			elem.Text = elem.Text + string(text)
			return nil
		}
		elem.Text = string(text)
	}
	return nil
}

// clone creates a new DataItem that can be appended by the ItemBuilder.
func (di *DataItem) clone() *DataItem {
	return &DataItem{
		Label:      di.Label,
		LangCode:   di.LangCode,
		ScriptCode: di.ScriptCode,
		Number:     di.Number,
		Units:      di.Units,
		Tag:        di.Tag,
		TagType:    di.TagType,
		Link:       di.Link,
		LinkType:   di.LinkType,
		Activate:   di.Activate,
		Normal:     di.Normal,
		Era:        di.Era,
		Calendar:   di.Calendar,
		Type:       di.Type,
	}
}

func (ib *itemBuilder) parse(b []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	total := 0

outer:
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break outer
		}

		switch se := t.(type) {
		case xml.StartElement:
			total++
			err := ib.push(se)
			if err != nil {
				return err
			}

		case xml.EndElement:
			err := ib.pop(se)
			if err != nil {
				return err
			}
		case xml.CharData:
			// trim space and when not empty create new data item from queue type
			text := bytes.TrimSpace(se)
			if len(text) != 0 {
				ib.addText(text)

			}
		default:
		}
	}

	return nil
}

// queuePath returns a path representation of the non-empty DataItems in the queue.
func queuePath(q *Deque) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("len (%d): ", q.Len()))
	for idx, elem := range q.List() {
		if elem != nil {
			sb.WriteString(fmt.Sprintf("%s", elem.(*DataItem).Tag))
			if idx != q.Len()-1 {
				sb.WriteString(" / ")
			}
		}
	}
	return sb.String()
}

func newItemBuilder(ctx context.Context) *itemBuilder {
	return &itemBuilder{
		counter: itemCounter{},
		items:   nil,
		q:       new(Deque),
	}
}

// NewDescription creates an Description from a Cead object.
func NewDescription(ead *Cead) (*Description, error) {
	desc := new(Description)
	if ead.Ceadheader != nil {
		desc.Summary.Profile = newProfile(ead.Ceadheader)
		desc.Summary.File = newFile(ead.Ceadheader)
		desc.Summary.FindingAid = newFindingAid(ead.Ceadheader)

		err := desc.Summary.FindingAid.AddUnit(ead.Carchdesc)
		if err != nil {
			return nil, err
		}
	}

	ib := newItemBuilder(context.Background())

	section := &DataItem{
		Type:  Section,
		Tag:   "eadid",
		Label: "Archief: ",
	}
	ib.append(section)

	// eadid
	if err := ib.parse(ead.Ceadheader.Ceadid.Raw); err != nil {
		return nil, err
	}

	// filedesc
	if err := ib.parse(ead.Ceadheader.Cfiledesc.Raw); err != nil {
		return nil, err
	}
	// Add sections
	info := &SectionInfo{
		Text:  "Archief",
		Start: int(section.Order),
		End:   int(ib.counter.GetCount()),
		Order: 1,
	}
	desc.Section = append(desc.Section, info)

	if len(ead.Carchdesc.Cdid) > 0 {
		section := &DataItem{
			Type: Section,
			Tag:  "archdesc-did",
		}
		ib.append(section)
		if err := ib.parse(ead.Carchdesc.Cdid[0].Raw); err != nil {
			return nil, err
		}

		info := &SectionInfo{
			Text:  section.Text,
			Start: int(section.Order),
			End:   int(ib.counter.GetCount()),
			Order: 2,
		}
		desc.Section = append(desc.Section, info)
	}

	if len(ead.Carchdesc.Cdescgrp) > 0 {

		for idx, grp := range ead.Carchdesc.Cdescgrp {
			section := &DataItem{
				Type:    Section,
				Tag:     "descgrp",
				TagType: grp.Attrtype,
			}
			ib.append(section)
			err := ib.parse(grp.Raw)
			if err != nil {
				return nil, err
			}

			// Add sections
			info := &SectionInfo{
				Text:  section.Text,
				Start: int(section.Order),
				End:   int(ib.counter.GetCount()),
				Order: idx + 3,
			}
			desc.Section = append(desc.Section, info)
		}
	}

	if ib.counter.GetCount() > uint64(0) {
		desc.Item = ib.items
		desc.NrSections = len(desc.Section)
		desc.NrItems = len(desc.Item)
	}

	return desc, nil
}

func (desc *Description) getSpec() string {
	return desc.Summary.FindingAid.ID
}

func GetDescription(spec string) (*Description, error) {
	f, err := os.Open(getDescriptionPath(spec))
	if err != nil {
		return nil, err
	}

	return decodeDescription(f)
}

func (desc *Description) encode(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(desc)
}

func decodeDescription(r io.Reader) (*Description, error) {
	var desc Description

	dec := gob.NewDecoder(r)
	err := dec.Decode(&desc)

	return &desc, err
}

func (desc *Description) Write() error {
	err := os.MkdirAll(GetDataPath(desc.getSpec()), os.ModePerm)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = desc.encode(&buf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		getDescriptionPath(desc.getSpec()),
		buf.Bytes(),
		0644,
	)
}

func (desc *Description) SaveDescription(cfg *NodeConfig, unitInfo *UnitInfo, bi BulkIndex) error {
	fg, _, err := desc.DescriptionGraph(cfg, unitInfo)
	if err != nil {
		return err
	}

	if bi == nil {
		return nil
	}

	m, err := fg.IndexMessage()
	if err != nil {
		return fmt.Errorf("unable to marshal fragment graph; %w", err)
	}

	bi.Publish(context.Background(), m)

	return nil
}

func (desc *Description) DescriptionGraph(cfg *NodeConfig, unitInfo *UnitInfo) (*fragments.FragmentGraph, *fragments.ResourceMap, error) {
	rm := fragments.NewEmptyResourceMap()
	id := "desc"
	subject := newSubject(cfg, id)
	header := &fragments.Header{
		OrgID:         cfg.OrgID,
		Spec:          cfg.Spec,
		Revision:      cfg.Revision,
		HubID:         fmt.Sprintf("%s_%s_%s", cfg.OrgID, cfg.Spec, id),
		DocType:       fragments.FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		Modified:      fragments.NowInMillis(),
		Tags:          []string{"eadDesc"},
	}

	tree := &fragments.Tree{}

	tree.HubID = header.HubID
	tree.ChildCount = 0
	tree.Type = "desc"
	tree.InventoryID = cfg.Spec

	if len(cfg.Title) > 0 {
		tree.Title = cfg.Title[0]
	}

	// TODO(kiivihal): replace raw this with <p> blocks. to align search
	for _, item := range desc.Item {
		tree.Description = append(tree.Description, item.Text)
	}

	tree.PeriodDesc = cfg.PeriodDesc

	if len(tree.PeriodDesc) == 0 {
		de := &DuplicateError{
			Spec:  cfg.Spec,
			Error: "ead period is empty",
		}
		cfg.Errors = append(cfg.Errors, de)
	}

	// add periodDesc to nodeConfig so they can be applied to each cLevel
	cfg.PeriodDesc = tree.PeriodDesc

	s := rdf.NewResource(subject)
	t := func(s rdf.Term, p, o string, oType convert, idx int) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			err := rm.AppendOrderedTriple(t, false, idx)
			if err != nil {
				log.Printf("unable to add triple: %#v", err)
			}
		}
		return
	}

	intType := func(value string) rdf.Term {
		return rdf.NewLiteralWithDatatype(value, rdf.NewResource("http://www.w3.org/2001/XMLSchema#integer"))
	}
	extractDigit := func(value string) string {
		parts := strings.Fields(value)
		for _, part := range parts {
			runes := []rune(value)
			if unicode.IsDigit(runes[0]) {
				return strings.ReplaceAll(part, ",", ".")
			}
		}

		return ""
	}
	floatType := func(value string) rdf.Term {
		return rdf.NewLiteralWithDatatype(value, rdf.NewResource("http://www.w3.org/2001/XMLSchema#float"))
	}
	// add total clevels
	t(s, "nrClevels", fmt.Sprintf("%d", cfg.Counter.GetCount()), intType, 0)

	if unitInfo != nil {
		t(s, "files", extractDigit(unitInfo.Files), intType, 0)
		t(s, "length", extractDigit(unitInfo.Length), floatType, 0)

		for _, abstract := range unitInfo.Abstract {
			t(s, "abstract", abstract, rdf.NewLiteral, 0)
		}

		t(s, "material", unitInfo.Material, rdf.NewLiteral, 0)
		t(s, "language", unitInfo.Language, rdf.NewLiteral, 0)

		for _, origin := range unitInfo.Origin {
			t(s, "origin", origin, rdf.NewLiteral, 0)
		}
	}

	// add period desc for range search from the archdesc > did > date
	for idx, p := range tree.PeriodDesc {
		t(s, "periodDesc", p, rdf.NewLiteral, idx)
	}

	fg := fragments.NewFragmentGraph()
	fg.Meta = header
	fg.Tree = tree

	// only set resources when the full graph is filled.
	fg.SetResources(rm)
	return fg, rm, nil
}

// newProfile creates a new *Profile from the eadheader profilestmt.
func newProfile(header *Ceadheader) *Profile {
	if header.Cprofiledesc != nil {
		profile := new(Profile)
		if header.Cprofiledesc.Clangusage != nil {
			profile.Language = sanitizeXMLAsString(header.Cprofiledesc.Clangusage.Raw)
		}
		if header.Cprofiledesc.Ccreation != nil {
			profile.Creation = sanitizeXMLAsString(header.Cprofiledesc.Ccreation.Raw)
		}
		return profile
	}
	return nil
}

// newFile creates a new *File from the eadheader filestmt.
func newFile(header *Ceadheader) *File {
	fileDesc := header.Cfiledesc
	if fileDesc != nil {
		file := new(File)
		if fileDesc.Ctitlestmt != nil {
			if fileDesc.Ctitlestmt.Ctitleproper != nil {
				file.Title = sanitizeXMLAsString(
					fileDesc.Ctitlestmt.Ctitleproper.Raw,
				)
			}
			if fileDesc.Ctitlestmt.Cauthor != nil {
				file.Author = sanitizeXMLAsString(
					fileDesc.Ctitlestmt.Cauthor.Raw,
				)
			}
		}
		if fileDesc.Ceditionstmt != nil {
			file.Edition = append(
				file.Edition,
				sanitizeXMLAsString(fileDesc.Ceditionstmt.Raw),
			)
		}
		if fileDesc.Cpublicationstmt != nil {
			if fileDesc.Cpublicationstmt.Cpublisher != nil {
				file.Publisher = fileDesc.Cpublicationstmt.Cpublisher.Publisher
			}
			if fileDesc.Cpublicationstmt.Cdate != nil && len(fileDesc.Cpublicationstmt.Cdate) != 0 {
				file.PublicationDate = fileDesc.Cpublicationstmt.Cdate[0].Date
			}
			if len(fileDesc.Cpublicationstmt.Cp) > 0 {
				for _, p := range fileDesc.Cpublicationstmt.Cp {
					if p.Attrid == "copyright" && len(p.Cextref) != 0 {
						file.Copyright = p.Cextref[0].Extref
						file.CopyrightURI = p.Cextref[0].Attrhref
					}
				}
			}
		}
		return file
	}
	return nil
}

// newFindingAid creates a new FindingAid with information from the EadHeader.
// You must call AddUnit to populate the *UnitInfo
func newFindingAid(header *Ceadheader) *FindingAid {
	if header.Ceadid != nil {
		aid := new(FindingAid)
		aid.ID = header.Ceadid.EadID
		aid.Country = header.Ceadid.Attrcountrycode
		aid.AgencyCode = header.Ceadid.Attrmainagencycode
		return aid
	}
	return nil
}

// AddUnit adds the DID information from the ArchDesc to the FindingAid.
func (fa *FindingAid) AddUnit(archdesc *Carchdesc) error {
	if len(archdesc.Cdid) != 0 && fa != nil {
		did := archdesc.Cdid[0]
		for _, title := range did.Cunittitle {
			if title.Attrtype == "short" {
				fa.ShortTitle = sanitizeXMLAsString(title.Raw)
				continue
			}

			fa.Title = append(
				fa.Title,
				sanitizeXMLAsString(title.Raw),
			)
		}

		unit := new(UnitInfo)

		// only write one ID, only clevel unitids have more than one
		for _, unitid := range did.Cunitid {
			unit.ID = unitid.Unitid
		}

		for _, date := range did.Cunitdate {
			if date != nil {
				switch date.Attrtype {
				case "bulk":
					unit.DateBulk = date.Unitdate
				default:
					unit.Date = append(unit.Date, date.Unitdate)
				}
			}

		}

		if len(did.Cphysdesc) != 0 {
			for _, extent := range did.Cphysdesc[0].Cextent {
				switch extent.Attrunit {
				case "files":
					unit.Files = extent.Extent
				case "meter", "metre", "metres":
					unit.Length = extent.Extent
				}
			}
			unit.Physical = sanitizeXMLAsString(did.Cphysdesc[0].Raw)
		}

		if did.Clangmaterial != nil {
			unit.Language = sanitizeXMLAsString(did.Clangmaterial.Raw)
		}

		if len(did.Cmaterialspec) != 0 {
			unit.Material = sanitizeXMLAsString(did.Cmaterialspec[0].Raw)
		}

		if did.Crepository != nil {
			unit.Repository = sanitizeXMLAsString(did.Crepository.Raw)
		}

		if len(did.Cphysloc) != 0 {
			unit.PhysicalLocation = sanitizeXMLAsString(did.Cphysloc[0].Raw)
		}

		if did.Corigination != nil {
			parts := bytes.Split(did.Corigination.Raw, []byte("<corpname>"))
			for _, part := range parts {
				if len(bytes.TrimSpace(part)) == 0 {
					continue
				}

				unit.Origin = append(
					unit.Origin,
					sanitizeXMLAsString(bytes.ReplaceAll(part, []byte(" , "), []byte(" "))),
				)
			}
		}

		if did.Cabstract != nil {
			unit.Abstract = did.Cabstract.CleanAbstract()
		}

		fa.UnitInfo = unit
	}

	return nil
}

func ResaveDescriptions(eadPath string) error {
	dirs, err := ioutil.ReadDir(eadPath)
	if err != nil {
		return err
	}

	var seen int

	for _, ead := range dirs {
		if !ead.IsDir() {
			continue
		}
		spec := ead.Name()

		fname := filepath.Join(eadPath, spec, fmt.Sprintf("%s.xml", spec))
		fmt.Printf("%s\n", fname)
		if _, err := os.Stat(fname); err != nil {
			continue
		}

		cead, err := ReadEAD(fname)
		if err != nil {
			return err
		}
		// create desciption
		desc, err := NewDescription(cead)
		if err != nil {
			return fmt.Errorf("unable to create description; %w", err)
		}

		descIndex := NewDescriptionIndex(spec)

		err = descIndex.CreateFrom(desc)
		if err != nil {
			return fmt.Errorf("unable to create DescriptionIndex; %w", err)
		}

		err = descIndex.Write()
		if err != nil {
			return fmt.Errorf("unable to write DescriptionIndex; %w", err)
		}

		err = desc.Write()
		if err != nil {
			return fmt.Errorf("unable to write description; %w", err)
		}

		meta, _, err := GetOrCreateMeta(spec)
		if err != nil {
			return fmt.Errorf("unable to retrieve meta; %w", err)
		}

		// set basics for ead
		meta.Label = cead.Ceadheader.GetTitle()
		meta.Period = cead.Carchdesc.GetPeriods()

		if err := meta.Write(); err != nil {
			return fmt.Errorf("unable to write meta; %w", err)
		}

		seen++
	}

	log.Printf("updated %d eads", seen)

	return nil
}

func getRemoteDescriptionCount(spec, query string) (int, error) {
	type hits struct {
		Total int
	}

	var netClient = &http.Client{
		Timeout: time.Second * 1,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/ead/%s/desc/index?q=%s", c.Config.DataNodeURL, spec, query), nil)
	if err != nil {
		return 0, err
	}

	resp, err := netClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var descriptionHits hits

	decodeErr := json.NewDecoder(resp.Body).Decode(&descriptionHits)
	if decodeErr != nil {
		return 0, decodeErr
	}

	return descriptionHits.Total, nil
}

func GetDescriptionCount(spec, query string) (int, error) {
	if !c.Config.IsDataNode() {
		return getRemoteDescriptionCount(spec, query)
	}

	var hits int

	descriptionIndex, getErr := GetDescriptionIndex(spec)
	if getErr != nil && !errors.Is(getErr, ErrNoDescriptionIndex) {
		c.Config.Logger.Error().Err(getErr).
			Str("subquery", "description").
			Msg("error with retrieving description index")

		return 0, getErr
	}

	if descriptionIndex != nil {
		searhHits, searchErr := descriptionIndex.SearchWithString(query)
		if searchErr != nil && !errors.Is(searchErr, memory.ErrSearchNoMatch) {
			c.Config.Logger.Error().Err(searchErr).
				Str("subquery", "description").
				Msg("unable to search description")

			return 0, searchErr
		}

		hits = searhHits.Total()
	}

	return hits, nil

}

// HightlightSummary applied query highlights to the ead.Summary.
// func (dq *DescriptionQuery) HightlightSummary(s Summary) Summary {
// if s.Profile != nil {
// s.Profile.Creation, _ = dq.highlightQuery(s.Profile.Creation)
// s.Profile.Language, _ = dq.highlightQuery(s.Profile.Language)
// }

// if s.File != nil {
// s.File.Author, _ = dq.highlightQuery(s.File.Author)
// s.File.Copyright, _ = dq.highlightQuery(s.File.Copyright)
// s.File.PublicationDate, _ = dq.highlightQuery(s.File.PublicationDate)
// s.File.Publisher, _ = dq.highlightQuery(s.File.Publisher)
// s.File.Title, _ = dq.highlightQuery(s.File.Title)

// var editions []string

// for _, e := range s.File.Edition {
// edition, _ := dq.highlightQuery(e)
// editions = append(editions, edition)
// }

// if len(editions) != 0 {
// s.File.Edition = editions
// }
// }

// if s.FindingAid != nil {
// s.FindingAid.AgencyCode, _ = dq.highlightQuery(s.FindingAid.AgencyCode)
// s.FindingAid.Country, _ = dq.highlightQuery(s.FindingAid.Country)
// s.FindingAid.ID, _ = dq.highlightQuery(s.FindingAid.ID)
// s.FindingAid.ShortTitle, _ = dq.highlightQuery(s.FindingAid.ShortTitle)

// var titles []string

// for _, t := range s.FindingAid.Title {
// title, _ := dq.highlightQuery(t)
// titles = append(titles, title)
// }

// if len(titles) != 0 {
// s.FindingAid.Title = titles
// }
// }

// if s.FindingAid != nil && s.FindingAid.UnitInfo != nil {
// unit := s.FindingAid.UnitInfo

// unit.ID, _ = dq.highlightQuery(unit.ID)
// unit.Language, _ = dq.highlightQuery(unit.Language)
// unit.DateBulk, _ = dq.highlightQuery(unit.DateBulk)
// unit.Files, _ = dq.highlightQuery(unit.Files)
// unit.Length, _ = dq.highlightQuery(unit.Length)
// unit.Material, _ = dq.highlightQuery(unit.Material)
// unit.Physical, _ = dq.highlightQuery(unit.Physical)
// unit.PhysicalLocation, _ = dq.highlightQuery(unit.PhysicalLocation)
// unit.Repository, _ = dq.highlightQuery(unit.Repository)

// var origins []string

// for _, o := range unit.Origin {
// origin, _ := dq.highlightQuery(o)
// origins = append(origins, origin)
// }

// if len(origins) != 0 {
// unit.Origin = origins
// }

// var dates []string

// for _, d := range unit.Date {
// date, _ := dq.highlightQuery(d)
// dates = append(dates, date)
// }

// if len(dates) != 0 {
// unit.Date = dates
// }

// var abstracts []string

// for _, a := range unit.Abstract {
// abstract, _ := dq.highlightQuery(a)
// abstracts = append(abstracts, abstract)
// }

// if len(abstracts) != 0 {
// unit.Abstract = abstracts
// }

// s.FindingAid.UnitInfo = unit
// }

// return s
// }
