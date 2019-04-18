package namespace

import (
	"sort"
	"strings"

	"github.com/segmentio/ksuid"
)

// URI represents a NameSpace URI.
type URI string

// NameSpace is a container for URI conversions for RDF- and XML-namespaces.
type NameSpace struct {

	// UUID is the unique identifier of a namespace
	UUID string `json:"uuid"`

	// Base is the default base-URI for a namespace
	Base string `json:"base"`

	// Prefix is the default short version that identifies the base-URI
	Prefix string `json:"prefix"`

	// BaseAlt are alternative base-URI for the same prefix.
	// Sometimes historically the base-URIs for a namespace changes and we still
	// have to correctly resolve both.
	BaseAlt []string `json:"baseAlt"`

	// PrefixAlt are altenative prefixes for the default base URI.
	// Different content-providers and organisations have at time selected alternative
	// prefixes for the same base URI. We need to support both entry entry-points.
	PrefixAlt []string `json:"prefixAlt"`

	// Schema is an URL to the RDFS or OWL definition of namespace
	Schema string `json:"schema"`
}

// String returns a string representation of URI
func (uri URI) String() string {
	return string(uri)
}

// SplitURI takes a given URI and splits it into a base-URI and a localname.
// When the URI can't be split, the full URI is returned as the label with an
// empty base.
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

// AddPrefix adds a prefix to the list of prefix alternatives.
//
// When the prefix is already present in PrefixAlt no error is thrown.
func (ns *NameSpace) AddPrefix(prefix string) error {
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
func (ns *NameSpace) AddBase(base string) error {
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
func (ns NameSpace) Prefixes() []string {
	prefixes := append(ns.PrefixAlt, ns.Prefix)
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i] < prefixes[j]
	})
	return prefixes
}

// BaseURIs returns all namespace base-URIs linked to this NameSpace.
// This includes the default Base and all alternative base-URIs.
func (ns NameSpace) BaseURIs() []string {
	baseURIs := append(ns.BaseAlt, ns.Base)
	sort.Slice(baseURIs, func(i, j int) bool {
		return baseURIs[i] < baseURIs[j]
	})
	return baseURIs
}

// GetID returns a string representation of a UUID.
// When no UUID is sit, this function will generate it and update the NameSpace.
func (ns *NameSpace) GetID() string {
	if ns.UUID == "" {
		uuid := ksuid.New()
		ns.UUID = uuid.String()
	}
	return ns.UUID
}

// Merge merges the values of two NameSpace objects.
// The prefixes and alternative base URIs of the other NameSpace are merged into ns.
func (ns *NameSpace) Merge(other *NameSpace) error {
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
