package jsonld

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// Triples with a relative IRI are still left out, but no longer silently (F23).
func TestParseWarnsOnRelativeIRIs(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	doc := `{
	  "@id": "http://example.org/r",
	  "http://schema.org/url": {"@id": "toBeRemoved"},
	  "http://schema.org/image": {"@id": "img/1.jpg"},
	  "http://schema.org/sameAs": {"@id": "http://example.org/other"},
	  "http://schema.org/name": "Naam"
	}`
	g, err := Parse(strings.NewReader(doc), nil)
	if err != nil {
		t.Fatal(err)
	}
	if n := g.Len(); n != 2 {
		t.Errorf("graph has %d triples, want 2 (sameAs, name)", n)
	}
	out := logs.String()
	if !strings.Contains(out, "relative IRI") || !strings.Contains(out, "count=2") ||
		!(strings.Contains(out, "example=toBeRemoved") || strings.Contains(out, "example=img/1.jpg")) {
		t.Errorf("want a warning with dropped=2 and an example, got %q", out)
	}
}
