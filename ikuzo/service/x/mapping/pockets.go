package mapping

import (
	"encoding/xml"
	"io"
)

type Source struct {
	XMLName    xml.Name `xml:"pockets,omitempty" json:"pockets,omitempty"`
	OrgID      string   `xml:"orgID,attr"  json:",omitempty"`
	DatasetID  string   `xml:"datasetID,attr"  json:",omitempty"`
	RecdefName string   `xml:"recdef,attr"  json:",omitempty"`
	Pockets    []Pocket `xml:"pocket,omitempty" json:"pocket,omitempty"`
}

type Pocket struct {
	XMLName xml.Name `xml:"pocket,omitempty" json:"pocket,omitempty"`
	ID      string   `xml:"id,attr"  json:",omitempty"`
	Content []byte   `xml:",innerxml" json:",omitempty"`
}

// ParseSource parses original Sip-Creator XML source file.
//
// If the source file is compressed it must be wrapped
// decompressing io.Reader such as gzip.NewReader.
func ParseSource(r io.Reader) (*Source, error) {
	var s Source
	if err := xml.NewDecoder(r).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// MarshalToXML creates a Sip-Creator compliant XML source file.
func (s *Source) MarshalToXML(w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(s)
}
