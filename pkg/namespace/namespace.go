package namespace

import "strings"

// URI represents a NameSpace URI.
type URI string

// NameSpace is a container for URI conversions for RDF- and XML-namespaces.
type NameSpace struct {
	// Base is the default base-URI for a namespace
	Base URI `json:"base"`

	// Prefix is the default short version that identifies the base-URI
	Prefix string `json:"prefix"`

	// BaseAlt are alternative base-URI for the same prefix.
	// Sometimes historically the base-URIs for a namespace changes and we still
	// have to correctly resolve both.
	BaseAlt []URI `json:"alternatives"`

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
