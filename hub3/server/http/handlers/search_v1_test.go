package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
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
