// Copyright 2020 Delving B.V.
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

package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
)

var (
	ErrNameSpaceNotFound       = errors.New("namespace not found")
	ErrNameSpaceDuplicateEntry = errors.New("prefix and base stored in different entries")
	ErrNameSpaceNotValid       = errors.New("prefix or base not valid")
)

// URI represents a NameSpace URI.
type URI string

// String returns a string representation of URI
func (uri URI) String() string {
	return string(uri)
}

// Namespace is a container for URI conversions for RDF- and XML-namespaces.
type Namespace struct {

	// ID is the unique identifier of a namespace.
	// This identifier will be generated when empty.
	//
	// example: "f9ca66c45c2c0a61"
	ID string `json:"uuid"`

	// Base is the default base-URI for a namespace
	// example: "http://purl.org/dc/elements/1.1/"
	Base string `json:"base"`

	// Prefix is the default short version that identifies the base-URI
	// example: "dc"
	Prefix string `json:"prefix"`

	// BaseAlt are alternative base-URI for the same prefix.
	// Sometimes historically the base-URIs for a namespace changes and we still
	// have to correctly resolve both.
	//
	// example: "["http://purl.org/dc/elements/1.1/"]"
	BaseAlt []string `json:"baseAlt,omitempty"`

	// PrefixAlt are altenative prefixes for the default base URI.
	// Different content-providers and organizations have at time selected alternative
	// prefixes for the same base URI. We need to support both entry-points.
	//
	// example: "["dce", "dc11"]"
	PrefixAlt []string `json:"prefixAlt,omitempty"`

	// Schema is an URL to the RDFS or OWL definition of namespace
	// example: "https://www.dublincore.org/specifications/dublin-core/dcmi-terms/dublin_core_terms.ttl"
	Schema string `json:"schema,omitempty"`

	// Temporary defines if the NameSpace has been given a temporary prefix because
	// only the base-URI was known when the NameSpace was created.
	// Namespaces with prefix collissions will also be given a temporary prefix
	//
	// example: "true"
	// default: "false"
	Temporary bool `json:"temporary,omitempty"`

	// TODO(kiivihal): add function for custom hashing similar to isIdentRune
}

func (ns *Namespace) XMLNS() string {
	return "xmlns:" + ns.Prefix
}

func (ns *Namespace) String() string {
	return fmt.Sprintf("%s: %s", ns.Prefix, ns.Base)
}

// SplitURI takes a given URI and splits it into a base-URI and a localname.
// When the URI can't be split, the full URI is returned as the label with an
// empty base.
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

// AddPrefix adds a prefix to the list of prefix alternatives.
//
// When the prefix is already present in PrefixAlt no error is thrown.
func (ns *Namespace) AddPrefix(prefix string) error {
	if ns.Temporary {
		ns.Temporary = false
		ns.Prefix = prefix

		return nil
	}

	for _, p := range ns.PrefixAlt {
		if p == prefix {
			return nil
		}
	}

	ns.PrefixAlt = append(ns.PrefixAlt, prefix)

	return nil
}

// AddBase adds a base-URI to the list of base alternatives.
//
// When the base-URI is already present in BaseAlt no error is thrown.
func (ns *Namespace) AddBase(base string) error {
	for _, b := range ns.BaseAlt {
		if b == base {
			return nil
		}
	}

	ns.BaseAlt = append(ns.BaseAlt, base)

	return nil
}

// Prefixes returns all namespace prefix linked to this NameSpace.
// This includes the default Prefix and all alternative prefixes.
func (ns *Namespace) Prefixes() []string {
	prefixes := append(ns.PrefixAlt, ns.Prefix)
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i] < prefixes[j]
	})

	return prefixes
}

// BaseURIs returns all namespace base-URIs linked to this NameSpace.
// This includes the default Base and all alternative base-URIs.
func (ns *Namespace) BaseURIs() []string {
	baseURIs := append(ns.BaseAlt, ns.Base)
	sort.Slice(baseURIs, func(i, j int) bool {
		return baseURIs[i] < baseURIs[j]
	})

	return baseURIs
}

// GetID returns a string representation of a UUID.
// When no UUID is set, this function will generate it and update the NameSpace.
func (ns *Namespace) GetID() string {
	if ns.ID == "" {
		ns.ID = generateID()
	}

	return ns.ID
}

func generateID() string {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		// TODO(kiivihal): how can this be tested
		log.Fatalf("unable to generate random uuid for namespace; %s", err)
		return ""
	}

	s := fmt.Sprintf("%x", b)

	return s
}

// Merge merges the values of two NameSpace objects.
// The prefixes and alternative base URIs of the other NameSpace are merged into ns.
func (ns *Namespace) Merge(other *Namespace) error {
	ns.PrefixAlt = mergeSlice(ns.PrefixAlt, other.Prefixes(), ns.Prefix)
	ns.BaseAlt = mergeSlice(ns.BaseAlt, other.BaseURIs(), ns.Base)

	return nil
}

func mergeSlice(first, second []string, exclude string) []string {
	keys := map[string]bool{}

	for _, items := range [][]string{first, second} {
		for _, p := range items {
			if p != exclude {
				keys[p] = true
			}
		}
	}

	i := 0

	merged := make([]string, len(keys))
	for k := range keys {
		merged[i] = k
		i++
	}

	return merged
}
