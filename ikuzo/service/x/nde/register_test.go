package nde

import "testing"

func TestRegisterConfig_newCatalog_ID(t *testing.T) {
	r := &RegisterConfig{RDFBaseURL: "https://example.org", URLPrefix: "toegang"}

	// datasets point to their catalog with IncludedInDataCatalog, so the catalog ID must match it
	want := "https://example.org/id/datacatalog/toegang"
	if got := r.newCatalog().ID; got != want {
		t.Errorf("newCatalog().ID = %q, want %q", got, want)
	}
}
