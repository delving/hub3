package ead

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"sync/atomic"
)

const (
	devStart = 76
	devEnd   = devStart + 100
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
	counter    *DescriptionCounter
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
	Depth int `json:"depth,omitempty"`

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
	Origin           string   `json:"origin,omitempty"`
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
	counter  itemCounter
	desc     *DescriptionCounter
	items    []*DataItem
	q        *Deque
	sections []*DataItem
}

func (ib *itemBuilder) append(item *DataItem) {
	ib.counter.Increment()

	item.Order = ib.counter.GetCount()
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
		desc:    NewDescriptionCounter(),
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

	if len(ead.Carchdesc.Cdescgrp) > 0 {
		ib := newItemBuilder(context.Background())

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
				Order: idx + 1,
			}
			desc.Section = append(desc.Section, info)
		}
		if ib.counter.GetCount() > uint64(0) {
			//desc.Item = ib.items[devStart:devEnd]
			desc.Item = ib.items
			desc.NrSections = len(desc.Section)
			desc.NrItems = len(desc.Item)
		}
		desc.counter = ib.desc
	}

	return desc, nil
}

// newProfile creates a new *Profile from the eadheader profilestmt.
func newProfile(header *Ceadheader) *Profile {
	if header.Cprofiledesc != nil {
		profile := new(Profile)
		if header.Cprofiledesc.Clangusage != nil {
			profile.Language = sanitizeXMLAsString(header.Cprofiledesc.Clangusage.LangUsage)
		}
		if header.Cprofiledesc.Ccreation != nil {
			profile.Creation = sanitizeXMLAsString(header.Cprofiledesc.Ccreation.Creation)
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
					fileDesc.Ctitlestmt.Ctitleproper.TitleProper,
				)
			}
			if fileDesc.Ctitlestmt.Cauthor != nil {
				file.Author = sanitizeXMLAsString(
					fileDesc.Ctitlestmt.Cauthor.Author,
				)
			}
		}
		if fileDesc.Ceditionstmt != nil {
			for _, edition := range fileDesc.Ceditionstmt.Cedition {
				file.Edition = append(
					file.Edition,
					sanitizeXMLAsString(edition.Edition),
				)
			}
		}
		if fileDesc.Cpublicationstmt != nil {
			if fileDesc.Cpublicationstmt.Cpublisher != nil {
				file.Publisher = fileDesc.Cpublicationstmt.Cpublisher.Publisher
			}
			if fileDesc.Cpublicationstmt.Cdate != nil {
				file.PublicationDate = fileDesc.Cpublicationstmt.Cdate.Date
			}
			if len(fileDesc.Cpublicationstmt.Cp) > 0 {
				for _, p := range fileDesc.Cpublicationstmt.Cp {
					if p.Attrid == "copyright" && p.Cextref != nil {
						file.Copyright = p.Cextref.ExtRef
						file.CopyrightURI = p.Cextref.Attrhref
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
	if archdesc.Cdid != nil && fa != nil {
		did := archdesc.Cdid
		for _, title := range did.Cunittitle {
			if title.Attrtype == "short" {
				fa.ShortTitle = sanitizeXMLAsString(title.RawTitle)
				continue
			}
			fa.Title = append(
				fa.Title,
				sanitizeXMLAsString(title.RawTitle),
			)
		}

		unit := new(UnitInfo)

		// only write one ID, only clevel unitids have more than one
		for _, unitid := range did.Cunitid {
			unit.ID = unitid.ID
		}

		for _, date := range did.Cunitdate {
			if date != nil {
				switch date.Attrtype {
				case "bulk":
					unit.DateBulk = date.Date
				default:
					unit.Date = append(unit.Date, date.Date)
				}
			}

		}

		if did.Cphysdesc != nil {
			for _, extent := range did.Cphysdesc.Cextent {
				switch extent.Attrunit {
				case "files":
					unit.Files = extent.Extent
				case "meter", "metre", "metres":
					unit.Length = extent.Extent
				}
			}
			unit.Physical = sanitizeXMLAsString(did.Cphysdesc.Raw)
		}

		if did.Clangmaterial != nil {
			unit.Language = sanitizeXMLAsString(did.Clangmaterial.Raw)
		}

		if did.Cmaterialspec != nil {
			unit.Material = sanitizeXMLAsString(did.Cmaterialspec.Raw)
		}

		if did.Crepository != nil {
			unit.Repository = sanitizeXMLAsString(did.Crepository.Raw)
		}

		if did.Cphysloc != nil {
			unit.PhysicalLocation = sanitizeXMLAsString(did.Cphysloc.Raw)
		}

		if did.Corigination != nil {
			unit.Origin = sanitizeXMLAsString(did.Corigination.Raw)
		}

		if did.Cabstract != nil {
			unit.Abstract = did.Cabstract.Abstract()
		}

		fa.UnitInfo = unit
	}

	return nil
}
