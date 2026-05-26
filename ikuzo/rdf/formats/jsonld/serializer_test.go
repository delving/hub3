package jsonld

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/matryer/is"
)

func TestSerialize(t *testing.T) {
	t.Run("serialize jsonld with graph", func(t *testing.T) {
		is := is.New(t)

		g := rdf.NewGraph()
		is.Equal(g.Len(), 0)
		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)

		var buf bytes.Buffer
		err = Serialize(g, &buf, nil)
		is.NoErr(err)

		is.True(buf.String() != "")
	})
}

const (
	subjectIRI = "http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458"
	typeIRI    = "http://www.europeana.eu/schemas/edm/ProvidedCHO"
	titleIRI   = "http://purl.org/dc/elements/1.1/title"
)

func loadFlatGraph(t *testing.T) *rdf.Graph {
	t.Helper()
	r, err := getReader("flat")
	if err != nil {
		t.Fatalf("open testdata: %v", err)
	}
	g := rdf.NewGraph()
	if _, err := Parse(r, g); err != nil {
		t.Fatalf("parse testdata: %v", err)
	}
	subj, err := rdf.NewIRI(subjectIRI)
	if err != nil {
		t.Fatalf("build subject IRI: %v", err)
	}
	g.Subject = rdf.Subject(subj)
	return g
}

func TestFrame_AnchorsOnSubject(t *testing.T) {
	is := is.New(t)
	g := loadFlatGraph(t)

	framed, err := Frame(g, subjectIRI, nil, nil)
	is.NoErr(err)
	is.True(framed != nil)
	is.Equal(framed["@id"], subjectIRI)

	// Default frame uses @embed:@always, so embedded structure should not be
	// collapsed: at least one predicate is present alongside @id/@type.
	if len(framed) < 3 {
		t.Fatalf("expected embedded predicates, got keys: %v", keys(framed))
	}
}

func TestFrame_UsesProvidedContext(t *testing.T) {
	is := is.New(t)
	g := loadFlatGraph(t)

	ctx := map[string]any{
		"dc":   "http://purl.org/dc/elements/1.1/",
		"edm":  "http://www.europeana.eu/schemas/edm/",
		"name": map[string]any{"@id": titleIRI, "@container": "@language"},
	}

	framed, err := Frame(g, subjectIRI, ctx, nil)
	is.NoErr(err)
	is.True(framed["@context"] != nil)

	// dc:title should be compacted via the named alias, not the full IRI.
	if _, full := framed[titleIRI]; full {
		t.Fatalf("expected title to be compacted via context, but raw IRI was emitted: %v", keys(framed))
	}
	if _, named := framed["name"]; !named {
		t.Fatalf("expected compacted name predicate, got keys: %v", keys(framed))
	}
}

func TestFrame_ExplicitFrameTakesPrecedence(t *testing.T) {
	is := is.New(t)
	g := loadFlatGraph(t)

	custom := map[string]any{
		"@context": map[string]any{},
		"@type":    typeIRI,
		"@embed":   "@always",
	}

	framed, err := Frame(g, subjectIRI, nil, custom)
	is.NoErr(err)
	is.True(framed != nil)

	// Explicit frame restricts results by @type. The result must therefore
	// expose the matching @type for the anchor.
	if got, _ := framed["@type"].(string); got != typeIRI {
		// could also be []any with the type
		if arr, ok := framed["@type"].([]any); ok && len(arr) > 0 && arr[0] == typeIRI {
			return
		}
		t.Fatalf("expected @type %q, got %#v", typeIRI, framed["@type"])
	}
}

func TestSerializeFramed_WritesValidJSON(t *testing.T) {
	is := is.New(t)
	g := loadFlatGraph(t)

	var buf bytes.Buffer
	err := SerializeFramed(g, &buf, nil, nil)
	is.NoErr(err)
	is.True(buf.Len() > 0)

	var out map[string]any
	is.NoErr(json.NewDecoder(&buf).Decode(&out))
	is.Equal(out["@id"], subjectIRI)
}

func TestFrame_NoSubjectReturnsGraph(t *testing.T) {
	is := is.New(t)
	g := loadFlatGraph(t)
	g.Subject = nil

	framed, err := Frame(g, "", nil, nil)
	is.NoErr(err)
	is.True(framed != nil)

	// Without an explicit @id, framing emits a @graph array of subjects.
	graph, ok := framed["@graph"].([]any)
	if !ok {
		t.Fatalf("expected @graph array when no subject is provided, got keys: %v", keys(framed))
	}
	is.True(len(graph) > 0)
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

