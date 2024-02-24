package internal

import (
	"encoding/xml"
	"fmt"
	"io"
)

func ParseResponse(r io.Reader) (*CadlibXML, error) {
	var resp CadlibXML
	decodeErr := xml.NewDecoder(r).Decode(&resp)
	if decodeErr != nil {
		return nil, fmt.Errorf("unable to decode Adlib API Response; %w", decodeErr)
	}
	return &resp, nil
}

type CadlibXML struct {
	XMLName     xml.Name     `xml:"adlibXML,omitempty" json:"adlibXML,omitempty"`
	Cdiagnostic *Cdiagnostic `xml:"diagnostic,omitempty" json:"diagnostic,omitempty"`
	CrecordList *CrecordList `xml:"recordList,omitempty" json:"recordList,omitempty"`
}

func (a *CadlibXML) GetError() string {
	if a.Cdiagnostic == nil {
		return ""
	}
	if a.Cdiagnostic.Cerror == nil {
		return ""
	}

	return a.Cdiagnostic.Cerror.Message
}

type CrecordList struct {
	XMLName xml.Name   `xml:"recordList,omitempty" json:"recordList,omitempty"`
	Crecord []*Crecord `xml:"record,omitempty" json:"record,omitempty"`
}

type Crecord struct {
	XMLName          xml.Name `xml:"record,omitempty" json:"record,omitempty"`
	Attrcreated      string   `xml:"created,attr"  json:",omitempty"`
	Attrmodification string   `xml:"modification,attr"  json:",omitempty"`
	Attrpriref       string   `xml:"priref,attr"  json:",omitempty"`
	Attrselected     string   `xml:"selected,attr"  json:",omitempty"`
	Raw              []byte   `xml:",innerxml" json:",omitempty"`
}

type Cdiagnostic struct {
	XMLName          xml.Name          `xml:"diagnostic,omitempty" json:"diagnostic,omitempty"`
	Ccgistring       *Ccgistring       `xml:"cgistring,omitempty" json:"cgistring,omitempty"`
	Cdbname          *Cdbname          `xml:"dbname,omitempty" json:"dbname,omitempty"`
	Cdsname          *Cdsname          `xml:"dsname,omitempty" json:"dsname,omitempty"`
	CfirstItem       *CfirstItem       `xml:"first_item,omitempty" json:"first_item,omitempty"`
	Chits            *Chits            `xml:"hits,omitempty" json:"hits,omitempty"`
	ChitsOnDisplay   *ChitsOnDisplay   `xml:"hits_on_display,omitempty" json:"hits_on_display,omitempty"`
	Climit           *Climit           `xml:"limit,omitempty" json:"limit,omitempty"`
	ClinkResolveTime *ClinkResolveTime `xml:"link_resolve_time,omitempty" json:"link_resolve_time,omitempty"`
	CresponseTime    *CresponseTime    `xml:"response_time,omitempty" json:"response_time,omitempty"`
	Csearch          *Csearch          `xml:"search,omitempty" json:"search,omitempty"`
	Csort            *Csort            `xml:"sort,omitempty" json:"sort,omitempty"`
	CxmlCreationTime *CxmlCreationTime `xml:"xml_creation_time,omitempty" json:"xml_creation_time,omitempty"`
	Cxmltype         *Cxmltype         `xml:"xmltype,omitempty" json:"xmltype,omitempty"`
	Cerror           *Cerror
}

type Cerror struct {
	XMLName xml.Name `xml:"error,omitempty" json:"error,omitempty"`
	Info    string   `xml:"info"  json:"info,omitempty"`
	Message string   `xml:"message"  json:"message,omitempty"`
}

type Chits struct {
	XMLName xml.Name `xml:"hits,omitempty" json:"hits,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Cxmltype struct {
	XMLName xml.Name `xml:"xmltype,omitempty" json:"xmltype,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type ClinkResolveTime struct {
	XMLName     xml.Name `xml:"link_resolve_time,omitempty" json:"link_resolve_time,omitempty"`
	Attrculture string   `xml:"culture,attr"  json:",omitempty"`
	Attrunit    string   `xml:"unit,attr"  json:",omitempty"`
	Text        string   `xml:",chardata" json:",omitempty"`
}

type CfirstItem struct {
	XMLName xml.Name `xml:"first_item,omitempty" json:"first_item,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Csearch struct {
	XMLName xml.Name `xml:"search,omitempty" json:"search,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Csort struct {
	XMLName xml.Name `xml:"sort,omitempty" json:"sort,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Climit struct {
	XMLName xml.Name `xml:"limit,omitempty" json:"limit,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type ChitsOnDisplay struct {
	XMLName xml.Name `xml:"hits_on_display,omitempty" json:"hits_on_display,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type CresponseTime struct {
	XMLName     xml.Name `xml:"response_time,omitempty" json:"response_time,omitempty"`
	Attrculture string   `xml:"culture,attr"  json:",omitempty"`
	Attrunit    string   `xml:"unit,attr"  json:",omitempty"`
	Text        string   `xml:",chardata" json:",omitempty"`
}

type CxmlCreationTime struct {
	XMLName     xml.Name `xml:"xml_creation_time,omitempty" json:"xml_creation_time,omitempty"`
	Attrculture string   `xml:"culture,attr"  json:",omitempty"`
	Attrunit    string   `xml:"unit,attr"  json:",omitempty"`
	Text        string   `xml:",chardata" json:",omitempty"`
}

type Cdbname struct {
	XMLName xml.Name `xml:"dbname,omitempty" json:"dbname,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Cdsname struct {
	XMLName xml.Name `xml:"dsname,omitempty" json:"dsname,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type Ccgistring struct {
	XMLName xml.Name `xml:"cgistring,omitempty" json:"cgistring,omitempty"`
	Cparam  *Cparam  `xml:"param,omitempty" json:"param,omitempty"`
}

type Cparam struct {
	XMLName  xml.Name `xml:"param,omitempty" json:"param,omitempty"`
	Attrname string   `xml:"name,attr"  json:",omitempty"`
	Text     string   `xml:",chardata" json:",omitempty"`
}
