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

package ead

import (
	"bytes"
	"encoding/json"
)

type FlowType int

const (
	LineBreak FlowType = iota
	Inline
	Next
)

type DataType int

const (
	Paragraph DataType = iota
	Date
	Image
	Link
	List
	ListItem
	DefItem
	ListLabel
	Table
	TableHead
	TableRow
	TableCel
	Unit
	Language
	Repository
	Nested
	Enum
	Section
	SubSection
	Note
	ChronList
	ChronItem
	Event
)

var dataTypetoString = map[DataType]string{
	Paragraph:  "paragraph",
	Date:       "date",
	Image:      "image",
	Link:       "link",
	List:       "list",
	ListItem:   "listitem",
	DefItem:    "defitem",
	ListLabel:  "listlabel",
	Table:      "table",
	TableHead:  "tablehead",
	TableRow:   "tablerow",
	TableCel:   "tablecel",
	Unit:       "unit",
	Language:   "language",
	Repository: "repository",
	Nested:     "nested",
	Enum:       "enum",
	Section:    "section",
	SubSection: "subsection",
	Note:       "note",
	ChronList:  "chronlist",
	ChronItem:  "chronitem",
	Event:      "event",
}

var dataTypetoID = map[string]DataType{
	"paragraph":  Paragraph,
	"date":       Date,
	"image":      Image,
	"link":       Link,
	"list":       List,
	"listitem":   ListItem,
	"defitem":    DefItem,
	"listlabel":  ListLabel,
	"table":      Table,
	"tablerow":   TableRow,
	"tablehead":  TableHead,
	"tablecel":   TableCel,
	"unit":       Unit,
	"language":   Language,
	"repository": Repository,
	"nested":     Nested,
	"enum":       Enum,
	"section":    Section,
	"subsection": SubSection,
	"note":       Note,
	"chronlist":  ChronList,
	"chronitem":  ChronItem,
	"event":      Event,
}

// MarshalJSON marshals the enum as a quoted json string
func (dt DataType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(dataTypetoString[dt])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (dt *DataType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Paragraph' in this case.
	*dt = dataTypetoID[j]
	return nil
}

var flowTypetoString = map[FlowType]string{
	LineBreak: "linebreak",
	Inline:    "inline",
	Next:      "next",
}

var flowTypetoID = map[string]FlowType{
	"linebreak": LineBreak,
	"inline":    Inline,
	"next":      Next,
}

// MarshalJSON marshals the enum as a quoted json string
func (ft FlowType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(flowTypetoString[ft])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (ft *FlowType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Paragraph' in this case.
	*ft = flowTypetoID[j]
	return nil
}
