package index

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/validator"
)

type Header struct {
	// The tenant identifier for this Graph.
	OrgID string `json:"orgID,omitempty"`

	// The spec is the unique dataset to which the Graph belongs
	Spec string `json:"spec,omitempty"`

	// The hubId is the unique identifier for any document record in hub3
	HubID string `json:"hubID,omitempty"`

	// Each Graph can be tagged with additional metadata. This can be queried for.
	Tags []string `json:"tags,omitempty"`

	// The document type for ElasticSearch. This is a constant value
	DocType string `json:"docType,omitempty"`

	// The subject of the graph stored
	EntryURI string `json:"entryURI,omitempty"`

	// the namedgraph URI of the graph stored
	NamedGraphURI string `json:"namedGraphURI,omitempty"`

	// miliseconds since epoch
	Modified int64 `json:"modified,omitempty"`

	// sourceID is the unique fingerprint of the Graph data. This excludes the header.
	SourceID string `protobuf:"bytes,10,opt,name=sourceID,proto3" json:"sourceID,omitempty"`

	// The revision is used to determine which version is an orphan and should be removed.
	Revision int32 `json:"revision,omitempty"`
}

// AddTags adds a tag string to the tags array of the Header
func (m *Header) AddTags(tags ...string) {
	m.Tags = appendUnique(m.Tags, tags...)
}

// Valid checks the header for invalid content.
func (m *Header) Valid() error {
	v := validator.New()
	checks := map[string]string{
		"OrgID":         m.OrgID,
		"HubID":         m.HubID,
		"Spec":          m.Spec,
		"DocType":       m.DocType,
		"EntryURI":      m.EntryURI,
		"NamedGraphURI": m.NamedGraphURI,
	}

	for label, val := range checks {
		v.Check(val != "", label, fmt.Errorf("%s cannot be empty", label), fmt.Sprintf("%s must always be set", label))
	}

	return v.ErrorOrNil()
}

// addDefaults adds default values to the Header of they are empty.
func (m *Header) addDefaults() {
	if m.Modified == 0 {
		m.Modified = NowInMillis()
	}
	if m.NamedGraphURI == "" {
		m.NamedGraphURI = fmt.Sprintf("%s/graph", m.EntryURI)
	}
	m.DocType = GraphDocType
}
