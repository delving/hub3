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

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/models"
	proto "github.com/golang/protobuf/proto"
)

// ReadEAD reads an ead2002 XML from a path
func ReadEAD(path string) (*Cead, error) {
	rawEAD, err := ioutil.ReadFile(path)
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

func ProcessUpload(r *http.Request, spec string) (uint64, error) {

	ds, _, err := models.GetOrCreateDataSet(spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", spec)
		return uint64(0), err
	}

	err = ds.IncrementRevision()
	if err != nil {
		log.Printf("Unable to increment %s\n", spec)
		return uint64(0), err
	}
	basePath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s", spec))
	f, err := os.Create(basePath + ".xml")
	defer f.Close()

	if err != nil {
		log.Printf("Unable to create output file %s; %s", spec, err)
		return uint64(0), err
	}

	in, header, err := r.FormFile("ead")
	if err != nil {
		return uint64(0), err
	}
	defer in.Close()

	buf := bytes.NewBuffer(make([]byte, 0, header.Size))
	_, err = io.Copy(f, io.TeeReader(in, buf))

	if err != nil {
		return uint64(0), err
	}

	cead, err := eadParse(buf.Bytes())
	if err != nil {
		return uint64(0), err
	}

	cfg := NewNodeConfig(context.Background(), true)
	nl, _, err := cead.Carchdesc.Cdsc.NewNodeList(cfg)
	if err != nil {
		return uint64(0), err
	}

	b, err := json.Marshal(nl)
	if err != nil {
		return uint64(0), err
	}

	err = ioutil.WriteFile(basePath+".sparse.nodelist.json", b, 0644)
	if err != nil {
		return uint64(0), err
	}

	// save protobuf
	b, err = proto.Marshal(nl)
	if err != nil {
		return uint64(0), err
	}
	err = ioutil.WriteFile(basePath+".sparse.nodelist.pb", b, 0644)
	if err != nil {
		return uint64(0), err
	}

	// write the header and archDesc without the cLevels
	cead.Carchdesc.Cdsc = nil
	b, err = json.Marshal(cead)
	if err != nil {
		return uint64(0), err
	}

	err = ioutil.WriteFile(basePath+".headers.json", b, 0644)
	if err != nil {
		return uint64(0), err
	}

	return cfg.Counter.GetCount(), nil
}

///////////////////////////
/// structs
///////////////////////////

type Cabstract struct {
	XMLName   xml.Name `xml:"abstract,omitempty" json:"abstract,omitempty"`
	Attrlabel string   `xml:"label,attr"  json:",omitempty"`
	Clb       []*Clb   `xml:"lb,omitempty" json:"lb,omitempty"`
	Abstract  string   `xml:",chardata" json:",omitempty"`
}

type Caccessrestrict struct {
	XMLName      xml.Name      `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Attrid       string        `xml:"id,attr"  json:",omitempty"`
	Attrtype     string        `xml:"type,attr"  json:",omitempty"`
	Chead        []*Chead      `xml:"head,omitempty" json:"head,omitempty"`
	Clegalstatus *Clegalstatus `xml:"legalstatus,omitempty" json:"legalstatus,omitempty"`
	Cp           []*Cp         `xml:"p,omitempty" json:"p,omitempty"`
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
	XMLName   xml.Name    `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
	Attrlevel string      `xml:"level,attr"  json:",omitempty"`
	Attrtype  string      `xml:"type,attr"  json:",omitempty"`
	Cdescgrp  []*Cdescgrp `xml:"descgrp,omitempty" json:"descgrp,omitempty"`
	Cdid      *Cdid       `xml:"did,omitempty" json:"did,omitempty"`
	Cdsc      *Cdsc       `xml:"dsc,omitempty" json:"dsc,omitempty"`
}

type Carrangement struct {
	XMLName xml.Name `xml:"arrangement,omitempty" json:"arrangement,omitempty"`
	Chead   []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cauthor struct {
	XMLName xml.Name `xml:"author,omitempty" json:"author,omitempty"`
	Author  string   `xml:",chardata" json:",omitempty"`
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
}

type Cblockquote struct {
	XMLName xml.Name `xml:"blockquote,omitempty" json:"blockquote,omitempty"`
	Cnote   *Cnote   `xml:"note,omitempty" json:"note,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

type Cc01 struct {
	XMLName         xml.Name         `xml:"c01,omitempty" json:"c01,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc02          `xml:"c02,omitempty" json:"c02,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc02 struct {
	XMLName         xml.Name         `xml:"c02,omitempty" json:"c02,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc03          `xml:"c03,omitempty" json:"c03,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc03 struct {
	XMLName         xml.Name         `xml:"c03,omitempty" json:"c03,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc04          `xml:"c04,omitempty" json:"c04,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc04 struct {
	XMLName         xml.Name         `xml:"c04,omitempty" json:"c04,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc05          `xml:"c05,omitempty" json:"c05,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc05 struct {
	XMLName         xml.Name         `xml:"c05,omitempty" json:"c05,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc06          `xml:"c06,omitempty" json:"c06,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc06 struct {
	XMLName         xml.Name         `xml:"c06,omitempty" json:"c06,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc07          `xml:"c07,omitempty" json:"c07,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc07 struct {
	XMLName         xml.Name         `xml:"c07,omitempty" json:"c07,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc08          `xml:"c08,omitempty" json:"c08,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc08 struct {
	XMLName         xml.Name         `xml:"c08,omitempty" json:"c08,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc09          `xml:"c09,omitempty" json:"c09,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc09 struct {
	XMLName         xml.Name         `xml:"c09,omitempty" json:"c09,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc10          `xml:"c10,omitempty" json:"c10,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc10 struct {
	XMLName         xml.Name         `xml:"c10,omitempty" json:"c10,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc11          `xml:"c11,omitempty" json:"c11,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc11 struct {
	XMLName         xml.Name         `xml:"c11,omitempty" json:"c11,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc12          `xml:"c12,omitempty" json:"c12,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cc12 struct {
	XMLName         xml.Name         `xml:"c12,omitempty" json:"c12,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc12          `xml:"c12,omitempty" json:"c12,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
}

type Cchange struct {
	XMLName xml.Name `xml:"change,omitempty" json:"change,omitempty"`
	Cdate   *Cdate   `xml:"date,omitempty" json:"date,omitempty"`
	Citem   []*Citem `xml:"item,omitempty" json:"item,omitempty"`
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
}

type Ccorpname struct {
	XMLName  xml.Name `xml:"corpname,omitempty" json:"corpname,omitempty"`
	CorpName string   `xml:",chardata" json:",omitempty"`
}

type Ccreation struct {
	XMLName      xml.Name  `xml:"creation,omitempty" json:"creation,omitempty"`
	Attraudience string    `xml:"audience,attr"  json:",omitempty"`
	Cdate        *Cdate    `xml:"date,omitempty" json:"date,omitempty"`
	Ctitle       []*Ctitle `xml:"title,omitempty" json:"title,omitempty"`
	Creation     string    `xml:",chardata" json:",omitempty"`
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
}

type Cdescrules struct {
	XMLName      xml.Name   `xml:"descrules,omitempty" json:"descrules,omitempty"`
	Attraudience string     `xml:"audience,attr"  json:",omitempty"`
	Cbibref      []*Cbibref `xml:"bibref,omitempty" json:"bibref,omitempty"`
	Descrrules   string     `xml:",chardata" json:",omitempty"`
}

type Cdid struct {
	XMLName       xml.Name       `xml:"did,omitempty" json:"did,omitempty"`
	Cabstract     *Cabstract     `xml:"abstract,omitempty" json:"abstract,omitempty"`
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
}

type Cdsc struct {
	XMLName  xml.Name `xml:"dsc,omitempty" json:"dsc,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Nested   []*Cc01  `xml:"c01,omitempty" json:"c01,omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
}

type Cead struct {
	XMLName      xml.Name    `xml:"ead,omitempty" json:"ead,omitempty"`
	Attraudience string      `xml:"audience,attr"  json:",omitempty"`
	Carchdesc    *Carchdesc  `xml:"archdesc,omitempty" json:"archdesc,omitempty"`
	Ceadheader   *Ceadheader `xml:"eadheader,omitempty" json:"eadheader,omitempty"`
}

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

type Cedition struct {
	XMLName xml.Name `xml:"edition,omitempty" json:"edition,omitempty"`
	Edition string   `xml:",chardata" json:",omitempty"`
}

type Ceditionstmt struct {
	XMLName  xml.Name  `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Cedition *Cedition `xml:"edition,omitempty" json:"edition,omitempty"`
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

type Cfiledesc struct {
	XMLName          xml.Name          `xml:"filedesc,omitempty" json:"filedesc,omitempty"`
	Ceditionstmt     *Ceditionstmt     `xml:"editionstmt,omitempty" json:"editionstmt,omitempty"`
	Cpublicationstmt *Cpublicationstmt `xml:"publicationstmt,omitempty" json:"publicationstmt,omitempty"`
	Ctitlestmt       *Ctitlestmt       `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
}

type Chead struct {
	XMLName xml.Name `xml:"head,omitempty" json:"head,omitempty"`
	Head    string   `xml:",chardata" json:",omitempty"`
}

type Citem struct {
	XMLName xml.Name `xml:"item,omitempty" json:"item,omitempty"`
	Cemph   []*Cemph `xml:"emph,omitempty" json:"emph,omitempty"`
	Cextref *Cextref `xml:"extref,omitempty" json:"extref,omitempty"`
	Clist   *Clist   `xml:"list,omitempty" json:"list,omitempty"`
	Item    string   `xml:",chardata" json:",omitempty"`
}

type Clangmaterial struct {
	XMLName   xml.Name   `xml:"langmaterial,omitempty" json:"langmaterial,omitempty"`
	Attrlabel string     `xml:"label,attr"  json:",omitempty"`
	Clanguage *Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	Lang      string     `xml:",chardata" json:",omitempty"`
}

type Clanguage struct {
	XMLName        xml.Name `xml:"language,omitempty" json:"language,omitempty"`
	Attrlangcode   string   `xml:"langcode,attr"  json:",omitempty"`
	Attrscriptcode string   `xml:"scriptcode,attr"  json:",omitempty"`
	Language       string   `xml:",chardata" json:",omitempty"`
}

type Clangusage struct {
	XMLName   xml.Name   `xml:"langusage,omitempty" json:"langusage,omitempty"`
	Clanguage *Clanguage `xml:"language,omitempty" json:"language,omitempty"`
	LangUsage string     `xml:",chardata" json:",omitempty"`
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
}

type Corigination struct {
	XMLName     xml.Name   `xml:"origination,omitempty" json:"origination,omitempty"`
	Attrlabel   string     `xml:"label,attr"  json:",omitempty"`
	Ccorpname   *Ccorpname `xml:"corpname,omitempty" json:"corpname,omitempty"`
	Origination string     `xml:",chardata" json:",omitempty"`
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
	XMLName   xml.Name   `xml:"physdesc,omitempty" json:"physdesc,omitempty"`
	Attrlabel string     `xml:"label,attr"  json:",omitempty"`
	Cextent   []*Cextent `xml:"extent,omitempty" json:"extent,omitempty"`
	PhyscDesc string     `xml:",chardata" json:",omitempty"`
}

type Cphysloc struct {
	XMLName  xml.Name `xml:"physloc,omitempty" json:"physloc,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	PhysLoc  string   `xml:",chardata" json:",omitempty"`
}

type Cphystech struct {
	XMLName  xml.Name `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
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

type Cprofiledesc struct {
	XMLName    xml.Name    `xml:"profiledesc,omitempty" json:"profiledesc,omitempty"`
	Ccreation  *Ccreation  `xml:"creation,omitempty" json:"creation,omitempty"`
	Cdescrules *Cdescrules `xml:"descrules,omitempty" json:"descrules,omitempty"`
	Clangusage *Clangusage `xml:"langusage,omitempty" json:"langusage,omitempty"`
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
}

type Crevisiondesc struct {
	XMLName      xml.Name `xml:"revisiondesc,omitempty" json:"revisiondesc,omitempty"`
	Attraudience string   `xml:"audience,attr"  json:",omitempty"`
	Cchange      *Cchange `xml:"change,omitempty" json:"change,omitempty"`
}

type Cscopecontent struct {
	XMLName xml.Name `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Cp      []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
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

type Ctitleproper struct {
	XMLName     xml.Name `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
	TitleProper string   `xml:",chardata" json:",omitempty"`
}

type Ctitlestmt struct {
	XMLName      xml.Name      `xml:"titlestmt,omitempty" json:"titlestmt,omitempty"`
	Cauthor      *Cauthor      `xml:"author,omitempty" json:"author,omitempty"`
	Ctitleproper *Ctitleproper `xml:"titleproper,omitempty" json:"titleproper,omitempty"`
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
	Title     string       `xml:",chardata" json:",omitempty"`
}

type Cuserestrict struct {
	XMLName  xml.Name `xml:"userestrict,omitempty" json:"userestrict,omitempty"`
	Attrtype string   `xml:"type,attr"  json:",omitempty"`
	Chead    []*Chead `xml:"head,omitempty" json:"head,omitempty"`
	Cp       []*Cp    `xml:"p,omitempty" json:"p,omitempty"`
}

///////////////////////////
