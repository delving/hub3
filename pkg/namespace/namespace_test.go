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
