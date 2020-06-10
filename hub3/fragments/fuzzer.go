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

package fragments

import (
	"encoding/json"
	"encoding/xml"
	fmt "fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/delving/hub3/config"
	fuzz "github.com/google/gofuzz"
	r "github.com/kiivihal/rdf2go"
)

// Fuzzer is the builder for building fuzzed records based on a record definition
type Fuzzer struct {
	nm       *config.NameSpaceMap
	resource []*FuzzResource
	BaseURL  string
	f        *fuzz.Fuzzer
}

// CreateRecords for n number of fuzzed records
func (fz *Fuzzer) CreateRecords(n int) ([]string, error) {
	records := []string{}
	for i := 0; i < n; i++ {
		ld := []map[string]interface{}{}
		fr := &FuzzRecord{fz, i, NewEmptyResourceMap()}
		err := fr.AddTriples()
		if err != nil {
			return nil, err
		}
		for _, rsc := range fr.rm.ResourcesList(nil) {
			ld = append(ld, rsc.GenerateJSONLD())
		}
		graph, err := json.Marshal(ld)
		if err != nil {
			return nil, err
		}
		records = append(records, string(graph))
	}
	return records, nil
}

type FuzzRecord struct {
	fz   *Fuzzer
	seed int
	rm   *ResourceMap
}

func (fr *FuzzRecord) AddTriples() error {
	order := 0
	for _, rsc := range fr.fz.resource {
		subject := fr.fz.NewURI(rsc.SearchLabel, fr.seed)
		// add type
		t := r.NewTriple(
			r.NewResource(subject),
			r.NewResource(RDFType),
			r.NewResource(rsc.Type),
		)
		fr.rm.AppendOrderedTriple(t, false, order)
		order++

		// add entries
		for _, fe := range rsc.Predicates {
			for _, t := range fr.fz.CreateTriples(subject, fe) {
				fr.rm.AppendOrderedTriple(t, false, order)
				order++
			}
		}
	}
	return nil
}

// ExpandNameSpace converts prefix xml label to fully qualified URLs
func (fz *Fuzzer) ExpandNameSpace(xmlLabel string) (string, error) {
	if xmlLabel == "" {
		return "", fmt.Errorf("can't expand empty string")
	}
	parts := strings.Split(xmlLabel, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed xml label: %s", xmlLabel)
	}
	base, ok := fz.nm.GetBaseURI(parts[0])
	if !ok {
		return "", fmt.Errorf("unable to find namespace: %s", parts[0])
	}
	return fmt.Sprintf("%s%s", base, parts[1]), nil
}

// NewURI created new Fuzzed URI. When the label is given that is used for the URI
func (fz *Fuzzer) NewURI(label string, key int) string {
	uri := fmt.Sprintf("%s/%s/%d", strings.TrimSuffix(fz.BaseURL, "/"), label, key)
	return uri
}

// NewString creates a new fuzzed string
func (fz *Fuzzer) NewString(label string) string {
	if label == "" {
		var s string
		fz.f.Fuzz(&s)
		return s
	}
	return fmt.Sprintf("%s %d", label, rand.Intn(10))
}

// FuzzResource holds all the information to generate a Fuzzed RDF-resource
type FuzzResource struct {
	Subject     string       `json:"subject"`
	Type        string       `json:"type"`
	SearchLabel string       `json:"searchLabel"`
	Predicates  []*FuzzEntry `json:"predicates"`
	Order       int          `json:"order"`
}

// NewFuzzResource creates a FuzzResource
func (fz *Fuzzer) NewFuzzResource(order int, elem *Celem) (*FuzzResource, error) {
	fr := &FuzzResource{Order: order}
	if elem.Attrtag != "" {
		rType, err := fz.ExpandNameSpace(elem.Attrtag)
		if err != nil {
			return nil, err
		}
		fr.Type = rType
		searchLabel, err := fz.nm.GetSearchLabel(rType)
		if err != nil {
			return nil, err
		}
		fr.SearchLabel = searchLabel

	}

	for idx, cElem := range elem.Celem {
		// todo how to deal with type
		fe, err := fz.NewFuzzEntry(idx, cElem)
		if err != nil {
			return nil, err
		}
		fr.Predicates = append(fr.Predicates, fe)
	}

	return fr, nil
}

// NewFuzzEntry creates a FuzzEntry from a child elem in the Record Definition
func (fz *Fuzzer) NewFuzzEntry(order int, elem *Celem) (*FuzzEntry, error) {
	tags := strings.Split(elem.Attrattrs, ",")
	predicate, err := fz.ExpandNameSpace(elem.Attrtag)
	if err != nil {
		return nil, err
	}
	fe := &FuzzEntry{
		Predicate:   predicate,
		Tags:        tags,
		Order:       order,
		SearchLabel: strings.Replace(elem.Attrtag, ":", "_", 0),
	}
	return fe, nil
}

// FuzzEntry holds all the information to generate a Fuzzed Triple
type FuzzEntry struct {
	Predicate   string   `json:"predicate"`
	Tags        []string `json:"tags"`
	Order       int      `json:"order"`
	SearchLabel string   `json:"searchLabel"`
}

// CreateTriples creates fuzzed Triples for a FuzzEntry
func (fz *Fuzzer) CreateTriples(subject string, fe *FuzzEntry) []*r.Triple {
	triples := []*r.Triple{}
	for _, tag := range fe.Tags {
		var t *r.Triple
		switch tag {
		case "rdf:resource":
			t = r.NewTriple(
				r.NewResource(subject),
				r.NewResource(fe.Predicate),
				r.NewResource(fz.NewURI(fe.SearchLabel, rand.Intn(10))),
			)
		case "xml:lang":
			t = r.NewTriple(
				r.NewResource(subject),
				r.NewResource(fe.Predicate),
				r.NewLiteral(fz.NewString(fe.SearchLabel)),
			)
		}
		if t != nil {
			triples = append(triples, t)
		}
	}
	return triples
}

// NewFuzzer creates a Fuzzer for creating Records based on the Record Definition
func NewFuzzer(recDef *Crecord_dash_definition) (*Fuzzer, error) {
	fz := &Fuzzer{
		nm: config.NewNameSpaceMap(),
		f:  fuzz.New(),
	}

	// add namespaces
	for _, ns := range recDef.Cnamespaces.Cnamespace {
		fz.nm.Add(ns.Attrprefix, ns.Attruri)
	}

	// add Fuzz Resources
	for idx, elem := range recDef.Croot.Celem {
		fr, err := fz.NewFuzzResource(idx, elem)
		if err != nil {
			return nil, err
		}
		fz.resource = append(fz.resource, fr)
	}

	return fz, nil
}

// NewRecDef takes a []byte and creates a record definition
func NewRecDef(r io.Reader) (*Crecord_dash_definition, error) {
	var naa Crecord_dash_definition
	if err := xml.NewDecoder(r).Decode(&naa); err != nil {
		return nil, err
	}
	return &naa, nil
}

//// Generated structs for parsing the XML record definition
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
	Attrattrs          string              `xml:"attrs,attr"  json:",omitempty"`
	Attrtag            string              `xml:"tag,attr"  json:",omitempty"`
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

type Cnode_dash_mapping struct {
	XMLName           xml.Name           `xml:"node-mapping,omitempty" json:"node-mapping,omitempty"`
	AttrinputPath     string             `xml:"inputPath,attr"  json:",omitempty"`
	AttroutputPath    string             `xml:"outputPath,attr"  json:",omitempty"`
	Cgroovy_dash_code *Cgroovy_dash_code `xml:"groovy-code,omitempty" json:"groovy-code,omitempty"`
}

type Copts struct {
	XMLName xml.Name `xml:"opts,omitempty" json:"opts,omitempty"`
}

type Cpara struct {
	XMLName  xml.Name `xml:"para,omitempty" json:"para,omitempty"`
	Attrname string   `xml:"name,attr"  json:",omitempty"`
	string   string   `xml:",chardata" json:",omitempty"`
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

type Cstring struct {
	XMLName xml.Name `xml:"string,omitempty" json:"string,omitempty"`
	string  string   `xml:",chardata" json:",omitempty"`
}

///////////////////////////
