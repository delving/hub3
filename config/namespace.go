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
