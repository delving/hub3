// Package diwfragments stores and serves pre-rendered DIW HTML fragments
// under the frozen /api/ui/v1 contract (see the ecosystem repo's
// diw-served-components-evolution.md). At birth the store is an
// Elasticsearch index and fragments are written by the diw-core render
// worker after indexing; the contract must survive the Delivery rewrite.
package diwfragments

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Kind discriminates the two fragment shapes the v1 contract serves.
type Kind string

const (
	// KindItem is a record-detail fragment, keyed by RecordID (hubID).
	KindItem Kind = "item"
	// KindListing is the initial-listing fragment for a collection.
	KindListing Kind = "listing"
)

// Meta travels inside the served envelope and identifies exactly which
// renderer and config produced a fragment, for cache debugging.
type Meta struct {
	RenderedAt    string `json:"renderedAt"`
	DiwVersion    string `json:"diwVersion"`
	ConfigVersion string `json:"configVersion"`
}

// Fragment is one stored pre-rendered DIW fragment. OrgID scopes the
// storage index; Collection is the opaque DIW collection id Orchestra
// manages (the same key a data-diw embed attribute uses).
type Fragment struct {
	OrgID      string `json:"orgID"`
	Collection string `json:"collection"`
	Kind       Kind   `json:"kind"`
	RecordID   string `json:"recordID,omitempty"`
	Lang       string `json:"lang"`
	HeadTags   string `json:"headTags"`
	HTML       string `json:"html"`
	Meta       Meta   `json:"meta"`
}

// Envelope is the response body of the GET routes — the contract shape.
type Envelope struct {
	HeadTags string `json:"headTags"`
	HTML     string `json:"html"`
	Meta     Meta   `json:"meta"`
}

// DocID returns the deterministic storage id for this fragment so that
// re-rendering upserts in place: org, collection, kind, record and lang
// joined with '~' (all components are ES/URL-safe identifiers; listing
// fragments use '_' for the empty record slot).
func (f *Fragment) DocID() string {
	rec := f.RecordID
	if rec == "" {
		rec = "_"
	}
	return strings.Join([]string{f.OrgID, f.Collection, string(f.Kind), rec, f.Lang}, "~")
}

// ETag hashes everything a consumer can observe, so any re-render that
// changes output changes the tag and HTTP caches revalidate correctly.
func (f *Fragment) ETag() string {
	h := sha256.New()
	for _, part := range []string{f.HTML, f.HeadTags, f.Meta.DiwVersion, f.Meta.ConfigVersion} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Validate rejects fragments that could not be served back: unknown
// kinds, missing identifiers, or an item fragment without its record id.
func (f *Fragment) Validate() error {
	if f.Kind != KindItem && f.Kind != KindListing {
		return fmt.Errorf("unknown fragment kind %q", f.Kind)
	}
	if f.OrgID == "" || f.Collection == "" || f.Lang == "" {
		return fmt.Errorf("fragment requires orgID, collection and lang")
	}
	if f.Kind == KindItem && f.RecordID == "" {
		return fmt.Errorf("item fragment requires recordID")
	}
	if f.HTML == "" {
		return fmt.Errorf("fragment requires html")
	}
	return nil
}
