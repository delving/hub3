package jsonld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/piprate/json-gold/ld"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

// Serialize converts an RDF graph to JSON-LD and writes it to the provided writer.
func Serialize(g *rdf.Graph, w io.Writer, context map[string]any) error {
	var buf bytes.Buffer
	err := ntriples.Serialize(g, &buf)
	if err != nil {
		return err
	}

	if len(context) == 0 {
		context = map[string]any{}
	}

	// Create a JSON-LD processor
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	slog.Info("serialized string", "rdf", buf.String())

	// Convert the RDF dataset to JSON-LD
	jsonLdDoc, err := proc.FromRDF(buf.String(), options)
	if err != nil {
		return err
	}

	// Perform frame and compact operations to structure the JSON-LD
	// This will compact the JSON-LD document using a context
	compacted, err := proc.Compact(jsonLdDoc, context, options)
	if err != nil {
		return err
	}

	// Write the JSON-LD to the writer
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(compacted)
}

// SerializeFramed converts an RDF graph to JSON-LD, frames it, and writes the
// result to the provided writer.
//
// When frame is empty, a default frame is built from context and g.Subject
// (when g.Subject is an IRI). The default uses "@embed": "@always" so the
// embedded structure is preserved rather than collapsed to bare @id references.
func SerializeFramed(g *rdf.Graph, w io.Writer, context, frame map[string]any) error {
	framed, err := Frame(g, frameSubject(g), context, frame)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(framed)
}

// Frame converts an RDF graph to a framed JSON-LD document.
//
// Behavior:
//   - When customFrame is non-empty, it is used verbatim.
//   - Otherwise a default frame is constructed as
//     {"@context": context, "@embed": "@always", "@id": subject}
//     where "@id" is only included when subject is non-empty.
//
// context may be nil; an empty context yields full IRIs in the output.
func Frame(g *rdf.Graph, subject string, context, customFrame map[string]any) (map[string]any, error) {
	var buf bytes.Buffer
	if err := ntriples.Serialize(g, &buf); err != nil {
		return nil, err
	}

	if context == nil {
		context = map[string]any{}
	}

	proc := ld.NewJsonLdProcessor()
	fromRDFOptions := ld.NewJsonLdOptions("")
	fromRDFOptions.Format = "application/n-quads"
	fromRDFOptions.ProcessingMode = ld.JsonLd_1_1

	jsonLdDoc, err := proc.FromRDF(buf.String(), fromRDFOptions)
	if err != nil {
		return nil, err
	}

	frame := customFrame
	if len(frame) == 0 {
		frame = map[string]any{
			"@context": context,
			"@embed":   "@always",
		}
		if subject != "" {
			frame["@id"] = subject
		}
	}

	// Frame must run without the n-quads Format option set on the processor,
	// otherwise json-gold attempts to re-parse the Go map as n-quads input.
	// OmitGraph collapses the default `{"@graph": [...]}` wrapper when there
	// is a single matching subject, which is what callers expect when the
	// frame is anchored on an explicit @id.
	frameOptions := ld.NewJsonLdOptions("")
	frameOptions.ProcessingMode = ld.JsonLd_1_1
	frameOptions.OmitGraph = true

	framed, err := proc.Frame(jsonLdDoc, frame, frameOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to frame JSON-LD: %w", err)
	}

	return framed, nil
}

func frameSubject(g *rdf.Graph) string {
	if g == nil || g.Subject == nil {
		return ""
	}
	if g.Subject.Type() != rdf.TermIRI {
		return ""
	}
	return g.Subject.RawValue()
}
