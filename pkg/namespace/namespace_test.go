package namespace_test

import (
	"fmt"
	"testing"

	"github.com/delving/hub3/pkg/namespace"
)

func TestSplitURI(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name     string
		args     args
		wantBase string
		wantName string
	}{
		{
			"split by /",
			args{"http://purl.org/dc/elements/1.1/title"},
			"http://purl.org/dc/elements/1.1/",
			"title",
		},
		{
			"split by #",
			args{"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			"http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"type",
		},
		{
			"unable to split URI",
			args{"urn:123"},
			"",
			"urn:123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotName := namespace.SplitURI(tt.args.uri)
			if gotBase != tt.wantBase {
				t.Errorf("SplitURI() gotBase = %v, want %v", gotBase, tt.wantBase)
			}
			if gotName != tt.wantName {
				t.Errorf("SplitURI() gotName = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func ExampleSplitURI() {
	fmt.Println(namespace.SplitURI("http://purl.org/dc/elements/1.1/title"))
	// output: http://purl.org/dc/elements/1.1/ title
}

func TestNameSpace_GetID(t *testing.T) {
	type fields struct {
		UUID      string
		Base      namespace.URI
		Prefix    string
		BaseAlt   []namespace.URI
		PrefixAlt []string
		Schema    string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"known uuid",
			fields{UUID: "123", Prefix: "dc"},
		},
		{
			"unknown uuid",
			fields{Prefix: "dc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &namespace.NameSpace{
				UUID:      tt.fields.UUID,
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
				Schema:    tt.fields.Schema,
			}
			if got := ns.GetID(); got == "" {
				t.Errorf("NameSpace.GetID() = %v, it should not be empty", got)
			}
		})
	}
}
