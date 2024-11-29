package html

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/html/internal"
	"github.com/delving/hub3/ikuzo/rdf/formats/jsonld"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/go-chi/chi/v5"
)

// RDFHandler handles the rendering of RDF resources
type RDFHandler struct {
	logger *slog.Logger
}

// NewRDFHandler creates a new RDF handler
func NewRDFHandler(logger *slog.Logger) *RDFHandler {
	return &RDFHandler{
		logger: logger,
	}
}

// RegisterRoutes registers the routes for the RDF handler
func (h *RDFHandler) RegisterRoutes(r chi.Router) {
	r.Get("/resource", h.GetResource)
	r.Get("/resource/raw", h.GetRawRDF)
	r.Get("/resource/download", h.DownloadRDF)
}

// GetResource handles the rendering of an RDF resource
func (h *RDFHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get URI from query parameter
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "Missing URI parameter", http.StatusBadRequest)
		return
	}

	// Get the RDF graph for the URI
	graph, err := rdf.DefaultResourceGetter().GetGraph(ctx, uri)
	if err != nil {
		h.logger.Error("Failed to get graph", "error", err)
		render.Error(w, r, fmt.Errorf("failed to get graph: %w", err), nil)
		return
	}

	if strings.HasPrefix(uri, "urn:") && strings.HasSuffix(uri, "/graph") {
		uri = graph.Subject.RawValue()
	}

	// Find the subject term from the URI
	subject, found := findSubjectByURI(graph, uri)
	if !found {
		h.logger.Error("Resource not found", "uri", uri)
		render.Error(w, r, fmt.Errorf("resource not found: %s", uri), nil)
		return
	}

	// Get format from query parameters for content negotiation
	format := r.URL.Query().Get("format")

	// Check if we need to use content negotiation (non-HTML format or Accept header)
	if format != "" {
		// Format is explicitly specified in query parameter
		if format == "turtle" || format == "jsonld" || format == "n3" || format == "ntriples" || format == "rdfxml" {
			h.serveRDFFormat(w, r, graph, format)
			return
		}
	} else if r.Header.Get("HX-Request") != "true" && !isExplicitlyRequestingHTML(r) {
		// Content negotiation based on Accept header
		if err := RenderRDF(w, r, graph); err != nil {
			h.logger.Error("Failed to render RDF", "error", err)
			render.Error(w, r, fmt.Errorf("failed to render RDF: %w", err), nil)
		}
		return
	}

	// If we get here, we're rendering HTML
	// HTMX request gets just the resource component
	if r.Header.Get("HX-Request") == "true" {
		if err := internal.RDFResource(graph, subject).Render(ctx, w); err != nil {
			h.logger.Error("Failed to render RDF resource component", "error", err)
			render.Error(w, r, fmt.Errorf("failed to render resource: %w", err), nil)
		}
	} else {
		// Regular request gets the full page
		if err := internal.RDFView(graph, subject).Render(ctx, w); err != nil {
			h.logger.Error("Failed to render RDF view", "error", err)
			render.Error(w, r, fmt.Errorf("failed to render view: %w", err), nil)
		}
	}
}

// GetRawRDF handles requests for raw RDF data to display in the UI
func (h *RDFHandler) GetRawRDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get URI and format from query parameters
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "Missing URI parameter", http.StatusBadRequest)
		return
	}

	// Parse form data to get format from form fields if present
	if err := r.ParseForm(); err != nil {
		h.logger.Error("Failed to parse form", "error", err)
	}

	// Determine format
	format := r.URL.Query().Get("format")
	if format == "" {
		format = r.Form.Get("format")
		if format == "" {
			format = "jsonld" // Default
		}
	}

	// Get the RDF graph
	graph, err := rdf.DefaultResourceGetter().GetGraph(ctx, uri)
	if err != nil {
		h.logger.Error("Failed to get graph", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get graph: %v", err), http.StatusInternalServerError)
		return
	}

	// For HTMX requests, we need to return HTML with the proper structure for Prism.js
	if r.Header.Get("HX-Request") == "true" {
		// Change content type to HTML for proper rendering
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Create a buffer to capture the RDF output
		var buffer strings.Builder

		// Serialize RDF to the buffer based on format
		switch format {
		case "turtle":
			err = ntriples.Serialize(graph, &buffer)
		case "jsonld":
			err = jsonld.Serialize(graph, &buffer, nil)
		case "n3":
			err = ntriples.Serialize(graph, &buffer)
		case "ntriples":
			err = ntriples.Serialize(graph, &buffer)
		case "rdfxml":
			cfg := &mappingxml.FilterConfig{
				Subject:         graph.Subject,
				URIPrefixFilter: "urn:private",
			}
			err = mappingxml.Serialize(graph, &buffer, cfg)
		default:
			err = jsonld.Serialize(graph, &buffer, nil)
		}

		if err != nil {
			h.logger.Error("Failed to serialize RDF", "error", err, "format", format)
			http.Error(w, fmt.Sprintf("Failed to serialize RDF in %s format: %v", format, err), http.StatusInternalServerError)
			return
		}

		// Escape the content for HTML
		content := strings.ReplaceAll(buffer.String(), "<", "&lt;")
		content = strings.ReplaceAll(content, ">", "&gt;")

		// Determine the language class for Prism.js based on format
		var langClass string
		switch format {
		case "jsonld":
			langClass = "language-json"
		case "rdfxml":
			langClass = "language-markup"
		case "ntriples":
			langClass = "language-ntriples" // Specific class for N-Triples
		case "n3":
			langClass = "language-turtle" // N3 can use Turtle highlighting
		case "turtle":
			langClass = "language-turtle"
		default:
			langClass = "language-json"
		}

		// Create HTML response with proper structure for Prism.js
		htmlResponse := fmt.Sprintf(`
<div class="text-center py-8" id="raw-loading" style="display: none;">
	<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
		<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
		<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
	</svg>
	<p class="mt-2 text-gray-600 dark:text-gray-400">Loading RDF data...</p>
</div>
<pre class="%s"><code>%s</code></pre>
`, langClass, content)

		// Write the HTML response
		w.Write([]byte(htmlResponse))
		return
	}

	// For non-HTMX requests (e.g., direct API calls), serve plain text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.serveRDFFormat(w, r, graph, format)
}

// DownloadRDF handles requests to download RDF data in various formats
func (h *RDFHandler) DownloadRDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get URI and format from query parameters
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "Missing URI parameter", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "turtle" // Default
	}

	// Get the RDF graph
	graph, err := rdf.DefaultResourceGetter().GetGraph(ctx, uri)
	if err != nil {
		h.logger.Error("Failed to get graph", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get graph: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers for download
	filename := generateFilename(uri, format)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Set the correct content type
	switch format {
	case "turtle":
		w.Header().Set("Content-Type", "text/turtle")
	case "jsonld":
		w.Header().Set("Content-Type", "application/ld+json")
	case "n3":
		w.Header().Set("Content-Type", "text/n3")
	case "ntriples":
		w.Header().Set("Content-Type", "application/n-triples")
	case "rdfxml":
		w.Header().Set("Content-Type", "application/rdf+xml")
	default:
		w.Header().Set("Content-Type", "application/ld+json")
	}

	// Serve the RDF data
	h.serveRDFFormat(w, r, graph, format)
}

// serveRDFFormat outputs RDF in the specified format
func (h *RDFHandler) serveRDFFormat(w http.ResponseWriter, r *http.Request, graph *rdf.Graph, format string) {
	var err error

	switch format {
	case "turtle":
		// Use ntriples until a turtle serializer is available
		err = ntriples.Serialize(graph, w)
	case "jsonld":
		err = jsonld.Serialize(graph, w, nil)
	case "n3":
		// Fallback to ntriples for N3 format
		err = ntriples.Serialize(graph, w)
	case "ntriples":
		err = ntriples.Serialize(graph, w)
	case "rdfxml":
		cfg := &mappingxml.FilterConfig{
			Subject:         graph.Subject,
			URIPrefixFilter: "urn:private",
		}
		err = mappingxml.Serialize(graph, w, cfg)
	default:
		err = jsonld.Serialize(graph, w, nil)
	}

	if err != nil {
		h.logger.Error("Failed to serialize RDF", "error", err, "format", format)
		http.Error(w, fmt.Sprintf("Failed to serialize RDF in %s format: %v", format, err), http.StatusInternalServerError)
	}
}

// RenderRDFResource is a helper function to render an RDF resource
func RenderRDFResource(graph *rdf.Graph, resource rdf.Subject, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Determine if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		return internal.RDFResource(graph, resource).Render(ctx, w)
	}

	// Otherwise render the full page
	return internal.RDFView(graph, resource).Render(ctx, w)
}

// Helper function to check if request is explicitly asking for HTML
func isExplicitlyRequestingHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html") ||
		strings.Contains(accept, "application/xhtml+xml") ||
		r.Header.Get("HX-Request") == "true"
}

// Helper function to find a subject in the graph by its URI
func findSubjectByURI(graph *rdf.Graph, uri string) (rdf.Subject, bool) {
	for _, triple := range graph.Triples() {
		if triple.Subject.RawValue() == uri {
			return triple.Subject, true
		}
	}
	return nil, false
}

// Helper function to generate a safe filename for downloads
func generateFilename(uri string, format string) string {
	// Get the local part of the URI
	var localPart string
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' || uri[i] == '#' {
			localPart = uri[i+1:]
			break
		}
	}

	if localPart == "" {
		localPart = "resource"
	}

	// Replace problematic characters
	localPart = strings.ReplaceAll(localPart, "%", "_")
	localPart = strings.ReplaceAll(localPart, " ", "_")

	// Add appropriate extension
	var extension string
	switch format {
	case "turtle":
		extension = ".ttl"
	case "jsonld":
		extension = ".jsonld"
	case "n3":
		extension = ".n3"
	case "ntriples":
		extension = ".nt"
	case "rdfxml":
		extension = ".rdf"
	default:
		extension = ".jsonld"
	}

	return localPart + extension
}
