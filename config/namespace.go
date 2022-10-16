// Copyright 2017 Delving B.V.
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
	"log"
	"strings"
	"sync"

	"github.com/OneOfOne/xxhash"
)

const (
	ebuCoreURN = "urn:ebu:metadata-schema:ebuCore_2014"
)

// Namespace is a container for Namespaces base URLs and prefixes
// This is used by resolving namespaces in the RDF conversions
type Namespace struct {
	Base   string `json:"base"`
	Prefix string `json:"prefix"`
}

// NamespaceMap contains all the namespaces
type NamespaceMap struct {
	rw          sync.RWMutex
	prefix2base map[string]string
	base2prefix map[string]string
}

// NewNamespaceMap creates a new NameSpaceMap
func NewNamespaceMap() *NamespaceMap {
	return &NamespaceMap{
		prefix2base: make(map[string]string),
		base2prefix: make(map[string]string),
	}
}

// Len counts the number of keys in the Map
func (n *NamespaceMap) Len() (prefixes, baseURIs int) {
	return len(n.prefix2base), len(n.base2prefix)
}

// Add adds a namespace to the namespace Map
func (n *NamespaceMap) Add(prefix, base string) {
	n.rw.Lock()
	n.prefix2base[prefix] = base
	n.base2prefix[base] = prefix
	n.rw.Unlock()
}

// AddNamespace is a convenience function to add NameSpace objects to the Map
func (n *NamespaceMap) AddNamespace(ns Namespace) {
	n.Add(ns.Prefix, ns.Base)
}

// Load loads the namespaces from the config object
func (n *NamespaceMap) Load(c *RawConfig) {
	for _, ns := range c.Namespaces {
		n.AddNamespace(ns)
	}
}

// NewConfigNamespaceMap creates a map from the NameSpaces defined in the config
func NewConfigNamespaceMap(c *RawConfig) *NamespaceMap {
	nsMap := NewNamespaceMap()
	nsMap.setDefaultNamespaces()
	nsMap.Load(c)

	return nsMap
}

// GetBaseURI returns the base URI from the prefix
func (n *NamespaceMap) GetBaseURI(prefix string) (base string, ok bool) {
	n.rw.RLock()
	base, ok = n.prefix2base[prefix]
	n.rw.RUnlock()

	return base, ok
}

// GetPrefix returns the prefix for a base URI
func (n *NamespaceMap) GetPrefix(baseURI string) (prefix string, ok bool) {
	n.rw.RLock()
	prefix, ok = n.base2prefix[baseURI]
	n.rw.RUnlock()

	return prefix, ok
}

// DeletePrefix removes a namespace from the NameSpaceMap
func (n *NamespaceMap) DeletePrefix(prefix string) {
	n.rw.Lock()

	base, ok := n.prefix2base[prefix]
	if ok {
		delete(n.base2prefix, base)
	}

	delete(n.prefix2base, prefix)
	n.rw.Unlock()
}

// DeleteBaseURI removes a namespace from the NameSpaceMap
func (n *NamespaceMap) DeleteBaseURI(base string) {
	n.rw.Lock()

	prefix, ok := n.base2prefix[base]
	if ok {
		delete(n.prefix2base, prefix)
	}

	delete(n.base2prefix, base)
	n.rw.Unlock()
}

// ByPrefix returns the map with prefixes as keys
func (n *NamespaceMap) ByPrefix() map[string]string {
	return n.prefix2base
}

// SplitURI takes a given URI and splits it into a base URI and a local name
func SplitURI(uri string) (base, name string) {
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
func (n *NamespaceMap) GetSearchLabel(uri string) (string, error) {
	if strings.HasPrefix(uri, ebuCoreURN) {
		uri = strings.TrimPrefix(uri, ebuCoreURN)
		uri = strings.TrimLeft(uri, "/")
		uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
	}

	base, label := SplitURI(uri)

	prefix, ok := n.GetPrefix(base)
	if !ok {
		hash := xxhash.Checksum64([]byte(base))
		prefix = fmt.Sprintf("%016x", hash)
		n.Add(prefix, base)
		// return "", fmt.Errorf("no prefix found in ns map for %s + %s", base, label)
		log.Printf("Added default prefix %s for %s", prefix, base)
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
	"ebucore":     "http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#",
	"foaf":        "http://xmlns.com/foaf/0.1/",
	"gn":          "http://www.geonames.org/ontology#",
	"icn":         "http://www.icn.nl/schemas/icn/",
	"narthex":     "http://schemas.delving.eu/narthex/terms/",
	"nave":        "http://schemas.delving.eu/nave/terms/",
	"rapid":       "http://data.rapid.org/",
	"ore":         "http://www.openarchives.org/ore/terms/",
	"owl":         "http://www.w3.org/2002/07/owl#",
	"raw":         "http://delving.eu/namespaces/raw/",
	"rda":         "http://rdvocab.info/ElementsGr2/",
	"rdf":         "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":        "http://www.w3.org/2000/01/rdf-schema#",
	"skos":        "http://www.w3.org/2004/02/skos/core#",
	"tib":         "http://schemas.delving.eu/resource/ns/tib/",
	"wgs84_pos":   "http://www.w3.org/2003/01/geo/wgs84_pos#",
	"naa":         "https://archief.nl/def/",
	"ead-rdf":     "https://archief.nl/def/ead/",
	"ead-mets":    "https://archief.nl/def/mets/",
}

// setDefaultNamespaces sets the default namespaces that are supported
func (n *NamespaceMap) setDefaultNamespaces() {
	for prefix, baseURI := range defaultNameSpaces {
		n.Add(prefix, baseURI)
	}
}
