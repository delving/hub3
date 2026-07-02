package bulk

import (
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/delving/hub3/hub3/fragments"
)

func TestSNKCreatorsSeparate(t *testing.T) {
	is := is.New(t)

	b, err := os.ReadFile("./testdata/snk_teruggegeven_werken.jsonld")
	is.NoErr(err)

	req := &Request{
		HubID:         "nk_snk-teruggegeven-werken_29541",
		OrgID:         "nk",
		DatasetID:     "snk-teruggegeven-werken",
		Graph:         string(b),
		GraphMimeType: "application/ld+json",
		RecDefID:      "nkc",
		aboutTypeURI:  []string{"https://wo2.collectienederland.nl/nk/terms/NKRecord"},
	}

	fb, err := req.createFragmentBuilder(10)
	is.NoErr(err)

	// Verify the RDF graph was parsed
	t.Logf("Graph has %d triples", fb.Graph.Len())
	is.True(fb.Graph.Len() > 0)

	rm, err := fb.ResourceMap()
	is.NoErr(err)

	t.Logf("ResourceMap has %d resources", rm.Len())

	// Verify the CHO resource exists
	choRsc, found := rm.GetResource("https://wo2.collectienederland.nl/id/cho/snk-29541")
	is.True(found)
	is.Equal(choRsc.Types, []string{"https://wo2.collectienederland.nl/nk/terms/CHO"})

	// Count Creator resources in the ResourceMap
	creatorResources := []string{}
	for id, rsc := range rm.Resources() {
		for _, typ := range rsc.Types {
			if typ == "https://wo2.collectienederland.nl/nk/terms/Creator" {
				creatorResources = append(creatorResources, id)
				t.Logf("Creator resource: %s", id)

				for pred := range rsc.Predicates() {
					t.Logf("  predicate: %s", pred)
				}
			}
		}
	}

	// CRITICAL: There must be exactly 2 separate Creator resources
	// _:b2 = "Onbekend" (no RKD link)
	// _:b3 = "Robert, H." (with RKD link)
	is.Equal(len(creatorResources), 2) // each Creator must be a separate resource

	// Build the FragmentGraph document
	fg, err := fb.Doc()
	is.NoErr(err)

	// Find Creator resources in the FragmentGraph
	var creators []*fragments.FragmentResource
	for _, rsc := range fg.Resources {
		for _, typ := range rsc.Types {
			if typ == "https://wo2.collectienederland.nl/nk/terms/Creator" {
				creators = append(creators, rsc)
				break
			}
		}
	}

	t.Logf("Found %d Creator resources in FragmentGraph", len(creators))
	is.Equal(len(creators), 2) // must stay 2 separate creators in final output

	// Verify each creator has exactly one creatorName
	for _, creator := range creators {
		nameCount := 0
		var names []string
		var hasRKDLink bool

		for _, entry := range creator.Entries {
			predicate := entry.Predicate
			if strings.HasSuffix(predicate, "creatorName") {
				nameCount++
				names = append(names, entry.Value)
			}
			if strings.HasSuffix(predicate, "rkdArtistLink") {
				hasRKDLink = true
			}
		}

		t.Logf("Creator %s: names=%v, hasRKDLink=%v", creator.ID, names, hasRKDLink)

		// Each creator node must have exactly one name
		is.Equal(nameCount, 1) // creator must not have merged names from different nodes

		// Verify RKD link is only on "Robert, H.", not on "Onbekend"
		if len(names) > 0 && names[0] == "Onbekend" {
			is.True(!hasRKDLink) // "Onbekend" must NOT have an RKD link
		}
		if len(names) > 0 && names[0] == "Robert, H." {
			is.True(hasRKDLink) // "Robert, H." must have an RKD link
		}
	}
}
