package html

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestContentNegotiation tests the content negotiation function
func TestContentNegotiation(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedFormat string
	}{
		{"Empty Accept", "", "html"},
		{"HTML Accept", "text/html", "html"},
		{"XHTML Accept", "application/xhtml+xml", "html"},
		{"RDF/XML Accept", "application/rdf+xml", "rdf/xml"},
		{"Turtle Accept", "text/turtle", "turtle"},
		{"N-Triples Accept", "application/n-triples", "n-triples"},
		{"JSON-LD Accept", "application/ld+json", "jsonld"},
		{"Multiple Accept", "text/html,application/rdf+xml", "rdf/xml"},
		{"Quality Values", "application/ld+json;q=0.9,text/html;q=0.8", "jsonld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/resource", nil)
			if test.acceptHeader != "" {
				req.Header.Set("Accept", test.acceptHeader)
			}

			format := contentNegotiation(req)
			if format != test.expectedFormat {
				t.Errorf("Expected format %s, got %s", test.expectedFormat, format)
			}
		})
	}
}

// TestAddFormatLinks tests the function that adds format links to response headers
func TestAddFormatLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/resource?uri=urn:test:123", nil)
	w := httptest.NewRecorder()

	supportedFormats := []string{
		"text/html",
		"application/rdf+xml",
		"text/turtle",
	}

	addFormatLinks(w, req, supportedFormats)

	// Check Vary header
	if w.Header().Get("Vary") != "Accept" {
		t.Errorf("Expected Vary: Accept header, got %s", w.Header().Get("Vary"))
	}

	// Check Link headers
	links := w.Header()["Link"]
	if len(links) != len(supportedFormats) {
		t.Errorf("Expected %d Link headers, got %d", len(supportedFormats), len(links))
	}

	// Verify each format is represented in Link headers
	for _, format := range supportedFormats {
		found := false
		for _, link := range links {
			if strings.Contains(link, "type=\""+format+"\"") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Link header for format %s not found", format)
		}
	}

	// Links carry the resource in ?uri= and the target format in &format=,
	// which is what contentNegotiation consumes.
	for _, link := range links {
		if !strings.Contains(link, "<https://example.com/resource?uri=urn:test:123&format=") {
			t.Errorf("Link header contains incorrect URI: %s", link)
		}
	}
}
