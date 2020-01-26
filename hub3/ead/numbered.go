package ead

import "encoding/xml"

type Cc01 struct {
	XMLName         xml.Name         `xml:"c01,omitempty" json:"c01,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc02          `xml:"c02,omitempty" json:"c02,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc01) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc01) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc01) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc01) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc01) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc01) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc01) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc01) GetOdd() []*Codd                      { return c.Codd }
func (c Cc01) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc01) GetMaterial() string                  { return "" }
func (c Cc01) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc01) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc02 struct {
	XMLName         xml.Name         `xml:"c02,omitempty" json:"c02,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc03          `xml:"c03,omitempty" json:"c03,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc02) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc02) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc02) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc02) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc02) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc02) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc02) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc02) GetOdd() []*Codd                      { return c.Codd }
func (c Cc02) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc02) GetMaterial() string                  { return "" }
func (c Cc02) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc02) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc03 struct {
	XMLName         xml.Name         `xml:"c03,omitempty" json:"c03,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc04          `xml:"c04,omitempty" json:"c04,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc03) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc03) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc03) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc03) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc03) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc03) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc03) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc03) GetOdd() []*Codd                      { return c.Codd }
func (c Cc03) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc03) GetMaterial() string                  { return "" }
func (c Cc03) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc03) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc04 struct {
	XMLName         xml.Name         `xml:"c04,omitempty" json:"c04,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc05          `xml:"c05,omitempty" json:"c05,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc04) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc04) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc04) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc04) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc04) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc04) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc04) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc04) GetOdd() []*Codd                      { return c.Codd }
func (c Cc04) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc04) GetMaterial() string                  { return "" }
func (c Cc04) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc04) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc05 struct {
	XMLName         xml.Name         `xml:"c05,omitempty" json:"c05,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc06          `xml:"c06,omitempty" json:"c06,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc05) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc05) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc05) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc05) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc05) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc05) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc05) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc05) GetOdd() []*Codd                      { return c.Codd }
func (c Cc05) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc05) GetMaterial() string                  { return "" }
func (c Cc05) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc05) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc06 struct {
	XMLName         xml.Name         `xml:"c06,omitempty" json:"c06,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc07          `xml:"c07,omitempty" json:"c07,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc06) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc06) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc06) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc06) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc06) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc06) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc06) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc06) GetOdd() []*Codd                      { return c.Codd }
func (c Cc06) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc06) GetMaterial() string                  { return "" }
func (c Cc06) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc06) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc07 struct {
	XMLName         xml.Name         `xml:"c07,omitempty" json:"c07,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc08          `xml:"c08,omitempty" json:"c08,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc07) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc07) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc07) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc07) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc07) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc07) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc07) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc07) GetOdd() []*Codd                      { return c.Codd }
func (c Cc07) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc07) GetMaterial() string                  { return "" }
func (c Cc07) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc07) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc08 struct {
	XMLName         xml.Name         `xml:"c08,omitempty" json:"c08,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc09          `xml:"c09,omitempty" json:"c09,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc08) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc08) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc08) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc08) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc08) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc08) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc08) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc08) GetOdd() []*Codd                      { return c.Codd }
func (c Cc08) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc08) GetMaterial() string                  { return "" }
func (c Cc08) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc08) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc09 struct {
	XMLName         xml.Name         `xml:"c09,omitempty" json:"c09,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc10          `xml:"c10,omitempty" json:"c10,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc09) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc09) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc09) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc09) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc09) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc09) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc09) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc09) GetOdd() []*Codd                      { return c.Codd }
func (c Cc09) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc09) GetMaterial() string                  { return "" }
func (c Cc09) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc09) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc10 struct {
	XMLName         xml.Name         `xml:"c10,omitempty" json:"c10,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc11          `xml:"c11,omitempty" json:"c11,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc10) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc10) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc10) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc10) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc10) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc10) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc10) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc10) GetOdd() []*Codd                      { return c.Codd }
func (c Cc10) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc10) GetMaterial() string                  { return "" }
func (c Cc10) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc10) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc11 struct {
	XMLName         xml.Name         `xml:"c11,omitempty" json:"c11,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc12          `xml:"c12,omitempty" json:"c12,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc11) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc11) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc11) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc11) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc11) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc11) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc11) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc11) GetOdd() []*Codd                      { return c.Codd }
func (c Cc11) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc11) GetMaterial() string                  { return "" }
func (c Cc11) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc11) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc12 struct {
	XMLName         xml.Name         `xml:"c12,omitempty" json:"c12,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc12          `xml:"c13,omitempty" json:"c13,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc12) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc12) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc12) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc12) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc12) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc12) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc12) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc12) GetOdd() []*Codd                      { return c.Codd }
func (c Cc12) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc12) GetMaterial() string                  { return "" }
func (c Cc12) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc12) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc13 struct {
	XMLName         xml.Name         `xml:"c13,omitempty" json:"c13,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc14          `xml:"c14,omitempty" json:"c14,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc13) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc13) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc13) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc13) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc13) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc13) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc13) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc13) GetOdd() []*Codd                      { return c.Codd }
func (c Cc13) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc13) GetMaterial() string                  { return "" }
func (c Cc13) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc13) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc14 struct {
	XMLName         xml.Name         `xml:"c14,omitempty" json:"c14,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc15          `xml:"c15,omitempty" json:"c15,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc14) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc14) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc14) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc14) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc14) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc14) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc14) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc14) GetOdd() []*Codd                      { return c.Codd }
func (c Cc14) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc14) GetMaterial() string                  { return "" }
func (c Cc14) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc14) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc15 struct {
	XMLName         xml.Name         `xml:"c15,omitempty" json:"c15,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc16          `xml:"c16,omitempty" json:"c16,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc15) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc15) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc15) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc15) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc15) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc15) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc15) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc15) GetOdd() []*Codd                      { return c.Codd }
func (c Cc15) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc15) GetMaterial() string                  { return "" }
func (c Cc15) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc15) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc16 struct {
	XMLName         xml.Name         `xml:"c16,omitempty" json:"c16,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc17          `xml:"c17,omitempty" json:"c17,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc16) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc16) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc16) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc16) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc16) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc16) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc16) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc16) GetOdd() []*Codd                      { return c.Codd }
func (c Cc16) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc16) GetMaterial() string                  { return "" }
func (c Cc16) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc16) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc17 struct {
	XMLName         xml.Name         `xml:"c17,omitempty" json:"c17,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc18          `xml:"c18,omitempty" json:"c18,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc17) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc17) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc17) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc17) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc17) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc17) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc17) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc17) GetOdd() []*Codd                      { return c.Codd }
func (c Cc17) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc17) GetMaterial() string                  { return "" }
func (c Cc17) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc17) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc18 struct {
	XMLName         xml.Name         `xml:"c18,omitempty" json:"c18,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc19          `xml:"c19,omitempty" json:"c19,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc18) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc18) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc18) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc18) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc18) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc18) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc18) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc18) GetOdd() []*Codd                      { return c.Codd }
func (c Cc18) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc18) GetMaterial() string                  { return "" }
func (c Cc18) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc18) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc19 struct {
	XMLName         xml.Name         `xml:"c19,omitempty" json:"c19,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc20          `xml:"c20,omitempty" json:"c20,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc19) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc19) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc19) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc19) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc19) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc19) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc19) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc19) GetOdd() []*Codd                      { return c.Codd }
func (c Cc19) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc19) GetMaterial() string                  { return "" }
func (c Cc19) GetNested() []CLevel {
	levels := make([]CLevel, len(c.Nested))
	for i, v := range c.Nested {
		levels[i] = CLevel(v)
	}
	return levels
}
func (c Cc19) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}

type Cc20 struct {
	XMLName         xml.Name         `xml:"c20,omitempty" json:"c20,omitempty"`
	Attrlevel       string           `xml:"level,attr"  json:",omitempty"`
	Attrotherlevel  string           `xml:"otherlevel,attr"  json:",omitempty"`
	Attraltrender   string           `xml:"altrender,attr"  json:",omitempty"`
	Caccessrestrict *Caccessrestrict `xml:"accessrestrict,omitempty" json:"accessrestrict,omitempty"`
	Nested          []*Cc            `xml:"c,omitempty" json:"c,omitempty"`
	Cdid            *Cdid            `xml:"did,omitempty" json:"did,omitempty"`
	Cscopecontent   *Cscopecontent   `xml:"scopecontent,omitempty" json:"scopecontent,omitempty"`
	Codd            []*Codd          `xml:"odd,omitempty" json:"odd,omitempty"`
	Cphystech       []*Cphystech     `xml:"phystech,omitempty" json:"phystech,omitempty"`
	Ccontrolaccess  *Ccontrolaccess  `xml:"controlaccess,omitempty" json:"controlaccess,omitempty"`
}

func (c Cc20) GetXMLName() xml.Name                 { return c.XMLName }
func (c Cc20) GetAttrlevel() string                 { return c.Attrlevel }
func (c Cc20) GetAttrotherlevel() string            { return c.Attrotherlevel }
func (c Cc20) GetAttraltrender() string             { return c.Attraltrender }
func (c Cc20) GetCaccessrestrict() *Caccessrestrict { return c.Caccessrestrict }
func (c Cc20) GetCdid() *Cdid                       { return c.Cdid }
func (c Cc20) GetScopeContent() *Cscopecontent      { return c.Cscopecontent }
func (c Cc20) GetOdd() []*Codd                      { return c.Codd }
func (c Cc20) GetPhystech() []*Cphystech            { return c.Cphystech }
func (c Cc20) GetMaterial() string                  { return "" }
func (c Cc20) GetNested() []CLevel {
	levels := make([]CLevel, 0)
	return levels
}
func (c Cc20) GetGenreform() string {
	if c.Ccontrolaccess != nil && c.Ccontrolaccess.Cgenreform != nil {
		return c.Ccontrolaccess.Cgenreform.Genreform
	}

	return ""
}
