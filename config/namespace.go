// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"strings"
	"sync"
)

// NameSpace is a container for Namespaces base URLs and prefixes
// This is used by resolving namespaces in the RDF conversions
type NameSpace struct {
	Base   string `json:"base"`
	Prefix string `json:"prefix"`
}

// NameSpaceMap contains all the namespaces
type NameSpaceMap struct {
	sync.RWMutex
	prefix2base map[string]string
	base2prefix map[string]string
}

// NewNameSpaceMap creates a new NameSpaceMap
func NewNameSpaceMap() *NameSpaceMap {
	return &NameSpaceMap{
		prefix2base: make(map[string]string),
		base2prefix: make(map[string]string),
	}
}

// Len counts the number of keys in the Map
func (n *NameSpaceMap) Len() (int, int) {
	return len(n.prefix2base), len(n.base2prefix)
}

// Add adds a namespace to the namespace Map
func (n *NameSpaceMap) Add(prefix, base string) {
	n.Lock()
	n.prefix2base[prefix] = base
	n.base2prefix[base] = prefix
	n.Unlock()
}

// AddNameSpace is a convenience function to add NameSpace objects to the Map
func (n *NameSpaceMap) AddNameSpace(ns NameSpace) {
	n.Add(ns.Prefix, ns.Base)
}

// Load loads the namespaces from the config object
func (n *NameSpaceMap) Load(c *RawConfig) {
	for _, ns := range c.NameSpaces {
		n.AddNameSpace(ns)
	}
}

// NewConfigNameSpaceMap creates a map from the NameSpaces defined in the config
func NewConfigNameSpaceMap(c *RawConfig) *NameSpaceMap {
	nsMap := NewNameSpaceMap()
	nsMap.setDefaultNameSpaces()
	nsMap.Load(c)
	return nsMap
}

// GetBaseURI returns the base URI from the prefix
func (n *NameSpaceMap) GetBaseURI(prefix string) (base string, ok bool) {
	n.RLock()
	base, ok = n.prefix2base[prefix]
	n.RUnlock()
	return base, ok
}

// GetPrefix returns the prefix for a base URI
func (n *NameSpaceMap) GetPrefix(baseURI string) (prefix string, ok bool) {
	n.RLock()
	prefix, ok = n.base2prefix[baseURI]
	n.RUnlock()
	return prefix, ok
}

// DeletePrefix removes a namespace from the NameSpaceMap
func (n *NameSpaceMap) DeletePrefix(prefix string) {
	n.Lock()
	base, ok := n.prefix2base[prefix]
	if ok {
		delete(n.base2prefix, base)
	}
	delete(n.prefix2base, prefix)
	n.Unlock()
}

// DeleteBaseURI removes a namespace from the NameSpaceMap
func (n *NameSpaceMap) DeleteBaseURI(base string) {
	n.Lock()
	prefix, ok := n.base2prefix[base]
	if ok {
		delete(n.prefix2base, prefix)
	}
	delete(n.base2prefix, base)
	n.Unlock()
}

// ByPrefix returns the map with prefixes as keys
func (n *NameSpaceMap) ByPrefix() map[string]string {
	return n.prefix2base
}

// SplitURI takes a given URI and splits it into a base URI and a local name
func SplitURI(uri string) (base string, name string) {
	index := strings.LastIndex(uri, "#") + 1

	if index > 0 {
		return uri[:index], uri[index:]
	}

	index = strings.LastIndex(uri, "/") + 1

	if index > 0 {
		return uri[:index], uri[index:]
	}

	return "", uri
}

// GetSearchLabel returns the search label for a Predicate URI
func (n *NameSpaceMap) GetSearchLabel(uri string) (string, error) {
	base, label := SplitURI(uri)
	prefix, ok := n.GetPrefix(base)
	if !ok {
		return "", fmt.Errorf("no prefix found in ns map for %s", base)
	}
	return fmt.Sprintf("%s_%s", prefix, label), nil
}

var defaultNameSpaces = map[string]string{
	"abc":         "http://www.ab-c.nl/",
	"abm":         "http://purl.org/abm/sen",
	"cc":          "http://creativecommons.org/ns#",
	"custom":      "http://www.delving.eu/namespaces/custom/",
	"dbpedia-owl": "http://dbpedia.org/ontology/",
	"dc":          "http://purl.org/dc/elements/1.1/",
	"dcterms":     "http://purl.org/dc/terms/",
	"delving":     "http://schemas.delving.eu/",
	"devmode":     "http://localhost:8000/resource/",
	"edm":         "http://www.europeana.eu/schemas/edm/",
	"europeana":   "http://www.europeana.eu/schemas/ese/",
	"foaf":        "http://xmlns.com/foaf/0.1/",
	"gn":          "http://www.geonames.org/ontology#",
	"icn":         "http://www.icn.nl/schemas/icn/",
	"narthex":     "http://schemas.delving.eu/narthex/terms/",
	"nave":        "http://schemas.delving.eu/nave/terms/",
	"ore":         "http://www.openarchives.org/ore/terms/",
	"owl":         "http://www.w3.org/2002/07/owl#",
	"raw":         "http://delving.eu/namespaces/raw/",
	"rda":         "http://rdvocab.info/ElementsGr2/",
	"rdf":         "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":        "http://www.w3.org/2000/01/rdf-schema#",
	"skos":        "http://www.w3.org/2004/02/skos/core#",
	"tib":         "http://schemas.delving.eu/resource/ns/tib/",
	"wgs84_pos":   "http://www.w3.org/2003/01/geo/wgs84_pos#",
}

// setDefaultNameSpaces sets the default namespaces that are supported
func (n *NameSpaceMap) setDefaultNameSpaces() {
	for prefix, baseURI := range defaultNameSpaces {
		n.Add(prefix, baseURI)
	}
}
