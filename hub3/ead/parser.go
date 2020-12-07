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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	r "github.com/kiivihal/rdf2go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
)

type BulkIndex interface {
	Publish(ctx context.Context, message ...*domainpb.IndexMessage) error
}

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
func (cead *Cead) SaveDescription(cfg *NodeConfig, unitInfo *UnitInfo) error {
	fg, _, err := cead.DescriptionGraph(cfg, unitInfo)
	if err != nil {
		return err
	}

	return writeResourceFile(cfg, fg, "")
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
		HubID:         fmt.Sprintf("%s_%s_%s", cfg.OrgID, cfg.Spec, id),
		DocType:       fragments.FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		Tags:          []string{"eadDesc"},
	}

	if len(cfg.Tags) != 0 {
		header.Tags = append(header.Tags, cfg.Tags...)
	}

	if tags, ok := c.Config.DatasetTagMap.Get(header.Spec); ok {
		header.Tags = append(header.Tags, tags...)
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

	var idx int
	increment := func() int {
		idx++
		return idx
	}

	s := r.NewResource(subject)
	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			idx := increment()
			err := rm.AppendOrderedTriple(t, false, idx)
			if err != nil {
				log.Error().Err(err).Msg("unable to add triple: %#v")
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
	t(s, "nrClevels", fmt.Sprintf("%d", cfg.Counter.GetCount()), intType)

	if unitInfo != nil {
		t(s, "files", extractDigit(unitInfo.Files), intType)
		t(s, "length", extractDigit(unitInfo.Length), floatType)

		for _, abstract := range unitInfo.Abstract {
			t(s, "abstract", abstract, r.NewLiteral)
		}

		t(s, "material", unitInfo.Material, r.NewLiteral)
		t(s, "language", unitInfo.Language, r.NewLiteral)

		for _, origin := range unitInfo.Origin {
			t(s, "origin", origin, r.NewLiteral)
		}
	}

	// add period desc for range search from the archdesc > did > date
	for _, p := range tree.PeriodDesc {
		t(s, "periodDesc", p, r.NewLiteral)
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

// NewClevel creates a fake c level series struct from the paragraph text.
func (cp *Cp) NewClevel() (*Cc, error) {
	title := strings.Replace(cp.P, ". .", ".", 1)
	odd := ""
	if len(cp.Cextref) > 0 {
		refs := make([]string, 0)
		for _, cex := range cp.Cextref {
			cex.Extref = ""
			x, err := xml.Marshal(cex)
			if err != nil {
				return nil, err
			}
			refs = append(refs, fmt.Sprintf("<p>%s</p>", string(x)))
		}
		odd = fmt.Sprintf("<odd>%s</odd>", strings.Join(refs, ""))
	}
	fakeC := fmt.Sprintf(`<c level="file"><did><unittitle>%s</unittitle></did>%s</c>`, title, odd)
	cc := &Cc{}
	err := xml.Unmarshal([]byte(fakeC), cc)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
