package jsonld

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/matryer/is"
	"github.com/piprate/json-gold/ld"
)

func getReader(testname string) (r io.Reader, err error) {
	path := "./testdata/" + testname + ".jsonld"

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// nolint:gocritic
func TestParse(t *testing.T) {
	t.Run("parse jsonld with graph", func(t *testing.T) {
		is := is.New(t)

		g := rdf.NewGraph()
		is.Equal(g.Len(), 0)
		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)
	})

	t.Run("parse jsonld without graph", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("flat")
		is.NoErr(err)

		returnedGraph, err := Parse(r, nil)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 47)
	})
}

// testdata/with_context.jsonld references @context documents published by a
// third party (apidg.gent.be). Resolving those over the network made this
// test fail whenever that host had a hiccup -- it returned 502 for a stretch
// on 2026-09-21 and took CI down with it. The documents are vendored under
// testdata/context/ and mapped here, so the test exercises the same parsing
// path without depending on anyone's uptime.
func vendoredContextLoader() ld.DocumentLoader {
	const base = "https://apidg.gent.be/opendata/adlib2eventstream/v1/context/"

	names := []string{
		"cultureel-erfgoed-object-ap",
		"persoon-basis",
		"cultureel-erfgoed-event-ap",
		"organisatie-basis",
		"generiek-basis",
		"dossier",
	}

	mapping := make(map[string]string, len(names))
	for _, n := range names {
		mapping[base+n+".jsonld"] = "./testdata/context/" + n + ".jsonld"
	}

	// The fallback loader is deliberately nil-clienting nothing: every URL the
	// fixture uses is in the mapping, so a network call means the fixture and
	// this list drifted apart, and the test should say so.
	loader := ld.NewCachingDocumentLoader(ld.NewDefaultDocumentLoader(nil))
	loader.PreloadWithMapping(mapping)

	return loader
}

func TestParseWithContext(t *testing.T) {
	t.Run("parse with external context", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("with_context")
		is.NoErr(err)

		returnedGraph, err := ParseWithContextLoader(r, nil, vendoredContextLoader())
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 85)
	})
}

func TestParseWithCollections(t *testing.T) {
	t.Run("collections with blank nodes", func(t *testing.T) {
		is := is.New(t)

		r, err := getReader("with_collections")
		is.NoErr(err)

		returnedGraph, err := Parse(r, nil)
		is.NoErr(err)

		stats := returnedGraph.Stats()
		is.Equal(stats.Predicates, 25)

		var b bytes.Buffer
		err = ntriples.Serialize(returnedGraph, &b)
		is.NoErr(err)

		t.Logf("triples: \n %s", b.String())
		is.Equal(returnedGraph.Len(), 47)
	})
}
