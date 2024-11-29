package pid

import (
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestExtract(t *testing.T) {
	is := is.New(t)

	f, err := os.Open("./testdata/rdf.xml")
	is.NoErr(err)
	defer f.Close()

	g, err := rdfxml.Parse(f, nil, "")
	is.NoErr(err)

	hubID, err := domain.NewHubID("brabantcloud_enb-389-objecten_enb-389.objecten-e0bddc04-6202-fd3d-4ea1-4f725ef3da83-006b90dd-43a9-60bc-0e79-6b38c4d094ea")
	is.NoErr(err)

	got, err := Extract(hubID, g)
	is.NoErr(err)

	want := &PID{
		ID:         "http://data.brabantcloud.nl/resource/aggregation/enb-389-objecten/enb-389.objecten-e0bddc04-6202-fd3d-4ea1-4f725ef3da83-006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		ExternalID: "ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		Type:       Ark,
		ReplacedBy: "",
		Tombstone:  false,
		ModifiedAt: got.ModifiedAt,
		IsShownAt:  "https://delooierij.nl/collectie/objecten/?diw-id=brabantcloud_enb-389-objecten_enb-389.objecten-e0bddc04-6202-fd3d-4ea1-4f725ef3da83-006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		Meta: Meta{
			hubID.OrgID(),
			hubID.DatasetID.String(),
			hubID.String(),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Extract() mismatch (-want +got):\n%s", diff)
	}
}
