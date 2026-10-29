package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/delving/hub3/hub3/server/http/handlers/legacy"
	"github.com/delving/hub3/ikuzo/domain"
)

// nave's delving_hasDigitalObject is hub3's meta.hasDigitalObject; the
// websites filter on it (qf=delving_hasDigitalObject:true).
func TestManipulateV1RequestHasDigitalObject(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/search/v1/?qf=delving_hasDigitalObject:true&qf=delving_spec:princessehof", nil)

	if err := manipulateV1Request(r); err != nil {
		t.Fatal(err)
	}

	q := r.URL.Query()["qf"]
	if strings.Join(q, ",") != "meta.hasDigitalObject:true,meta.spec:princessehof" {
		t.Errorf("qf = %v", q)
	}
}

// nave adds its default facets (FACET_CONFIG) only when the request names
// none; a request with facet.field gets just those.
func TestManipulateV1RequestDefaultFacets(t *testing.T) {
	cfg := &domain.V1Config{DefaultFacets: []domain.FacetConfig{{Field: "icn_material_facet"}}}
	cfg.Convertor.Path = "../../../../convertors.toml"
	cfg.Convertor.Type = "icn"
	if err := legacy.RegisterOrgConverterWithTypeString("v1facets", cfg); err != nil {
		t.Fatal(err)
	}

	orgID, _ := domain.NewOrganizationID("v1facets")

	for query, want := range map[string]string{
		"q=x":                              "nave_material",
		"q=x&facet.field=dc_creator_facet": "dc_creator",
	} {
		r := httptest.NewRequest("GET", "/api/search/v1/?"+query, nil)
		r = domain.SetOrganization(r, &domain.Organization{ID: orgID})

		if err := manipulateV1Request(r); err != nil {
			t.Fatal(err)
		}

		if got := strings.Join(r.URL.Query()["facet.field"], ","); got != want {
			t.Errorf("%s: facet.field = %q, want %q", query, got, want)
		}
	}
}
