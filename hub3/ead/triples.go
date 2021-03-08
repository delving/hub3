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
	"fmt"
	"regexp"
	"strings"

	"github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
)

const (
	eadDomainNS = "https://archief.nl/def/ead"
)

var space = regexp.MustCompile(`\s+`)

func NewResource(label string) r.Term {
	return r.NewResource(fmt.Sprintf("%s/%s", eadDomainNS, label))
}

func (cdid *Cdid) Triples(referrerSubject r.Term) ([]*r.Triple, error) {
	triples := []*r.Triple{}
	s := r.NewResource(referrerSubject.RawValue() + "/did")

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	extract := func(s r.Term, raw []byte) {
		e, _ := NewExtractor(raw)

		for _, token := range e.Tokens() {
			switch token.Type {
			case Person:
				t(s, "persname", token.Text, r.NewLiteral)
			case GeoLocation:
				t(s, "geogname", token.Text, r.NewLiteral)
			case DateText:
				t(s, "datetext", token.Text, r.NewLiteral)
			case DateIso:
				t(s, "dateiso", token.Text, r.NewLiteral)
			}
		}

	}

	str := func(b []byte) string {
		return string(bytes.TrimSpace(space.ReplaceAll(b, []byte(" "))))
	}

	for _, id := range cdid.Cunitid {
		if id.Attraudience != "internal" {
			t(s, "unitID", id.Unitid, r.NewLiteral)
		}
	}

	for _, title := range cdid.Cunittitle {
		t(s, "unitTitle", str(title.Raw), r.NewLiteral)
		extract(s, title.Raw)
	}

	for _, date := range cdid.Cunitdate {
		t(s, "unitDate", date.Unitdate, r.NewLiteral)
	}

	for _, physDesc := range cdid.Cphysdesc {
		for _, extend := range physDesc.Cextent {
			t(s, "physdescExtent", extend.Extent, r.NewLiteral)
		}

		for _, physFacet := range physDesc.Cphysfacet {
			t(s, "physdescPhysfacet", physFacet.Physfacet, r.NewLiteral)
		}

		for _, dimensions := range physDesc.Cdimensions {
			t(s, "physdescDimension", dimensions.Dimensions, r.NewLiteral)
		}

		if physDesc.Physdesc != "" {
			t(s, "physdesc", strings.TrimSpace(physDesc.Physdesc), r.NewLiteral)
		}

	}

	for _, physloc := range cdid.Cphysloc {
		t(s, "physloc", physloc.Physloc, r.NewLiteral)
	}

	for _, materialspec := range cdid.Cmaterialspec {
		t(s, "materialspec", str(materialspec.Raw), r.NewLiteral)
	}

	if cdid.Corigination != nil {
		t(s, "origination", str(cdid.Corigination.Raw), r.NewLiteral)
	}

	if cdid.Cabstract != nil {
		t(s, "abstract", str(cdid.Cabstract.Raw), r.NewLiteral)
	}

	if cdid.Clangmaterial != nil {
		t(s, "langmaterial", str(cdid.Clangmaterial.Raw), r.NewLiteral)
	}

	for _, dao := range cdid.Cdao {
		t(s, "dao", dao.Attrhref, r.NewLiteral)
	}

	return triples, nil
}

func (cc *Cc) Triples(s r.Term) ([]*r.Triple, error) {

	didSubject := r.NewResource(s.RawValue() + "/did")

	triples := []*r.Triple{
		r.NewTriple(
			s,
			NewResource("hasDid"),
			didSubject,
		),
		r.NewTriple(
			didSubject,
			r.NewResource(fragments.RDFType),
			NewResource("Did"),
		),
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			NewResource("Clevel"),
		),
	}

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	extract := func(s r.Term, raw []byte) {
		e, _ := NewExtractor(raw)

		for _, token := range e.Tokens() {
			switch token.Type {
			case Person:
				t(s, "persname", token.Text, r.NewLiteral)
			case GeoLocation:
				t(s, "geogname", token.Text, r.NewLiteral)
			case DateText:
				t(s, "datetext", token.Text, r.NewLiteral)
			case DateIso:
				t(s, "dateiso", token.Text, r.NewLiteral)
			}
		}

	}

	str := func(b []byte) string {
		return string(bytes.TrimSpace(space.ReplaceAll(b, []byte(" "))))
	}

	for _, accessrestrict := range cc.Caccessrestrict {
		t(s, "accessrestrict", str(accessrestrict.Raw), r.NewLiteral)
	}

	for _, controlaccess := range cc.Ccontrolaccess {
		t(s, "controlaccess", str(controlaccess.Raw), r.NewLiteral)
	}

	for _, odd := range cc.Codd {
		t(s, "odd", str(odd.Raw), r.NewLiteral)
		extract(s, odd.Raw)
	}

	for _, scopecontent := range cc.Cscopecontent {
		t(s, "scopecontent", str(scopecontent.Raw), r.NewLiteral)
		extract(s, scopecontent.Raw)
	}

	for _, phystech := range cc.Cphystech {
		t(s, "phystech", str(phystech.Raw), r.NewLiteral)
	}

	for _, custodhist := range cc.Ccustodhist {
		t(s, "custodhist", str(custodhist.Raw), r.NewLiteral)
	}

	for _, altformavail := range cc.Caltformavail {
		t(s, "altformavail", str(altformavail.Raw), r.NewLiteral)
	}

	for _, info := range cc.Cacqinfo {
		t(s, "acqinfo", str(info.Raw), r.NewLiteral)
	}

	for _, userestrict := range cc.Cuserestrict {
		t(s, "userestrict", str(userestrict.Raw), r.NewLiteral)
	}

	for _, accruals := range cc.Caccruals {
		t(s, "accruals", str(accruals.Raw), r.NewLiteral)
	}

	for _, appraisal := range cc.Cappraisal {
		t(s, "appraisal", str(appraisal.Raw), r.NewLiteral)
	}

	for _, bioghist := range cc.Cbioghist {
		t(s, "bioghist", str(bioghist.Raw), r.NewLiteral)
		extract(s, bioghist.Raw)
	}

	for _, relatedmaterial := range cc.Crelatedmaterial {
		t(s, "relatedmaterial", str(relatedmaterial.Raw), r.NewLiteral)
	}

	for _, arrangement := range cc.Carrangement {
		t(s, "arrangement", str(arrangement.Raw), r.NewLiteral)
	}

	for _, separatedmaterial := range cc.Cseparatedmaterial {
		t(s, "separatedmaterial", str(separatedmaterial.Raw), r.NewLiteral)
	}

	for _, processinfo := range cc.Cprocessinfo {
		t(s, "processinfo", str(processinfo.Raw), r.NewLiteral)
	}

	for _, other := range cc.Cotherfindaid {
		t(s, "otherfindaid", str(other.Raw), r.NewLiteral)
	}

	if cc.Coriginalsloc != nil {
		t(s, "originalsloc", str(cc.Coriginalsloc.Raw), r.NewLiteral)
	}

	if cc.Cfileplan != nil {
		t(s, "fileplan", str(cc.Cfileplan.Raw), r.NewLiteral)
	}

	for _, grp := range cc.Cdescgrp {
		t(s, "descgrp", str(grp.Raw), r.NewLiteral)
	}

	return triples, nil
}
