package ead

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/render"
	r "github.com/kiivihal/rdf2go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
)

var sanitizer *bluemonday.Policy

func init() {
	sanitizer = bluemonday.StrictPolicy()
}

func sanitizeXML(b []byte) []byte {
	return bytes.TrimSpace(sanitizer.SanitizeBytes(b))
}

func sanitizeXMLAsString(b []byte) string {
	return string(sanitizeXML(b))
}

// ReadEAD reads an ead2002 XML from a path
func ReadEAD(path string) (*Cead, error) {
	rawEAD, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return eadParse(rawEAD)
}

func Unmarshal(src []byte) (*Cead, error) {
	return eadParse(src)
}

// Parse parses a ead2002 XML file into a set of Go structures
func eadParse(src []byte) (*Cead, error) {
	ead := new(Cead)
	err := xml.Unmarshal(src, ead)
	return ead, err
}

func ProcessEAD(r io.Reader, headerSize int64, spec string, p *elastic.BulkProcessor) (*NodeConfig, error) {
	os.MkdirAll(c.Config.EAD.CacheDir, os.ModePerm)

	f, err := ioutil.TempFile(c.Config.EAD.CacheDir, "*")
	defer f.Close()
	if err != nil {
		log.Printf("Unable to create output file %s; %s", spec, err)
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, headerSize))
	_, err = io.Copy(f, io.TeeReader(r, buf))

	if err != nil {
		return nil, err
	}

	cead, err := eadParse(buf.Bytes())
	if err != nil {
		log.Printf("Error during parsing; %s", err)
		return nil, err
	}

	if spec == "" {
		spec = cead.Ceadheader.Ceadid.EadID
	}

	f.Close()
	basePath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s", spec))
	os.Rename(f.Name(), fmt.Sprintf("%s.xml", basePath))

	ds, _, err := models.GetOrCreateDataSet(spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", spec)
		return nil, err
	}

	ds, err = ds.IncrementRevision()
	if err != nil {
		log.Printf("Unable to increment %s\n", spec)
		return nil, err
	}

	// set basics for ead
	ds.Label = cead.Ceadheader.GetTitle()
	// TODO enable again for born digital as well
	//ds.Owner = cead.Ceadheader.GetOwner()
	//ds.Abstract = cead.Carchdesc.GetAbstract()
	ds.Period = cead.Carchdesc.GetPeriods()

	cfg := NewNodeConfig(context.Background())
	cfg.CreateTree = CreateTree
	cfg.Spec = spec
	cfg.OrgID = c.Config.OrgID
	cfg.Revision = int32(ds.Revision)

	// create desciption
	desc, err := NewDescription(cead)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create description")
	}

	jsonOutput, err := json.MarshalIndent(desc, "", " ")
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to marshall description to JSON")
	}

	err = ioutil.WriteFile(
		fmt.Sprintf("%s.json", basePath),
		jsonOutput,
		0644,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to JSON description to disk")
	}

	// save description
	var unitInfo *UnitInfo
	if desc.Summary.FindingAid != nil && desc.Summary.FindingAid.UnitInfo != nil {
		unitInfo = desc.Summary.FindingAid.UnitInfo
		ds.Length = unitInfo.Length
		ds.Files = unitInfo.Files
		ds.Abstract = unitInfo.Abstract
		ds.Language = unitInfo.Language
		ds.Material = unitInfo.Material
		ds.ArchiveCreator = unitInfo.Origin
	}

	nl, _, err := cead.Carchdesc.Cdsc.NewNodeList(cfg)
	if err != nil {
		log.Printf("Error during parsing; %s", err)
		return cfg, err
	}

	ds.MetsFiles = int(cfg.MetsCounter.GetCount())
	ds.Clevels = int(cfg.Counter.GetCount())
	ds.Description = string(cead.RawDescription())

	err = ds.Save()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to save dataset")
	}

	err = cead.SaveDescription(cfg, unitInfo, p)
	if err != nil {
		log.Printf("Unable to save description for %s; %#v", spec, err)
		return nil, errors.Wrapf(err, "Unable to create index representation of the description")
	}

	if p != nil {
		go func() {
			start := time.Now()
			err := nl.ESSave(cfg, p)
			if err != nil {
				log.Printf("Unable to save nodes; %s", err)
			}

			_, err = ds.DropOrphans(context.TODO(), p, nil)
			if err != nil {
				log.Printf("Unable to drop orphans; %s", err)
			}
			end := time.Since(start)
			log.Printf("saving %s with %d records took: %s", spec, cfg.Counter.GetCount(), end)
		}()
	}

	return cfg, nil
}

func ProcessUpload(r *http.Request, w http.ResponseWriter, spec string, p *elastic.BulkProcessor) (uint64, error) {

	in, header, err := r.FormFile("ead")
	if err != nil {
		return uint64(0), err
	}
	defer in.Close()

	cfg, err := ProcessEAD(in, header.Size, spec, p)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return 0, err
	}

	render.PlainText(w, r, fmt.Sprintf("Processed %d for dataset %s\n", cfg.Counter.GetCount(), spec))
	log.Printf("nr of errors: %d", len(cfg.Errors))
	if len(cfg.Errors) > 0 {
		render.PlainText(w, r, fmt.Sprintf("Duplicate inventory numbers %d for dataset %s\n", len(cfg.Errors), spec))
		d, err := cfg.ErrorToCSV()
		if err != nil {
			return uint64(0), err
		}
		w.Write(d)
	}

	return cfg.Counter.GetCount(), nil

}

//////////////////////////////
//// clevels
/////////////////////////////

type CLevel interface {
	GetXMLName() xml.Name
	GetAttrlevel() string
	GetAttrotherlevel() string
	GetCaccessrestrict() *Caccessrestrict
	GetNested() []CLevel
	GetCdid() *Cdid
	GetScopeContent() *Cscopecontent
	GetOdd() []*Codd
	GetMaterial() string
	GetRaw() []byte
}

type Cc struct {
	XMLName         xml.Name         `xml:"c,omitempty" json:"c,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Cc              []*Cc            `xml:"c,omitempty" json:"c,omitempty"`
	Ccustodhist     *Ccustodhist     `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Cdid            []*Cdid          `xml:"did,omitempty" json:"did,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Raw             []byte           `xml:",innerxml" json:",omitempty"`
}

func (c Cc) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc) GetCdid() *Cdid                       { return c.Cdid[0] }
func (c Cc) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc) GetOdd() []*Codd                      { return c.Codd }
func (c Cc) GetNested() []CLevel                  { return c.Nested() }
func (c Cc) Nested() []CLevel {
	levels := make([]CLevel, len(c.Cc))
	for i, v := range c.Cc {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc) GetMaterial() string {
	if c.Ccontrolaccess != nil && len(c.Ccontrolaccess.Cp) > 0 {
		return c.Ccontrolaccess.Cp[0].P
	}
	cdid := c.GetCdid()

	if cdid.Cphysdesc != nil && cdid.Cphysdesc.Cphysfacet != nil {
		return cdid.Cphysdesc.Cphysfacet.PhysFacet
	}
	return ""
}
func (c Cc) GetRaw() []byte { return c.Raw }

///////////////////////////
/// structs
///////////////////////////

type Cead struct {
	XMLName      xml.Name    `xml:"ead,omitempty" json:"ead,omitempty"`
	Attraudience string      `xml:"audience,attr"  json:",omitempty"`
	Ceadheader   *Ceadheader `xml:"eadheader,omitempty" json:"eadheader,omitempty"`
	Carchdesc    *Carchdesc  `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
}

// SaveDescription stores the FragmentGraph of the EAD description in ElasticSearch
func (cead *Cead) SaveDescription(cfg *NodeConfig, unitInfo *UnitInfo, p *elastic.BulkProcessor) error {
	fg, _, err := cead.DescriptionGraph(cfg, unitInfo)
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

	return nil
}

// RawDescription returns the EAD description stripped of all markup.
func (cead *Cead) RawDescription() []byte {
	description := cead.Ceadheader.Raw
	description = append(description, cead.Carchdesc.Cdid.Raw...)
	for _, dscGrp := range cead.Carchdesc.Cdescgrp {
		description = append(description, dscGrp.Raw...)
	}
	for _, bioghist := range cead.Carchdesc.Cbioghist {
		description = append(description, bioghist.Raw...)
	}
	if cead.Carchdesc.Cuserestrict != nil {
		description = append(description, cead.Carchdesc.Cuserestrict.Raw...)
	}

	// strip all tags
	regex := regexp.MustCompile(`\s+`)
	description = regex.ReplaceAll(description, []byte(" "))
	return sanitizer.SanitizeBytes(description)
}

// DescriptionGraph returns the graph of the Description section (archdesc, descgroups, desc/did) as a FragmentGraph
func (cead *Cead) DescriptionGraph(cfg *NodeConfig, unitInfo *UnitInfo) (*fragments.FragmentGraph, *fragments.ResourceMap, error) {
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

	// TODO store triples later from n.Triples
	tree := &fragments.Tree{}

	tree.HubID = header.HubID
	tree.ChildCount = 0
	tree.Type = "desc"
	tree.InventoryID = cfg.Spec
	tree.Title = cead.Ceadheader.GetTitle()
	tree.AgencyCode = cead.Ceadheader.Ceadid.Attrmainagencycode
	tree.Description = string(cead.RawDescription())
	tree.PeriodDesc = cead.Carchdesc.GetNormalPeriods()

	// add periodDesc to nodeConfig so they can be applied to each cLevel
	cfg.PeriodDesc = tree.PeriodDesc

	s := r.NewResource(subject)
	t := func(s r.Term, p, o string, oType convert, idx int) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			err := rm.AppendOrderedTriple(t, false, idx)
			if err != nil {
				log.Printf("unable to add triple: %#v", err)
			}
		}
		return
	}

	intType := func(value string) r.Term {
		return r.NewLiteralWithDatatype(value, r.NewResource("http://www.w3.org/2001/XMLSchema#integer"))
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
	floatType := func(value string) r.Term {
		return r.NewLiteralWithDatatype(value, r.NewResource("http://www.w3.org/2001/XMLSchema#float"))
	}
	// add total clevels
	t(s, "nrClevels", fmt.Sprintf("%d", cfg.Counter.GetCount()), intType, 0)
	if unitInfo != nil {
		t(s, "files", extractDigit(unitInfo.Files), intType, 0)
		t(s, "length", extractDigit(unitInfo.Length), floatType, 0)
		for _, abstract := range unitInfo.Abstract {
			t(s, "abstract", abstract, r.NewLiteral, 0)
		}
		t(s, "material", unitInfo.Material, r.NewLiteral, 0)
		t(s, "language", unitInfo.Language, r.NewLiteral, 0)
		for _, origin := range unitInfo.Origin {
			t(s, "origin", origin, r.NewLiteral, 0)
		}
	}

	// add period desc for range search from the archdesc > did > date
	for idx, p := range tree.PeriodDesc {
		t(s, "periodDesc", p, r.NewLiteral, idx)
	}

	fg := fragments.NewFragmentGraph()
	fg.Meta = header
	fg.Tree = tree

	// only set resources when the full graph is filled.
	fg.SetResources(rm)
	return fg, rm, nil
}

/////////////////////////
///     EAD Header
///
/// Structs ordered by EAD order
////////////////////////

type Ceadheader struct {
	XMLName                xml.Name       `xml:"eadheader,omitempty" json:"eadheader,omitempty"`
	Attrcountryencoding    string         `xml:"countryencoding,attr"  json:",omitempty"`
	Attrdateencoding       string         `xml:"dateencoding,attr"  json:",omitempty"`
	Attrfindaidstatus      string         `xml:"findaidstatus,attr"  json:",omitempty"`
	Attrlangencoding       string         `xml:"langencoding,attr"  json:",omitempty"`
	Attrrepositoryencoding string         `xml:"repositoryencoding,attr"  json:",omitempty"`
	Attrscriptencoding     string         `xml:"scriptencoding,attr"  json:",omitempty"`
	Ceadid                 *Ceadid        `xml:"eadid,omitempty" json:"eadid,omitempty"`
	Cfiledesc              *Cfiledesc     `xml:"filedesc,omitempty" json:"filedesc,omitempty"`
	Cprofiledesc           *Cprofiledesc  `xml:"profiledesc,omitempty" json:"profiledesc,omitempty"`
	Crevisiondesc          *Crevisiondesc `xml:"revisiondesc,omitempty" json:"revisiondesc,omitempty"`
	Raw                    []byte         `xml:",innerxml" json:",omitempty"`
}

// GetTitle returns the title of the EAD
func (eh Ceadheader) GetTitle() string {
	if eh.Cfiledesc != nil && eh.Cfiledesc.Ctitlestmt != nil && eh.Cfiledesc.Ctitlestmt.Ctitleproper != nil {
		return string(eh.Cfiledesc.Ctitlestmt.Ctitleproper.TitleProper)
	}
	return ""
}

// GetOwner returns the owner of the EAD
func (eh Ceadheader) GetOwner() string {
	if eh.Cfiledesc != nil && eh.Cfiledesc.Cpublicationstmt != nil && eh.Cfiledesc.Cpublicationstmt.Cpublisher != nil {
		return eh.Cfiledesc.Cpublicationstmt.Cpublisher.Publisher
	}
	return ""
}

type Ceadid struct {
	XMLName            xml.Name `xml:"eadid,omitempty" json:"eadid,omitempty"`
	Attrcountrycode    string   `xml:"countrycode,attr"  json:",omitempty"`
	Attrmainagencycode string   `xml:"mainagencycode,attr"  json:",omitempty"`
	Attrpublicid       string   `xml:"publicid,attr"  json:",omitempty"`
	Attrurl            string   `xml:"url,attr"  json:",omitempty"`
	Attrurn            string   `xml:"urn,attr"  json:",omitempty"`
	EadID              string   `xml:",chardata" json:",omitempty"`
}

type Cfiledesc struct {
	XMLName          xml.Name          `xml:"filedesc,omitempty" json:"filedesc,omitempty"`
	Ctitlestmt       *Ctitlestmt       `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
	Ceditionstmt     *Ceditionstmt     `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Cpublicationstmt *Cpublicationstmt `xml:"publicationstmt,omitempty" json:"publicationstmt,omitempty"`
}

type Ctitlestmt struct {
	XMLName      xml.Name      `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
	Ctitleproper *Ctitleproper `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
	Cauthor      *Cauthor      `xml:"author,omitempty" json:"author,omitempty"`
}

type Ctitleproper struct {
	XMLName     xml.Name `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
	TitleProper []byte   `xml:",innerxml" json:",omitempty"`
}

type Cauthor struct {
	XMLName xml.Name `xml:"author,omitempty" json:"author,omitempty"`
	Author  []byte   `xml:",innerxml" json:",omitempty"`
}

type Ceditionstmt struct {
	XMLName  xml.Name    `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Cedition []*Cedition `xml:"edition,omitempty" json:"edition,omitempty"`
}

type Cedition struct {
	XMLName xml.Name `xml:"edition,omitempty" json:"edition,omitempty"`
	Edition []byte   `xml:",innerxml" json:",omitempty"`
}

type Cprofiledesc struct {
	XMLName    xml.Name    `xml:"profiledesc,omitempty" json:"profiledesc,omitempty"`
	Ccreation  *Ccreation  `xml:"creation,omitempty" json:"creation,omitempty"`
	Clangusage *Clangusage `xml:"langusage,omitempty" json:"langusage,omitempty"`
	Cdescrules *Cdescrules `xml:"descrules,omitempty" json:"descrules,omitempty"`
}

type Ccreation struct {
	XMLName      xml.Name `xml:"creation,omitempty" json:"creation,omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Creation     []byte   `xml:",innerxml" json:",omitempty"`
}

type Clangusage struct {
	XMLName   xml.Name   `xml:"langusage,omitempty" json:"langusage,omitempty"`
	Clanguage *Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	LangUsage []byte     `xml:",innerxml" json:",omitempty"`
}

type Cdescrules struct {
	XMLName      xml.Name `xml:"descrules,omitempty" json:"descrules,omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	//Cbibref      []*Cbibref `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Descrrules string `xml:",innerxml" json:",omitempty"`
}

type Crevisiondesc struct {
	XMLName      xml.Name `xml:"revisiondesc,omitempty" json:"revisiondesc,omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Cchange      *Cchange `xml:"change,omitempty" json:"change,omitempty"`
}

type Cchange struct {
	XMLName xml.Name `xml:"change,omitempty" json:"change,omitempty"`
	Cdate   *Cdate   `xml:"date,omitempty" json:"date,omitempty"`
	Citem   []*Citem `xml:"item,omitempty" json:"item,omitempty"`
}
type Citem struct {
	XMLName xml.Name `xml:"item,omitempty" json:"item,omitempty"`
	Cemph   []*Cemph `xml:"emph,omitempty" json:"emph,omitempty"`
	Cextref *Cextref `xml:"extref,omitempty" json:"extref,omitempty"`
	Clist   *Clist   `xml:"list,omitempty" json:"list,omitempty"`
	Item    string   `xml:",chardata" json:",omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
}

type Cabstract struct {
	XMLName   xml.Name `xml:"abstract,omitempty" json:"abstract,omitempty"`
	Attrlabel string   `xml:"label,attr"  json:",omitempty"`
	Clb       []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Raw       []byte   `xml:",innerxml" json:",omitempty"`
}

//////////////////////////////////////////////////
////     ArchDesc
//////////////////////////////////////////////////

// orderd by occurence in the EAD

func (ad Carchdesc) GetAbstract() []string {
	return ad.Cdid.Cabstract.Abstract()
}

func (ad Carchdesc) GetPeriods() []string {
	dates := []string{}
	for _, date := range ad.Cdid.Cunitdate {
		if date.Date != "" {
			dates = append(dates, date.Date)
		}
	}
	return dates
}

func (ad Carchdesc) GetNormalPeriods() []string {
	dates := []string{}
	for _, date := range ad.Cdid.Cunitdate {
		if date.Attrnormal != "" && date.Attrtype != "bulk" {
			dates = append(dates, date.Attrnormal)
		}
	}
	return dates
}

// Abstract returns the Abstract split on EAD '<lb />', i.e. line-break
func (ca Cabstract) Abstract() []string {
	if len(ca.Raw) == 0 {
		return []string{}
	}
	raw := bytes.ReplaceAll(ca.Raw, []byte("extref"), []byte("a"))
	raw = bytes.ReplaceAll(raw, []byte(" />"), []byte("/>"))

	parts := strings.Split(
		fmt.Sprintf("%s", raw),
		"<lb/>",
	)
	trimmed := []string{}
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if len(t) != 0 {
			trimmed = append(trimmed, t)
		}
	}

	return trimmed

}

type Caccessrestrict struct {
	XMLName      xml.Name      `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Attrid       string        `xml:"id,attr"  json:",omitempty"`
	Attrtype     string        `xml:"type,attr"  json:",omitempty"`
	Chead        []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clegalstatus *Clegalstatus `xml:"legalstatus,omitempty" json:"legalstatus,omitempty"`
	Cp           []*Cp         `xml:"p,omitempty" json:"p,omitempty"`
	Raw          []byte        `xml:",innerxml" json:",omitempty"`
}

type Caccruals struct {
	XMLName xml.Name `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cacqinfo struct {
	XMLName xml.Name `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Caltformavail struct {
	XMLName  xml.Name `xml:"altformavail,omitempty" json:"altformavail,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cappraisal struct {
	XMLName xml.Name `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Carchdesc struct {
	XMLName      xml.Name      `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
	Attrlevel    string        `xml:"level,attr"  json:",omitempty"`
	Attrtype     string        `xml:"type,attr"  json:",omitempty"`
	Cdescgrp     []*Cdescgrp   `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
	Cdid         *Cdid         `xml:"did,omitempty" json:"did,omitempty"`
	Cdsc         *Cdsc         `xml:"dsc,omitempty" json:"dsc,omitempty"`
	Cbioghist    []*Cbioghist  `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Cuserestrict *Cuserestrict `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
}

type Carrangement struct {
	XMLName xml.Name `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cbibref struct {
	XMLName xml.Name  `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Ctitle  []*Ctitle `xml:"title,omitempty" json:"title,omitempty"`
	Bibref  string    `xml:",chardata" json:",omitempty"`
}

type Cbioghist struct {
	XMLName   xml.Name     `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Cbioghist []*Cbioghist `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Chead     []*Chead     `xml:"head,omitempty" json:"head,omitempty"`
	Cp        []*Cp        `xml:"p,omitempty" json:"p,omitempty"`
	Raw       []byte       `xml:",innerxml" json:",omitempty"`
}

type Cblockquote struct {
	XMLName xml.Name `xml:"blockquote,omitempty" json:"blockquote,omitempty"`
	Cnote   *Cnote   `xml:"note,omitempty" json:"note,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cchronitem struct {
	XMLName xml.Name `xml:"chronitem,omitempty" json:"chronitem,omitempty"`
	Cdate   *Cdate   `xml:"date,omitempty" json:"date,omitempty"`
	Cevent  *Cevent  `xml:"event,omitempty" json:"event,omitempty"`
}

type Cchronlist struct {
	XMLName    xml.Name      `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Cchronitem []*Cchronitem `xml:"chronitem,omitempty" json:"chronitem,omitempty"`
	Chead      []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
}

type Ccontrolaccess struct {
	XMLName      xml.Name    `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Attraudience string      `xml:"audience,attr"  json:",omitempty"`
	Csubject     []*Csubject `xml:"subject,omitempty" json:"subject,omitempty"`
	Cnote        *Cnote      `xml:"note,omitempty" json:"note,omitempty"`
	Cp           []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
}

type Ccorpname struct {
	XMLName  xml.Name `xml:"corpname,omitempty" json:"corpname,omitempty"`
	CorpName string   `xml:",chardata" json:",omitempty"`
}

type Ccustodhist struct {
	XMLName  xml.Name    `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Cacqinfo []*Cacqinfo `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Chead    []*Chead    `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
}

type Cdate struct {
	XMLName      xml.Name `xml:"date,omitempty" json:"date,omitempty"`
	Attrcalendar string   `xml:"calendar,attr"  json:",omitempty"`
	Attrera      string   `xml:"era,attr"  json:",omitempty"`
	Attrnormal   string   `xml:"normal,attr"  json:",omitempty"`
	Date         string   `xml:",chardata" json:",omitempty"`
}

type Cdescgrp struct {
	XMLName          xml.Name          `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
	Attrtype         string            `xml:"type,attr"  json:",omitempty"`
	Caccessrestrict  *Caccessrestrict  `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Caccruals        *Caccruals        `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Caltformavail    *Caltformavail    `xml:"altformavail,omitempty" json:"altformavail,omitempty"`
	Cappraisal       *Cappraisal       `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Carrangement     *Carrangement     `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Cbioghist        []*Cbioghist      `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Ccontrolaccess   *Ccontrolaccess   `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Ccustodhist      *Ccustodhist      `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Chead            []*Chead          `xml:"head,omitempty" json:"head,omitempty"`
	Codd             []*Codd           `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech        *Cphystech        `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Cprefercite      *Cprefercite      `xml:"prefercite,omitempty" json:"prefercite,omitempty"`
	Cprocessinfo     *Cprocessinfo     `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Crelatedmaterial *Crelatedmaterial `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Cuserestrict     *Cuserestrict     `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
	Raw              []byte            `xml:",innerxml" json:",omitempty"`
}

type Cdid struct {
	XMLName       xml.Name       `xml:"did,omitempty" json:"did,omitempty"`
	Cabstract     *Cabstract     `xml:"abstract,omitempty" json:"abstract,omitempty"`
	Cdao          *Cdao          `xml:"dao,omitempty" json:"dao,omitempty"`
	Chead         []*Chead       `xml:"head,omitempty" json:"head,omitempty"`
	Clangmaterial *Clangmaterial `xml:"langmaterial,omitempty" json:"langmaterial,omitempty"`
	Cmaterialspec *Cmaterialspec `xml:"materialspec,omitempty" json:"materialspec,omitempty"`
	Corigination  *Corigination  `xml:"origination,omitempty" json:"origination,omitempty"`
	Cphysdesc     *Cphysdesc     `xml:"physdesc,omitempty" json:"physdesc,omitempty"`
	Cphysloc      *Cphysloc      `xml:"physloc,omitempty" json:"physloc,omitempty"`
	Crepository   *Crepository   `xml:"repository,omitempty" json:"repository,omitempty"`
	Cunitdate     []*Cunitdate   `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Cunitid       []*Cunitid     `xml:"unitid,omitempty" json:"unitid,omitempty"`
	Cunittitle    []*Cunittitle  `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
	Raw           []byte         `xml:",innerxml" json:",omitempty"`
}

type Cdao struct {
	XMLName      xml.Name `xml:"dao,omitempty" json:"dao,omitempty"`
	Attractuate  string   `xml:"actuate,attr"  json:",omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Attrhref     string   `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrrole     string   `xml:"role,attr"  json:",omitempty"`
	Attrshow     string   `xml:"show,attr"  json:",omitempty"`
}

type Cdsc struct {
	XMLName  xml.Name `xml:"dsc,omitempty" json:"dsc,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Nested   []*Cc01  `xml:"c01,omitempty" json:"c01,omitempty"`
	Cc       []*Cc    `xml:"c,omitempty" json:"c,omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
}

type Cemph struct {
	XMLName    xml.Name `xml:"emph,omitempty" json:"emph,omitempty"`
	Attrrender string   `xml:"render,attr"  json:",omitempty"`
	Emph       string   `xml:",chardata" json:",omitempty"`
}

type Cevent struct {
	XMLName xml.Name `xml:"event,omitempty" json:"event,omitempty"`
	Event   string   `xml:",chardata" json:",omitempty"`
}

type Cextent struct {
	XMLName  xml.Name `xml:"extent,omitempty" json:"extent,omitempty"`
	Attrunit string   `xml:"unit,attr"  json:",omitempty"`
	Extent   string   `xml:",chardata" json:",omitempty"`
}

type Cextref struct {
	XMLName      xml.Name `xml:"extref,omitempty" json:"extref,omitempty"`
	Attractuate  string   `xml:"actuate,attr"  json:",omitempty"`
	Attrhref     string   `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string   `xml:"show,attr"  json:",omitempty"`
	ExtRef       string   `xml:",chardata" json:",omitempty"`
}

type Chead struct {
	XMLName xml.Name `xml:"head,omitempty" json:"head,omitempty"`
	Head    string   `xml:",chardata" json:",omitempty"`
}

type Clangmaterial struct {
	XMLName   xml.Name   `xml:"langmaterial,omitempty" json:"langmaterial,omitempty"`
	Attrlabel string     `xml:"label,attr"  json:",omitempty"`
	Clanguage *Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	Lang      string     `xml:",chardata" json:",omitempty"`
	Raw       []byte     `xml:",innerxml" json:",omitempty"`
}

type Clanguage struct {
	XMLName        xml.Name `xml:"language,omitempty" json:"language,omitempty"`
	Attrlangcode   string   `xml:"langcode,attr"  json:",omitempty"`
	Attrscriptcode string   `xml:"scriptcode,attr"  json:",omitempty"`
	Language       string   `xml:",chardata" json:",omitempty"`
}

type Clb struct {
	XMLName xml.Name `xml:"lb,omitempty" json:"lb,omitempty"`
}

type Clegalstatus struct {
	XMLName     xml.Name `xml:"legalstatus,omitempty" json:"legalstatus,omitempty"`
	Attrtype    string   `xml:"type,attr"  json:",omitempty"`
	LegalStatus string   `xml:",chardata" json:",omitempty"`
}

type Clist struct {
	XMLName        xml.Name `xml:"list,omitempty" json:"list,omitempty"`
	Attrmark       string   `xml:"mark,attr"  json:",omitempty"`
	Attrnumeration string   `xml:"numeration,attr"  json:",omitempty"`
	Attrtype       string   `xml:"type,attr"  json:",omitempty"`
	Chead          []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Citem          []*Citem `xml:"item,omitempty" json:"item,omitempty"`
}

type Cmaterialspec struct {
	XMLName      xml.Name `xml:"materialspec,omitempty" json:"materialspec,omitempty"`
	Attrlabel    string   `xml:"label,attr"  json:",omitempty"`
	MaterialSpec string   `xml:",chardata" json:",omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
}

type Cnote struct {
	XMLName xml.Name `xml:"note,omitempty" json:"note,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Codd struct {
	XMLName  xml.Name `xml:"odd,omitempty" json:"odd,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Codd     []*Codd  `xml:"odd,omitempty" json:"odd,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
}

type Corigination struct {
	XMLName     xml.Name   `xml:"origination,omitempty" json:"origination,omitempty"`
	Attrlabel   string     `xml:"label,attr"  json:",omitempty"`
	Ccorpname   *Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Origination string     `xml:",chardata" json:",omitempty"`
	Raw         []byte     `xml:",innerxml" json:",omitempty"`
}

type Cp struct {
	XMLName     xml.Name     `xml:"p,omitempty" json:"p,omitempty"`
	Attrid      string       `xml:"id,attr"  json:",omitempty"`
	Cblockquote *Cblockquote `xml:"blockquote,omitempty" json:"blockquote,omitempty"`
	Cchronlist  *Cchronlist  `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Cextref     *Cextref     `xml:"extref,omitempty" json:"extref,omitempty"`
	Clb         []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Clist       *Clist       `xml:"list,omitempty" json:"list,omitempty"`
	Cnote       *Cnote       `xml:"note,omitempty" json:"note,omitempty"`
	Cref        *Cref        `xml:"ref,omitempty" json:"ref,omitempty"`
	Ctitle      []*Ctitle    `xml:"title,omitempty" json:"title,omitempty"`
	P           string       `xml:",chardata" json:",omitempty"`
}

type Cphysdesc struct {
	XMLName    xml.Name    `xml:"physdesc,omitempty" json:"physdesc,omitempty"`
	Attrlabel  string      `xml:"label,attr"  json:",omitempty"`
	Cextent    []*Cextent  `xml:"extent,omitempty" json:"extent,omitempty"`
	PhyscDesc  string      `xml:",chardata" json:",omitempty"`
	Cphysfacet *Cphysfacet `xml:"physfacet,omitempty" json:"physfacet,omitempty"`
	Raw        []byte      `xml:",innerxml" json:",omitempty"`
}

type Cphysfacet struct {
	XMLName   xml.Name `xml:"physfacet,omitempty" json:"physfacet,omitempty"`
	Attrtype  string   `xml:"type,attr"  json:",omitempty"`
	PhysFacet string   `xml:",chardata" json:",omitempty"`
}

type Cphysloc struct {
	XMLName  xml.Name `xml:"physloc,omitempty" json:"physloc,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	PhysLoc  string   `xml:",chardata" json:",omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
}

type Cphystech struct {
	XMLName  xml.Name `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
}

type Cprefercite struct {
	XMLName xml.Name `xml:"prefercite,omitempty" json:"prefercite,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cprocessinfo struct {
	XMLName xml.Name `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cpublicationstmt struct {
	XMLName    xml.Name    `xml:"publicationstmt,omitempty" json:"publicationstmt,omitempty"`
	Cdate      *Cdate      `xml:"date,omitempty" json:"date,omitempty"`
	Cp         []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
	Cpublisher *Cpublisher `xml:"publisher,omitempty" json:"publisher,omitempty"`
}

type Cpublisher struct {
	XMLName   xml.Name `xml:"publisher,omitempty" json:"publisher,omitempty"`
	Publisher string   `xml:",chardata" json:",omitempty"`
}

type Cref struct {
	XMLName      xml.Name `xml:"ref,omitempty" json:"ref,omitempty"`
	Attractuate  string   `xml:"actuate,attr"  json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string   `xml:"show,attr"  json:",omitempty"`
	Attrtarget   string   `xml:"target,attr"  json:",omitempty"`
	Cdate        *Cdate   `xml:"date,omitempty" json:"date,omitempty"`
	Ref          string   `xml:",chardata" json:",omitempty"`
}

type Crelatedmaterial struct {
	XMLName xml.Name `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Crepository struct {
	XMLName    xml.Name `xml:"repository,omitempty" json:"repository,omitempty"`
	Attrlabel  string   `xml:"label,attr"  json:",omitempty"`
	Repository string   `xml:",chardata" json:",omitempty"`
	Raw        []byte   `xml:",innerxml" json:",omitempty"`
}

type Cscopecontent struct {
	XMLName xml.Name `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
}

type Csubject struct {
	XMLName    xml.Name `xml:"subject,omitempty" json:"subject,omitempty"`
	Attrsource string   `xml:"source,attr"  json:",omitempty"`
	Subject    string   `xml:",chardata" json:",omitempty"`
}

type Ctitle struct {
	XMLName      xml.Name `xml:"title,omitempty" json:"title,omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Clb          []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Title        string   `xml:",chardata" json:",omitempty"`
}

type Cunitdate struct {
	XMLName      xml.Name `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Attrcalendar string   `xml:"calendar,attr"  json:",omitempty"`
	Attrera      string   `xml:"era,attr"  json:",omitempty"`
	Attrlabel    string   `xml:"label,attr"  json:",omitempty"`
	Attrnormal   string   `xml:"normal,attr"  json:",omitempty"`
	Attrtype     string   `xml:"type,attr"  json:",omitempty"`
	Date         string   `xml:",chardata" json:",omitempty"`
}

type Cunitid struct {
	XMLName            xml.Name `xml:"unitid,omitempty" json:"unitid,omitempty"`
	Attraudience       string   `xml:"audience,attr"  json:",omitempty"`
	Attrcountrycode    string   `xml:"countrycode,attr"  json:",omitempty"`
	Attridentifier     string   `xml:"identifier,attr"  json:",omitempty"`
	Attrlabel          string   `xml:"label,attr"  json:",omitempty"`
	Attrrepositorycode string   `xml:"repositorycode,attr"  json:",omitempty"`
	Attrtype           string   `xml:"type,attr"  json:",omitempty"`
	ID                 string   `xml:",chardata" json:",omitempty"`
}

type Cunittitle struct {
	XMLName   xml.Name     `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
	Attrlabel string       `xml:"label,attr"  json:",omitempty"`
	Attrtype  string       `xml:"type,attr"  json:",omitempty"`
	Cunitdate []*Cunitdate `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	RawTitle  []byte       `xml:",innerxml" json:",omitempty"`
}

func (ut Cunittitle) Title() string {
	return sanitizer.Sanitize(strings.TrimSpace(fmt.Sprintf("%s", ut.RawTitle)))
}

type Cuserestrict struct {
	XMLName  xml.Name `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
}

///////////////////////////
