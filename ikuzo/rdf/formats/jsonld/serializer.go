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

// SerializeFramed converts an RDF graph to JSON-LD and writes it to the provided writer.
func SerializeFramed(g *rdf.Graph, w io.Writer, context, frame map[string]any) error {
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
	options.Format = "application/n-quads"
	options.ProcessingMode = ld.JsonLd_1_1

	// Convert the RDF dataset to JSON-LD
	jsonLdDoc, err := proc.FromRDF(buf.String(), options)
	if err != nil {
		return err
	}

	if len(frame) == 0 {
		frame = map[string]any{
			"@context": context["@context"],
			// "@embed":   "@always",
		}
	}

	framed, err := proc.Frame(jsonLdDoc, frame, options)
	if err != nil {
		return fmt.Errorf("failed to frame JSON-LD: %w", err)
	}

	// Perform frame and compact operations to structure the JSON-LD
	// This will compact the JSON-LD document using a context
	compacted, err := proc.Compact(framed, context, options)
	if err != nil {
		return err
	}

	// Write the JSON-LD to the writer
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(compacted)
}
