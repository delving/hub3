package internal

import "encoding/xml"

type Cdyn_dash_opt struct {
	XMLName   xml.Name `xml:"dyn-opt,omitempty" json:"dyn-opt,omitempty"`
	Attrpath  string   `xml:"path,attr"  json:",omitempty"`
	Attrvalue string   `xml:"value,attr"  json:",omitempty"`
}

type Cdyn_dash_opts struct {
	XMLName       xml.Name         `xml:"dyn-opts,omitempty" json:"dyn-opts,omitempty"`
	Cdyn_dash_opt []*Cdyn_dash_opt `xml:"dyn-opt,omitempty" json:"dyn-opt,omitempty"`
}

type Centry struct {
	XMLName xml.Name   `xml:"entry,omitempty" json:"entry,omitempty"`
	Cstring []*Cstring `xml:"string,omitempty" json:"string,omitempty"`
}

type Cfacts struct {
	XMLName xml.Name  `xml:"facts,omitempty" json:"facts,omitempty"`
	Centry  []*Centry `xml:"entry,omitempty" json:"entry,omitempty"`
}

type Cnode_dash_mapping struct {
	XMLName           xml.Name           `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
	AttrinputPath     string             `xml:"inputPath,attr"  json:",omitempty"`
	AttroutputPath    string             `xml:"outputPath,attr"  json:",omitempty"`
	Cgroovy_dash_code *Cgroovy_dash_code `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
}

type Cnode_dash_mappings struct {
	XMLName            xml.Name              `xml:"node-mappings,omitempty" json:"node-mappings,omitempty"`
	Cnode_dash_mapping []*Cnode_dash_mapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type Crec_dash_mapping struct {
	XMLName             xml.Name             `xml:"rec-mapping,omitempty" json:"rec-mapping,omitempty"`
	Attrlocked          string               `xml:"locked,attr"  json:",omitempty"`
	Attrprefix          string               `xml:"prefix,attr"  json:",omitempty"`
	AttrschemaVersion   string               `xml:"schemaVersion,attr"  json:",omitempty"`
	Cdyn_dash_opts      *Cdyn_dash_opts      `xml:"dyn-opts,omitempty" json:"dyn-opts,omitempty"`
	Cfacts              *Cfacts              `xml:"facts,omitempty" json:"facts,omitempty"`
	Cfunctions          *Cfunctions          `xml:"functions,omitempty" json:"functions,omitempty"`
	Cnode_dash_mappings *Cnode_dash_mappings `xml:"node-mappings,omitempty" json:"node-mappings,omitempty"`
}

type Cstring struct {
	XMLName xml.Name `xml:"string,omitempty" json:"string,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}
