package bulk

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"

	"github.com/delving/hub3/hub3/fragments"
)

func TestParseGraph(t *testing.T) {
	is := is.New(t)

	b, err := os.ReadFile("./testdata/rdf_nk.nt")
	is.NoErr(err)

	req := &Request{
		HubID:         "nk_nk-rce-adlib_NK587-D1-2",
		Graph:         string(b),
		GraphMimeType: "application/n-triples",
		RecDefID:      "nkc",
		aboutTypeURI:  []string{"https://wo2.collectienederland.nl/nk/terms/NKRecord"},
	}

	fb, err := req.createFragmentBuilder(10)
	is.NoErr(err)

	is.Equal(fb.Graph.Len(), 63)

	rm, err := fb.ResourceMap()
	is.NoErr(err)
	is.Equal(rm.Len(), 9)

	rsc, found := rm.GetResource("https://wo2.collectienederland.nl/id/cho/NK587-D1-2")
	is.True(found)
	is.Equal(rsc.Types, []string{"https://wo2.collectienederland.nl/nk/terms/CHO"})
	is.Equal(len(rsc.Predicates()), 12)
	is.Equal(len(rsc.Entries), 0)

	fg, err := fb.Doc()
	is.NoErr(err)

	var cho *fragments.FragmentResource
	for _, rsc := range fg.Resources {
		if rsc.ID == "https://wo2.collectienederland.nl/id/cho/NK587-D1-2" {
			cho = rsc
			break
		}
	}
	is.Equal(len(cho.Entries), 15)
}

func TestRequestDecodesHubIDsField(t *testing.T) {
	is := is.New(t)
	body := `{"action":"drop_records","orgID":"o","dataset":"d","hubIds":["o_d_a","o_d_b"]}`
	var req Request
	err := json.Unmarshal([]byte(body), &req)
	is.NoErr(err)
	is.Equal(req.Action, "drop_records")
	is.Equal(len(req.HubIDs), 2)
	is.Equal(req.HubIDs[0], "o_d_a")
	is.Equal(req.HubIDs[1], "o_d_b")
}
