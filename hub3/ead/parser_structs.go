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

import "encoding/xml"

type Cabstract struct {
	XMLName   xml.Name   `xml:"abstract,omitempty" json:"abstract,omitempty"`
	Raw       []byte     `xml:",innerxml" json:",omitempty"`
	Attrlabel string     `xml:"label,attr"  json:",omitempty"`
	Abstract  string     `xml:",chardata" json:",omitempty"`
	Cextref   []*Cextref `xml:"extref,omitempty" json:"extref,omitempty"`
	Clb       []*Clb     `xml:"lb,omitempty" json:"lb,omitempty"`
}

type Caccessrestrict struct {
	XMLName         xml.Name           `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Raw             []byte             `xml:",innerxml" json:",omitempty"`
	Attrid          string             `xml:"id,attr"  json:",omitempty"`
	Attrtype        string             `xml:"type,attr"  json:",omitempty"`
	Caccessrestrict []*Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Chead           []*Chead           `xml:"head,omitempty" json:"head,omitempty"`
	Clegalstatus    *Clegalstatus      `xml:"legalstatus,omitempty" json:"legalstatus,omitempty"`
	Clist           []*Clist           `xml:"list,omitempty" json:"list,omitempty"`
	Cp              []*Cp              `xml:"p,omitempty" json:"p,omitempty"`
}

type Caccruals struct {
	XMLName xml.Name  `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Raw     []byte    `xml:",innerxml" json:",omitempty"`
	Attrid  string    `xml:"id,attr"  json:",omitempty"`
	Chead   []*Chead  `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp     `xml:"p,omitempty" json:"p,omitempty"`
	Ctable  []*Ctable `xml:"table,omitempty" json:"table,omitempty"`
}

type Cacqinfo struct {
	XMLName       xml.Name    `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Raw           []byte      `xml:",innerxml" json:",omitempty"`
	Attraltrender string      `xml:"altrender,attr"  json:",omitempty"`
	Cacqinfo      []*Cacqinfo `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Chead         []*Chead    `xml:"head,omitempty" json:"head,omitempty"`
	Clist         []*Clist    `xml:"list,omitempty" json:"list,omitempty"`
	Cp            []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
	Ctable        []*Ctable   `xml:"table,omitempty" json:"table,omitempty"`
}

type Caltformavail struct {
	XMLName  xml.Name `xml:"altformavail,omitempty" json:"altformavail,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Clist    []*Clist `xml:"list,omitempty" json:"list,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cappraisal struct {
	XMLName      xml.Name      `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Raw          []byte        `xml:",innerxml" json:",omitempty"`
	Attraudience string        `xml:"audience,attr"  json:",omitempty"`
	Cappraisal   []*Cappraisal `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Chead        []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clist        []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Cp           []*Cp         `xml:"p,omitempty" json:"p,omitempty"`
	Ctable       []*Ctable     `xml:"table,omitempty" json:"table,omitempty"`
}

type Carchdesc struct {
	XMLName       xml.Name         `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
	Raw           []byte           `xml:",innerxml" json:",omitempty"`
	Attrlevel     string           `xml:"level,attr"  json:",omitempty"`
	Attrtype      string           `xml:"type,attr"  json:",omitempty"`
	Caccruals     []*Caccruals     `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Cbibliography []*Cbibliography `xml:"bibliography,omitempty" json:"bibliography,omitempty"`
	Cbioghist     []*Cbioghist     `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Cdescgrp      []*Cdescgrp      `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
	Cdid          []*Cdid          `xml:"did,omitempty" json:"did,omitempty"`
	Cdsc          *Cdsc            `xml:"dsc,omitempty" json:"dsc,omitempty"`
	Cuserestrict  []*Cuserestrict  `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
}

type Carchref struct {
	XMLName      xml.Name      `xml:"archref,omitempty" json:"archref,omitempty"`
	Raw          []byte        `xml:",innerxml" json:",omitempty"`
	Attractuate  string        `xml:"actuate,attr"  json:",omitempty"`
	Attrhref     string        `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string        `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string        `xml:"show,attr"  json:",omitempty"`
	Archref      string        `xml:",chardata" json:",omitempty"`
	Clb          []*Clb        `xml:"lb,omitempty" json:"lb,omitempty"`
	Ctitle       []*Ctitle     `xml:"title,omitempty" json:"title,omitempty"`
	Cunitid      []*Cunitid    `xml:"unitid,omitempty" json:"unitid,omitempty"`
	Cunittitle   []*Cunittitle `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
}

type Carrangement struct {
	XMLName      xml.Name        `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Raw          []byte          `xml:",innerxml" json:",omitempty"`
	Carrangement []*Carrangement `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Chead        []*Chead        `xml:"head,omitempty" json:"head,omitempty"`
	Clist        []*Clist        `xml:"list,omitempty" json:"list,omitempty"`
	Cp           []*Cp           `xml:"p,omitempty" json:"p,omitempty"`
	Ctable       []*Ctable       `xml:"table,omitempty" json:"table,omitempty"`
}

type Cauthor struct {
	XMLName xml.Name `xml:"author,omitempty" json:"author,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Author  string   `xml:",chardata" json:",omitempty"`
	Clb     []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
}

type Cbibliography struct {
	XMLName       xml.Name         `xml:"bibliography,omitempty" json:"bibliography,omitempty"`
	Raw           []byte           `xml:",innerxml" json:",omitempty"`
	Cbibliography []*Cbibliography `xml:"bibliography,omitempty" json:"bibliography,omitempty"`
	Chead         []*Chead         `xml:"head,omitempty" json:"head,omitempty"`
	Clist         []*Clist         `xml:"list,omitempty" json:"list,omitempty"`
	Cp            []*Cp            `xml:"p,omitempty" json:"p,omitempty"`
	Ctable        []*Ctable        `xml:"table,omitempty" json:"table,omitempty"`
}

type Cbibref struct {
	XMLName      xml.Name     `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Raw          []byte       `xml:",innerxml" json:",omitempty"`
	Attractuate  string       `xml:"actuate,attr"  json:",omitempty"`
	Attrhref     string       `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string       `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string       `xml:"show,attr"  json:",omitempty"`
	Bibref       string       `xml:",chardata" json:",omitempty"`
	Cextref      []*Cextref   `xml:"extref,omitempty" json:"extref,omitempty"`
	Cimprint     []*Cimprint  `xml:"imprint,omitempty" json:"imprint,omitempty"`
	Clb          []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Cname        []*Cname     `xml:"name,omitempty" json:"name,omitempty"`
	Cpersname    []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Ctitle       []*Ctitle    `xml:"title,omitempty" json:"title,omitempty"`
}

type Cbioghist struct {
	XMLName    xml.Name      `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Raw        []byte        `xml:",innerxml" json:",omitempty"`
	Cbioghist  []*Cbioghist  `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Cchronlist []*Cchronlist `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Chead      []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clist      []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Codd       []*Codd       `xml:"odd,omitempty" json:"odd,omitempty"`
	Cp         []*Cp         `xml:"p,omitempty" json:"p,omitempty"`
	Ctable     []*Ctable     `xml:"table,omitempty" json:"table,omitempty"`
}

type Cblockquote struct {
	XMLName xml.Name `xml:"blockquote,omitempty" json:"blockquote,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Cnote   []*Cnote `xml:"note,omitempty" json:"note,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cc struct {
	XMLName            xml.Name              `xml:"c,omitempty" json:"c,omitempty"`
	Raw                []byte                `xml:",innerxml" json:",omitempty"`
	Attraltrender      string                `xml:"altrender,attr"  json:",omitempty"`
	Attrlevel          string                `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel     string                `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict    []*Caccessrestrict    `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Caccruals          []*Caccruals          `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Cacqinfo           []*Cacqinfo           `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Caltformavail      []*Caltformavail      `xml:"altformavail,omitempty" json:"altformavail,omitempty"`
	Cappraisal         []*Cappraisal         `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Carrangement       []*Carrangement       `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Cbibliography      []*Cbibliography      `xml:"bibliography,omitempty" json:"bibliography,omitempty"`
	Cbioghist          []*Cbioghist          `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Ccontrolaccess     []*Ccontrolaccess     `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Ccustodhist        []*Ccustodhist        `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Cdao               []*Cdao               `xml:"dao,omitempty" json:"dao,omitempty"`
	Cdid               []*Cdid               `xml:"did,omitempty" json:"did,omitempty"`
	Codd               []*Codd               `xml:"odd,omitempty" json:"odd,omitempty"`
	Coriginalsloc      *Coriginalsloc        `xml:"originalsloc,omitempty" json:"originalsloc,omitempty"`
	Cotherfindaid      []*Cotherfindaid      `xml:"otherfindaid,omitempty" json:"otherfindaid,omitempty"`
	Cphystech          []*Cphystech          `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Cprocessinfo       []*Cprocessinfo       `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Crelatedmaterial   []*Crelatedmaterial   `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Cscopecontent      []*Cscopecontent      `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Cseparatedmaterial []*Cseparatedmaterial `xml:"separatedmaterial,omitempty" json:"separatedmaterial,omitempty"`
	Cuserestrict       []*Cuserestrict       `xml:"userestrict,omitempty" json:"userestrict,omitempty"`

	Cc []*Cc `xml:"c,omitempty"`
	// not supported by data
	Cfileplan *Cfileplan  `xml:"fileplan,omitempty" json:"fileplan,omitempty"`
	Cdescgrp  []*Cdescgrp `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
}

type Cchange struct {
	XMLName xml.Name `xml:"change,omitempty" json:"change,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Cdate   []*Cdate `xml:"date,omitempty" json:"date,omitempty"`
	Citem   []*Citem `xml:"item,omitempty" json:"item,omitempty"`
}

type Cchronitem struct {
	XMLName   xml.Name   `xml:"chronitem,omitempty" json:"chronitem,omitempty"`
	Raw       []byte     `xml:",innerxml" json:",omitempty"`
	Cdate     []*Cdate   `xml:"date,omitempty" json:"date,omitempty"`
	Cevent    []*Cevent  `xml:"event,omitempty" json:"event,omitempty"`
	Ceventgrp *Ceventgrp `xml:"eventgrp,omitempty" json:"eventgrp,omitempty"`
}

type Cchronlist struct {
	XMLName    xml.Name      `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Raw        []byte        `xml:",innerxml" json:",omitempty"`
	Cchronitem []*Cchronitem `xml:"chronitem,omitempty" json:"chronitem,omitempty"`
	Chead      []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clisthead  *Clisthead    `xml:"listhead,omitempty" json:"listhead,omitempty"`
}

type Ccolspec struct {
	XMLName      xml.Name `xml:"colspec,omitempty" json:"colspec,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attralign    string   `xml:"align,attr"  json:",omitempty"`
	Attrcolname  string   `xml:"colname,attr"  json:",omitempty"`
	Attrcolnum   string   `xml:"colnum,attr"  json:",omitempty"`
	Attrcolsep   string   `xml:"colsep,attr"  json:",omitempty"`
	Attrcolwidth string   `xml:"colwidth,attr"  json:",omitempty"`
}

type Ccontrolaccess struct {
	XMLName      xml.Name    `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Raw          []byte      `xml:",innerxml" json:",omitempty"`
	Attraudience string      `xml:"audience,attr"  json:",omitempty"`
	Cgenreform   *Cgenreform `xml:"genreform,omitempty" json:"genreform,omitempty"`
	Cnote        []*Cnote    `xml:"note,omitempty" json:"note,omitempty"`
	Cp           []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
	Csubject     *Csubject   `xml:"subject,omitempty" json:"subject,omitempty"`
}

type Ccorpname struct {
	XMLName            xml.Name `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attrauthfilenumber string   `xml:"authfilenumber,attr"  json:",omitempty"`
	Attrencodinganalog string   `xml:"encodinganalog,attr"  json:",omitempty"`
	Attrnormal         string   `xml:"normal,attr"  json:",omitempty"`
	Attrrole           string   `xml:"role,attr"  json:",omitempty"`
	Attrsource         string   `xml:"source,attr"  json:",omitempty"`
	Corpname           string   `xml:",chardata" json:",omitempty"`
}

type Ccreation struct {
	XMLName      xml.Name  `xml:"creation,omitempty" json:"creation,omitempty"`
	Raw          []byte    `xml:",innerxml" json:",omitempty"`
	Attraudience string    `xml:"audience,attr"  json:",omitempty"`
	Cdate        []*Cdate  `xml:"date,omitempty" json:"date,omitempty"`
	Clb          []*Clb    `xml:"lb,omitempty" json:"lb,omitempty"`
	Creation     string    `xml:",chardata" json:",omitempty"`
	Ctitle       []*Ctitle `xml:"title,omitempty" json:"title,omitempty"`
}

type Ccustodhist struct {
	XMLName     xml.Name       `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Raw         []byte         `xml:",innerxml" json:",omitempty"`
	Cacqinfo    []*Cacqinfo    `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Ccustodhist []*Ccustodhist `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Chead       []*Chead       `xml:"head,omitempty" json:"head,omitempty"`
	Clist       []*Clist       `xml:"list,omitempty" json:"list,omitempty"`
	Codd        []*Codd        `xml:"odd,omitempty" json:"odd,omitempty"`
	Cp          []*Cp          `xml:"p,omitempty" json:"p,omitempty"`
	Ctable      []*Ctable      `xml:"table,omitempty" json:"table,omitempty"`
}

type Cdao struct {
	XMLName      xml.Name `xml:"dao,omitempty" json:"dao,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attractuate  string   `xml:"actuate,attr"  json:",omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Attrhref     string   `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrrole     string   `xml:"role,attr"  json:",omitempty"`
	Attrshow     string   `xml:"show,attr"  json:",omitempty"`
}

type Cdate struct {
	XMLName            xml.Name `xml:"date,omitempty" json:"date,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attrcalendar       string   `xml:"calendar,attr"  json:",omitempty"`
	Attrencodinganalog string   `xml:"encodinganalog,attr"  json:",omitempty"`
	Attrera            string   `xml:"era,attr"  json:",omitempty"`
	Attrnormal         string   `xml:"normal,attr"  json:",omitempty"`
	Attrtype           string   `xml:"type,attr"  json:",omitempty"`
	Date               string   `xml:",chardata" json:",omitempty"`
}

type Cdefitem struct {
	XMLName xml.Name `xml:"defitem,omitempty" json:"defitem,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Citem   []*Citem `xml:"item,omitempty" json:"item,omitempty"`
	Clabel  *Clabel  `xml:"label,omitempty" json:"label,omitempty"`
}

type Cdescgrp struct {
	XMLName            xml.Name              `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
	Raw                []byte                `xml:",innerxml" json:",omitempty"`
	Attrtype           string                `xml:"type,attr"  json:",omitempty"`
	Caccessrestrict    []*Caccessrestrict    `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Caccruals          []*Caccruals          `xml:"accruals,omitempty" json:"accruals,omitempty"`
	Cacqinfo           []*Cacqinfo           `xml:"acqinfo,omitempty" json:"acqinfo,omitempty"`
	Caltformavail      []*Caltformavail      `xml:"altformavail,omitempty" json:"altformavail,omitempty"`
	Cappraisal         []*Cappraisal         `xml:"appraisal,omitempty" json:"appraisal,omitempty"`
	Carrangement       []*Carrangement       `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Cbibliography      []*Cbibliography      `xml:"bibliography,omitempty" json:"bibliography,omitempty"`
	Cbioghist          []*Cbioghist          `xml:"bioghist,omitempty" json:"bioghist,omitempty"`
	Ccontrolaccess     []*Ccontrolaccess     `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
	Ccustodhist        []*Ccustodhist        `xml:"custodhist,omitempty" json:"custodhist,omitempty"`
	Cfileplan          *Cfileplan            `xml:"fileplan,omitempty" json:"fileplan,omitempty"`
	Chead              []*Chead              `xml:"head,omitempty" json:"head,omitempty"`
	Cindex             []*Cindex             `xml:"index,omitempty" json:"index,omitempty"`
	Clist              []*Clist              `xml:"list,omitempty" json:"list,omitempty"`
	Codd               []*Codd               `xml:"odd,omitempty" json:"odd,omitempty"`
	Coriginalsloc      *Coriginalsloc        `xml:"originalsloc,omitempty" json:"originalsloc,omitempty"`
	Cotherfindaid      []*Cotherfindaid      `xml:"otherfindaid,omitempty" json:"otherfindaid,omitempty"`
	Cp                 []*Cp                 `xml:"p,omitempty" json:"p,omitempty"`
	Cphystech          []*Cphystech          `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Cprefercite        *Cprefercite          `xml:"prefercite,omitempty" json:"prefercite,omitempty"`
	Cprocessinfo       []*Cprocessinfo       `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Crelatedmaterial   []*Crelatedmaterial   `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Cscopecontent      []*Cscopecontent      `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Cseparatedmaterial []*Cseparatedmaterial `xml:"separatedmaterial,omitempty" json:"separatedmaterial,omitempty"`
	Cuserestrict       []*Cuserestrict       `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
}

type Cdescrules struct {
	XMLName      xml.Name   `xml:"descrules,omitempty" json:"descrules,omitempty"`
	Raw          []byte     `xml:",innerxml" json:",omitempty"`
	Attraudience string     `xml:"audience,attr"  json:",omitempty"`
	Cbibref      []*Cbibref `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Clb          []*Clb     `xml:"lb,omitempty" json:"lb,omitempty"`
	Ctitle       []*Ctitle  `xml:"title,omitempty" json:"title,omitempty"`
	Descrules    string     `xml:",chardata" json:",omitempty"`
}

type Cdid struct {
	XMLName       xml.Name         `xml:"did,omitempty" json:"did,omitempty"`
	Raw           []byte           `xml:",innerxml" json:",omitempty"`
	Attrid        string           `xml:"id,attr"  json:",omitempty"`
	Cabstract     *Cabstract       `xml:"abstract,omitempty" json:"abstract,omitempty"`
	Cdao          []*Cdao          `xml:"dao,omitempty" json:"dao,omitempty"`
	Chead         []*Chead         `xml:"head,omitempty" json:"head,omitempty"`
	Clangmaterial *Clangmaterial   `xml:"langmaterial,omitempty" json:"langmaterial,omitempty"`
	Cmaterialspec []*Cmaterialspec `xml:"materialspec,omitempty" json:"materialspec,omitempty"`
	Corigination  *Corigination    `xml:"origination,omitempty" json:"origination,omitempty"`
	Cphysdesc     []*Cphysdesc     `xml:"physdesc,omitempty" json:"physdesc,omitempty"`
	Cphysloc      []*Cphysloc      `xml:"physloc,omitempty" json:"physloc,omitempty"`
	Crepository   *Crepository     `xml:"repository,omitempty" json:"repository,omitempty"`
	Cunitdate     []*Cunitdate     `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Cunitid       []*Cunitid       `xml:"unitid,omitempty" json:"unitid,omitempty"`
	Cunittitle    []*Cunittitle    `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
}

type Cdimensions struct {
	XMLName    xml.Name `xml:"dimensions,omitempty" json:"dimensions,omitempty"`
	Raw        []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype   string   `xml:"type,attr"  json:",omitempty"`
	Dimensions string   `xml:",chardata" json:",omitempty"`
}

type Cdsc struct {
	XMLName  xml.Name `xml:"dsc,omitempty" json:"dsc,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Cc       []*Cc    `xml:"c,omitempty" json:"c,omitempty"`
	Numbered []*Cc01  `xml:"c01,omitempty" json:"c,omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
	// Nested   []*Cc01  `xml:"c01,omitempty" json:"c01,omitempty"` // todo: support numbered??
}

// do not include Raw
type Cead struct {
	XMLName                    xml.Name    `xml:"ead,omitempty" json:"ead,omitempty"`
	Attraudience               string      `xml:"audience,attr,omitempty"  json:",omitempty"`
	AttrXsiSpaceschemaLocation string      `xml:"xsi:schemaLocation,attr,omitempty"  json:",omitempty"`
	AttrXmlnsxlink             string      `xml:"xmlns:xlink,attr,omitempty"  json:",omitempty"`
	Attrxmlns                  string      `xml:"xmlns,attr,omitempty"  json:",omitempty"`
	AttrXmlnsxsi               string      `xml:"xmlns:xsi,attr,omitempty"  json:",omitempty"`
	Ceadheader                 *Ceadheader `xml:"eadheader,omitempty" json:"eadheader,omitempty"`
	Carchdesc                  *Carchdesc  `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
}

type Ceadheader struct {
	XMLName                xml.Name       `xml:"eadheader,omitempty" json:"eadheader,omitempty"`
	Raw                    []byte         `xml:",innerxml" json:",omitempty"`
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
}

type Ceadid struct {
	XMLName            xml.Name `xml:"eadid,omitempty" json:"eadid,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attrcountrycode    string   `xml:"countrycode,attr"  json:",omitempty"`
	Attrmainagencycode string   `xml:"mainagencycode,attr"  json:",omitempty"`
	Attrpublicid       string   `xml:"publicid,attr"  json:",omitempty"`
	Attrurl            string   `xml:"url,attr"  json:",omitempty"`
	Attrurn            string   `xml:"urn,attr"  json:",omitempty"`
	EadID              string   `xml:",chardata" json:",omitempty"`
}

type Cedition struct {
	XMLName xml.Name `xml:"edition,omitempty" json:"edition,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Edition string   `xml:",chardata" json:",omitempty"`
}

type Ceditionstmt struct {
	XMLName  xml.Name  `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Raw      []byte    `xml:",innerxml" json:",omitempty"`
	Cedition *Cedition `xml:"edition,omitempty" json:"edition,omitempty"`
}

type Cemph struct {
	XMLName       xml.Name `xml:"emph,omitempty" json:"emph,omitempty"`
	Raw           []byte   `xml:",innerxml" json:",omitempty"`
	Attraltrender string   `xml:"altrender,attr"  json:",omitempty"`
	Attrrender    string   `xml:"render,attr"  json:",omitempty"`
	Emph          string   `xml:",chardata" json:",omitempty"`
}

type Centry struct {
	XMLName     xml.Name      `xml:"entry,omitempty" json:"entry,omitempty"`
	Raw         []byte        `xml:",innerxml" json:",omitempty"`
	Attralign   string        `xml:"align,attr"  json:",omitempty"`
	Attrchar    string        `xml:"char,attr"  json:",omitempty"`
	Attrcharoff string        `xml:"charoff,attr"  json:",omitempty"`
	Attrcolname string        `xml:"colname,attr"  json:",omitempty"`
	Attrcolsep  string        `xml:"colsep,attr"  json:",omitempty"`
	Attrnameend string        `xml:"nameend,attr"  json:",omitempty"`
	Attrnamest  string        `xml:"namest,attr"  json:",omitempty"`
	Attrrowsep  string        `xml:"rowsep,attr"  json:",omitempty"`
	Attrvalign  string        `xml:"valign,attr"  json:",omitempty"`
	Carchref    []*Carchref   `xml:"archref,omitempty" json:"archref,omitempty"`
	Cbibref     []*Cbibref    `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Ccorpname   []*Ccorpname  `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cdate       []*Cdate      `xml:"date,omitempty" json:"date,omitempty"`
	Cemph       []*Cemph      `xml:"emph,omitempty" json:"emph,omitempty"`
	Cfamname    []*Cfamname   `xml:"famname,omitempty" json:"famname,omitempty"`
	Cfunction   *Cfunction    `xml:"function,omitempty" json:"function,omitempty"`
	Cgeogname   []*Cgeogname  `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Clb         []*Clb        `xml:"lb,omitempty" json:"lb,omitempty"`
	Clist       []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Cname       []*Cname      `xml:"name,omitempty" json:"name,omitempty"`
	Cnote       []*Cnote      `xml:"note,omitempty" json:"note,omitempty"`
	Cpersname   []*Cpersname  `xml:"persname,omitempty" json:"persname,omitempty"`
	Cref        []*Cref       `xml:"ref,omitempty" json:"ref,omitempty"`
	Csubject    *Csubject     `xml:"subject,omitempty" json:"subject,omitempty"`
	Ctitle      []*Ctitle     `xml:"title,omitempty" json:"title,omitempty"`
	Cunitdate   []*Cunitdate  `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Cunittitle  []*Cunittitle `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
	Entry       string        `xml:",chardata" json:",omitempty"`
}

type Cevent struct {
	XMLName   xml.Name     `xml:"event,omitempty" json:"event,omitempty"`
	Raw       []byte       `xml:",innerxml" json:",omitempty"`
	Clb       []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Clist     []*Clist     `xml:"list,omitempty" json:"list,omitempty"`
	Cnote     []*Cnote     `xml:"note,omitempty" json:"note,omitempty"`
	Cpersname []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Event     string       `xml:",chardata" json:",omitempty"`
}

type Ceventgrp struct {
	XMLName xml.Name  `xml:"eventgrp,omitempty" json:"eventgrp,omitempty"`
	Raw     []byte    `xml:",innerxml" json:",omitempty"`
	Cevent  []*Cevent `xml:"event,omitempty" json:"event,omitempty"`
}

type Cextent struct {
	XMLName  xml.Name `xml:"extent,omitempty" json:"extent,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrunit string   `xml:"unit,attr"  json:",omitempty"`
	Extent   string   `xml:",chardata" json:",omitempty"`
}

type Cextptr struct {
	XMLName      xml.Name `xml:"extptr,omitempty" json:"extptr,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attractuate  string   `xml:"actuate,attr"  json:",omitempty"`
	Attrhref     string   `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string   `xml:"show,attr"  json:",omitempty"`
}

type Cextref struct {
	XMLName      xml.Name  `xml:"extref,omitempty" json:"extref,omitempty"`
	Raw          []byte    `xml:",innerxml" json:",omitempty"`
	Attractuate  string    `xml:"actuate,attr"  json:",omitempty"`
	Attrhref     string    `xml:"href,attr"  json:",omitempty"`
	Attrlinktype string    `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string    `xml:"show,attr"  json:",omitempty"`
	Ctitle       []*Ctitle `xml:"title,omitempty" json:"title,omitempty"`
	Extref       string    `xml:",chardata" json:",omitempty"`
}

type Cfamname struct {
	XMLName    xml.Name `xml:"famname,omitempty" json:"famname,omitempty"`
	Raw        []byte   `xml:",innerxml" json:",omitempty"`
	Attrnormal string   `xml:"normal,attr"  json:",omitempty"`
	Famname    string   `xml:",chardata" json:",omitempty"`
}

type Cfiledesc struct {
	XMLName          xml.Name          `xml:"filedesc,omitempty" json:"filedesc,omitempty"`
	Raw              []byte            `xml:",innerxml" json:",omitempty"`
	Ceditionstmt     *Ceditionstmt     `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Cnotestmt        *Cnotestmt        `xml:"notestmt,omitempty" json:"notestmt,omitempty"`
	Cpublicationstmt *Cpublicationstmt `xml:"publicationstmt,omitempty" json:"publicationstmt,omitempty"`
	Ctitlestmt       *Ctitlestmt       `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
}

type Cfileplan struct {
	XMLName   xml.Name   `xml:"fileplan,omitempty" json:"fileplan,omitempty"`
	Raw       []byte     `xml:",innerxml" json:",omitempty"`
	Cfileplan *Cfileplan `xml:"fileplan,omitempty" json:"fileplan,omitempty"`
	Chead     []*Chead   `xml:"head,omitempty" json:"head,omitempty"`
	Clist     []*Clist   `xml:"list,omitempty" json:"list,omitempty"`
	Cp        []*Cp      `xml:"p,omitempty" json:"p,omitempty"`
	Ctable    []*Ctable  `xml:"table,omitempty" json:"table,omitempty"`
}

type Cfunction struct {
	XMLName  xml.Name `xml:"function,omitempty" json:"function,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Function string   `xml:",chardata" json:",omitempty"`
}

type Cgenreform struct {
	XMLName   xml.Name `xml:"genreform,omitempty" json:"genreform,omitempty"`
	Raw       []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype  string   `xml:"type,attr"  json:",omitempty"`
	Genreform string   `xml:",chardata" json:",omitempty"`
}

type Cgeogname struct {
	XMLName            xml.Name `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attrencodinganalog string   `xml:"encodinganalog,attr"  json:",omitempty"`
	Attrnormal         string   `xml:"normal,attr"  json:",omitempty"`
	Geogname           string   `xml:",chardata" json:",omitempty"`
}

type Chead struct {
	XMLName xml.Name `xml:"head,omitempty" json:"head,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Attrid  string   `xml:"id,attr"  json:",omitempty"`
	Clb     []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Head    string   `xml:",chardata" json:",omitempty"`
}

type Chead01 struct {
	XMLName xml.Name `xml:"head01,omitempty" json:"head01,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Head01  string   `xml:",chardata" json:",omitempty"`
}

type Chead02 struct {
	XMLName xml.Name `xml:"head02,omitempty" json:"head02,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Head02  string   `xml:",chardata" json:",omitempty"`
}

type Cimprint struct {
	XMLName    xml.Name     `xml:"imprint,omitempty" json:"imprint,omitempty"`
	Raw        []byte       `xml:",innerxml" json:",omitempty"`
	Cdate      []*Cdate     `xml:"date,omitempty" json:"date,omitempty"`
	Cgeogname  []*Cgeogname `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Cpublisher *Cpublisher  `xml:"publisher,omitempty" json:"publisher,omitempty"`
	Imprint    string       `xml:",chardata" json:",omitempty"`
}

type Cindex struct {
	XMLName     xml.Name       `xml:"index,omitempty" json:"index,omitempty"`
	Raw         []byte         `xml:",innerxml" json:",omitempty"`
	Chead       []*Chead       `xml:"head,omitempty" json:"head,omitempty"`
	Cindexentry []*Cindexentry `xml:"indexentry,omitempty" json:"indexentry,omitempty"`
	Clisthead   *Clisthead     `xml:"listhead,omitempty" json:"listhead,omitempty"`
	Cp          []*Cp          `xml:"p,omitempty" json:"p,omitempty"`
}

type Cindexentry struct {
	XMLName   xml.Name     `xml:"indexentry,omitempty" json:"indexentry,omitempty"`
	Raw       []byte       `xml:",innerxml" json:",omitempty"`
	Ccorpname []*Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cgeogname []*Cgeogname `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Cname     []*Cname     `xml:"name,omitempty" json:"name,omitempty"`
	Cpersname []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Cptrgrp   *Cptrgrp     `xml:"ptrgrp,omitempty" json:"ptrgrp,omitempty"`
	Cref      []*Cref      `xml:"ref,omitempty" json:"ref,omitempty"`
	Csubject  *Csubject    `xml:"subject,omitempty" json:"subject,omitempty"`
	Ctitle    []*Ctitle    `xml:"title,omitempty" json:"title,omitempty"`
}

type Citem struct {
	XMLName    xml.Name      `xml:"item,omitempty" json:"item,omitempty"`
	Raw        []byte        `xml:",innerxml" json:",omitempty"`
	Carchref   []*Carchref   `xml:"archref,omitempty" json:"archref,omitempty"`
	Cbibref    []*Cbibref    `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Cchronlist []*Cchronlist `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Ccorpname  []*Ccorpname  `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cdate      []*Cdate      `xml:"date,omitempty" json:"date,omitempty"`
	Cemph      []*Cemph      `xml:"emph,omitempty" json:"emph,omitempty"`
	Cextref    []*Cextref    `xml:"extref,omitempty" json:"extref,omitempty"`
	Cgeogname  []*Cgeogname  `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Citem      []*Citem      `xml:"item,omitempty" json:"item,omitempty"`
	Clb        []*Clb        `xml:"lb,omitempty" json:"lb,omitempty"`
	Clist      []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Cname      []*Cname      `xml:"name,omitempty" json:"name,omitempty"`
	Cnote      []*Cnote      `xml:"note,omitempty" json:"note,omitempty"`
	Cnum       []*Cnum       `xml:"num,omitempty" json:"num,omitempty"`
	Cpersname  []*Cpersname  `xml:"persname,omitempty" json:"persname,omitempty"`
	Cref       []*Cref       `xml:"ref,omitempty" json:"ref,omitempty"`
	Csubject   *Csubject     `xml:"subject,omitempty" json:"subject,omitempty"`
	Ctitle     []*Ctitle     `xml:"title,omitempty" json:"title,omitempty"`
	Cunitdate  []*Cunitdate  `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Cunittitle []*Cunittitle `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
	Item       string        `xml:",chardata" json:",omitempty"`
}

type Clabel struct {
	XMLName xml.Name `xml:"label,omitempty" json:"label,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Label   string   `xml:",chardata" json:",omitempty"`
}

type Clangmaterial struct {
	XMLName      xml.Name     `xml:"langmaterial,omitempty" json:"langmaterial,omitempty"`
	Raw          []byte       `xml:",innerxml" json:",omitempty"`
	Attrlabel    string       `xml:"label,attr"  json:",omitempty"`
	Clanguage    []*Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	Clb          []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Langmaterial string       `xml:",chardata" json:",omitempty"`
}

type Clanguage struct {
	XMLName        xml.Name `xml:"language,omitempty" json:"language,omitempty"`
	Raw            []byte   `xml:",innerxml" json:",omitempty"`
	Attrlangcode   string   `xml:"langcode,attr"  json:",omitempty"`
	Attrscriptcode string   `xml:"scriptcode,attr"  json:",omitempty"`
	Language       string   `xml:",chardata" json:",omitempty"`
}

type Clangusage struct {
	XMLName   xml.Name     `xml:"langusage,omitempty" json:"langusage,omitempty"`
	Raw       []byte       `xml:",innerxml" json:",omitempty"`
	Clanguage []*Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	Langusage string       `xml:",chardata" json:",omitempty"`
}

type Clb struct {
	XMLName xml.Name `xml:"lb,omitempty" json:"lb,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
}

type Clegalstatus struct {
	XMLName     xml.Name `xml:"legalstatus,omitempty" json:"legalstatus,omitempty"`
	Raw         []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype    string   `xml:"type,attr"  json:",omitempty"`
	Legalstatus string   `xml:",chardata" json:",omitempty"`
}

type Clist struct {
	XMLName          xml.Name    `xml:"list,omitempty" json:"list,omitempty"`
	Raw              []byte      `xml:",innerxml" json:",omitempty"`
	Attrcontinuation string      `xml:"continuation,attr"  json:",omitempty"`
	Attrmark         string      `xml:"mark,attr"  json:",omitempty"`
	Attrnumeration   string      `xml:"numeration,attr"  json:",omitempty"`
	Attrtype         string      `xml:"type,attr"  json:",omitempty"`
	Cdefitem         []*Cdefitem `xml:"defitem,omitempty" json:"defitem,omitempty"`
	Chead            []*Chead    `xml:"head,omitempty" json:"head,omitempty"`
	Citem            []*Citem    `xml:"item,omitempty" json:"item,omitempty"`
	Clisthead        *Clisthead  `xml:"listhead,omitempty" json:"listhead,omitempty"`
}

type Clisthead struct {
	XMLName xml.Name `xml:"listhead,omitempty" json:"listhead,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Chead01 *Chead01 `xml:"head01,omitempty" json:"head01,omitempty"`
	Chead02 *Chead02 `xml:"head02,omitempty" json:"head02,omitempty"`
}

type Cmaterialspec struct {
	XMLName      xml.Name `xml:"materialspec,omitempty" json:"materialspec,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attrlabel    string   `xml:"label,attr"  json:",omitempty"`
	Attrtype     string   `xml:"type,attr"  json:",omitempty"`
	Clb          []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Materialspec string   `xml:",chardata" json:",omitempty"`
}

type Cname struct {
	XMLName  xml.Name `xml:"name,omitempty" json:"name,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrrole string   `xml:"role,attr"  json:",omitempty"`
	Name     string   `xml:",chardata" json:",omitempty"`
}

type Cnote struct {
	XMLName  xml.Name  `xml:"note,omitempty" json:"note,omitempty"`
	Raw      []byte    `xml:",innerxml" json:",omitempty"`
	Attrtype string    `xml:"type,attr"  json:",omitempty"`
	Cp       []*Cp     `xml:"p,omitempty" json:"p,omitempty"`
	Ctitle   []*Ctitle `xml:"title,omitempty" json:"title,omitempty"`
	Note     string    `xml:",chardata" json:",omitempty"`
}

type Cnotestmt struct {
	XMLName  xml.Name   `xml:"notestmt,omitempty" json:"notestmt,omitempty"`
	Raw      []byte     `xml:",innerxml" json:",omitempty"`
	Cextref  []*Cextref `xml:"extref,omitempty" json:"extref,omitempty"`
	Clb      []*Clb     `xml:"lb,omitempty" json:"lb,omitempty"`
	Cnote    []*Cnote   `xml:"note,omitempty" json:"note,omitempty"`
	Notestmt string     `xml:",chardata" json:",omitempty"`
}

type Cnum struct {
	XMLName  xml.Name `xml:"num,omitempty" json:"num,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Num      string   `xml:",chardata" json:",omitempty"`
}

type Codd struct {
	XMLName    xml.Name      `xml:"odd,omitempty" json:"odd,omitempty"`
	Raw        []byte        `xml:",innerxml" json:",omitempty"`
	Attrtype   string        `xml:"type,attr"  json:",omitempty"`
	Cchronlist []*Cchronlist `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Chead      []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clist      []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Codd       []*Codd       `xml:"odd,omitempty" json:"odd,omitempty"`
	Cp         []*Cp         `xml:"p,omitempty" json:"p,omitempty"`
	Csubject   *Csubject     `xml:"subject,omitempty" json:"subject,omitempty"`
	Ctable     []*Ctable     `xml:"table,omitempty" json:"table,omitempty"`
}

type Coriginalsloc struct {
	XMLName xml.Name  `xml:"originalsloc,omitempty" json:"originalsloc,omitempty"`
	Raw     []byte    `xml:",innerxml" json:",omitempty"`
	Chead   []*Chead  `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp     `xml:"p,omitempty" json:"p,omitempty"`
	Ctable  []*Ctable `xml:"table,omitempty" json:"table,omitempty"`
}

type Corigination struct {
	XMLName     xml.Name     `xml:"origination,omitempty" json:"origination,omitempty"`
	Raw         []byte       `xml:",innerxml" json:",omitempty"`
	Attrlabel   string       `xml:"label,attr"  json:",omitempty"`
	Ccorpname   []*Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cfamname    []*Cfamname  `xml:"famname,omitempty" json:"famname,omitempty"`
	Cpersname   []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Origination string       `xml:",chardata" json:",omitempty"`
}

type Cotherfindaid struct {
	XMLName xml.Name `xml:"otherfindaid,omitempty" json:"otherfindaid,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Clist   []*Clist `xml:"list,omitempty" json:"list,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cp struct {
	XMLName       xml.Name      `xml:"p,omitempty" json:"p,omitempty"`
	Raw           []byte        `xml:",innerxml" json:",omitempty"`
	Attraltrender string        `xml:"altrender,attr"  json:",omitempty"`
	Attrid        string        `xml:"id,attr"  json:",omitempty"`
	Carchref      []*Carchref   `xml:"archref,omitempty" json:"archref,omitempty"`
	Cbibref       []*Cbibref    `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Cblockquote   *Cblockquote  `xml:"blockquote,omitempty" json:"blockquote,omitempty"`
	Cchronlist    []*Cchronlist `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Ccorpname     []*Ccorpname  `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cdate         []*Cdate      `xml:"date,omitempty" json:"date,omitempty"`
	Cemph         []*Cemph      `xml:"emph,omitempty" json:"emph,omitempty"`
	Cextptr       []*Cextptr    `xml:"extptr,omitempty" json:"extptr,omitempty"`
	Cextref       []*Cextref    `xml:"extref,omitempty" json:"extref,omitempty"`
	Cgeogname     []*Cgeogname  `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Clb           []*Clb        `xml:"lb,omitempty" json:"lb,omitempty"`
	Clist         []*Clist      `xml:"list,omitempty" json:"list,omitempty"`
	Cname         []*Cname      `xml:"name,omitempty" json:"name,omitempty"`
	Cnote         []*Cnote      `xml:"note,omitempty" json:"note,omitempty"`
	Cnum          []*Cnum       `xml:"num,omitempty" json:"num,omitempty"`
	Cpersname     []*Cpersname  `xml:"persname,omitempty" json:"persname,omitempty"`
	Cref          []*Cref       `xml:"ref,omitempty" json:"ref,omitempty"`
	Ctable        []*Ctable     `xml:"table,omitempty" json:"table,omitempty"`
	Ctitle        []*Ctitle     `xml:"title,omitempty" json:"title,omitempty"`
	Cunitdate     []*Cunitdate  `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	P             string        `xml:",chardata" json:",omitempty"`
}

type Cpersname struct {
	XMLName      xml.Name `xml:"persname,omitempty" json:"persname,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Attrid       string   `xml:"id,attr"  json:",omitempty"`
	Attrnormal   string   `xml:"normal,attr"  json:",omitempty"`
	Attrrole     string   `xml:"role,attr"  json:",omitempty"`
	Clb          []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Persname     string   `xml:",chardata" json:",omitempty"`
}

type Cphysdesc struct {
	XMLName     xml.Name       `xml:"physdesc,omitempty" json:"physdesc,omitempty"`
	Raw         []byte         `xml:",innerxml" json:",omitempty"`
	Attrlabel   string         `xml:"label,attr"  json:",omitempty"`
	Cdimensions []*Cdimensions `xml:"dimensions,omitempty" json:"dimensions,omitempty"`
	Cextent     []*Cextent     `xml:"extent,omitempty" json:"extent,omitempty"`
	Cgenreform  *Cgenreform    `xml:"genreform,omitempty" json:"genreform,omitempty"`
	Clb         []*Clb         `xml:"lb,omitempty" json:"lb,omitempty"`
	Cphysfacet  []*Cphysfacet  `xml:"physfacet,omitempty" json:"physfacet,omitempty"`
	Physdesc    string         `xml:",chardata" json:",omitempty"`
}

type Cphysfacet struct {
	XMLName            xml.Name `xml:"physfacet,omitempty" json:"physfacet,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attrencodinganalog string   `xml:"encodinganalog,attr"  json:",omitempty"`
	Attrtype           string   `xml:"type,attr"  json:",omitempty"`
	Physfacet          string   `xml:",chardata" json:",omitempty"`
}

type Cphysloc struct {
	XMLName   xml.Name `xml:"physloc,omitempty" json:"physloc,omitempty"`
	Raw       []byte   `xml:",innerxml" json:",omitempty"`
	Attrlabel string   `xml:"label,attr"  json:",omitempty"`
	Attrtype  string   `xml:"type,attr"  json:",omitempty"`
	Physloc   string   `xml:",chardata" json:",omitempty"`
}

type Cphystech struct {
	XMLName  xml.Name `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Raw      []byte   `xml:",innerxml" json:",omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cprefercite struct {
	XMLName xml.Name `xml:"prefercite,omitempty" json:"prefercite,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cprocessinfo struct {
	XMLName      xml.Name        `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Raw          []byte          `xml:",innerxml" json:",omitempty"`
	Chead        []*Chead        `xml:"head,omitempty" json:"head,omitempty"`
	Clist        []*Clist        `xml:"list,omitempty" json:"list,omitempty"`
	Cp           []*Cp           `xml:"p,omitempty" json:"p,omitempty"`
	Cprocessinfo []*Cprocessinfo `xml:"processinfo,omitempty" json:"processinfo,omitempty"`
	Ctable       []*Ctable       `xml:"table,omitempty" json:"table,omitempty"`
}

type Cprofiledesc struct {
	XMLName    xml.Name    `xml:"profiledesc,omitempty" json:"profiledesc,omitempty"`
	Raw        []byte      `xml:",innerxml" json:",omitempty"`
	Ccreation  *Ccreation  `xml:"creation,omitempty" json:"creation,omitempty"`
	Cdescrules *Cdescrules `xml:"descrules,omitempty" json:"descrules,omitempty"`
	Clangusage *Clangusage `xml:"langusage,omitempty" json:"langusage,omitempty"`
}

type Cptrgrp struct {
	XMLName xml.Name `xml:"ptrgrp,omitempty" json:"ptrgrp,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Cref    []*Cref  `xml:"ref,omitempty" json:"ref,omitempty"`
}

type Cpublicationstmt struct {
	XMLName    xml.Name    `xml:"publicationstmt,omitempty" json:"publicationstmt,omitempty"`
	Raw        []byte      `xml:",innerxml" json:",omitempty"`
	Cdate      []*Cdate    `xml:"date,omitempty" json:"date,omitempty"`
	Cp         []*Cp       `xml:"p,omitempty" json:"p,omitempty"`
	Cpublisher *Cpublisher `xml:"publisher,omitempty" json:"publisher,omitempty"`
}

type Cpublisher struct {
	XMLName   xml.Name `xml:"publisher,omitempty" json:"publisher,omitempty"`
	Raw       []byte   `xml:",innerxml" json:",omitempty"`
	Publisher string   `xml:",chardata" json:",omitempty"`
}

type Cref struct {
	XMLName      xml.Name     `xml:"ref,omitempty" json:"ref,omitempty"`
	Raw          []byte       `xml:",innerxml" json:",omitempty"`
	Attractuate  string       `xml:"actuate,attr"  json:",omitempty"`
	Attrlinktype string       `xml:"linktype,attr"  json:",omitempty"`
	Attrshow     string       `xml:"show,attr"  json:",omitempty"`
	Attrtarget   string       `xml:"target,attr"  json:",omitempty"`
	Cdate        []*Cdate     `xml:"date,omitempty" json:"date,omitempty"`
	Cnote        []*Cnote     `xml:"note,omitempty" json:"note,omitempty"`
	Cpersname    []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Ref          string       `xml:",chardata" json:",omitempty"`
}

type Crelatedmaterial struct {
	XMLName          xml.Name            `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Raw              []byte              `xml:",innerxml" json:",omitempty"`
	Chead            []*Chead            `xml:"head,omitempty" json:"head,omitempty"`
	Clist            []*Clist            `xml:"list,omitempty" json:"list,omitempty"`
	Cp               []*Cp               `xml:"p,omitempty" json:"p,omitempty"`
	Crelatedmaterial []*Crelatedmaterial `xml:"relatedmaterial,omitempty" json:"relatedmaterial,omitempty"`
	Ctable           []*Ctable           `xml:"table,omitempty" json:"table,omitempty"`
}

type Crepository struct {
	XMLName    xml.Name     `xml:"repository,omitempty" json:"repository,omitempty"`
	Raw        []byte       `xml:",innerxml" json:",omitempty"`
	Attrlabel  string       `xml:"label,attr"  json:",omitempty"`
	Ccorpname  []*Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Repository string       `xml:",chardata" json:",omitempty"`
}

type Crevisiondesc struct {
	XMLName      xml.Name   `xml:"revisiondesc,omitempty" json:"revisiondesc,omitempty"`
	Raw          []byte     `xml:",innerxml" json:",omitempty"`
	Attraudience string     `xml:"audience,attr"  json:",omitempty"`
	Cchange      []*Cchange `xml:"change,omitempty" json:"change,omitempty"`
}

type Crow struct {
	XMLName xml.Name  `xml:"row,omitempty" json:"row,omitempty"`
	Raw     []byte    `xml:",innerxml" json:",omitempty"`
	Centry  []*Centry `xml:"entry,omitempty" json:"entry,omitempty"`
}

type Cscopecontent struct {
	XMLName       xml.Name         `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Raw           []byte           `xml:",innerxml" json:",omitempty"`
	Attraltrender string           `xml:"altrender,attr"  json:",omitempty"`
	Cchronlist    []*Cchronlist    `xml:"chronlist,omitempty" json:"chronlist,omitempty"`
	Chead         []*Chead         `xml:"head,omitempty" json:"head,omitempty"`
	Clist         []*Clist         `xml:"list,omitempty" json:"list,omitempty"`
	Cp            []*Cp            `xml:"p,omitempty" json:"p,omitempty"`
	Cscopecontent []*Cscopecontent `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Ctable        []*Ctable        `xml:"table,omitempty" json:"table,omitempty"`
}

type Cseparatedmaterial struct {
	XMLName            xml.Name              `xml:"separatedmaterial,omitempty" json:"separatedmaterial,omitempty"`
	Raw                []byte                `xml:",innerxml" json:",omitempty"`
	Attrtype           string                `xml:"type,attr"  json:",omitempty"`
	Chead              []*Chead              `xml:"head,omitempty" json:"head,omitempty"`
	Clist              []*Clist              `xml:"list,omitempty" json:"list,omitempty"`
	Cp                 []*Cp                 `xml:"p,omitempty" json:"p,omitempty"`
	Cseparatedmaterial []*Cseparatedmaterial `xml:"separatedmaterial,omitempty" json:"separatedmaterial,omitempty"`
	Ctable             []*Ctable             `xml:"table,omitempty" json:"table,omitempty"`
}

type Csubject struct {
	XMLName    xml.Name `xml:"subject,omitempty" json:"subject,omitempty"`
	Raw        []byte   `xml:",innerxml" json:",omitempty"`
	Attrsource string   `xml:"source,attr"  json:",omitempty"`
	Subject    string   `xml:",chardata" json:",omitempty"`
}

type Ctable struct {
	XMLName    xml.Name   `xml:"table,omitempty" json:"table,omitempty"`
	Raw        []byte     `xml:",innerxml" json:",omitempty"`
	Attrcolsep string     `xml:"colsep,attr"  json:",omitempty"`
	Attrframe  string     `xml:"frame,attr"  json:",omitempty"`
	Attrid     string     `xml:"id,attr"  json:",omitempty"`
	Attrpgwide string     `xml:"pgwide,attr"  json:",omitempty"`
	Attrrowsep string     `xml:"rowsep,attr"  json:",omitempty"`
	Chead      []*Chead   `xml:"head,omitempty" json:"head,omitempty"`
	Ctgroup    []*Ctgroup `xml:"tgroup,omitempty" json:"tgroup,omitempty"`
}

type Ctbody struct {
	XMLName xml.Name `xml:"tbody,omitempty" json:"tbody,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Crow    []*Crow  `xml:"row,omitempty" json:"row,omitempty"`
}

type Ctgroup struct {
	XMLName  xml.Name    `xml:"tgroup,omitempty" json:"tgroup,omitempty"`
	Raw      []byte      `xml:",innerxml" json:",omitempty"`
	Attrcols string      `xml:"cols,attr"  json:",omitempty"`
	Ccolspec []*Ccolspec `xml:"colspec,omitempty" json:"colspec,omitempty"`
	Ctbody   *Ctbody     `xml:"tbody,omitempty" json:"tbody,omitempty"`
	Cthead   *Cthead     `xml:"thead,omitempty" json:"thead,omitempty"`
}

type Cthead struct {
	XMLName xml.Name `xml:"thead,omitempty" json:"thead,omitempty"`
	Raw     []byte   `xml:",innerxml" json:",omitempty"`
	Crow    []*Crow  `xml:"row,omitempty" json:"row,omitempty"`
}

type Ctitle struct {
	XMLName      xml.Name `xml:"title,omitempty" json:"title,omitempty"`
	Raw          []byte   `xml:",innerxml" json:",omitempty"`
	Attrlinktype string   `xml:"linktype,attr"  json:",omitempty"`
	Attrrender   string   `xml:"render,attr"  json:",omitempty"`
	Attrtype     string   `xml:"type,attr"  json:",omitempty"`
	Cdate        []*Cdate `xml:"date,omitempty" json:"date,omitempty"`
	Clb          []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Title        string   `xml:",chardata" json:",omitempty"`
}

type Ctitleproper struct {
	XMLName     xml.Name `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
	Raw         []byte   `xml:",innerxml" json:",omitempty"`
	Clb         []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Titleproper string   `xml:",chardata" json:",omitempty"`
}

type Ctitlestmt struct {
	XMLName      xml.Name      `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
	Raw          []byte        `xml:",innerxml" json:",omitempty"`
	Cauthor      *Cauthor      `xml:"author,omitempty" json:"author,omitempty"`
	Ctitleproper *Ctitleproper `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
}

type Cunitdate struct {
	XMLName       xml.Name     `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Raw           []byte       `xml:",innerxml" json:",omitempty"`
	Attrcalendar  string       `xml:"calendar,attr"  json:",omitempty"`
	Attrcertainty string       `xml:"certainty,attr"  json:",omitempty"`
	Attrera       string       `xml:"era,attr"  json:",omitempty"`
	Attrlabel     string       `xml:"label,attr"  json:",omitempty"`
	Attrnormal    string       `xml:"normal,attr"  json:",omitempty"`
	Attrtype      string       `xml:"type,attr"  json:",omitempty"`
	Clb           []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Cunitdate     []*Cunitdate `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Unitdate      string       `xml:",chardata" json:",omitempty"`
}

type Cunitid struct {
	XMLName            xml.Name `xml:"unitid,omitempty" json:"unitid,omitempty"`
	Raw                []byte   `xml:",innerxml" json:",omitempty"`
	Attraudience       string   `xml:"audience,attr,omitempty"  json:",omitempty"`
	Attrcountrycode    string   `xml:"countrycode,attr,omitempty"  json:",omitempty"`
	Attrencodinganalog string   `xml:"encodinganalog,attr,omitempty"  json:",omitempty"`
	Attrid             string   `xml:"id,attr,omitempty"  json:",omitempty"`
	Attridentifier     string   `xml:"identifier,attr,omitempty"  json:",omitempty"`
	Attrlabel          string   `xml:"label,attr,omitempty"  json:",omitempty"`
	Attrrepositorycode string   `xml:"repositorycode,attr,omitempty"  json:",omitempty"`
	Attrtype           string   `xml:"type,attr,omitempty"  json:",omitempty"`
	Unitid             string   `xml:",chardata" json:",omitempty"`
}

type Cunittitle struct {
	XMLName    xml.Name     `xml:"unittitle,omitempty" json:"unittitle,omitempty"`
	Raw        []byte       `xml:",innerxml" json:",omitempty"`
	Attrlabel  string       `xml:"label,attr,omitempty"  json:",omitempty"`
	Attrtype   string       `xml:"type,attr"  json:",omitempty"`
	Carchref   []*Carchref  `xml:"archref,omitempty" json:"archref,omitempty"`
	Cbibref    []*Cbibref   `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Ccorpname  []*Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Cdate      []*Cdate     `xml:"date,omitempty" json:"date,omitempty"`
	Cemph      []*Cemph     `xml:"emph,omitempty" json:"emph,omitempty"`
	Cextref    []*Cextref   `xml:"extref,omitempty" json:"extref,omitempty"`
	Cfamname   []*Cfamname  `xml:"famname,omitempty" json:"famname,omitempty"`
	Cgenreform *Cgenreform  `xml:"genreform,omitempty" json:"genreform,omitempty"`
	Cgeogname  []*Cgeogname `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Clb        []*Clb       `xml:"lb,omitempty" json:"lb,omitempty"`
	Cname      []*Cname     `xml:"name,omitempty" json:"name,omitempty"`
	Cnum       []*Cnum      `xml:"num,omitempty" json:"num,omitempty"`
	Cpersname  []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Cref       []*Cref      `xml:"ref,omitempty" json:"ref,omitempty"`
	Ctitle     []*Ctitle    `xml:"title,omitempty" json:"title,omitempty"`
	Cunitdate  []*Cunitdate `xml:"unitdate,omitempty" json:"unitdate,omitempty"`
	Unittitle  string       `xml:",chardata" json:",omitempty"`
}

type Cuserestrict struct {
	XMLName      xml.Name        `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
	Raw          []byte          `xml:",innerxml" json:",omitempty"`
	Attrtype     string          `xml:"type,attr"  json:",omitempty"`
	Chead        []*Chead        `xml:"head,omitempty" json:"head,omitempty"`
	Clist        []*Clist        `xml:"list,omitempty" json:"list,omitempty"`
	Cp           []*Cp           `xml:"p,omitempty" json:"p,omitempty"`
	Cuserestrict []*Cuserestrict `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
}
