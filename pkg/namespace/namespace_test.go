package namespace_test

import (
	"fmt"
	"testing"

	"github.com/delving/hub3/pkg/namespace"
	"github.com/google/go-cmp/cmp"
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
		Base      string
		Prefix    string
		BaseAlt   []string
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

func TestNameSpace_Merge(t *testing.T) {
	type fields struct {
		Base      string
		Prefix    string
		BaseAlt   []string
		PrefixAlt []string
	}
	type args struct {
		other *namespace.NameSpace
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		prefixAlt []string
		baseAlt   []string
		wantErr   bool
	}{
		{
			"merge without overlap",
			fields{"http://purl.org/dc/elements/1.1/", "dc", []string{}, []string{}},
			args{&namespace.NameSpace{
				Base:      "http://purl.org/dc/elements/1.2/",
				Prefix:    "dce",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc", "dce"},
			[]string{"http://purl.org/dc/elements/1.1/", "http://purl.org/dc/elements/1.2/"},
			false,
		},
		{
			"merge with prefix overlap",
			fields{"http://purl.org/dc/elements/1.1/", "dc", []string{}, []string{}},
			args{&namespace.NameSpace{
				Base:      "http://purl.org/dc/elements/1.2/",
				Prefix:    "dc",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc"},
			[]string{"http://purl.org/dc/elements/1.1/", "http://purl.org/dc/elements/1.2/"},
			false,
		},
		{
			"merge with base overlap",
			fields{"http://purl.org/dc/elements/1.1/", "dc", []string{}, []string{}},
			args{&namespace.NameSpace{
				Base:      "http://purl.org/dc/elements/1.1/",
				Prefix:    "dce",
				BaseAlt:   []string{},
				PrefixAlt: []string{},
			}},
			[]string{"dc", "dce"},
			[]string{"http://purl.org/dc/elements/1.1/"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &namespace.NameSpace{
				Base:      tt.fields.Base,
				Prefix:    tt.fields.Prefix,
				BaseAlt:   tt.fields.BaseAlt,
				PrefixAlt: tt.fields.PrefixAlt,
			}
			if err := ns.Merge(tt.args.other); (err != nil) != tt.wantErr {
				t.Errorf("NameSpace.Merge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.prefixAlt, ns.Prefixes()) {
				t.Errorf("NameSpace.Merge() got %v; want %v", ns.Prefixes(), tt.prefixAlt)
			}
			if !cmp.Equal(tt.baseAlt, ns.BaseURIs()) {
				t.Errorf("NameSpace.Merge() got %v; want %v", ns.BaseURIs(), tt.baseAlt)
			}
			if ns.Prefix != tt.fields.Prefix {
				t.Errorf("NameSpace.Merge() should not change Prefix got %v; want %v", ns.Prefix, tt.fields.Prefix)
			}
			if ns.Base != tt.fields.Base {
				t.Errorf("NameSpace.Merge() should not change Base got %v; want %v", ns.Base, tt.fields.Base)
			}
		})
	}
}
