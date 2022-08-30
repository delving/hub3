package internal

import "encoding/xml"

type Cattr struct {
	XMLName            xml.Name            `xml:"attr,omitempty" json:"attr,omitempty"`
	Attrtag            string              `xml:"tag,attr"  json:",omitempty"`
	AttruriCheck       string              `xml:"uriCheck,attr"  json:",omitempty"`
	Cnode_dash_mapping *Cnode_dash_mapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type Cattrs struct {
	XMLName xml.Name `xml:"attrs,omitempty" json:"attrs,omitempty"`
	Cattr   []*Cattr `xml:"attr,omitempty" json:"attr,omitempty"`
}

type Cdoc struct {
	XMLName  xml.Name `xml:"doc,omitempty" json:"doc,omitempty"`
	Attrpath string   `xml:"path,attr"  json:",omitempty"`
	Cpara    []*Cpara `xml:"para,omitempty" json:"para,omitempty"`
}

type Cdocs struct {
	XMLName xml.Name `xml:"docs,omitempty" json:"docs,omitempty"`
	Cdoc    []*Cdoc  `xml:"doc,omitempty" json:"doc,omitempty"`
}

type Celem struct {
	XMLName            xml.Name            `xml:"elem,omitempty" json:"elem,omitempty"`
	Attrattribute      string              `xml:"attribute,attr,omitempty"  json:",omitempty"`
	Attrattrs          string              `xml:"attrs,attr,omitempty"  json:",omitempty"`
	Attrfunction       string              `xml:"function,attr,omitempty"  json:",omitempty"`
	Attrhidden         string              `xml:"hidden,attr,omitempty"  json:",omitempty"`
	Attrtag            string              `xml:"tag,attr,omitempty"  json:",omitempty"`
	Attrlabel          string              `xml:"label,attr,omitempty"  json:",omitempty"`
	Attrunmappable     string              `xml:"unmappable,attr,omitempty"  json:",omitempty"`
	AttruriCheck       string              `xml:"uriCheck,attr,omitempty"  json:",omitempty"`
	Cattr              []*Cattr            `xml:"attr,omitempty" json:"attr,omitempty"`
	Celem              []*Celem            `xml:"elem,omitempty" json:"elem,omitempty"`
	Cnode_dash_mapping *Cnode_dash_mapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type Cfield_dash_markers struct {
	XMLName xml.Name `xml:"field-markers,omitempty" json:"field-markers,omitempty"`
}

type Cfunctions struct {
	XMLName                xml.Name                  `xml:"functions,omitempty" json:"functions,omitempty"`
	Cmapping_dash_function []*Cmapping_dash_function `xml:"mapping-function,omitempty" json:"mapping-function,omitempty"`
}

type Cgroovy_dash_code struct {
	XMLName xml.Name   `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
	Cstring []*Cstring `xml:"string,omitempty" json:"string,omitempty"`
}

type CisRequiredBy struct {
	XMLName xml.Name `xml:"isRequiredBy,omitempty" json:"isRequiredBy,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

type Cmapping_dash_function struct {
	XMLName            xml.Name            `xml:"mapping-function,omitempty" json:"mapping-function,omitempty"`
	Attrname           string              `xml:"name,attr"  json:",omitempty"`
	Cgroovy_dash_code  *Cgroovy_dash_code  `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
	Csample_dash_input *Csample_dash_input `xml:"sample-input,omitempty" json:"sample-input,omitempty"`
}

type Cnamespace struct {
	XMLName    xml.Name `xml:"namespace,omitempty" json:"namespace,omitempty"`
	Attrprefix string   `xml:"prefix,attr"  json:",omitempty"`
	Attruri    string   `xml:"uri,attr"  json:",omitempty"`
}

type Cnamespaces struct {
	XMLName    xml.Name      `xml:"namespaces,omitempty" json:"namespaces,omitempty"`
	Cnamespace []*Cnamespace `xml:"namespace,omitempty" json:"namespace,omitempty"`
}

type Copt struct {
	XMLName   xml.Name `xml:"opt,omitempty" json:"opt,omitempty"`
	Attrvalue string   `xml:"value,attr"  json:",omitempty"`
}

type Copt_dash_list struct {
	XMLName         xml.Name `xml:"opt-list,omitempty" json:"opt-list,omitempty"`
	Attrdictionary  string   `xml:"dictionary,attr"  json:",omitempty"`
	AttrdisplayName string   `xml:"displayName,attr"  json:",omitempty"`
	Attrpath        string   `xml:"path,attr"  json:",omitempty"`
	Copt            []*Copt  `xml:"opt,omitempty" json:"opt,omitempty"`
}

type Copts struct {
	XMLName        xml.Name          `xml:"opts,omitempty" json:"opts,omitempty"`
	Copt_dash_list []*Copt_dash_list `xml:"opt-list,omitempty" json:"opt-list,omitempty"`
}

type Cpara struct {
	XMLName       xml.Name       `xml:"para,omitempty" json:"para,omitempty"`
	Attrname      string         `xml:"name,attr"  json:",omitempty"`
	CisRequiredBy *CisRequiredBy `xml:"isRequiredBy,omitempty" json:"isRequiredBy,omitempty"`
	string        string         `xml:",chardata" json:",omitempty"`
}

type Crecord_dash_definition struct {
	XMLName             xml.Name             `xml:"record-definition,omitempty" json:"record-definition,omitempty"`
	Attrflat            string               `xml:"flat,attr"  json:",omitempty"`
	Attrprefix          string               `xml:"prefix,attr"  json:",omitempty"`
	Attrversion         string               `xml:"version,attr"  json:",omitempty"`
	Cattrs              *Cattrs              `xml:"attrs,omitempty" json:"attrs,omitempty"`
	Cdocs               *Cdocs               `xml:"docs,omitempty" json:"docs,omitempty"`
	Cfield_dash_markers *Cfield_dash_markers `xml:"field-markers,omitempty" json:"field-markers,omitempty"`
	Cfunctions          *Cfunctions          `xml:"functions,omitempty" json:"functions,omitempty"`
	Cnamespaces         *Cnamespaces         `xml:"namespaces,omitempty" json:"namespaces,omitempty"`
	Copts               *Copts               `xml:"opts,omitempty" json:"opts,omitempty"`
	Croot               *Croot               `xml:"root,omitempty" json:"root,omitempty"`
}

type Croot struct {
	XMLName            xml.Name            `xml:"root,omitempty" json:"root,omitempty"`
	Attrtag            string              `xml:"tag,attr"  json:",omitempty"`
	Celem              []*Celem            `xml:"elem,omitempty" json:"elem,omitempty"`
	Cnode_dash_mapping *Cnode_dash_mapping `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
}

type Csample_dash_input struct {
	XMLName xml.Name   `xml:"sample-input,omitempty" json:"sample-input,omitempty"`
	Cstring []*Cstring `xml:"string,omitempty" json:"string,omitempty"`
}
