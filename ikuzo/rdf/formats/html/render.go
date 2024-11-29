package html

import (
	"net/http"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/jsonld"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

// contentNegotiation determines the response format based on Accept header
func contentNegotiation(r *http.Request) string {
	// Get the Accept header from the request
	accept := r.Header.Get("Accept")

	// Helper function to check if the Accept header contains a specific type
	containsType := func(mimeType string) bool {
		return strings.Contains(accept, mimeType)
	}

	// Check for format query parameter, which takes precedence over Accept header
	if format := r.URL.Query().Get("format"); format != "" {
		switch format {
		case "turtle":
			return "turtle"
		case "jsonld":
			return "jsonld"
		case "n3":
			return "n3"
		case "ntriples":
			return "n-triples"
		case "rdfxml":
			return "rdf/xml"
		}
	}

	// Use a switch with true as the expression to allow for complex conditions
	switch {
	case containsType("application/rdf+xml"):
		return "rdf/xml"
	case containsType("text/turtle"):
		return "turtle"
	case containsType("text/n3"):
		return "n3"
	case containsType("application/n-triples"):
		return "n-triples"
	case containsType("application/ld+json"):
		return "jsonld"
	case containsType("text/html") || containsType("application/xhtml+xml"):
		return "html"
	default:
		// Default to HTML if Accept header is empty or no match
		return "html"
	}
}

// addFormatLinks adds Link headers for available formats
func addFormatLinks(w http.ResponseWriter, r *http.Request, supportedFormats []string) {
	// Get the current resource URI
	baseURI := "https://" + r.Host + r.URL.Path
	resourceURI := r.URL.Query().Get("uri")

	// Add Link headers for each supported format
	for _, format := range supportedFormats {
		// Create a link to the current resource but with format specified as a query parameter
		formatURI := baseURI + "?uri=" + resourceURI + "&format=" + formatToQueryParam(format)
		w.Header().Add("Link", "<"+formatURI+">; rel=\"alternate\"; type=\""+format+"\"")
	}

	// Add Vary header to ensure proper caching behavior
	w.Header().Add("Vary", "Accept")
}

// Helper to convert MIME type to query parameter format value
func formatToQueryParam(mimeType string) string {
	switch mimeType {
	case "text/turtle":
		return "turtle"
	case "application/ld+json":
		return "jsonld"
	case "text/n3":
		return "n3"
	case "application/n-triples":
		return "ntriples"
	case "application/rdf+xml":
		return "rdfxml"
	default:
		return "html"
	}
}

// getAllSupportedFormats returns all formats supported by the application
func getAllSupportedFormats() []string {
	return []string{
		"text/html",
		"application/rdf+xml",
		"text/turtle",
		"text/n3",
		"application/n-triples",
		"application/ld+json",
	}
}

// RenderRDF renders the graph in the requested format
func RenderRDF(w http.ResponseWriter, r *http.Request, g *rdf.Graph) error {
	supportedFormats := getAllSupportedFormats()

	// Add Link headers for all supported formats
	addFormatLinks(w, r, supportedFormats)

	// Determine format based on Accept header or query parameter
	format := contentNegotiation(r)

	var err error
	// Set appropriate Content-Type
	switch format {
	case "rdf/xml":
		cfg := &mappingxml.FilterConfig{
			Subject:         g.Subject,
			URIPrefixFilter: "urn:private",
		}
		w.Header().Set("Content-Type", "application/rdf+xml")
		err = mappingxml.Serialize(g, w, cfg)
	case "turtle":
		w.Header().Set("Content-Type", "text/turtle")
		err = ntriples.Serialize(g, w) // Replace with turtle serializer when available
	case "n3":
		w.Header().Set("Content-Type", "text/n3")
		err = ntriples.Serialize(g, w) // Fallback to ntriples if n3 not available
	case "n-triples":
		w.Header().Set("Content-Type", "application/n-triples")
		err = ntriples.Serialize(g, w)
	case "jsonld":
		w.Header().Set("Content-Type", "application/ld+json")
		err = jsonld.Serialize(g, w, nil)
		// Serialize to JSON-LD
	default: // html
		w.Header().Set("Content-Type", "text/html")
		err = RenderRDFResource(g, g.Subject, w, r)
	}

	return err
}

