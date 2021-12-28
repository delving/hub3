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

// nolint:lll
package ead

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/ead/eadpb"
	"github.com/delving/hub3/hub3/fragments"
	rdf "github.com/kiivihal/rdf2go"
	"github.com/rs/zerolog"
)

// readMETS reads an METS ML from a path
func readMETS(filename string) (*Cmets, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	return metsParse(r)
}

// metsParse parses a METS XML file into a set of Go structures
func metsParse(r io.Reader) (*Cmets, error) {
	mets := new(Cmets)
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(mets)

	return mets, err
}

func getMimeTypes(fa *eadpb.FindingAid) []string {
	mimeTypes := []string{}
	for k := range fa.GetMimeTypes() {
		mimeTypes = append(mimeTypes, k)
	}

	return mimeTypes
}

func (mets *Cmets) extractFiles() (map[string]*eadpb.File, error) {
	files := map[string]*eadpb.File{}

	physical := mets.CstructMap.Cdiv
	if physical == nil || len(physical.Cdiv) == 0 {
		return files, fmt.Errorf("no physical items found in METS file")
	}

	for _, item := range physical.Cdiv {
		id := strings.TrimPrefix(item.AttrID, "ID")
		parts := strings.Split(item.AttrLABEL, "/")
		label := parts[len(parts)-1]
		file := &eadpb.File{
			Filename: label,
			Fileuuid: id,
		}
		files[id] = file
	}

	return files, nil
}

func updateFileInfo(files map[string]*eadpb.File, fg []*CfileGrp, fa *eadpb.FindingAid) error {
	var defaultGrp *CfileGrp

	for _, grp := range fg {
		if strings.EqualFold(grp.AttrUSE, "default") {
			defaultGrp = grp
		} else if strings.EqualFold(grp.AttrUSE, "thumbs") {
			for _, metsFile := range grp.Cfile {
				id := strings.TrimSuffix(strings.TrimPrefix(metsFile.AttrID, "ID"), "THB")
				file, ok := files[id]
				if !ok {
					return fmt.Errorf("id should always be in webresource map: %s", id)
				}
				if metsFile.CFLocat != nil {
					file.ThumbnailURI = metsFile.CFLocat.AttrXlinkSpacehref
				}
			}
		}
	}

	if defaultGrp == nil {
		metsErr := fmt.Errorf("mets without a default filegroup is invalid")
		c.Config.Logger.Error().
			Err(metsErr).
			Msg("processingerror due to invalid mets file")

		return metsErr
	}

	for _, metsFile := range defaultGrp.Cfile {
		id := strings.TrimSuffix(strings.TrimPrefix(metsFile.AttrID, "ID"), "DEF")

		file, ok := files[id]
		if !ok {
			return fmt.Errorf("id should always be in webresource map: %s", id)
		}

		file.MimeType = metsFile.AttrMIMETYPE
		fa.GetMimeTypes()[file.MimeType]++

		size, err := strconv.Atoi(metsFile.AttrSIZE)
		if err != nil {
			return err
		}

		file.FileSize = int32(size)

		if metsFile.CFLocat != nil {
			file.DownloadURI = metsFile.CFLocat.AttrXlinkSpacehref
			file.DeepzoomURI = createDeepZoomURI(file, fa.Duuid)
		}
	}

	return nil
}

func (mets *Cmets) newFindingAid(cfg *DaoConfig) (eadpb.FindingAid, error) {
	fa := eadpb.FindingAid{
		ArchiveID:      cfg.ArchiveID,
		InventoryID:    cfg.InventoryID,
		InventoryPath:  cfg.InventoryPath,
		InventoryTitle: cfg.InventoryTitle,
		HasOnlyTiles:   false,
		MimeTypes:      map[string]int32{},
		FileCount:      0,
		FilterTypes:    cfg.FilterTypes,
	}

	if mets.CmetsHdr.CaltRecordID != nil {
		fa.Duuid = mets.CmetsHdr.CaltRecordID.Text
	}

	fa.ArchiveTitle = cfg.getArchiveTitle()

	files, err := mets.extractFiles()
	if err != nil {
		return fa, err
	}

	err = updateFileInfo(files, mets.CfileSec.CfileGrp, &fa)
	if err != nil {
		return fa, err
	}

	if len(files) == 0 {
		return fa, nil
	}

	fa.Files = []*eadpb.File{}

	for _, value := range files {
		fa.Files = append(fa.Files, value)
	}

	sort.Slice(fa.Files, func(i, j int) bool { return fa.Files[i].Filename < fa.Files[j].Filename })

	// update order
	for idx, file := range fa.Files {
		file.SortKey = int32(idx + 1)
	}

	fa.FileCount = int32(len(fa.Files))

loop:
	for k := range fa.MimeTypes {
		switch {
		case isTileMimeType(k):
			fa.HasOnlyTiles = true
		default:
			// non-file so set to false and break
			fa.HasOnlyTiles = false
			break loop
		}
	}

	return fa, nil
}

var imageMimeTypes = []string{"image/tiff", "image/jpeg", "image/jpg"}

func isTileMimeType(mimeType string) bool {
	for _, allowed := range imageMimeTypes {
		if strings.EqualFold(mimeType, allowed) {
			return true
		}
	}

	return false
}

func createDeepZoomURI(file *eadpb.File, duuid string) string {
	if !isTileMimeType(file.MimeType) {
		return ""
	}

	if file.ThumbnailURI == "" {
		return ""
	}

	serviceURL, err := url.Parse(file.ThumbnailURI)
	if err != nil {
		return ""
	}

	chunkedUUID := chunkString(strings.ReplaceAll(duuid, "-", ""), 2)

	return fmt.Sprintf(
		"%s://%s/iip?IIIF=/%s/%s.jp2/info.json",
		serviceURL.Scheme,
		serviceURL.Host,
		strings.Join(chunkedUUID, "/"),
		file.Fileuuid,
	)
}

func chunkString(s string, chunkSize int) []string {
	var chunks []string

	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}

		chunks = append(chunks, string(runes[i:nn]))
	}

	return chunks
}

func findingAidTriples(subject string, fa *eadpb.FindingAid, daoCfg *DaoConfig) []*rdf.Triple {
	s := rdf.NewResource(subject)
	triples := []*rdf.Triple{
		rdf.NewTriple(
			s,
			rdf.NewResource(fragments.RDFType),
			rdf.NewResource("https://archief.nl/def/ead/mets/FindingAid"),
		),
	}
	t := func(s rdf.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
	}

	t(s, "duuid", fa.Duuid, rdf.NewLiteral)
	t(s, "archiveID", fa.ArchiveID, rdf.NewLiteral)
	t(s, "archiveTitle", fa.ArchiveTitle, rdf.NewLiteral)
	t(s, "inventoryID", fa.InventoryID, rdf.NewLiteral)
	t(s, "inventoryTitle", fa.InventoryTitle, rdf.NewLiteral)

	for _, filterType := range fa.FilterTypes {
		t(s, "filterTypes", filterType, rdf.NewLiteral)
	}

	for _, file := range fa.Files {
		t(s, "fileUUID", file.Fileuuid, rdf.NewLiteral)
		t(s, "recordUUID", file.Filename, rdf.NewLiteral)
	}
	for _, period := range daoCfg.Periods {
		t(s, "periodDesc", period, rdf.NewLiteral)
	}
	t(s, "dao", daoCfg.Link, rdf.NewLiteral)

	return triples
}

// saveFileFragmentGraphs saves all eadpb.File and eadpb.FindingAid graphs to ElasticSearch.
// Note that during the process the files are removed from the eadpb.FindingAid.
// func saveFileFragmentGraphs(cfg *NodeConfig, fa *eadpb.FindingAid) error {
// for _, file := range fa.Files {
// fg, err := fragmentGraph(cfg, fa, file)
// if err != nil {
// return err
// }

// if err := writeDaoFile(cfg, fg); err != nil {
// return err
// }
// }

// fg, err := findingAidFragmentGraph(cfg, fa)
// if err != nil {
// return err
// }

// if err := writeDaoFile(cfg, fg); err != nil {
// return err
// }

// return nil
// }

type METSEventStatusEnum int

const (
	METSCreate METSEventStatusEnum = iota
	METSUpdate
)

func (m METSEventStatusEnum) String() string {
	return [...]string{
		"create",
		"update",
	}[m]
}

func getEventMETS(level zerolog.Level, archiveID, inventoryID string, eventType METSEventStatusEnum) *zerolog.Event {
	return c.Config.Logger.WithLevel(level).
		Str("application", "hub3").
		Str("app_process", "METS").
		Str("event_type", eventType.String()).
		Str("archive_id", archiveID).
		Str("inventory_id", inventoryID)
}

func logMETSError(archiveID, inventoryID, format string, a ...interface{}) {
	getEventMETS(zerolog.ErrorLevel, archiveID, inventoryID, METSCreate).
		Str("error.message", fmt.Sprint(format, a)).
		Str("metsStatus", "failed").
		Send()
}

type CContextRts struct {
	XMLName          xml.Name         `xml:"Context,omitempty" json:"Context,omitempty"`
	AttrCONTEXTCLASS string           `xml:"CONTEXTCLASS,attr"  json:",omitempty"`
	CPermissionsRts  *CPermissionsRts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ Permissions,omitempty" json:"Permissions,omitempty"`
}

type CPermissionsRts struct {
	XMLName       xml.Name `xml:"Permissions,omitempty" json:"Permissions,omitempty"`
	AttrCOPY      string   `xml:"COPY,attr"  json:",omitempty"`
	AttrDELETE    string   `xml:"DELETE,attr"  json:",omitempty"`
	AttrDISCOVER  string   `xml:"DISCOVER,attr"  json:",omitempty"`
	AttrDISPLAY   string   `xml:"DISPLAY,attr"  json:",omitempty"`
	AttrDUPLICATE string   `xml:"DUPLICATE,attr"  json:",omitempty"`
	AttrMODIFY    string   `xml:"MODIFY,attr"  json:",omitempty"`
	AttrPRINT     string   `xml:"PRINT,attr"  json:",omitempty"`
}

type CRightsDeclarationRts struct {
	XMLName     xml.Name `xml:"RightsDeclaration,omitempty" json:"RightsDeclaration,omitempty"`
	AttrCONTEXT string   `xml:"CONTEXT,attr"  json:",omitempty"`
}

type CRightsDeclarationMDRts struct {
	XMLName                    xml.Name               `xml:"RightsDeclarationMD,omitempty" json:"RightsDeclarationMD,omitempty"`
	AttrRIGHTSCATEGORY         string                 `xml:"RIGHTSCATEGORY,attr"  json:",omitempty"`
	AttrRIGHTSDECID            string                 `xml:"RIGHTSDECID,attr"  json:",omitempty"`
	AttrXsiSpaceschemaLocation string                 `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"  json:",omitempty"`
	Attrxmlns                  string                 `xml:"xmlns,attr"  json:",omitempty"`
	AttrXmlnsxsi               string                 `xml:"xmlns xsi,attr"  json:",omitempty"`
	CContextRts                *CContextRts           `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ Context,omitempty" json:"Context,omitempty"`
	CRightsDeclarationRts      *CRightsDeclarationRts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsDeclaration,omitempty" json:"RightsDeclaration,omitempty"`
	CRightsHolderRts           *CRightsHolderRts      `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolder,omitempty" json:"RightsHolder,omitempty"`
}

type CRightsHolderRts struct {
	XMLName                  xml.Name                  `xml:"RightsHolder,omitempty" json:"RightsHolder,omitempty"`
	CRightsHolderCommentsRts *CRightsHolderCommentsRts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderComments,omitempty" json:"RightsHolderComments,omitempty"`
	CRightsHolderContactRts  *CRightsHolderContactRts  `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContact,omitempty" json:"RightsHolderContact,omitempty"`
	CRightsHolderNameRts     *CRightsHolderNameRts     `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderName,omitempty" json:"RightsHolderName,omitempty"`
}

type CRightsHolderCommentsRts struct {
	XMLName xml.Name `xml:"RightsHolderComments,omitempty" json:"RightsHolderComments,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactRts struct {
	XMLName                            xml.Name                            `xml:"RightsHolderContact,omitempty" json:"RightsHolderContact,omitempty"`
	CRightsHolderContactAddressRts     *CRightsHolderContactAddressRts     `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactAddress,omitempty" json:"RightsHolderContactAddress,omitempty"`
	CRightsHolderContactDesignationRts *CRightsHolderContactDesignationRts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactDesignation,omitempty" json:"RightsHolderContactDesignation,omitempty"`
	CRightsHolderContactEmailRts       *CRightsHolderContactEmailRts       `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactEmail,omitempty" json:"RightsHolderContactEmail,omitempty"`
	CRightsHolderContactPhoneRts       *CRightsHolderContactPhoneRts       `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactPhone,omitempty" json:"RightsHolderContactPhone,omitempty"`
}

type CRightsHolderContactAddressRts struct {
	XMLName xml.Name `xml:"RightsHolderContactAddress,omitempty" json:"RightsHolderContactAddress,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactDesignationRts struct {
	XMLName xml.Name `xml:"RightsHolderContactDesignation,omitempty" json:"RightsHolderContactDesignation,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactEmailRts struct {
	XMLName xml.Name `xml:"RightsHolderContactEmail,omitempty" json:"RightsHolderContactEmail,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactPhoneRts struct {
	XMLName       xml.Name `xml:"RightsHolderContactPhone,omitempty" json:"RightsHolderContactPhone,omitempty"`
	AttrPHONETYPE string   `xml:"PHONETYPE,attr"  json:",omitempty"`
	Text          string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderNameRts struct {
	XMLName xml.Name `xml:"RightsHolderName,omitempty" json:"RightsHolderName,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CFLocat struct {
	XMLName            xml.Name `xml:"FLocat,omitempty" json:"FLocat,omitempty"`
	AttrLOCTYPE        string   `xml:"LOCTYPE,attr"  json:",omitempty"`
	AttrXlinkSpacehref string   `xml:"http://www.w3.org/1999/xlink href,attr"  json:",omitempty"`
	AttrXlinkSpacetype string   `xml:"http://www.w3.org/1999/xlink type,attr"  json:",omitempty"`
}

type Cagent struct {
	XMLName  xml.Name `xml:"agent,omitempty" json:"agent,omitempty"`
	AttrROLE string   `xml:"ROLE,attr"  json:",omitempty"`
	Cmname   *Cmname  `xml:"http://www.loc.gov/METS/ name,omitempty" json:"name,omitempty"`
}

type CaltRecordID struct {
	XMLName  xml.Name `xml:"altRecordID,omitempty" json:"altRecordID,omitempty"`
	AttrTYPE string   `xml:"TYPE,attr"  json:",omitempty"`
	Text     string   `xml:",chardata" json:",omitempty"`
}

type CamdSec struct {
	XMLName   xml.Name   `xml:"amdSec,omitempty" json:"amdSec,omitempty"`
	CrightsMD *CrightsMD `xml:"http://www.loc.gov/METS/ rightsMD,omitempty" json:"rightsMD,omitempty"`
}

// nolint:govet
type Cdiv struct {
	XMLName        xml.Name `xml:"div,omitempty" json:"div,omitempty"`
	AttrID         string   `xml:"ID,attr"  json:",omitempty"`
	AttrLABEL      string   `xml:"LABEL,attr"  json:",omitempty"`
	AttrORDER      string   `xml:"ORDER,attr"  json:",omitempty"`
	AttrORDERLABEL string   `xml:"ORDERLABEL,attr"  json:",omitempty"`
	Cdiv           []*Cdiv  `xml:"http://www.loc.gov/METS/ div,omitempty" json:"div,omitempty"`
	Cfptr          []*Cfptr `xml:"http://www.loc.gov/METS/ fptr,omitempty" json:"fptr,omitempty"`
}

type Cfile struct {
	XMLName      xml.Name `xml:"file,omitempty" json:"file,omitempty"`
	AttrID       string   `xml:"ID,attr"  json:",omitempty"`
	AttrMIMETYPE string   `xml:"MIMETYPE,attr"  json:",omitempty"`
	AttrSIZE     string   `xml:"SIZE,attr"  json:",omitempty"`
	AttrUSE      string   `xml:"USE,attr"  json:",omitempty"`
	CFLocat      *CFLocat `xml:"http://www.loc.gov/METS/ FLocat,omitempty" json:"FLocat,omitempty"`
}

type CfileGrp struct {
	XMLName xml.Name `xml:"fileGrp,omitempty" json:"fileGrp,omitempty"`
	AttrUSE string   `xml:"USE,attr"  json:",omitempty"`
	Cfile   []*Cfile `xml:"http://www.loc.gov/METS/ file,omitempty" json:"file,omitempty"`
}

type CfileSec struct {
	XMLName  xml.Name    `xml:"fileSec,omitempty" json:"fileSec,omitempty"`
	CfileGrp []*CfileGrp `xml:"http://www.loc.gov/METS/ fileGrp,omitempty" json:"fileGrp,omitempty"`
}

type Cfptr struct {
	XMLName    xml.Name `xml:"fptr,omitempty" json:"fptr,omitempty"`
	AttrFILEID string   `xml:"FILEID,attr"  json:",omitempty"`
}

type CmdWrap struct {
	XMLName         xml.Name  `xml:"mdWrap,omitempty" json:"mdWrap,omitempty"`
	AttrMDTYPE      string    `xml:"MDTYPE,attr"  json:",omitempty"`
	AttrOTHERMDTYPE string    `xml:"OTHERMDTYPE,attr"  json:",omitempty"`
	CxmlData        *CxmlData `xml:"http://www.loc.gov/METS/ xmlData,omitempty" json:"xmlData,omitempty"`
}

type Cmets struct {
	XMLName                    xml.Name    `xml:"mets,omitempty" json:"mets,omitempty"`
	AttrPROFILE                string      `xml:"PROFILE,attr"  json:",omitempty"`
	AttrXmlnsrts               string      `xml:"xmlns rts,attr"  json:",omitempty"`
	AttrXsiSpaceschemaLocation string      `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"  json:",omitempty"`
	AttrXmlnsxlink             string      `xml:"xmlns xlink,attr"  json:",omitempty"`
	Attrxmlns                  string      `xml:"xmlns,attr"  json:",omitempty"`
	AttrXmlnsxs                string      `xml:"xmlns xs,attr"  json:",omitempty"`
	AttrXmlnsxsi               string      `xml:"xmlns xsi,attr"  json:",omitempty"`
	CamdSec                    *CamdSec    `xml:"http://www.loc.gov/METS/ amdSec,omitempty" json:"amdSec,omitempty"`
	CfileSec                   *CfileSec   `xml:"http://www.loc.gov/METS/ fileSec,omitempty" json:"fileSec,omitempty"`
	CmetsHdr                   *CmetsHdr   `xml:"http://www.loc.gov/METS/ metsHdr,omitempty" json:"metsHdr,omitempty"`
	CstructMap                 *CstructMap `xml:"http://www.loc.gov/METS/ structMap,omitempty" json:"structMap,omitempty"`
}

type CmetsDocumentID struct {
	XMLName xml.Name `xml:"metsDocumentID,omitempty" json:"metsDocumentID,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CmetsHdr struct {
	XMLName          xml.Name         `xml:"metsHdr,omitempty" json:"metsHdr,omitempty"`
	AttrCREATEDATE   string           `xml:"CREATEDATE,attr"  json:",omitempty"`
	AttrLASTMODDATE  string           `xml:"LASTMODDATE,attr"  json:",omitempty"`
	AttrRECORDSTATUS string           `xml:"RECORDSTATUS,attr"  json:",omitempty"`
	Cagent           []*Cagent        `xml:"http://www.loc.gov/METS/ agent,omitempty" json:"agent,omitempty"`
	CaltRecordID     *CaltRecordID    `xml:"http://www.loc.gov/METS/ altRecordID,omitempty" json:"altRecordID,omitempty"`
	CmetsDocumentID  *CmetsDocumentID `xml:"http://www.loc.gov/METS/ metsDocumentID,omitempty" json:"metsDocumentID,omitempty"`
}

type Cmname struct {
	XMLName xml.Name `xml:"name,omitempty" json:"name,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CrightsMD struct {
	XMLName xml.Name `xml:"rightsMD,omitempty" json:"rightsMD,omitempty"`
	AttrID  string   `xml:"ID,attr"  json:",omitempty"`
	CmdWrap *CmdWrap `xml:"http://www.loc.gov/METS/ mdWrap,omitempty" json:"mdWrap,omitempty"`
}

type CstructMap struct {
	XMLName   xml.Name `xml:"structMap,omitempty" json:"structMap,omitempty"`
	AttrLABEL string   `xml:"LABEL,attr"  json:",omitempty"`
	AttrTYPE  string   `xml:"TYPE,attr"  json:",omitempty"`
	Cdiv      *Cdiv    `xml:"http://www.loc.gov/METS/ div,omitempty" json:"div,omitempty"`
}

type CxmlData struct {
	XMLName                 xml.Name                 `xml:"xmlData,omitempty" json:"xmlData,omitempty"`
	CRightsDeclarationMDRts *CRightsDeclarationMDRts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsDeclarationMD,omitempty" json:"RightsDeclarationMD,omitempty"`
}
