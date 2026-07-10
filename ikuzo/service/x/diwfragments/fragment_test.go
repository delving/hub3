package diwfragments

import (
	"strings"
	"testing"
)

func validFragment() Fragment {
	return Fragment{
		OrgID: "demo", Collection: "coll1", Kind: KindItem,
		RecordID: "demo_spec_158", Lang: "nl",
		HeadTags: "<link rel=\"canonical\" href=\"https://x\">", HTML: "<div>x</div>",
		Meta: Meta{RenderedAt: "2026-07-10T12:00:00Z", DiwVersion: "1.0.0", ConfigVersion: "abc123"},
	}
}

func TestDocIDIsDeterministicAndDistinct(t *testing.T) {
	a, b := validFragment(), validFragment()
	if a.DocID() != b.DocID() {
		t.Fatal("same fragment must produce same DocID")
	}
	b.Lang = "en"
	if a.DocID() == b.DocID() {
		t.Fatal("different lang must produce different DocID")
	}
	if strings.ContainsAny(a.DocID(), "/ ") {
		t.Fatalf("DocID must be URL-safe, got %q", a.DocID())
	}
}

func TestETagChangesWithContent(t *testing.T) {
	a, b := validFragment(), validFragment()
	if a.ETag() != b.ETag() {
		t.Fatal("same content must produce same ETag")
	}
	b.HTML = "<div>y</div>"
	if a.ETag() == b.ETag() {
		t.Fatal("different html must produce different ETag")
	}
}

func TestValidate(t *testing.T) {
	f := validFragment()
	if err := f.Validate(); err != nil {
		t.Fatalf("valid fragment must validate, got %v", err)
	}
	f.Kind = "bogus"
	if err := f.Validate(); err == nil {
		t.Fatal("bogus kind must fail validation")
	}
	g := validFragment()
	g.Kind = KindItem
	g.RecordID = ""
	if err := g.Validate(); err == nil {
		t.Fatal("item fragment without recordID must fail validation")
	}
	h := validFragment()
	h.Kind = KindListing
	h.RecordID = ""
	if err := h.Validate(); err != nil {
		t.Fatalf("listing fragment needs no recordID, got %v", err)
	}
}
