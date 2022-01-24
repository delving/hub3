package ntriples

import (
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/resource"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestParse(t *testing.T) {
	t.Run("parse ntriples with graph", func(t *testing.T) {
		is := is.New(t)

		g := resource.NewGraph()
		is.Equal(g.Len(), 0)
		returnedGraph, err := Parse(strings.NewReader(testNtriples), g)
		is.NoErr(err)
		is.Equal(g, returnedGraph)

		is.Equal(g.Len(), 47)
	})

	t.Run("parse ntriples without graph", func(t *testing.T) {
		is := is.New(t)

		returnedGraph, err := Parse(strings.NewReader(testNtriples), nil)
		is.NoErr(err)

		is.Equal(returnedGraph.Len(), 47)
	})
}

// nolint:lll // this is test data
var testNtriples = `
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://creativecommons.org/ns#attributionName> "museum-klok-en-peel" .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://schemas.delving.eu/narthex/terms/belongsTo> <http://data.brabantcloud.nl/resource/dataset/museum-klok-en-peel> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://schemas.delving.eu/narthex/terms/contentHash> "de8bc9366bacd77ed1d3060f0ba2b73e124c74f0" .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://schemas.delving.eu/narthex/terms/saveTime> "2018-02-12T18:36:30Z" .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://schemas.delving.eu/narthex/terms/synced> "false"^^<http://www.w3.org/2001/XMLSchema#boolean> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://xmlns.com/foaf/0.1/Document> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about_this> <http://xmlns.com/foaf/0.1/primaryTopic> <http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/aggregatedCHO> <http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/dataProvider> "Museum Klok & Peel" .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/isShownAt> <http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/isShownBy> <https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/500> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/object> <https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/220> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/provider> "Erfgoed Brabant" .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/rights> <http://creativecommons.org/publicdomain/zero/1.0/> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.openarchives.org/ore/terms/aggregates> _:b0 .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.openarchives.org/ore/terms/aggregates> _:b1 .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schemas.delving.eu/narthex/terms/Record> .
<http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.openarchives.org/ore/terms/Aggregation> .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/elements/1.1/date> "-481 t/m -221" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/elements/1.1/description> "Bellen uit terracotta zoals deze waren grafgiften. Zij werden gemaakt ter vervanging van het originele object dat in voorgaande perioden de dode meegegeven werd. Het bekendste voorbeeld van dit gebruik is het terracotta-leger van de eerste keizer van China." .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/elements/1.1/identifier> "2458" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/elements/1.1/title> "Terracottabel tpe bo [Periode van de Strijdende Staten]"@nl .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/created> "-0481-01-01T00:00:01" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/createdEnd> "-0221-01-01T00:00:01" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/createdRaw> "-481 t/m -221" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/extent> "Hoogte: 186 mm, diameter: 148-165 mm" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/medium> "keramiek" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://purl.org/dc/terms/spatial> "China, Azie" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://www.europeana.eu/schemas/edm/type> "IMAGE" .
<http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.europeana.eu/schemas/edm/ProvidedCHO> .
_:b0 <http://schemas.delving.eu/nave/terms/allowDeepZoom> "true" .
_:b0 <http://schemas.delving.eu/nave/terms/allowLinkedOpenData> "true" .
_:b0 <http://schemas.delving.eu/nave/terms/allowSourceDownload> "false" .
_:b0 <http://schemas.delving.eu/nave/terms/deepZoomUrl> "https://media.delving.org/iip/deepzoom/mnt/tib/tiles/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant.tif.dzi" .
_:b0 <http://schemas.delving.eu/nave/terms/featured> "false" .
_:b0 <http://schemas.delving.eu/nave/terms/public> "true" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schemas.delving.eu/nave/terms/DelvingResource> .
_:b1 <http://schemas.delving.eu/nave/terms/collection> "Museum Klok & Peel" .
_:b1 <http://schemas.delving.eu/nave/terms/collectionPart> "opgravingen" .
_:b1 <http://schemas.delving.eu/nave/terms/collectionType> "Algemeen" .
_:b1 <http://schemas.delving.eu/nave/terms/creatorRole> "gieter" .
_:b1 <http://schemas.delving.eu/nave/terms/dimension> "Hoogte: 186 mm, diameter: 148-165 mm" .
_:b1 <http://schemas.delving.eu/nave/terms/material> "keramiek" .
_:b1 <http://schemas.delving.eu/nave/terms/objectNumber> "2458" .
_:b1 <http://schemas.delving.eu/nave/terms/thumbLarge> "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/500" .
_:b1 <http://schemas.delving.eu/nave/terms/thumbSmall> "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/220" .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schemas.delving.eu/nave/terms/BrabantCloudResource> .
`
