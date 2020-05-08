package cmd

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

// nolint:lll
var testRec = v1{
	OrgID: "brabantcloud",
	Spec:  "ton-smits-huis",
	HubID: "brabantcloud_ton-smits-huis_001",
	System: struct {
		Graph         string "json:\"source_graph\""
		NamedGraphURI string "json:\"graph_name\""
	}{
		Graph:         `[{"@id":"_:b0","@type":["http://schemas.delving.eu/nave/terms/BrabantCloudResource"],"http://schemas.delving.eu/nave/terms/collection":[{"@value":"Ton Smits Huis"}],"http://schemas.delving.eu/nave/terms/collectionPart":[{"@value":"Schilderijen"}],"http://schemas.delving.eu/nave/terms/collectionType":[{"@value":"Algemeen"}],"http://schemas.delving.eu/nave/terms/color":[{"@value":"Figuratief"}],"http://schemas.delving.eu/nave/terms/creatorRole":[{"@value":"kunstschilder"}],"http://schemas.delving.eu/nave/terms/date":[{"@value":"1960 – 1970"}],"http://schemas.delving.eu/nave/terms/place":[{"@value":"Eindhoven"}],"http://schemas.delving.eu/nave/terms/productionEnd":[{"@value":"1969"}],"http://schemas.delving.eu/nave/terms/productionPlace":[{"@value":"Eindhoven"}],"http://schemas.delving.eu/nave/terms/productionStart":[{"@value":"1969"}],"http://schemas.delving.eu/nave/terms/technique":[{"@value":"Geschilderd"}]},{"@id":"http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/001","@type":["http://www.openarchives.org/ore/terms/Aggregation"],"http://www.europeana.eu/schemas/edm/aggregatedCHO":[{"@id":"http://data.brabantcloud.nl/resource/document/ton-smits-huis/001"}],"http://www.europeana.eu/schemas/edm/dataProvider":[{"@value":"Ton Smits Huis"}],"http://www.europeana.eu/schemas/edm/hasView":[{"@id":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/500"}],"http://www.europeana.eu/schemas/edm/isShownAt":[{"@id":"https://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/001"}],"http://www.europeana.eu/schemas/edm/isShownBy":[{"@id":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/500"}],"http://www.europeana.eu/schemas/edm/object":[{"@id":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/220"}],"http://www.europeana.eu/schemas/edm/provider":[{"@value":"Erfgoed Brabant"}],"http://www.europeana.eu/schemas/edm/rights":[{"@id":"https://rightsstatements.org/vocab/InC/1.0/"}],"http://www.openarchives.org/ore/terms/aggregates":[{"@id":"_:b0"},{"@id":"_:b1"}]},{"@id":"http://data.brabantcloud.nl/resource/document/ton-smits-huis/001","@type":["http://www.europeana.eu/schemas/edm/ProvidedCHO"],"http://purl.org/dc/elements/1.1/creator":[{"@value":"Ton Smits"}],"http://purl.org/dc/elements/1.1/description":[{"@value":"Een ballerina die dansend op een figuratief paard staat"}],"http://purl.org/dc/elements/1.1/identifier":[{"@value":"001"}],"http://purl.org/dc/elements/1.1/rights":[{"@value":"© L. Smits-Zoetmulder, info@tonsmitshuis.nl"}],"http://purl.org/dc/elements/1.1/subject":[{"@value":"Bloemen"},{"@value":"Paarden"},{"@value":"Dansers"},{"@value":"Ballet"},{"@value":"Maan"},{"@value":"Bergen"},{"@value":"Gelaat"}],"http://purl.org/dc/elements/1.1/title":[{"@value":"Droom van een ballerina"}],"http://purl.org/dc/elements/1.1/type":[{"@value":"Schilderij"}],"http://purl.org/dc/terms/created":[{"@value":"1969"}],"http://purl.org/dc/terms/medium":[{"@value":"Slotvernis"},{"@value":"Olieverf"},{"@value":"Paneel"}],"http://www.europeana.eu/schemas/edm/type":[{"@value":"IMAGE"}]},{"@id":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/500","@type":["http://www.europeana.eu/schemas/edm/WebResource"],"http://schemas.delving.eu/nave/terms/allowDeepZoom":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/allowPublicWebView":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/allowSourceDownload":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/deepZoomUrl":[{"@value":"https://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/001.tif.dzi"}],"http://schemas.delving.eu/nave/terms/resourceSortOrder":[{"@value":"1"}],"http://schemas.delving.eu/nave/terms/thumbLarge":[{"@value":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/500"}],"http://schemas.delving.eu/nave/terms/thumbSmall":[{"@value":"https://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/001/220"}],"http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#hasMimeType":[{"@value":"image/jpeg"}]},{"@id":"_:b1","@type":["http://schemas.delving.eu/nave/terms/DelvingResource"],"http://schemas.delving.eu/nave/terms/allowDeepZoom":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/allowLinkedOpenData":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/allowSourceDownload":[{"@value":"true"}],"http://schemas.delving.eu/nave/terms/featured":[{"@value":"false"}],"http://schemas.delving.eu/nave/terms/public":[{"@value":"true"}]}]`,
		NamedGraphURI: "http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/001/graph",
	},
}

// nolint:gocritic
func TestNewV1(t *testing.T) {
	is := is.New(t)

	input, err := os.Open("./testdata/v1-001.json")
	is.NoErr(err)

	record, err := newV1(input)
	is.NoErr(err)
	is.Equal(record.HubID, "brabantcloud_ton-smits-huis_001")

	if diff := cmp.Diff(&testRec, record); diff != "" {
		t.Errorf("newV1() = mismatch (-want +got):\n%s", diff)
	}

}
