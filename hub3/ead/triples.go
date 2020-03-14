package ead

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
)

const (
	eadDomainNS = "https://archief.nl/def/ead"
)

var space = regexp.MustCompile(`\s+`)

func EADResource(label string) r.Term {
	return r.NewResource(fmt.Sprintf("%s/%s", eadDomainNS, label))
}

func NewSubject(spec, eadType, id string) string {
	identifier := strings.Join([]string{eadType, id}, "/")

	return fmt.Sprintf(
		"%s/%s/archive/%s/%s",
		config.Config.RDF.BaseURL,
		config.Config.OrgID,
		spec,
		identifier,
	)
}

func (did *Cdid) Triples(s r.Term) ([]*r.Triple, error) {
	triples := []*r.Triple{}
	// TODO add type triple

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	str := func(b []byte) string {
		return string(bytes.TrimSpace(space.ReplaceAll(b, []byte(" "))))
	}

	for _, id := range did.Cunitid {
		if id.Attraudience != "internal" {
			t(s, "unitID", id.ID, r.NewLiteral)
		}
	}

	for _, title := range did.Cunittitle {
		t(s, "unitTitle", str(title.RawTitle), r.NewLiteral)
	}

	for _, date := range did.Cunitdate {
		t(s, "unitDate", date.Date, r.NewLiteral)
	}

	if physDesc := did.Cphysdesc; physDesc != nil {
		for _, extend := range physDesc.Cextent {
			t(s, "physdescExtent", extend.Extent, r.NewLiteral)
		}

		for _, dimension := range physDesc.Cdimensions {
			t(s, "physdescDimension", dimension.Dimension, r.NewLiteral)
		}

		if physDesc.Cphysfacet != nil {
			t(s, "physdescPhysfacet", physDesc.Cphysfacet.PhysFacet, r.NewLiteral)
		}

		if physDesc.PhyscDesc != "" {
			t(s, "physdesc", strings.TrimSpace(physDesc.PhyscDesc), r.NewLiteral)
		}
	}

	if did.Cphysloc != nil {
		t(s, "physloc", did.Cphysloc.PhysLoc, r.NewLiteral)
	}

	if did.Cmaterialspec != nil {
		t(s, "materialspec", str(did.Cmaterialspec.Raw), r.NewLiteral)
	}

	if did.Corigination != nil {
		t(s, "origination", str(did.Corigination.Raw), r.NewLiteral)
	}

	if did.Cabstract != nil {
		t(s, "abstract", str(did.Cabstract.Raw), r.NewLiteral)
	}

	if did.Clangmaterial != nil {
		t(s, "langmaterial", str(did.Clangmaterial.Raw), r.NewLiteral)
	}

	if did.Cdao != nil {
		t(s, "dao", did.Cdao.Attrhref, r.NewLiteral)
	}

	return triples, nil
}

func (cc *Cc) Triples(s r.Term) ([]*r.Triple, error) {

	didSubject := r.NewResource(s.RawValue() + "/did")

	triples := []*r.Triple{
		r.NewTriple(
			s,
			EADResource("hasDid"),
			didSubject,
		),
		r.NewTriple(
			didSubject,
			r.NewResource(fragments.RDFType),
			EADResource("Did"),
		),
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			EADResource("Clevel"),
		),
	}

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	str := func(b []byte) string {
		return string(bytes.TrimSpace(space.ReplaceAll(b, []byte(" "))))
	}

	if cc.Caccessrestrict != nil {
		t(s, "accessrestrict", str(cc.Caccessrestrict.Raw), r.NewLiteral)
	}

	if cc.Ccontrolaccess != nil {
		t(s, "controlaccess", str(cc.Ccontrolaccess.Raw), r.NewLiteral)
	}

	for _, odd := range cc.Codd {
		t(s, "odd", str(odd.Raw), r.NewLiteral)
	}

	if cc.Cscopecontent != nil {
		t(s, "scopecontent", str(cc.Cscopecontent.Raw), r.NewLiteral)
	}

	for _, phystech := range cc.Cphystech {
		t(s, "phystech", str(phystech.Raw), r.NewLiteral)
	}

	if cc.Ccustodhist != nil {
		t(s, "custodhist", str(cc.Ccustodhist.Raw), r.NewLiteral)
	}

	if cc.Caltformavail != nil {
		t(s, "altformavail", str(cc.Caltformavail.Raw), r.NewLiteral)
	}

	for _, info := range cc.Cacqinfo {
		t(s, "acqinfo", str(info.Raw), r.NewLiteral)
	}

	if cc.Cuserestrict != nil {
		t(s, "userestrict", str(cc.Cuserestrict.Raw), r.NewLiteral)
	}

	if cc.Caccruals != nil {
		t(s, "accruals", str(cc.Caccruals.Raw), r.NewLiteral)
	}

	if cc.Cappraisal != nil {
		t(s, "appraisal", str(cc.Cappraisal.Raw), r.NewLiteral)
	}

	for _, bioghist := range cc.Cbioghist {
		t(s, "bioghist", str(bioghist.Raw), r.NewLiteral)
	}

	if cc.Crelatedmaterial != nil {
		t(s, "relatedmaterial", str(cc.Crelatedmaterial.Raw), r.NewLiteral)
	}

	if cc.Carrangement != nil {
		t(s, "arrangement", str(cc.Carrangement.Raw), r.NewLiteral)
	}

	for _, separatedmaterial := range cc.Cseparatedmaterial {
		t(s, "separatedmaterial", str(separatedmaterial.Raw), r.NewLiteral)
	}

	if cc.Cprocessinfo != nil {
		t(s, "processinfo", str(cc.Cprocessinfo.Raw), r.NewLiteral)
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
