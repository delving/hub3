package ead

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/go-chi/render"
	r "github.com/kiivihal/rdf2go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/src-d/go-git/plumbing"
)

type BulkIndex interface {
	Publish(ctx context.Context, message ...*domainpb.IndexMessage) error
}

var sanitizer *bluemonday.Policy

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sanitizer = bluemonday.StrictPolicy()
}

func sanitizeXML(b []byte) []byte {
	return bytes.TrimSpace(sanitizer.SanitizeBytes(b))
}

func sanitizeXMLAsString(b []byte) string {
	return string(sanitizeXML(b))
}

// ReadEAD reads an ead2002 XML from a path
func ReadEAD(fpath string) (*Cead, error) {
	rawEAD, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	return eadParse(rawEAD)
}

// Parse parses a ead2002 XML file into a set of Go structures
func eadParse(src []byte) (*Cead, error) {
	ead := new(Cead)
	err := xml.Unmarshal(src, ead)
	return ead, err
}

func ProcessEAD(r io.Reader, headerSize int64, spec string, bi BulkIndex) (*NodeConfig, error) {
	os.MkdirAll(config.Config.EAD.CacheDir, os.ModePerm)

	f, err := ioutil.TempFile(config.Config.EAD.CacheDir, "*")
	if err != nil {
		log.Printf("Unable to create output file %s; %s", spec, err)
		return nil, err
	}
	defer f.Close()

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
		if strings.Contains(spec, "/") {
			spec = strings.ReplaceAll(spec, "/", ".")
		}
	}

	f.Close()

	ds, _, err := models.GetOrCreateDataSet(spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", spec)
		return nil, err
	}
	// check for cache entry
	hash := plumbing.ComputeHash(plumbing.BlobObject, buf.Bytes())
	ds.Fingerprint = hash.String()

	basePath := GetDataPath(spec)
	os.MkdirAll(basePath, os.ModePerm)
	os.Rename(f.Name(), fmt.Sprintf("%s/%s.xml", basePath, spec))

	ds, err = ds.IncrementRevision()
	if err != nil {
		log.Printf("Unable to increment %s\n", spec)
		return nil, err
	}

	// set basics for ead
	ds.Label = cead.Ceadheader.GetTitle()
	ds.Period = cead.Carchdesc.GetPeriods()

	// description must be set to empty
	ds.Description = ""

	cfg := NewNodeConfig(context.Background())
	cfg.CreateTree = CreateTree
	cfg.Spec = spec
	cfg.OrgID = config.Config.OrgID
	cfg.Revision = int32(ds.Revision)

	// create desciption
	desc, err := NewDescription(cead)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create description")
	}

	cfg.Title = []string{desc.Summary.File.Title}
	cfg.TitleShort = desc.Summary.FindingAid.ShortTitle

	descIndex := NewDescriptionIndex(spec)
	err = descIndex.CreateFrom(desc)
	if err != nil {
		return nil, fmt.Errorf("unable to create DescriptionIndex; %w", err)
	}

	err = descIndex.Write()
	if err != nil {
		return nil, fmt.Errorf("unable to write DescriptionIndex; %w", err)
	}

	err = desc.Write()
	if err != nil {
		return nil, fmt.Errorf("unable to write description; %w", err)
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

	err = ds.Save()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to save dataset")
	}

	err = cead.SaveDescription(cfg, unitInfo, bi)
	if err != nil {
		log.Printf("Unable to save description for %s; %#v", spec, err)
		return nil, errors.Wrapf(err, "Unable to create index representation of the description")
	}

	// write error log
	if len(cfg.Errors) != 0 {
		errs, err := cfg.ErrorToCSV()
		if err != nil {
			log.Printf("unable to get error csv: %#v", err)
			return nil, err
		}

		err = ioutil.WriteFile(
			fmt.Sprintf("%s/errors.csv", basePath),
			errs,
			0644,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to EAD erros to disk")
		}

	}

	if bi != nil {
		start := time.Now()
		err := nl.ESSave(cfg, bi)
		if err != nil {
			log.Printf("Unable to save nodes; %s", err)
		}

		// TODO(kiivihal): decide what to do with drop orphans later
		// _, err = ds.DropOrphans(context.TODO(), bi, nil)
		// if err != nil {
		// log.Printf("Unable to drop orphans; %s", err)
		// }
		end := time.Since(start)
		log.Printf("saving %s with %d records took: %s", spec, cfg.Counter.GetCount(), end)
	}

	return cfg, nil
}

func ProcessUpload(r *http.Request, w http.ResponseWriter, spec string, bi BulkIndex) (uint64, error) {
	in, header, err := r.FormFile("ead")
	if err != nil {
		return uint64(0), err
	}

	defer in.Close()
	defer func() {
		err = r.MultipartForm.RemoveAll()
	}()

	cfg, err := ProcessEAD(in, header.Size, spec, bi)
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

type CLevel interface {
	GetNested() []CLevel
	GetCc() *Cc
}

func (c *Cc) GetXMLName() xml.Name                   { return c.XMLName }
func (c *Cc) GetAttrlevel() string                   { return c.Attrlevel }
func (c *Cc) GetAttrotherlevel() string              { return c.Attrotherlevel }
func (c *Cc) GetAttraltrender() string               { return c.Attraltrender }
func (c *Cc) GetCaccessrestrict() []*Caccessrestrict { return c.Caccessrestrict }
func (c *Cc) GetCdid() *Cdid                         { return c.Cdid[0] }
func (c *Cc) GetScopeContent() []*Cscopecontent      { return c.Cscopecontent }
func (c *Cc) GetOdd() []*Codd                        { return c.Codd }
func (c *Cc) GetPhystech() []*Cphystech              { return c.Cphystech }
func (c *Cc) GetNested() []CLevel {
	nested := []CLevel{}

	for _, v := range c.GetCc().Cc {
		nested = append(nested, v)
	}
	return nested
}
func (c *Cc) GetCc() *Cc { return c }

func (c *Cc) GetGenreform() string {
	if c.Ccontrolaccess != nil && len(c.Ccontrolaccess) != 0 {
		if c.Ccontrolaccess[0].Cgenreform != nil {
			return c.Ccontrolaccess[0].Cgenreform.Genreform
		}
	}

	return ""
}

func (c *Cc) GetMaterial() string {
	if c.Ccontrolaccess != nil {
		for _, ca := range c.Ccontrolaccess {
			if len(ca.Cp) != 0 {
				return ca.Cp[0].P
			}
		}
	}
	cdid := c.GetCdid()

	if cdid.Cphysdesc != nil {
		for _, physdesc := range cdid.Cphysdesc {
			for _, physfacet := range physdesc.Cphysfacet {
				if physfacet.Physfacet != "" {
					return physfacet.Physfacet
				}
			}
		}
	}

	return ""
}

// SaveDescription stores the FragmentGraph of the EAD description in ElasticSearch
func (cead *Cead) SaveDescription(cfg *NodeConfig, unitInfo *UnitInfo, bi BulkIndex) error {
	fg, _, err := cead.DescriptionGraph(cfg, unitInfo)
	if err != nil {
		return err
	}

	if bi == nil {
		return nil
	}

	m, err := fg.IndexMessage()
	if err != nil {
		return fmt.Errorf("unable to marshal fragment graph: %w", err)
	}

	if err := bi.Publish(context.Background(), m); err != nil {
		return err
	}

	atomic.AddUint64(&cfg.RecordsPublishedCounter, 1)

	return nil
}

// RawDescription returns the EAD description stripped of all markup.
func (cead *Cead) RawDescription() []byte {
	description := cead.Ceadheader.Raw
	for _, did := range cead.Carchdesc.Cdid {
		description = append(description, did.Raw...)
	}
	for _, dscGrp := range cead.Carchdesc.Cdescgrp {
		description = append(description, dscGrp.Raw...)
	}
	for _, bioghist := range cead.Carchdesc.Cbioghist {
		description = append(description, bioghist.Raw...)
	}
	for _, userestrict := range cead.Carchdesc.Cuserestrict {
		description = append(description, userestrict.Raw...)
	}

	// strip all tags
	description = space.ReplaceAll(description, []byte(" "))
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

	tree := &fragments.Tree{}

	tree.HubID = header.HubID
	tree.ChildCount = 0
	tree.Type = "desc"
	tree.InventoryID = cfg.Spec
	tree.Title = cead.Ceadheader.GetTitle()
	tree.AgencyCode = cead.Ceadheader.Ceadid.Attrmainagencycode
	tree.Description = []string{string(cead.RawDescription())}
	tree.PeriodDesc = cead.Carchdesc.GetNormalPeriods()

	if len(tree.PeriodDesc) == 0 {
		de := &DuplicateError{
			Spec:  cfg.Spec,
			Error: "ead period is empty",
		}
		cfg.Errors = append(cfg.Errors, de)
	}

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

// GetTitle returns the title of the EAD
func (eh *Ceadheader) GetTitle() string {
	if eh.Cfiledesc != nil && eh.Cfiledesc.Ctitlestmt != nil && eh.Cfiledesc.Ctitlestmt.Ctitleproper != nil {
		return eh.Cfiledesc.Ctitlestmt.Ctitleproper.Titleproper
	}

	return ""
}

// GetOwner returns the owner of the EAD
func (eh *Ceadheader) GetOwner() string {
	if eh.Cfiledesc != nil && eh.Cfiledesc.Cpublicationstmt != nil && eh.Cfiledesc.Cpublicationstmt.Cpublisher != nil {
		return eh.Cfiledesc.Cpublicationstmt.Cpublisher.Publisher
	}
	return ""
}

func (ad *Carchdesc) GetAbstract() []string {
	if len(ad.Cdid) != 0 {
		return ad.Cdid[0].Cabstract.CleanAbstract()
	}

	return []string{}
}

func (ad *Carchdesc) GetPeriods() []string {
	dates := []string{}
	for _, did := range ad.Cdid {
		for _, date := range did.Cunitdate {
			if date.Unitdate != "" {
				dates = append(dates, date.Unitdate)
			}
		}

	}
	return dates
}

func (ad *Carchdesc) GetNormalPeriods() []string {
	dates := []string{}
	for _, did := range ad.Cdid {
		for _, date := range did.Cunitdate {
			if date.Attrnormal != "" && date.Attrtype != "bulk" {
				dates = append(dates, date.Attrnormal)
			}
		}
	}
	return dates
}

// CleanAbstract returns the Abstract split on EAD '<lb />', i.e. line-break
func (ca *Cabstract) CleanAbstract() []string {
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
		if t != "" {
			trimmed = append(trimmed, t)
		}
	}

	return trimmed

}

func (ut *Cunittitle) Title() string {
	return sanitizer.Sanitize(strings.TrimSpace(fmt.Sprintf("%s", ut.Raw)))
}
