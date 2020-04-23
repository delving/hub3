package ead

import (
	"bytes"
	"encoding/xml"
)

type NLPType int

const (
	Unknown NLPType = iota
	Person
	GeoLocation
	DateText
	DateIso
	Organization
)

type NLPToken struct {
	Type NLPType
	Text string
}

type Extractor struct {
	XMLName   xml.Name     `xml:"input,omitempty" json:"input,omitempty"`
	Cgeogname []*Cgeogname `xml:"geogname,omitempty" json:"geogname,omitempty"`
	Cpersname []*Cpersname `xml:"persname,omitempty" json:"persname,omitempty"`
	Cdate     []*Cdate     `xml:"date,omitempty" json:"date,omitempty"`
}

func NewExtractor(b []byte) (*Extractor, error) {
	var buf bytes.Buffer

	buf.WriteString("<input>")
	buf.Write(b)
	buf.WriteString("</input>")

	in := new(Extractor)

	err := xml.Unmarshal(buf.Bytes(), in)
	if err != nil {
		return nil, err
	}

	return in, err
}

func (e *Extractor) Tokens() []NLPToken {
	tokens := []NLPToken{}

	for _, geo := range e.Cgeogname {
		tokens = append(tokens, NLPToken{Text: geo.Geogname, Type: GeoLocation})
	}

	for _, pers := range e.Cpersname {
		tokens = append(tokens, NLPToken{Text: pers.Persname, Type: Person})
	}

	for _, date := range e.Cdate {
		tokens = append(tokens, NLPToken{Text: date.Date, Type: DateText})

		if date.Attrnormal != "" {
			tokens = append(tokens, NLPToken{Text: date.Attrnormal, Type: DateIso})
		}
	}

	return tokens
}
