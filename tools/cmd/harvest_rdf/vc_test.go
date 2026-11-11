package main

import (
	"reflect"
	"testing"
)

func TestFilterGroups(t *testing.T) {
	vc := &virtualCollection{
		Slug:  "stadsmusea",
		Query: `"nave_collection:Museum Helmond";;;"delving_spec_facet:bhic";;;"-dc_type:Bidprentje";;;"dc_title:a:b";;;"dc_title:x"y"`,
		OrQueries: []string{
			`"nave_collection:Tongerlohuys";;;"nave_collectionPart:Religie"`,
			`"nave_collection:ignored by nave"`,
		},
	}

	got, err := vc.filterGroups()
	if err != nil {
		t.Fatal(err)
	}

	want := [][]string{
		{"nave_collection:Museum Helmond", "meta.spec:bhic", "-dc_type:Bidprentje", "dc_title:a:b", "dc_title:x'y"},
		{"nave_collection:Tongerlohuys", "nave_collectionPart:Religie"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filterGroups()\n got %q\nwant %q", got, want)
	}

	for _, raw := range []string{`"(a OR b)"`, `"just text"`, `"delving_owner:x"`, `""`} {
		if _, err := naveToV2Filters(raw); err == nil {
			t.Errorf("naveToV2Filters(%s) should fail", raw)
		}
	}
}
