package ead

import "encoding/xml"

type CContext__rts struct {
	XMLName           xml.Name           `xml:"Context,omitempty" json:"Context,omitempty"`
	AttrCONTEXTCLASS  string             `xml:"CONTEXTCLASS,attr"  json:",omitempty"`
	CPermissions__rts *CPermissions__rts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ Permissions,omitempty" json:"Permissions,omitempty"`
}

type CPermissions__rts struct {
	XMLName       xml.Name `xml:"Permissions,omitempty" json:"Permissions,omitempty"`
	AttrCOPY      string   `xml:"COPY,attr"  json:",omitempty"`
	AttrDELETE    string   `xml:"DELETE,attr"  json:",omitempty"`
	AttrDISCOVER  string   `xml:"DISCOVER,attr"  json:",omitempty"`
	AttrDISPLAY   string   `xml:"DISPLAY,attr"  json:",omitempty"`
	AttrDUPLICATE string   `xml:"DUPLICATE,attr"  json:",omitempty"`
	AttrMODIFY    string   `xml:"MODIFY,attr"  json:",omitempty"`
	AttrPRINT     string   `xml:"PRINT,attr"  json:",omitempty"`
}

type CRightsDeclaration__rts struct {
	XMLName     xml.Name `xml:"RightsDeclaration,omitempty" json:"RightsDeclaration,omitempty"`
	AttrCONTEXT string   `xml:"CONTEXT,attr"  json:",omitempty"`
}

type CRightsDeclarationMD__rts struct {
	XMLName                    xml.Name                 `xml:"RightsDeclarationMD,omitempty" json:"RightsDeclarationMD,omitempty"`
	AttrRIGHTSCATEGORY         string                   `xml:"RIGHTSCATEGORY,attr"  json:",omitempty"`
	AttrRIGHTSDECID            string                   `xml:"RIGHTSDECID,attr"  json:",omitempty"`
	AttrXsiSpaceschemaLocation string                   `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"  json:",omitempty"`
	Attrxmlns                  string                   `xml:"xmlns,attr"  json:",omitempty"`
	AttrXmlnsxsi               string                   `xml:"xmlns xsi,attr"  json:",omitempty"`
	CContext__rts              *CContext__rts           `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ Context,omitempty" json:"Context,omitempty"`
	CRightsDeclaration__rts    *CRightsDeclaration__rts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsDeclaration,omitempty" json:"RightsDeclaration,omitempty"`
	CRightsHolder__rts         *CRightsHolder__rts      `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolder,omitempty" json:"RightsHolder,omitempty"`
}

type CRightsHolder__rts struct {
	XMLName                    xml.Name                    `xml:"RightsHolder,omitempty" json:"RightsHolder,omitempty"`
	CRightsHolderComments__rts *CRightsHolderComments__rts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderComments,omitempty" json:"RightsHolderComments,omitempty"`
	CRightsHolderContact__rts  *CRightsHolderContact__rts  `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContact,omitempty" json:"RightsHolderContact,omitempty"`
	CRightsHolderName__rts     *CRightsHolderName__rts     `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderName,omitempty" json:"RightsHolderName,omitempty"`
}

type CRightsHolderComments__rts struct {
	XMLName xml.Name `xml:"RightsHolderComments,omitempty" json:"RightsHolderComments,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContact__rts struct {
	XMLName                              xml.Name                              `xml:"RightsHolderContact,omitempty" json:"RightsHolderContact,omitempty"`
	CRightsHolderContactAddress__rts     *CRightsHolderContactAddress__rts     `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactAddress,omitempty" json:"RightsHolderContactAddress,omitempty"`
	CRightsHolderContactDesignation__rts *CRightsHolderContactDesignation__rts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactDesignation,omitempty" json:"RightsHolderContactDesignation,omitempty"`
	CRightsHolderContactEmail__rts       *CRightsHolderContactEmail__rts       `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactEmail,omitempty" json:"RightsHolderContactEmail,omitempty"`
	CRightsHolderContactPhone__rts       *CRightsHolderContactPhone__rts       `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsHolderContactPhone,omitempty" json:"RightsHolderContactPhone,omitempty"`
}

type CRightsHolderContactAddress__rts struct {
	XMLName xml.Name `xml:"RightsHolderContactAddress,omitempty" json:"RightsHolderContactAddress,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactDesignation__rts struct {
	XMLName xml.Name `xml:"RightsHolderContactDesignation,omitempty" json:"RightsHolderContactDesignation,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactEmail__rts struct {
	XMLName xml.Name `xml:"RightsHolderContactEmail,omitempty" json:"RightsHolderContactEmail,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderContactPhone__rts struct {
	XMLName       xml.Name `xml:"RightsHolderContactPhone,omitempty" json:"RightsHolderContactPhone,omitempty"`
	AttrPHONETYPE string   `xml:"PHONETYPE,attr"  json:",omitempty"`
	string        string   `xml:",chardata" json:",omitempty"`
}

type CRightsHolderName__rts struct {
	XMLName xml.Name `xml:"RightsHolderName,omitempty" json:"RightsHolderName,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
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
	Cname    *Cname   `xml:"http://www.loc.gov/METS/ name,omitempty" json:"name,omitempty"`
}

type CaltRecordID struct {
	XMLName  xml.Name `xml:"altRecordID,omitempty" json:"altRecordID,omitempty"`
	AttrTYPE string   `xml:"TYPE,attr"  json:",omitempty"`
	string   string   `xml:",chardata" json:",omitempty"`
}

type CamdSec struct {
	XMLName   xml.Name   `xml:"amdSec,omitempty" json:"amdSec,omitempty"`
	CrightsMD *CrightsMD `xml:"http://www.loc.gov/METS/ rightsMD,omitempty" json:"rightsMD,omitempty"`
}

type Cdiv struct {
	XMLName        xml.Name `xml:"div,omitempty" json:"div,omitempty"`
	AttrID         string   `xml:"ID,attr"  json:",omitempty"`
	AttrLABEL      string   `xml:"LABEL,attr"  json:",omitempty"`
	AttrORDER      string   `xml:"ORDER,attr"  json:",omitempty"`
	AttrORDERLABEL string   `xml:"ORDERLABEL,attr"  json:",omitempty"`
	Cdiv           *Cdiv    `xml:"http://www.loc.gov/METS/ div,omitempty" json:"div,omitempty"`
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
	Cfile   *Cfile   `xml:"http://www.loc.gov/METS/ file,omitempty" json:"file,omitempty"`
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
	string  string   `xml:",chardata" json:",omitempty"`
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

type Cname struct {
	XMLName xml.Name `xml:"name,omitempty" json:"name,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
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
	XMLName                   xml.Name                   `xml:"xmlData,omitempty" json:"xmlData,omitempty"`
	CRightsDeclarationMD__rts *CRightsDeclarationMD__rts `xml:"http://www.archivesportaleurope.net/Portal/profiles/rights/ RightsDeclarationMD,omitempty" json:"RightsDeclarationMD,omitempty"`
}
