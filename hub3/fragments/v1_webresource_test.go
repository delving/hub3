package fragments

import (
	"os"
	"strings"
	"testing"

	r "github.com/kiivihal/rdf2go"
	"github.com/matryer/is"
)

func TestCleanWebResourceGraph(t *testing.T) {
	is := is.New(t)
	fg := NewFragmentGraph()
	fb := NewFragmentBuilder(fg)

	// file, err := os.Open("./testdata/rdf-jsonld-dcn.json")
	file, err := os.Open("./testdata/rdf-ntriples-dcn.nt")
	defer file.Close()

	is.NoErr(err)
	// err = fb.ParseGraph(file, "application/ld+json")
	err = fb.ParseGraph(file, "text/turtle")
	is.NoErr(err)

	is.Equal(fb.Graph.Len(), 64)
	is.Equal(len(fb.GetUrns()), 1)

	urnTriples := `<urn:museum-sloten/D036a> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.europeana.eu/schemas/edm/WebResource> .
<urn:museum-sloten/D036a> <http://schemas.delving.eu/nave/terms/thumbLarge> "https://media.delving.org/thumbnail/dcn/museum-sloten/D036a/1000" .
<urn:museum-sloten/D036a> <http://schemas.delving.eu/nave/terms/allowDeepZoom> "true" .
<urn:museum-sloten/D036a> <http://schemas.delving.eu/nave/terms/deepZoomUrl> "https://media.delving.org/deepzoom/dcn/museum-sloten/D036a.tif.dzi" .
<urn:museum-sloten/D036a> <http://schemas.delving.eu/nave/terms/thumbSmall> "https://media.delving.org/thumbnail/dcn/museum-sloten/D036a/500" .
`

	err = fb.Graph.Parse(strings.NewReader(urnTriples), "text/turtle")
	is.NoErr(err)

	is.Equal(fb.Graph.Len(), 68)
	is.Equal(len(fb.GetUrns()), 1)

	hasUrns := len(fb.GetUrns()) != 0

	hasView := fb.ByPredicate(GetEDMField("hasView"))
	is.Equal(len(hasView), 2)

	oreNS := r.NewResource(GetNSField("ore", "aggregates"))
	hasAggregates := fb.ByPredicate(oreNS)
	is.Equal(len(hasAggregates), 2)

	resources, aggregates, webTriples := fb.CleanWebResourceGraph(hasUrns)
	is.Equal(fb.Graph.Len(), 56)
	is.Equal(len(resources), 1)
	t.Logf("webtriples: %#v", webTriples.triples)
	t.Logf("resources: %#v", resources)
	t.Logf("aggregates: %#v", aggregates)
	is.Equal(len(webTriples.triples), 1)
	is.Equal(len(aggregates), 2)
	hasAggregates = fb.ByPredicate(oreNS)
	is.Equal(len(hasAggregates), 2)

	hasView = fb.Graph.All(nil, GetEDMField("hasView"), nil)
	is.Equal(len(hasView), 0)
}
