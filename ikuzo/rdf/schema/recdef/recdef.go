package recdef

import (
	"encoding/xml"
	"io"
)

func Parse(r io.Reader) (*RecDef, error) {
	var rd RecDef
	if err := xml.NewDecoder(r).Decode(&rd); err != nil {
		return nil, err
	}

	return &rd, nil
}

func (r *RecDef) Write(w io.Writer) error {
	w.Write([]byte(xml.Header))

	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(r)
}

type Attr struct {
	XMLName      xml.Name     `xml:"attr,omitempty" json:"attr,omitempty"`
	Attrtag      string       `xml:"tag,attr"  json:",omitempty"`
	AttruriCheck string       `xml:"uriCheck,attr,omitempty"  json:",omitempty"`
	NodeMapping  *NodeMapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type Attrs struct {
	XMLName xml.Name `xml:"attrs,omitempty" json:"attrs,omitempty"`
	Attr    []*Attr  `xml:"attr,omitempty" json:"attr,omitempty"`
}

type Doc struct {
	XMLName  xml.Name `xml:"doc,omitempty" json:"doc,omitempty"`
	Attrpath string   `xml:"path,attr"  json:",omitempty"`
	Para     []*Para  `xml:"para,omitempty" json:"para,omitempty"`
}

type Docs struct {
	XMLName xml.Name `xml:"docs,omitempty" json:"docs,omitempty"`
	Doc     []*Doc   `xml:"doc,omitempty" json:"doc,omitempty"`
}

type Elem struct {
	XMLName        xml.Name     `yml:"elem,omitempty" json:"elem,omitempty"`
	Attrattribute  string       `xml:"attribute,attr,omitempty"  json:",omitempty"`
	Attrtag        string       `xml:"tag,attr,omitempty"  json:",omitempty"`
	Attrattrs      string       `xml:"attrs,attr,omitempty"  json:",omitempty"`
	Attrfunction   string       `xml:"function,attr,omitempty"  json:",omitempty"`
	Attrhidden     string       `xml:"hidden,attr,omitempty"  json:",omitempty"`
	Attrunmappable string       `xml:"unmappable,attr,omitempty"  json:",omitempty"`
	AttruriCheck   string       `xml:"uriCheck,attr,omitempty"  json:",omitempty"`
	Attr           []*Attr      `xml:"attr,omitempty" json:"attr,omitempty"`
	Elems          []*Elem      `xml:"elem,omitempty" json:"elem,omitempty"`
	NodeMapping    *NodeMapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type FieldMarkers struct {
	XMLName xml.Name `xml:"field-markers,omitempty" json:"field-markers,omitempty"`
}

type Functions struct {
	XMLName         xml.Name           `xml:"functions,omitempty" json:"functions,omitempty"`
	MappingFunction []*MappingFunction `xml:"mapping-function,omitempty" json:"mapping-function,omitempty"`
}

type GroovyCode struct {
	XMLName xml.Name  `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
	String  []*String `xml:"string,omitempty" json:"string,omitempty"`
}

type IsRequiredBy struct {
	XMLName xml.Name `xml:"isRequiredBy,omitempty" json:"isRequiredBy,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}

type MappingFunction struct {
	XMLName     xml.Name     `xml:"mapping-function,omitempty" json:"mapping-function,omitempty"`
	Attrname    string       `xml:"name,attr"  json:",omitempty"`
	SampleInput *SampleInput `xml:"sample-input,omitempty" json:"sample-input,omitempty"`
	GroovyCode  *GroovyCode  `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
}

type Namespace struct {
	XMLName    xml.Name `xml:"namespace,omitempty" json:"namespace,omitempty"`
	Attrprefix string   `xml:"prefix,attr"  json:",omitempty"`
	Attruri    string   `xml:"uri,attr"  json:",omitempty"`
}

type Namespaces struct {
	XMLName   xml.Name     `xml:"namespaces,omitempty" json:"namespaces,omitempty"`
	Namespace []*Namespace `xml:"namespace,omitempty" json:"namespace,omitempty"`
}

type NodeMapping struct {
	XMLName       xml.Name    `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
	AttrinputPath string      `xml:"inputPath,attr"  json:",omitempty"`
	GroovyCode    *GroovyCode `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
}

type Opt struct {
	XMLName   xml.Name `xml:"opt,omitempty" json:"opt,omitempty"`
	Attrvalue string   `xml:"value,attr"  json:",omitempty"`
}

type OptList struct {
	XMLName         xml.Name `xml:"opt-list,omitempty" json:"opt-list,omitempty"`
	Attrdictionary  string   `xml:"dictionary,attr"  json:",omitempty"`
	AttrdisplayName string   `xml:"displayName,attr"  json:",omitempty"`
	Attrpath        string   `xml:"path,attr"  json:",omitempty"`
	Opt             []*Opt   `xml:"opt,omitempty" json:"opt,omitempty"`
}

type Opts struct {
	XMLName xml.Name   `xml:"opts,omitempty" json:"opts,omitempty"`
	OptList []*OptList `xml:"opt-list,omitempty" json:"opt-list,omitempty"`
}

type Para struct {
	XMLName      xml.Name      `xml:"para,omitempty" json:"para,omitempty"`
	Attrname     string        `xml:"name,attr"  json:",omitempty"`
	IsRequiredBy *IsRequiredBy `xml:"isRequiredBy,omitempty" json:"isRequiredBy,omitempty"`
	Text         string        `xml:",chardata" json:",omitempty"`
}

type RecDef struct {
	XMLName      xml.Name      `xml:"record-definition,omitempty" json:"record-definition,omitempty"`
	Attrprefix   string        `xml:"prefix,attr"  json:",omitempty"`
	Attrversion  string        `xml:"version,attr"  json:",omitempty"`
	Attrflat     string        `xml:"flat,attr"  json:",omitempty"`
	Namespaces   *Namespaces   `xml:"namespaces,omitempty" json:"namespaces,omitempty"`
	Functions    *Functions    `xml:"functions,omitempty" json:"functions,omitempty"`
	Attrs        *Attrs        `xml:"attrs,omitempty" json:"attrs,omitempty"`
	Root         *Root         `xml:"root,omitempty" json:"root,omitempty"`
	FieldMarkers *FieldMarkers `xml:"field-markers,omitempty" json:"field-markers,omitempty"`
	Opts         *Opts         `xml:"opts,omitempty" json:"opts,omitempty"`
	Docs         *Docs         `xml:"docs,omitempty" json:"docs,omitempty"`
}

type Root struct {
	XMLName     xml.Name     `xml:"root,omitempty" json:"root,omitempty"`
	Attrtag     string       `xml:"tag,attr"  json:",omitempty"`
	Elem        []*Elem      `xml:"elem,omitempty" json:"elem,omitempty"`
	NodeMapping *NodeMapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type SampleInput struct {
	XMLName xml.Name  `xml:"sample-input,omitempty" json:"sample-input,omitempty"`
	String  []*String `xml:"string,omitempty" json:"string,omitempty"`
}

type String struct {
	XMLName xml.Name `xml:"string,omitempty" json:"string,omitempty"`
	Text    string   `xml:",chardata" json:",omitempty"`
}
