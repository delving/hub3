package sru

import (
	"encoding/xml"
	"io"
)

func newResponse(r io.Reader) (*SearchRetrieveResponse, error) {
	var resp SearchRetrieveResponse

	if err := xml.NewDecoder(r).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type Item struct {
	Body []byte
}

type Harvest struct {
	URL      string
	Callback func(item *Item) error
}

type SearchRetrieveResponse struct {
	XMLName                          xml.Name                          `xml:"searchRetrieveResponse,omitempty" json:"searchRetrieveResponse,omitempty"`
	AttrXmlnsdc                      string                            `xml:"xmlns dc,attr"  json:",omitempty"`
	AttrXmlnsdcx                     string                            `xml:"xmlns dcx,attr"  json:",omitempty"`
	AttrXmlnsddd                     string                            `xml:"xmlns ddd,attr"  json:",omitempty"`
	AttrXmlnsfacets                  string                            `xml:"xmlns facets,attr"  json:",omitempty"`
	AttrXmlnssrw                     string                            `xml:"xmlns srw,attr"  json:",omitempty"`
	AttrXmlnstel                     string                            `xml:"xmlns tel,attr"  json:",omitempty"`
	AttrXmlnsxsi                     string                            `xml:"xmlns xsi,attr"  json:",omitempty"`
	NumberOfRecords__srw             string                            `xml:"http://www.loc.gov/zing/srw/ numberOfRecords,omitempty" json:"numberOfRecords,omitempty"`
	EchoedSearchRetrieveRequest__srw *EchoedSearchRetrieveRequest__srw `xml:"http://www.loc.gov/zing/srw/ echoedSearchRetrieveRequest,omitempty" json:"echoedSearchRetrieveRequest,omitempty"`
	Records__srw                     *Records__srw                     `xml:"http://www.loc.gov/zing/srw/ records,omitempty" json:"records,omitempty"`
	// CkbmdoMilliSeconds__srw           *CkbmdoMilliSeconds__srw           `xml:"http://www.loc.gov/zing/srw/ kbmdoMilliSeconds,omitempty" json:"kbmdoMilliSeconds,omitempty"`
	// CsearchEngineMilliSeconds__srw    *CsearchEngineMilliSeconds__srw    `xml:"http://www.loc.gov/zing/srw/ searchEngineMilliSeconds,omitempty" json:"searchEngineMilliSeconds,omitempty"`
	// CtotalMilliSeconds__srw           *CtotalMilliSeconds__srw           `xml:"http://www.loc.gov/zing/srw/ totalMilliSeconds,omitempty" json:"totalMilliSeconds,omitempty"`
	// Cversion__srw                     *Cversion__srw                     `xml:"http://www.loc.gov/zing/srw/ version,omitempty" json:"version,omitempty"`
}

type EchoedSearchRetrieveRequest__srw struct {
	XMLName             xml.Name `xml:"echoedSearchRetrieveRequest,omitempty" json:"echoedSearchRetrieveRequest,omitempty"`
	MaximumRecords__srw string   `xml:"http://www.loc.gov/zing/srw/ maximumRecords,omitempty" json:"maximumRecords,omitempty"`
	Query__srw          string   `xml:"http://www.loc.gov/zing/srw/ query,omitempty" json:"query,omitempty"`
	RecordSchema__srw   string   `xml:"http://www.loc.gov/zing/srw/ recordSchema,omitempty" json:"recordSchema,omitempty"`
	ResultSetTTL__srw   string   `xml:"http://www.loc.gov/zing/srw/ resultSetTTL,omitempty" json:"resultSetTTL,omitempty"`
	StartRecord__srw    string   `xml:"http://www.loc.gov/zing/srw/ startRecord,omitempty" json:"startRecord,omitempty"`
	Version__srw        string   `xml:"http://www.loc.gov/zing/srw/ version,omitempty" json:"version,omitempty"`
}

type Records__srw struct {
	XMLName     xml.Name       `xml:"records,omitempty" json:"records,omitempty"`
	Record__srw []*Record__srw `xml:"http://www.loc.gov/zing/srw/ record,omitempty" json:"record,omitempty"`
}

type Record__srw struct {
	XMLName              xml.Name         `xml:"record,omitempty" json:"record,omitempty"`
	ExtraRecordData__srw string           `xml:"http://www.loc.gov/zing/srw/ extraRecordData,omitempty" json:"extraRecordData,omitempty"`
	RecordData__srw      *RecordData__srw `xml:"http://www.loc.gov/zing/srw/ recordData,omitempty" json:"recordData,omitempty"`
	RecordPacking__srw   string           `xml:"http://www.loc.gov/zing/srw/ recordPacking,omitempty" json:"recordPacking,omitempty"`
	RecordPosition__srw  string           `xml:"http://www.loc.gov/zing/srw/ recordPosition,omitempty" json:"recordPosition,omitempty"`
	RecordSchema__srw    string           `xml:"http://www.loc.gov/zing/srw/ recordSchema,omitempty" json:"recordSchema,omitempty"`
}

type RecordData__srw struct {
	XMLName xml.Name `xml:"recordData,omitempty" json:"recordData,omitempty"`
	Body    []byte   `xml:",innerxml" json:",omitempty"`
}
